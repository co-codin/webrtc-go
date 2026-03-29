package server

import (
	"flag"
	"os"
	"webrtc-go/internal/handlers"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v3"
)

var (
	addr = flag.String("addr", ":"+os.Getenv("PORT"), "")
	cert = flag.String("cert", "", "")
	key  = flag.String("key", "", "")
)

func Run() error {
	flag.Parse()

	if *addr == ":" || *addr == "" {
		*addr = ":8080"
	}

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{Views: engine})
	app.Use(logger.New())
	app.Use(cors.New())

	// HTTP routes
	app.Get("/", handlers.Welcome)
	app.Get("/room/create", handlers.RoomCreate)
	app.Get("/room/:uuid", handlers.Room)
	app.Get("/stream/:suuid", handlers.Stream)

	// WebSocket routes — room
	app.Get("/room/:uuid/websocket", handlers.RoomWebsocket)
	app.Get("/room/:uuid/chat/websocket", handlers.RoomChatWebsocket)
	app.Get("/room/:uuid/viewer/websocket", handlers.RoomViewerWebsocket)

	// WebSocket routes — stream
	app.Get("/stream/:suuid/websocket", handlers.StreamWebsocket)
	app.Get("/stream/:suuid/chat/websocket", handlers.StreamChatWebsocket)
	app.Get("/stream/:suuid/viewer/websocket", handlers.StreamViewerWebsocket)

	// Static assets
	app.Get("/*", static.New("./assets"))

	return app.Listen(*addr, fiber.ListenConfig{
		CertFile:    *cert,
		CertKeyFile: *key,
	})
}
