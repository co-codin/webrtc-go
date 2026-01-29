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

	if *addr == "" {
		*addr = ":8080"
	}

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{Views: engine})
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/", handlers.Welcome)
	app.Get("/room/create", handlers.RoomCreate)
	app.Get("/room/:uuid", handlers.Room)
	// app.Get("/room/:uuid/websocket", websocket.New(handlers.RoomWebsocket, websocket.Config{
	// 	HandshakeTimeout: 10 * time.Second,
	// }))
	// app.Get("/room/:uuid/chat", handlers.RoomChat)
	// app.Get("/room/:uuid/chat/websocket", websocket.New(handlers.RoomChatWebsocket))
	// app.Get("/room/:uuid/viewer/websocket", websocket.New(handlers.RoomViewerWebsocket))
	// app.Get("/stream/:suuid", handlers.Stream)
	// app.Get("/stream/:suuid/websocket", websocket.New(handlers.StreamWebsocket, websocket.Config{
	// 	HandshakeTimeout: 10 * time.Second,
	// }))
	// app.Get("/stream/:suuid/chat/websocket", websocket.New(handlers.StreamChatWebsocket))
	// app.Get("/stream/:suuid/viewer/websocket", websocket.New(handlers.StreamViewerWebsocket))
	app.Get("/*", static.New("./assets"))

	return app.Listen(*addr)
}
