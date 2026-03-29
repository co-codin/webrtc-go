package chat

// Client is a single chat participant connected via WebSocket.
type Client struct {
	Hub  *Hub
	Send chan []byte
}

func NewClient(hub *Hub) *Client {
	return &Client{
		Hub:  hub,
		Send: make(chan []byte, 256),
	}
}
