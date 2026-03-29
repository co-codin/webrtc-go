package handlers

import (
	"fmt"
	"log"

	fasthttpws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	pionwebrtc "github.com/pion/webrtc/v3"

	webrtcpkg "webrtc-go/pkg/webrtc"
)

// Stream renders the viewer page. If no room exists for the given UUID the
// template receives NoStream=true and shows an error message.
func Stream(c fiber.Ctx) error {
	suuid := c.Params("suuid")

	webrtcpkg.RoomsLock.RLock()
	_, exists := webrtcpkg.Rooms[suuid]
	webrtcpkg.RoomsLock.RUnlock()

	ws := wsScheme(c)
	host := c.Hostname()

	return c.Render("stream", fiber.Map{
		"Type":                "stream",
		"NoStream":            !exists,
		"StreamWebsocketAddr": fmt.Sprintf("%s://%s/stream/%s/websocket", ws, host, suuid),
		"ChatWebsocketAddr":   fmt.Sprintf("%s://%s/stream/%s/chat/websocket", ws, host, suuid),
		"ViewerWebsocketAddr": fmt.Sprintf("%s://%s/stream/%s/viewer/websocket", ws, host, suuid),
	}, "layouts/main")
}

// StreamWebsocket handles the WebRTC signaling WebSocket for stream viewers.
// Viewers are added to the room's Peers so they receive all published tracks
// but do not add transceivers for sending.
func StreamWebsocket(c fiber.Ctx) error {
	suuid := c.Params("suuid")
	return wsUpgrader.Upgrade(c.RequestCtx(), func(conn *fasthttpws.Conn) {
		webrtcpkg.RoomsLock.RLock()
		room, ok := webrtcpkg.Rooms[suuid]
		webrtcpkg.RoomsLock.RUnlock()
		if !ok {
			return
		}

		pc, err := pionwebrtc.NewPeerConnection(pionwebrtc.Configuration{
			ICEServers: []pionwebrtc.ICEServer{
				{URLs: []string{"stun:stun.l.google.com:19302"}},
			},
		})
		if err != nil {
			log.Println("StreamWebsocket NewPeerConnection:", err)
			return
		}
		defer pc.Close()

		writer := &webrtcpkg.ThreadSafeWriter{Conn: conn}
		room.Peers.AddPeerConnection(pc, writer)
		defer room.Peers.SignalPeerConnections()

		room.Peers.SignalPeerConnections()
		webrtcpkg.RunSignalingLoop(conn, pc, "StreamWebsocket")
	})
}
