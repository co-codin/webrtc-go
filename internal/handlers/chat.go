package handlers

import (
	"fmt"
	"log"

	fasthttpws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"

	"webrtc-go/pkg/chat"
	webrtcpkg "webrtc-go/pkg/webrtc"
)

// RoomChatWebsocket handles chat messages for room participants.
func RoomChatWebsocket(c fiber.Ctx) error {
	uuid := c.Params("uuid")
	return wsUpgrader.Upgrade(c.RequestCtx(), func(conn *fasthttpws.Conn) {
		room := webrtcpkg.CreateOrGetRoom(uuid)
		serveChat(conn, room.Hub)
	})
}

// StreamChatWebsocket handles chat messages for stream viewers.
func StreamChatWebsocket(c fiber.Ctx) error {
	suuid := c.Params("suuid")
	return wsUpgrader.Upgrade(c.RequestCtx(), func(conn *fasthttpws.Conn) {
		sr := webrtcpkg.CreateOrGetStreamRoom(suuid)
		serveChat(conn, sr.Hub)
	})
}

// serveChat registers the connection with the hub, pumps incoming messages
// to the broadcast channel, and writes outgoing messages back to the client.
func serveChat(conn *fasthttpws.Conn, hub *chat.Hub) {
	client := chat.NewClient(hub)
	hub.Register(client)
	defer hub.Unregister(client)

	// Write pump: forward hub messages to the WebSocket.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for msg := range client.Send {
			if err := conn.WriteMessage(fasthttpws.TextMessage, msg); err != nil {
				log.Println("chat write:", err)
				return
			}
		}
	}()

	// Read pump: forward incoming WebSocket messages to the hub.
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		hub.Broadcast <- msg
	}
	<-done
}

// RoomViewerWebsocket tracks the viewer count for a room and broadcasts
// updates to every connected viewer WebSocket.
func RoomViewerWebsocket(c fiber.Ctx) error {
	uuid := c.Params("uuid")
	return wsUpgrader.Upgrade(c.RequestCtx(), func(conn *fasthttpws.Conn) {
		room := webrtcpkg.CreateOrGetRoom(uuid)
		serveViewerCount(conn, room.ViewerHub, room.IncrementViewers, room.DecrementViewers)
	})
}

// StreamViewerWebsocket tracks the viewer count for a stream.
func StreamViewerWebsocket(c fiber.Ctx) error {
	suuid := c.Params("suuid")
	return wsUpgrader.Upgrade(c.RequestCtx(), func(conn *fasthttpws.Conn) {
		sr := webrtcpkg.CreateOrGetStreamRoom(suuid)
		serveViewerCount(conn, sr.ViewerHub, sr.IncrementViewers, sr.DecrementViewers)
	})
}

// serveViewerCount registers the connection, increments the count, broadcasts
// the new total, then waits until the connection closes and decrements.
func serveViewerCount(
	conn *fasthttpws.Conn,
	hub *chat.Hub,
	increment func() int32,
	decrement func() int32,
) {
	client := chat.NewClient(hub)
	hub.Register(client)
	defer func() {
		hub.Unregister(client)
		count := decrement()
		hub.Broadcast <- []byte(fmt.Sprintf("%d", count))
	}()

	// Write pump.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for msg := range client.Send {
			if err := conn.WriteMessage(fasthttpws.TextMessage, msg); err != nil {
				log.Println("viewer write:", err)
				return
			}
		}
	}()

	// Broadcast the updated count to all viewers.
	count := increment()
	hub.Broadcast <- []byte(fmt.Sprintf("%d", count))

	// Keep the connection open until the client disconnects.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
	<-done
}
