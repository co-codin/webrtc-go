package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/template/html/v3"

	webrtcpkg "webrtc-go/pkg/webrtc"
)

func newTestApp() *fiber.App {
	engine := html.New("../../views", ".html")
	app := fiber.New(fiber.Config{Views: engine})
	app.Get("/room/create", RoomCreate)
	app.Get("/room/:uuid", Room)
	app.Get("/stream/:suuid", Stream)
	return app
}

func TestRoomCreate_Redirects(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/room/create", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		t.Errorf("expected redirect (3xx), got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc == "" {
		t.Error("expected a Location header on redirect")
	}
	// Location should look like /room/<uuid>
	if len(loc) < len("/room/") {
		t.Errorf("Location %q is too short to be a room URL", loc)
	}
}

func TestRoom_RendersPage(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/room/test-uuid-123", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestStream_NoRoomShowsPage(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/stream/nonexistent-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	// Page renders (200) even when no room exists; NoStream=true drives the template.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRoom_ChatWebsocketAddrHasUsername(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/room/test-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "?username=User") {
		t.Error("expected ChatWebsocketAddr to contain ?username=User#### in rendered page")
	}
}

func TestStream_ChatWebsocketAddrHasUsername(t *testing.T) {
	// Rooms are only created when a WebSocket peer connects, not on page load.
	// Seed one directly so the stream page renders the full UI (NoStream=false).
	webrtcpkg.CreateOrGetRoom("stream-uuid")

	app := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/stream/stream-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "?username=User") {
		t.Error("expected ChatWebsocketAddr to contain ?username=User#### in rendered stream page")
	}
}


func TestWelcome_ReturnsOK(t *testing.T) {
	engine := html.New("../../views", ".html")
	app := fiber.New(fiber.Config{Views: engine})
	app.Get("/", Welcome)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
