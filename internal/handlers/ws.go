package handlers

import (
	fasthttpws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

// wsUpgrader is shared by all WebSocket handlers.
var wsUpgrader = fasthttpws.FastHTTPUpgrader{
	CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true },
}

// wsScheme returns "wss" when the request arrived over HTTPS, otherwise "ws".
func wsScheme(c fiber.Ctx) string {
	if c.Protocol() == "https" {
		return "wss"
	}
	return "ws"
}
