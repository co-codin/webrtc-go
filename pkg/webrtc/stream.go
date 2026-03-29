package webrtc

import "sync"

// StreamRoom holds the chat and viewer-count state for stream viewers.
// WebRTC tracks are served from the corresponding Room's Peers.
type StreamRoom struct {
	ViewerState
}

var (
	StreamRoomsLock sync.RWMutex
	StreamRooms     = make(map[string]*StreamRoom)
)

// CreateOrGetStreamRoom returns an existing stream room or creates a new one.
func CreateOrGetStreamRoom(suuid string) *StreamRoom {
	StreamRoomsLock.Lock()
	defer StreamRoomsLock.Unlock()

	if sr, ok := StreamRooms[suuid]; ok {
		return sr
	}

	sr := &StreamRoom{ViewerState: newViewerState()}
	StreamRooms[suuid] = sr
	return sr
}
