package webrtc

import (
	"sync"
	"time"
)

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

// CleanupRoomIfEmpty removes the room (and its associated stream room) from
// the global maps when all peer connections have closed. The hub goroutines
// are stopped after a brief delay to let any in-flight handlers drain.
func CleanupRoomIfEmpty(uuid string) {
	RoomsLock.Lock()

	r, ok := Rooms[uuid]
	if !ok {
		RoomsLock.Unlock()
		return
	}

	r.Peers.ListLock.RLock()
	empty := len(r.Peers.Connections) == 0
	r.Peers.ListLock.RUnlock()

	if !empty {
		RoomsLock.Unlock()
		return
	}

	delete(Rooms, uuid)
	RoomsLock.Unlock()

	// Stop hubs after a short delay so in-flight register/broadcast calls finish.
	go func() {
		time.Sleep(time.Second)
		r.Stop()
	}()

	// Remove the associated stream room under its own lock.
	CleanupStreamRoom(uuid)
}
