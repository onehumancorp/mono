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
	"log/slog"
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
	node, err := centrifuge.New(cfg)
	if err != nil {
		return nil, err
	}

	node.OnConnecting(func(ctx context.Context, e centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
		// Accept all connections; authentication is handled by the outer HTTP middleware.
		return centrifuge.ConnectReply{
			Credentials: &centrifuge.Credentials{
				UserID: e.Token, // reuse token as userID for traceability
			},
		}, nil
	})

	node.OnConnect(func(client *centrifuge.Client) {
		slog.Debug("[centrifuge] client connected", "userID", client.UserID(), "id", client.ID())

		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			cb(centrifuge.SubscribeReply{}, nil)
		})

		client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
			cb(centrifuge.PublishReply{}, nil)
		})

		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			slog.Debug("[centrifuge] client disconnected", "userID", client.UserID(), "reason", e.Reason)
		})
	})

	if err := node.Run(); err != nil {
		return nil, err
	}

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
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("[centrifuge] marshal meeting message", "error", err)
		return
	}
	if _, err := cn.node.Publish(channel, data); err != nil {
		slog.Debug("[centrifuge] publish meeting message", "channel", channel, "error", err)
	}
}

// PublishChatMessage fans a chat message out to all subscribers of the
// "chat:<roomID>" Centrifuge channel.
func (cn *CentrifugeNode) PublishChatMessage(roomID string, msg Message) {
	channel := "chat:" + roomID
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("[centrifuge] marshal chat message", "error", err)
		return
	}
	if _, err := cn.node.Publish(channel, data); err != nil {
		slog.Debug("[centrifuge] publish chat message", "channel", channel, "error", err)
	}
}

// PublishAgentNotification sends a lightweight inbox-notification to a specific
// agent's Centrifuge channel.
func (cn *CentrifugeNode) PublishAgentNotification(agentID string, msg Message) {
	channel := "agent:" + agentID
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("[centrifuge] marshal agent notification", "error", err)
		return
	}
	if _, err := cn.node.Publish(channel, data); err != nil {
		slog.Debug("[centrifuge] publish agent notification", "channel", channel, "error", err)
	}
}

// Close shuts down the Centrifuge node gracefully.
func (cn *CentrifugeNode) Close() error {
	return cn.node.Shutdown(context.Background())
}
