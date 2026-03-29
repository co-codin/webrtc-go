package handlers

import (
	"net"
	"strings"
	"testing"
	"time"

	fasthttpws "github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"

	"webrtc-go/pkg/chat"
)

// startChatServer launches a minimal fasthttp WebSocket server that calls
// serveChat with the given username and returns the listener address.
func startChatServer(t *testing.T, hub *chat.Hub, username string) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	upgrader := fasthttpws.FastHTTPUpgrader{
		CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true },
	}
	srv := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			_ = upgrader.Upgrade(ctx, func(conn *fasthttpws.Conn) {
				serveChat(conn, hub, username)
			})
		},
	}
	go srv.Serve(ln) //nolint:errcheck
	return ln.Addr().String()
}

func TestServeChat_PrefixesUsername(t *testing.T) {
	hub := chat.NewHub()
	go hub.Run()

	addr := startChatServer(t, hub, "Alice")

	// Connect receiver first, then sender (both share the same hub).
	rc, _, err := fasthttpws.DefaultDialer.Dial("ws://"+addr, nil)
	if err != nil {
		t.Fatalf("receiver dial: %v", err)
	}
	defer rc.Close()

	sc, _, err := fasthttpws.DefaultDialer.Dial("ws://"+addr, nil)
	if err != nil {
		t.Fatalf("sender dial: %v", err)
	}
	defer sc.Close()

	// Allow serveChat goroutines to register both clients with the hub.
	time.Sleep(30 * time.Millisecond)

	if err := sc.WriteMessage(fasthttpws.TextMessage, []byte("hello")); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := rc.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		t.Fatalf("deadline: %v", err)
	}
	_, got, err := rc.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.HasPrefix(string(got), "Alice: ") {
		t.Errorf("expected message prefixed with %q, got %q", "Alice: ", string(got))
	}
}

func TestServeChat_BroadcastsToAllClients(t *testing.T) {
	hub := chat.NewHub()
	go hub.Run()

	addr := startChatServer(t, hub, "Bob")

	conns := make([]*fasthttpws.Conn, 3)
	for i := range conns {
		c, _, err := fasthttpws.DefaultDialer.Dial("ws://"+addr, nil)
		if err != nil {
			t.Fatalf("dial[%d]: %v", i, err)
		}
		defer c.Close()
		conns[i] = c
	}

	time.Sleep(30 * time.Millisecond)

	// Send from the first connection.
	if err := conns[0].WriteMessage(fasthttpws.TextMessage, []byte("hi all")); err != nil {
		t.Fatalf("write: %v", err)
	}

	// All three connections should receive the message (sender included via hub broadcast).
	for i, c := range conns {
		if err := c.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
			t.Fatalf("deadline[%d]: %v", i, err)
		}
		_, msg, err := c.ReadMessage()
		if err != nil {
			t.Fatalf("read[%d]: %v", i, err)
		}
		if string(msg) != "Bob: hi all" {
			t.Errorf("conn[%d] expected %q, got %q", i, "Bob: hi all", string(msg))
		}
	}
}
