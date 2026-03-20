// Package chatwoot provides a REST API client for Chatwoot, the open-source
// customer-support / chat platform used by OHC as its default meeting-room
// and chat infrastructure.
//
// Environment variables consumed:
//
//	CHATWOOT_URL             – base URL of the Chatwoot instance (default: http://chatwoot:3000)
//	CHATWOOT_ADMIN_EMAIL     – admin account email used for auto-setup (default: admin@ohc.local)
//	CHATWOOT_ADMIN_PASSWORD  – admin account password (default: changeme)
package chatwoot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Summary: DefaultBaseURL is the in-cluster URL for the Chatwoot service.
// Intent: DefaultBaseURL is the in-cluster URL for the Chatwoot service.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
const DefaultBaseURL = "http://chatwoot:3000"

// Summary: Client interacts with the Chatwoot REST API v1.
// Intent: Client interacts with the Chatwoot REST API v1.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Client struct {
	BaseURL     string
	AccessToken string // api_access_token for authenticated requests
	AccountID   int
	httpClient  *http.Client
}

// Summary: NewClientFromEnv creates a Client using environment variables. CHATWOOT_URL overrides the base URL.
// Intent: NewClientFromEnv creates a Client using environment variables. CHATWOOT_URL overrides the base URL.
// Params: None
// Returns: *Client
// Errors: None
// Side Effects: None
func NewClientFromEnv() *Client {
	base := os.Getenv("CHATWOOT_URL")
	if base == "" {
		base = DefaultBaseURL
	}
	return &Client{
		BaseURL:    base,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Summary: NewClient creates a Client with an explicit base URL (useful in tests).
// Intent: NewClient creates a Client with an explicit base URL (useful in tests).
// Params: baseURL
// Returns: *Client
// Errors: None
// Side Effects: None
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// ── Auth ──────────────────────────────────────────────────────────────────────

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signInData struct {
	AccessToken string `json:"access_token"`
	AccountID   int    `json:"account_id"`
}

type signInResponse struct {
	Data signInData `json:"data"`
}

// Summary: SignIn authenticates with Chatwoot and stores the resulting access token and account ID on the Client.
// Intent: SignIn authenticates with Chatwoot and stores the resulting access token and account ID on the Client.
// Params: email, password
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (c *Client) SignIn(email, password string) error {
	var resp signInResponse
	if err := c.post("/auth/sign_in", signInRequest{Email: email, Password: password}, &resp); err != nil {
		return fmt.Errorf("chatwoot sign-in: %w", err)
	}
	if resp.Data.AccessToken == "" {
		return fmt.Errorf("chatwoot sign-in: empty access token in response")
	}
	c.AccessToken = resp.Data.AccessToken
	c.AccountID = resp.Data.AccountID
	return nil
}

// ── Inboxes ───────────────────────────────────────────────────────────────────

// Summary: Inbox represents a Chatwoot inbox (a communication channel).
// Intent: Inbox represents a Chatwoot inbox (a communication channel).
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Inbox struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ChannelID string `json:"channel_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type inboxListResponse struct {
	Payload []Inbox `json:"payload"`
}

type createInboxRequest struct {
	Name        string            `json:"name"`
	ChannelType string            `json:"channel_type"`
	Channel     map[string]string `json:"channel"`
}

// Summary: ListInboxes returns all inboxes in the account.
// Intent: ListInboxes returns all inboxes in the account.
// Params: None
// Returns: ([]Inbox, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (c *Client) ListInboxes() ([]Inbox, error) {
	var resp inboxListResponse
	if err := c.get(c.accountPath("/inboxes"), &resp); err != nil {
		return nil, fmt.Errorf("chatwoot list inboxes: %w", err)
	}
	return resp.Payload, nil
}

// Summary: CreateAPIInbox creates a new API-type inbox with the given name.
// Intent: CreateAPIInbox creates a new API-type inbox with the given name.
// Params: name
// Returns: (Inbox, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (c *Client) CreateAPIInbox(name string) (Inbox, error) {
	body := createInboxRequest{
		Name:        name,
		ChannelType: "Channel::Api",
		Channel:     map[string]string{"welcome_title": name},
	}
	var created Inbox
	if err := c.post(c.accountPath("/inboxes"), body, &created); err != nil {
		return Inbox{}, fmt.Errorf("chatwoot create inbox: %w", err)
	}
	return created, nil
}

// ── Contacts ──────────────────────────────────────────────────────────────────

// Summary: Contact represents a Chatwoot contact (a participant in conversations).
// Intent: Contact represents a Chatwoot contact (a participant in conversations).
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Contact struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type createContactRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Summary: CreateContact creates a new contact.
// Intent: CreateContact creates a new contact.
// Params: name, email
// Returns: (Contact, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (c *Client) CreateContact(name, email string) (Contact, error) {
	var contact Contact
	if err := c.post(c.accountPath("/contacts"), createContactRequest{Name: name, Email: email}, &contact); err != nil {
		return Contact{}, fmt.Errorf("chatwoot create contact: %w", err)
	}
	return contact, nil
}

// ── Conversations ─────────────────────────────────────────────────────────────

// Summary: Conversation represents a Chatwoot conversation (a chat thread).
// Intent: Conversation represents a Chatwoot conversation (a chat thread).
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Conversation struct {
	ID          int    `json:"id"`
	InboxID     int    `json:"inbox_id"`
	ContactID   int    `json:"contact_id,omitempty"`
	AccountID   int    `json:"account_id"`
	DisplayID   int    `json:"display_id"`
}

type createConversationRequest struct {
	InboxID          int    `json:"inbox_id"`
	ContactID        int    `json:"contact_id,omitempty"`
	AdditionalAttributes map[string]string `json:"additional_attributes,omitempty"`
}

// Summary: CreateConversation opens a new conversation in the given inbox.
// Intent: CreateConversation opens a new conversation in the given inbox.
// Params: inboxID, contactID
// Returns: (Conversation, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (c *Client) CreateConversation(inboxID, contactID int) (Conversation, error) {
	body := createConversationRequest{
		InboxID:   inboxID,
		ContactID: contactID,
	}
	var conv Conversation
	if err := c.post(c.accountPath("/conversations"), body, &conv); err != nil {
		return Conversation{}, fmt.Errorf("chatwoot create conversation: %w", err)
	}
	return conv, nil
}

// ── Messages ──────────────────────────────────────────────────────────────────

// Summary: Message represents a single message in a conversation.
// Intent: Message represents a single message in a conversation.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Message struct {
	ID             int    `json:"id"`
	Content        string `json:"content"`
	MessageType    int    `json:"message_type"` // 0=incoming, 1=outgoing
	CreatedAt      int64  `json:"created_at"`
	ConversationID int    `json:"conversation_id"`
}

type createMessageRequest struct {
	Content     string `json:"content"`
	MessageType string `json:"message_type"` // "outgoing" or "incoming"
	Private     bool   `json:"private"`
}

type messageListResponse struct {
	Payload []Message `json:"payload"`
}

// Summary: SendMessage posts a message into a conversation.
// Intent: SendMessage posts a message into a conversation.
// Params: conversationID, content, messageType
// Returns: (Message, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (c *Client) SendMessage(conversationID int, content, messageType string) (Message, error) {
	if messageType == "" {
		messageType = "outgoing"
	}
	body := createMessageRequest{Content: content, MessageType: messageType}
	var msg Message
	path := fmt.Sprintf("%s/conversations/%d/messages", c.accountBase(), conversationID)
	if err := c.post(path, body, &msg); err != nil {
		return Message{}, fmt.Errorf("chatwoot send message: %w", err)
	}
	return msg, nil
}

// Summary: ListMessages returns all messages in a conversation.
// Intent: ListMessages returns all messages in a conversation.
// Params: conversationID
// Returns: ([]Message, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (c *Client) ListMessages(conversationID int) ([]Message, error) {
	var resp messageListResponse
	path := fmt.Sprintf("%s/conversations/%d/messages", c.accountBase(), conversationID)
	if err := c.get(path, &resp); err != nil {
		return nil, fmt.Errorf("chatwoot list messages: %w", err)
	}
	return resp.Payload, nil
}

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func (c *Client) accountBase() string {
	return fmt.Sprintf("/api/v1/accounts/%d", c.AccountID)
}

func (c *Client) accountPath(suffix string) string {
	return c.accountBase() + suffix
}

func (c *Client) get(path string, dest interface{}) error {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	c.addHeaders(req)
	return c.do(req, dest)
}

func (c *Client) post(path string, body, dest interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.addHeaders(req)
	return c.do(req, dest)
}

func (c *Client) addHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	if c.AccessToken != "" {
		req.Header.Set("api_access_token", c.AccessToken)
	}
}

func (c *Client) do(req *http.Request, dest interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chatwoot API %s %s returned %d: %s", req.Method, req.URL.Path, resp.StatusCode, string(b))
	}

	if dest != nil {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return fmt.Errorf("chatwoot decode response: %w", err)
		}
	}
	return nil
}
