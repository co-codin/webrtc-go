package webrtc

import "sync"

// Room represents a single video-chat room with bidirectional peer connections.
type Room struct {
	Peers *Peers
	ViewerState
}

var (
	RoomsLock sync.RWMutex
	Rooms     = make(map[string]*Room)
)

// CreateOrGetRoom returns an existing room or creates a new one.
func CreateOrGetRoom(uuid string) *Room {
	RoomsLock.Lock()
	defer RoomsLock.Unlock()

	if r, ok := Rooms[uuid]; ok {
		return r
	}

	r := &Room{
		Peers:       NewPeers(),
		ViewerState: newViewerState(),
	}
	Rooms[uuid] = r
	return r
}
