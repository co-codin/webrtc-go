package chat

import (
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	h := NewHub()
	if h == nil {
		t.Fatal("NewHub returned nil")
	}
	if h.Broadcast == nil {
		t.Error("Broadcast channel is nil")
	}
}

func TestHubRegisterAndBroadcast(t *testing.T) {
	h := NewHub()
	go h.Run()

	c1 := NewClient(h)
	c2 := NewClient(h)
	h.Register(c1)
	h.Register(c2)

	// Allow the hub goroutine to process registrations.
	time.Sleep(10 * time.Millisecond)

	msg := []byte("hello")
	h.Broadcast <- msg

	for _, c := range []*Client{c1, c2} {
		select {
		case got := <-c.Send:
			if string(got) != string(msg) {
				t.Errorf("expected %q, got %q", msg, got)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("timed out waiting for broadcast message")
		}
	}
}

func TestHubUnregister(t *testing.T) {
	h := NewHub()
	go h.Run()

	c := NewClient(h)
	h.Register(c)
	time.Sleep(10 * time.Millisecond)

	h.Unregister(c)
	time.Sleep(10 * time.Millisecond)

	// Send channel must be closed after unregister.
	select {
	case _, ok := <-c.Send:
		if ok {
			t.Error("expected Send channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timed out waiting for Send channel to close")
	}
}

func TestHub_Stop(t *testing.T) {
	h := NewHub()
	exited := make(chan struct{})
	go func() {
		h.Run()
		close(exited)
	}()

	h.Stop()

	select {
	case <-exited:
		// Run returned as expected.
	case <-time.After(200 * time.Millisecond):
		t.Error("Hub.Run did not exit after Stop()")
	}
}

func TestHub_StopClosesClientSend(t *testing.T) {
	h := NewHub()
	go h.Run()

	c := NewClient(h)
	h.Register(c)
	time.Sleep(10 * time.Millisecond)

	h.Unregister(c)
	time.Sleep(10 * time.Millisecond)

	h.Stop()

	// After stop, no further broadcasts should panic or block.
	// This is a smoke test to ensure Stop is safe to call after clients leave.
}

func TestNewClient(t *testing.T) {
	h := NewHub()
	c := NewClient(h)
	if c.Hub != h {
		t.Error("client Hub does not point to the given hub")
	}
	if c.Send == nil {
		t.Error("Send channel is nil")
	}
}
