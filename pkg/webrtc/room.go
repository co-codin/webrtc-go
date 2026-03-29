package webrtc

import (
	"sync"
	"sync/atomic"

	"webrtc-go/pkg/chat"
)

// Room represents a single video-chat room with bidirectional peer connections.
type Room struct {
	Peers       *Peers
	Hub         *chat.Hub // chat messages for room participants
	ViewerHub   *chat.Hub // broadcasts viewer-count updates
	ViewerCount int32     // updated atomically
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

	hub := chat.NewHub()
	go hub.Run()

	viewerHub := chat.NewHub()
	go viewerHub.Run()

	r := &Room{
		Peers:     NewPeers(),
		Hub:       hub,
		ViewerHub: viewerHub,
	}
	Rooms[uuid] = r
	return r
}

// IncrementViewers adds 1 to the viewer count and returns the new value.
func (r *Room) IncrementViewers() int32 {
	return atomic.AddInt32(&r.ViewerCount, 1)
}

// DecrementViewers subtracts 1 from the viewer count and returns the new value.
func (r *Room) DecrementViewers() int32 {
	return atomic.AddInt32(&r.ViewerCount, -1)
}
