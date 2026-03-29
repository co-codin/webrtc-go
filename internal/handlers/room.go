package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	fasthttpws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	guuid "github.com/google/uuid"
	pionwebrtc "github.com/pion/webrtc/v3"
	"github.com/valyala/fasthttp"

	webrtcpkg "webrtc-go/pkg/webrtc"
)

var roomUpgrader = fasthttpws.FastHTTPUpgrader{
	CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true },
}

func RoomCreate(c fiber.Ctx) error {
	return c.Redirect().To(fmt.Sprintf("/room/%s", guuid.New().String()))
}

func Room(c fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == "" {
		return c.Status(fiber.StatusBadRequest).SendString("missing room id")
	}

	scheme := c.Protocol()
	ws := "ws"
	if scheme == "https" {
		ws = "wss"
	}
	host := c.Hostname()

	return c.Render("peer", fiber.Map{
		"Type":                "room",
		"RoomLink":            fmt.Sprintf("%s://%s/room/%s", scheme, host, uuid),
		"StreamLink":          fmt.Sprintf("%s://%s/stream/%s", scheme, host, uuid),
		"RoomWebsocketAddr":   fmt.Sprintf("%s://%s/room/%s/websocket", ws, host, uuid),
		"ChatWebsocketAddr":   fmt.Sprintf("%s://%s/room/%s/chat/websocket", ws, host, uuid),
		"ViewerWebsocketAddr": fmt.Sprintf("%s://%s/room/%s/viewer/websocket", ws, host, uuid),
	}, "layouts/main")
}

// RoomWebsocket handles the WebRTC signaling WebSocket for room peers.
func RoomWebsocket(c fiber.Ctx) error {
	uuid := c.Params("uuid")
	return roomUpgrader.Upgrade(c.RequestCtx(), func(conn *fasthttpws.Conn) {
		room := webrtcpkg.CreateOrGetRoom(uuid)

		pc, err := pionwebrtc.NewPeerConnection(pionwebrtc.Configuration{
			ICEServers: []pionwebrtc.ICEServer{
				{URLs: []string{"stun:stun.l.google.com:19302"}},
			},
		})
		if err != nil {
			log.Println("RoomWebsocket NewPeerConnection:", err)
			return
		}
		defer pc.Close()

		for _, kind := range []pionwebrtc.RTPCodecType{
			pionwebrtc.RTPCodecTypeVideo,
			pionwebrtc.RTPCodecTypeAudio,
		} {
			if _, err := pc.AddTransceiverFromKind(kind, pionwebrtc.RTPTransceiverInit{
				Direction: pionwebrtc.RTPTransceiverDirectionSendrecv,
			}); err != nil {
				log.Println("RoomWebsocket AddTransceiver:", err)
				return
			}
		}

		writer := &webrtcpkg.ThreadSafeWriter{Conn: conn}
		room.Peers.AddPeerConnection(pc, writer)
		defer room.Peers.SignalPeerConnections()

		room.Peers.SignalPeerConnections()

		msg := &webrtcpkg.WebsocketMessage{}
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := json.Unmarshal(raw, msg); err != nil {
				log.Println("RoomWebsocket unmarshal:", err)
				return
			}
			switch msg.Event {
			case "answer":
				answer := pionwebrtc.SessionDescription{}
				if err := json.Unmarshal([]byte(msg.Data), &answer); err != nil {
					log.Println("RoomWebsocket answer unmarshal:", err)
					continue
				}
				if err := pc.SetRemoteDescription(answer); err != nil {
					log.Println("RoomWebsocket SetRemoteDescription:", err)
					continue
				}
			case "candidate":
				candidate := pionwebrtc.ICECandidateInit{}
				if err := json.Unmarshal([]byte(msg.Data), &candidate); err != nil {
					log.Println("RoomWebsocket candidate unmarshal:", err)
					continue
				}
				if err := pc.AddICECandidate(candidate); err != nil {
					log.Println("RoomWebsocket AddICECandidate:", err)
					continue
				}
			}
		}
	})
}
