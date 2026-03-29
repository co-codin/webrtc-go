package webrtc

import (
	"sync/atomic"

	"webrtc-go/pkg/chat"
)

// ViewerState holds the chat hub, viewer-count hub, and atomic viewer count.
// It is embedded by both Room and StreamRoom to avoid duplication.
type ViewerState struct {
	Hub         *chat.Hub // chat messages
	ViewerHub   *chat.Hub // broadcasts viewer-count updates
	ViewerCount int32     // updated atomically
}

func newViewerState() ViewerState {
	hub := chat.NewHub()
	go hub.Run()
	viewerHub := chat.NewHub()
	go viewerHub.Run()
	return ViewerState{Hub: hub, ViewerHub: viewerHub}
}

func (v *ViewerState) IncrementViewers() int32 {
	return atomic.AddInt32(&v.ViewerCount, 1)
}

func (v *ViewerState) DecrementViewers() int32 {
	return atomic.AddInt32(&v.ViewerCount, -1)
}

// Stop shuts down both hub goroutines. Call only after all clients have left.
func (v *ViewerState) Stop() {
	v.Hub.Stop()
	v.ViewerHub.Stop()
}
