package webrtc

import (
	"sync"
	"testing"
)

func resetRooms() {
	RoomsLock.Lock()
	Rooms = make(map[string]*Room)
	RoomsLock.Unlock()
}

func resetStreamRooms() {
	StreamRoomsLock.Lock()
	StreamRooms = make(map[string]*StreamRoom)
	StreamRoomsLock.Unlock()
}

func TestCreateOrGetRoom_New(t *testing.T) {
	resetRooms()
	r := CreateOrGetRoom("uuid-1")
	if r == nil {
		t.Fatal("expected a room, got nil")
	}
	if r.Peers == nil {
		t.Error("Peers is nil")
	}
	if r.Hub == nil {
		t.Error("Hub is nil")
	}
	if r.ViewerHub == nil {
		t.Error("ViewerHub is nil")
	}
}

func TestCreateOrGetRoom_ExistingRoomReturned(t *testing.T) {
	resetRooms()
	r1 := CreateOrGetRoom("uuid-2")
	r2 := CreateOrGetRoom("uuid-2")
	if r1 != r2 {
		t.Error("expected same room pointer on second call")
	}
}

func TestRoom_ViewerCount(t *testing.T) {
	resetRooms()
	r := CreateOrGetRoom("uuid-3")

	if got := r.IncrementViewers(); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	if got := r.IncrementViewers(); got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
	if got := r.DecrementViewers(); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func TestCreateOrGetRoom_Concurrent(t *testing.T) {
	resetRooms()
	var wg sync.WaitGroup
	results := make([]*Room, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = CreateOrGetRoom("concurrent-uuid")
		}(i)
	}
	wg.Wait()

	first := results[0]
	for _, r := range results[1:] {
		if r != first {
			t.Error("concurrent calls returned different room pointers")
		}
	}
}

func TestCreateOrGetStreamRoom_New(t *testing.T) {
	resetStreamRooms()
	sr := CreateOrGetStreamRoom("suuid-1")
	if sr == nil {
		t.Fatal("expected a stream room, got nil")
	}
	if sr.Hub == nil {
		t.Error("Hub is nil")
	}
	if sr.ViewerHub == nil {
		t.Error("ViewerHub is nil")
	}
}

func TestCreateOrGetStreamRoom_ExistingReturned(t *testing.T) {
	resetStreamRooms()
	sr1 := CreateOrGetStreamRoom("suuid-2")
	sr2 := CreateOrGetStreamRoom("suuid-2")
	if sr1 != sr2 {
		t.Error("expected same stream room pointer on second call")
	}
}

func TestStreamRoom_ViewerCount(t *testing.T) {
	resetStreamRooms()
	sr := CreateOrGetStreamRoom("suuid-3")

	if got := sr.IncrementViewers(); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
	if got := sr.DecrementViewers(); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestNewPeers(t *testing.T) {
	p := NewPeers()
	if p == nil {
		t.Fatal("NewPeers returned nil")
	}
	if p.TrackLocals == nil {
		t.Error("TrackLocals map is nil")
	}
	if len(p.Connections) != 0 {
		t.Error("Connections should be empty initially")
	}
}
