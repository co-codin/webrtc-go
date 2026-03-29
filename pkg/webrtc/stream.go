package webrtc

import (
	"sync"
	"sync/atomic"

	"webrtc-go/pkg/chat"
)

// StreamRoom holds the chat and viewer-count state for stream viewers.
// WebRTC tracks are served from the corresponding Room's Peers.
type StreamRoom struct {
	Hub         *chat.Hub // chat messages for stream viewers
	ViewerHub   *chat.Hub // broadcasts viewer-count updates
	ViewerCount int32     // updated atomically
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

	hub := chat.NewHub()
	go hub.Run()

	viewerHub := chat.NewHub()
	go viewerHub.Run()

	sr := &StreamRoom{
		Hub:       hub,
		ViewerHub: viewerHub,
	}
	StreamRooms[suuid] = sr
	return sr
}

// IncrementViewers adds 1 to the viewer count and returns the new value.
func (sr *StreamRoom) IncrementViewers() int32 {
	return atomic.AddInt32(&sr.ViewerCount, 1)
}

// DecrementViewers subtracts 1 from the viewer count and returns the new value.
func (sr *StreamRoom) DecrementViewers() int32 {
	return atomic.AddInt32(&sr.ViewerCount, -1)
}
