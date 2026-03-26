package chat

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ChatBackendType defines the type of chat service.
type ChatBackendType string

const (
	BackendChatwoot ChatBackendType = "chatwoot"
	BackendSlack    ChatBackendType = "slack"
	BackendTelegram ChatBackendType = "telegram"
	BackendDiscord  ChatBackendType = "discord"
	BackendHub      ChatBackendType = "hub"
	BackendWebhook  ChatBackendType = "webhook"
)

// ChatBackend configuration.
type ChatBackend struct {
	Type ChatBackendType `json:"type"`
	URL  string          `json:"url,omitempty"`
}

// ChatChannel represents a single chat room or channel.
type ChatChannel struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organization_id"`
	Name           string            `json:"name"`
	Backend        ChatBackend       `json:"backend"`
	Config         map[string]string `json:"config"`
	Enabled        bool              `json:"enabled"`
	CreatedAt      time.Time         `json:"created_at"`
}

// NewChannel creates a new chat channel.
func NewChannel(orgID, name string, backend ChatBackend) *ChatChannel {
	return &ChatChannel{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		Name:           name,
		Backend:        backend,
		Config:         make(map[string]string),
		Enabled:        true,
		CreatedAt:      time.Now().UTC(),
	}
}

// ChatMessage represents a single message in a channel.
type ChatMessage struct {
	ID             string    `json:"id"`
	ChannelID      string    `json:"channel_id"`
	OrganizationID string    `json:"organization_id"`
	AuthorID       string    `json:"author_id"`
	AuthorName     string    `json:"author_name"`
	Body           string    `json:"body"`
	SentAt         time.Time `json:"sent_at"`
}

// NewMessage creates a new chat message.
func NewMessage(channelID, orgID, authorID, authorName, body string) *ChatMessage {
	return &ChatMessage{
		ID:             uuid.New().String(),
		ChannelID:      channelID,
		OrganizationID: orgID,
		AuthorID:       authorID,
		AuthorName:     authorName,
		Body:           body,
		SentAt:         time.Now().UTC(),
	}
}

// ChatTransport defines the interface for sending messages.
type ChatTransport interface {
	Send(channel *ChatChannel, message *ChatMessage) error
}

// NoopTransport is a mock transport for testing.
type NoopTransport struct{}

func (t *NoopTransport) Send(channel *ChatChannel, message *ChatMessage) error {
	return nil
}

// HubTransport handles real-time delivery via the orchestration Hub.
type HubTransport struct {
	// We use an interface here to avoid direct dependency on orchestration if possible,
	// but for now a direct reference is fine as long as there's no circularity.
	Publisher interface {
		Publish(from, to, room, msgType, content string) error
	}
}

func NewHubTransport(publisher interface {
	Publish(from, to, room, msgType, content string) error
}) *HubTransport {
	return &HubTransport{Publisher: publisher}
}

func (t *HubTransport) Send(channel *ChatChannel, message *ChatMessage) error {
	if t.Publisher == nil {
		return fmt.Errorf("hub publisher not initialized")
	}

	// Deliver via Hub
	return t.Publisher.Publish(message.AuthorID, "all", channel.ID, "chat", message.Body)
}

// ChatStore defines the interface for persisting chat data.
type ChatStore interface {
	SaveChannel(channel *ChatChannel) error
	GetChannel(id string) (*ChatChannel, error)
	ListChannels(orgID string) ([]*ChatChannel, error)
	SaveMessage(message *ChatMessage) error
	ListMessages(channelID, orgID string, limit int) ([]*ChatMessage, error)
}

// InMemoryChatStore is a thread-safe in-memory implementation of ChatStore.
type InMemoryChatStore struct {
	mu       sync.RWMutex
	channels map[string]*ChatChannel
	messages []*ChatMessage
}

func NewInMemoryStore() *InMemoryChatStore {
	return &InMemoryChatStore{
		channels: make(map[string]*ChatChannel),
		messages: make([]*ChatMessage, 0),
	}
}

func (s *InMemoryChatStore) SaveChannel(channel *ChatChannel) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[channel.ID] = channel
	return nil
}

func (s *InMemoryChatStore) GetChannel(id string) (*ChatChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ch, ok := s.channels[id]
	if !ok {
		return nil, fmt.Errorf("channel not found: %s", id)
	}
	return ch, nil
}

func (s *InMemoryChatStore) ListChannels(orgID string) ([]*ChatChannel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []*ChatChannel
	for _, ch := range s.channels {
		if ch.OrganizationID == orgID {
			results = append(results, ch)
		}
	}
	return results, nil
}

func (s *InMemoryChatStore) SaveMessage(message *ChatMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, message)
	return nil
}

func (s *InMemoryChatStore) ListMessages(channelID, orgID string, limit int) ([]*ChatMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []*ChatMessage
	for _, m := range s.messages {
		if m.ChannelID == channelID && m.OrganizationID == orgID {
			results = append(results, m)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].SentAt.Before(results[j].SentAt)
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

// ChatManager high-level manager for chat integrations.
type ChatManager struct {
	store     ChatStore
	transport ChatTransport
}

func NewChatManager(store ChatStore, transport ChatTransport) *ChatManager {
	return &ChatManager{store: store, transport: transport}
}

func (m *ChatManager) AddChannel(orgID, name string, backend ChatBackend) (*ChatChannel, error) {
	ch := NewChannel(orgID, name, backend)
	if err := m.store.SaveChannel(ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func (m *ChatManager) Send(channelID, orgID, authorID, authorName, body string) (*ChatMessage, error) {
	ch, err := m.store.GetChannel(channelID)
	if err != nil {
		return nil, err
	}
	if ch.OrganizationID != orgID {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	msg := NewMessage(channelID, orgID, authorID, authorName, body)
	if err := m.transport.Send(ch, msg); err != nil {
		return nil, err
	}
	if err := m.store.SaveMessage(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (m *ChatManager) Messages(channelID, orgID string, limit int) ([]*ChatMessage, error) {
	return m.store.ListMessages(channelID, orgID, limit)
}

func (m *ChatManager) ListChannels(orgID string) ([]*ChatChannel, error) {
	return m.store.ListChannels(orgID)
}
