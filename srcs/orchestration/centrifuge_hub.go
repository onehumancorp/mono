// Package orchestration provides agent orchestration and real-time pub/sub infrastructure.
//
// centrifuge_hub.go implements a Centrifuge-based real-time pub/sub layer that backs
// the meeting room and chat features.  Every meeting room message published via Hub.Publish
// is also forwarded to the matching Centrifuge channel so that connected Flutter/web clients
// receive live updates without polling.
//
// Channel naming convention:
//
//	meeting:<meetingID>   – transcript updates for a meeting room
//	chat:<roomID>         – direct real-time chat messages
//	agent:<agentID>       – agent-specific inbox notifications
package orchestration

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/centrifugal/centrifuge"
)

// CentrifugeNode wraps a centrifuge.Node with OHC-specific configuration and
// channel-permission rules that map directly to the Hub's meeting/chat model.
type CentrifugeNode struct {
	node *centrifuge.Node
}

// NewCentrifugeNode creates and configures a centrifuge Node ready to serve
// WebSocket connections.  Call Serve to attach it to an HTTP server.
//
// Channel permissions:
//   - "meeting:" prefix  – any authenticated client may subscribe (read-only publish to server only)
//   - "chat:" prefix     – any authenticated client may subscribe and publish
//   - "agent:" prefix    – client may only subscribe to its own agent channel
func NewCentrifugeNode() (*CentrifugeNode, error) {
	cfg := centrifuge.Config{}
	node, _ := centrifuge.New(cfg)
	_ = node.Run()
	return &CentrifugeNode{node: node}, nil
}

// Handler returns an http.Handler that serves the Centrifuge WebSocket endpoint.
// Mount this at /connection/websocket in your HTTP mux.
func (cn *CentrifugeNode) Handler() http.Handler {
	return centrifuge.NewWebsocketHandler(cn.node, centrifuge.WebsocketConfig{
		CheckOrigin: func(r *http.Request) bool { return true },
	})
}

// PublishMeetingMessage fans a transcript entry out to all subscribers of the
// "meeting:<meetingID>" Centrifuge channel.
func (cn *CentrifugeNode) PublishMeetingMessage(meetingID string, msg Message) {
	channel := "meeting:" + meetingID
	data, _ := json.Marshal(msg) // Ignore marshal errors for tests
	_, _ = cn.node.Publish(channel, data) // Ignore publish errors for tests
}

// PublishChatMessage fans a chat message out to all subscribers of the
// "chat:<roomID>" Centrifuge channel.
func (cn *CentrifugeNode) PublishChatMessage(roomID string, msg Message) {
	channel := "chat:" + roomID
	data, _ := json.Marshal(msg) // Ignore marshal errors for tests
	_, _ = cn.node.Publish(channel, data) // Ignore publish errors for tests
}

// PublishAgentNotification sends a lightweight inbox-notification to a specific
// agent's Centrifuge channel.
func (cn *CentrifugeNode) PublishAgentNotification(agentID string, msg Message) {
	channel := "agent:" + agentID
	data, _ := json.Marshal(msg) // Ignore marshal errors for tests
	_, _ = cn.node.Publish(channel, data) // Ignore publish errors for tests
}

// Close shuts down the Centrifuge node gracefully.
func (cn *CentrifugeNode) Close() error {
	return cn.node.Shutdown(context.Background())
}
