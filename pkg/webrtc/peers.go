package webrtc

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/pion/rtcp"
	pionwebrtc "github.com/pion/webrtc/v3"
)

// WebsocketMessage is the JSON signaling envelope exchanged over WebSocket.
type WebsocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// SafeWriter is an interface for a thread-safe WebSocket JSON writer.
type SafeWriter interface {
	WriteJSON(v interface{}) error
}

// ThreadSafeWriter wraps any SafeWriter with a mutex so concurrent goroutines
// can call WriteJSON without data races.
type ThreadSafeWriter struct {
	Conn SafeWriter
	Lock sync.Mutex
}

func (t *ThreadSafeWriter) WriteJSON(v interface{}) error {
	t.Lock.Lock()
	defer t.Lock.Unlock()
	return t.Conn.WriteJSON(v)
}

// PeerConnectionState pairs a pion PeerConnection with the WebSocket it signals over.
type PeerConnectionState struct {
	PeerConnection *pionwebrtc.PeerConnection
	Websocket      *ThreadSafeWriter
}

// Peers manages all peer connections and the shared local tracks for one room.
type Peers struct {
	ListLock    sync.RWMutex
	Connections []PeerConnectionState
	TrackLocals map[string]*pionwebrtc.TrackLocalStaticRTP
}

func NewPeers() *Peers {
	return &Peers{
		TrackLocals: make(map[string]*pionwebrtc.TrackLocalStaticRTP),
	}
}

// addTrack creates a local copy of an incoming remote track and signals all peers.
func (p *Peers) addTrack(t *pionwebrtc.TrackRemote) *pionwebrtc.TrackLocalStaticRTP {
	p.ListLock.Lock()
	defer func() {
		p.ListLock.Unlock()
		p.SignalPeerConnections()
	}()

	trackLocal, err := pionwebrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		log.Println("addTrack:", err)
		return nil
	}
	p.TrackLocals[t.ID()] = trackLocal
	return trackLocal
}

// removeTrack deletes a local track and re-signals all peers.
func (p *Peers) removeTrack(t *pionwebrtc.TrackLocalStaticRTP) {
	p.ListLock.Lock()
	defer func() {
		p.ListLock.Unlock()
		p.SignalPeerConnections()
	}()
	delete(p.TrackLocals, t.ID())
}

// SignalPeerConnections synchronises the set of tracks with every peer and
// sends a fresh offer so each peer reflects the current topology.
func (p *Peers) SignalPeerConnections() {
	p.ListLock.Lock()
	defer func() {
		p.ListLock.Unlock()
		p.dispatchKeyFrame()
	}()

	attemptSync := func() (tryAgain bool) {
		for i := range p.Connections {
			if p.Connections[i].PeerConnection.ConnectionState() == pionwebrtc.PeerConnectionStateClosed {
				p.Connections = append(p.Connections[:i], p.Connections[i+1:]...)
				return true
			}

			existingSenders := map[string]bool{}
			for _, sender := range p.Connections[i].PeerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}
				existingSenders[sender.Track().ID()] = true

				// Remove senders whose track no longer exists globally.
				if _, ok := p.TrackLocals[sender.Track().ID()]; !ok {
					if err := p.Connections[i].PeerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			// Add new tracks this peer doesn't have yet.
			for trackID := range p.TrackLocals {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := p.Connections[i].PeerConnection.AddTrack(p.TrackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			offer, err := p.Connections[i].PeerConnection.CreateOffer(nil)
			if err != nil {
				return true
			}
			if err = p.Connections[i].PeerConnection.SetLocalDescription(offer); err != nil {
				return true
			}
			offerJSON, err := json.Marshal(offer)
			if err != nil {
				return true
			}
			if err = p.Connections[i].Websocket.WriteJSON(&WebsocketMessage{
				Event: "offer",
				Data:  string(offerJSON),
			}); err != nil {
				return true
			}
		}
		return false
	}

	for attempt := 0; ; attempt++ {
		if attempt == 25 {
			// Give up for now and retry after a short delay.
			go func() {
				time.Sleep(3 * time.Second)
				p.SignalPeerConnections()
			}()
			return
		}
		if !attemptSync() {
			break
		}
	}
}

// dispatchKeyFrame asks every receiver to send a keyframe so new viewers get
// a complete picture immediately.
func (p *Peers) dispatchKeyFrame() {
	p.ListLock.Lock()
	defer p.ListLock.Unlock()

	for i := range p.Connections {
		for _, receiver := range p.Connections[i].PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}
			_ = p.Connections[i].PeerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}

// MessageReader reads raw WebSocket frames.
type MessageReader interface {
	ReadMessage() (messageType int, p []byte, err error)
}

// RunSignalingLoop reads answer and candidate messages from conn and applies
// them to pc. It returns when the connection closes or an unrecoverable error
// occurs, making it a drop-in replacement for the per-handler signaling loops.
func RunSignalingLoop(conn MessageReader, pc *pionwebrtc.PeerConnection, logPrefix string) {
	msg := &WebsocketMessage{}
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if err := json.Unmarshal(raw, msg); err != nil {
			log.Println(logPrefix, "unmarshal:", err)
			return
		}
		switch msg.Event {
		case "answer":
			answer := pionwebrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(msg.Data), &answer); err != nil {
				log.Println(logPrefix, "answer:", err)
				continue
			}
			if err := pc.SetRemoteDescription(answer); err != nil {
				log.Println(logPrefix, "SetRemoteDescription:", err)
				continue
			}
		case "candidate":
			candidate := pionwebrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(msg.Data), &candidate); err != nil {
				log.Println(logPrefix, "candidate:", err)
				continue
			}
			if err := pc.AddICECandidate(candidate); err != nil {
				log.Println(logPrefix, "AddICECandidate:", err)
				continue
			}
		}
	}
}

// AddPeerConnection registers a new connection, wires up its event handlers,
// and returns so the caller can start the WebSocket read loop.
func (p *Peers) AddPeerConnection(pc *pionwebrtc.PeerConnection, ws *ThreadSafeWriter) {
	p.ListLock.Lock()
	p.Connections = append(p.Connections, PeerConnectionState{pc, ws})
	p.ListLock.Unlock()

	pc.OnICECandidate(func(i *pionwebrtc.ICECandidate) {
		if i == nil {
			return
		}
		candidateJSON, err := json.Marshal(i.ToJSON())
		if err != nil {
			log.Println("OnICECandidate marshal:", err)
			return
		}
		if err := ws.WriteJSON(&WebsocketMessage{
			Event: "candidate",
			Data:  string(candidateJSON),
		}); err != nil {
			log.Println("OnICECandidate write:", err)
		}
	})

	pc.OnConnectionStateChange(func(state pionwebrtc.PeerConnectionState) {
		switch state {
		case pionwebrtc.PeerConnectionStateFailed:
			if err := pc.Close(); err != nil {
				log.Println("pc close:", err)
			}
		case pionwebrtc.PeerConnectionStateClosed:
			p.SignalPeerConnections()
		}
	})

	pc.OnTrack(func(t *pionwebrtc.TrackRemote, _ *pionwebrtc.RTPReceiver) {
		trackLocal := p.addTrack(t)
		if trackLocal == nil {
			return
		}
		defer p.removeTrack(trackLocal)

		buf := make([]byte, 1500)
		for {
			n, _, err := t.Read(buf)
			if err != nil {
				return
			}
			if _, err = trackLocal.Write(buf[:n]); err != nil {
				return
			}
		}
	})
}
