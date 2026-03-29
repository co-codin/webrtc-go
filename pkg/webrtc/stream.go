package webrtc

import (
	"sync"
	"time"
)

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

// CleanupStreamRoom removes a stream room from the global map and stops its
// hub goroutines. It is called by CleanupRoomIfEmpty when the parent room
// is deleted.
func CleanupStreamRoom(suuid string) {
	StreamRoomsLock.Lock()

	sr, ok := StreamRooms[suuid]
	if !ok {
		StreamRoomsLock.Unlock()
		return
	}

	delete(StreamRooms, suuid)
	StreamRoomsLock.Unlock()

	go func() {
		time.Sleep(time.Second)
		sr.Stop()
	}()
}
