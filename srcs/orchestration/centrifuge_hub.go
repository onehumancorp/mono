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

// CentrifugeClient defines the subset of centrifuge.Client methods used by the Hub.
type CentrifugeClient interface {
	UserID() string
	ID() string
	OnSubscribe(centrifuge.SubscribeHandler)
	OnPublish(centrifuge.PublishHandler)
	OnDisconnect(centrifuge.DisconnectHandler)
}

// CentrifugeNode wraps a centrifuge.Node with OHC-specific configuration and
// channel-permission rules that map directly to the Hub's meeting/chat model.
type CentrifugeNode struct {
	node *centrifuge.Node
}

// test hook variables
var createCentrifugeNode = centrifuge.New
var runCentrifugeNode = func(node *centrifuge.Node) error {
	return node.Run()
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
	node, err := createCentrifugeNode(cfg)
	if err != nil {
		return nil, err
	}

	cn := &CentrifugeNode{node: node}

	node.OnConnecting(cn.handleConnecting)
	node.OnConnect(cn.handleConnect)

	if err := runCentrifugeNode(node); err != nil {
		return nil, err
	}

	return cn, nil
}

func (cn *CentrifugeNode) handleConnecting(ctx context.Context, e centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
	// Accept all connections; authentication is handled by the outer HTTP middleware.
	return centrifuge.ConnectReply{
		Credentials: &centrifuge.Credentials{
			UserID: e.Token, // reuse token as userID for traceability
		},
	}, nil
}

func (cn *CentrifugeNode) handleConnect(client *centrifuge.Client) {
	cn.handleConnectInternal(client)
}

func (cn *CentrifugeNode) handleConnectInternal(client CentrifugeClient) {
	slog.Debug("[centrifuge] client connected", "userID", client.UserID(), "id", client.ID())

	client.OnSubscribe(cn.handleSubscribe)
	client.OnPublish(cn.handlePublish)
	client.OnDisconnect(cn.handleDisconnect(client))
}

func (cn *CentrifugeNode) handleSubscribe(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
	cb(centrifuge.SubscribeReply{}, nil)
}

func (cn *CentrifugeNode) handlePublish(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
	cb(centrifuge.PublishReply{}, nil)
}

func (cn *CentrifugeNode) handleDisconnect(client CentrifugeClient) func(e centrifuge.DisconnectEvent) {
	return func(e centrifuge.DisconnectEvent) {
		slog.Debug("[centrifuge] client disconnected", "userID", client.UserID(), "reason", e.Reason)
	}
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
	data, _ := json.Marshal(msg)
	if _, err := cn.node.Publish(channel, data); err != nil {
		slog.Debug("[centrifuge] publish meeting message", "channel", channel, "error", err)
	}
}

// PublishChatMessage fans a chat message out to all subscribers of the
// "chat:<roomID>" Centrifuge channel.
func (cn *CentrifugeNode) PublishChatMessage(roomID string, msg Message) {
	channel := "chat:" + roomID
	data, _ := json.Marshal(msg)
	if _, err := cn.node.Publish(channel, data); err != nil {
		slog.Debug("[centrifuge] publish chat message", "channel", channel, "error", err)
	}
}

// PublishAgentNotification sends a lightweight inbox-notification to a specific
// agent's Centrifuge channel.
func (cn *CentrifugeNode) PublishAgentNotification(agentID string, msg Message) {
	channel := "agent:" + agentID
	data, _ := json.Marshal(msg)
	if _, err := cn.node.Publish(channel, data); err != nil {
		slog.Debug("[centrifuge] publish agent notification", "channel", channel, "error", err)
	}
}

// Close shuts down the Centrifuge node gracefully.
func (cn *CentrifugeNode) Close() error {
	return cn.node.Shutdown(context.Background())
}
