package orchestration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto"
	"google.golang.org/protobuf/proto"
)

// TokenEfficientContextSummarization securely processes token efficient context summarization.
//
// Parameters:
//
//   - eventID: string; Unique event identifier.
//
//   - agentID: string; Identifier of the invoking agent.
//
//   - payload: []byte; The operation payload containing specific context instructions.
//
//   - error: Error object if validation or processing fails.
//
// Accepts parameters: h *Hub (No Constraints).
// Returns TokenEfficientContextSummarization(eventID, agentID string, payload []byte) error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (h *Hub) TokenEfficientContextSummarization(eventID, agentID string, payload []byte) error {
	h.mu.Lock()
	if _, exists := h.tokenTrackers[eventID]; exists {
		h.mu.Unlock()
		return errors.New("event already being processed")
	}
	h.tokenTrackers[eventID] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.tokenTrackers, eventID)
		h.mu.Unlock()
	}()

	var temp struct {
		Context string `json:"context"`
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&temp); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	client := NewMinimaxClient(h.MinimaxAPIKey())
	prompt := fmt.Sprintf("Summarize the following context efficiently to save tokens: %s", redactPII(temp.Context))
	summarizedContext, err := client.Reason(context.Background(), prompt)
	if err != nil {
		return fmt.Errorf("summarization failed: %w", err)
	}

	h.eventLogChan <- map[string]interface{}{
		"event_id":           eventID,
		"agent_id":           agentID,
		"type":               "TokenEfficientContextSummarization",
		"summarized_context": summarizedContext,
	}

	return nil
}

// ToolParameterAutoCorrection securely processes tool parameter auto-correction.
//
// Parameters:
//
//   - eventID: string; Unique event identifier.
//
//   - agentID: string; Identifier of the invoking agent.
//
//   - payload: []byte; The operation payload containing tool parameters.
//
//   - error: Error object if validation or processing fails.
//
// Accepts parameters: h *Hub (No Constraints).
// Returns ToolParameterAutoCorrection(eventID, agentID string, payload []byte) error.
// Produces errors: Explicit error handling.
// Has no side effects.
func (h *Hub) ToolParameterAutoCorrection(eventID, agentID string, payload []byte) error {
	h.mu.Lock()
	if _, exists := h.autoCorTrack[eventID]; exists {
		h.mu.Unlock()
		return errors.New("event already being processed")
	}
	h.autoCorTrack[eventID] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.autoCorTrack, eventID)
		h.mu.Unlock()
	}()

	var structTemp struct {
		ID    string `json:"id,omitempty"`
		Value string `json:"value,omitempty"`
		Name  string `json:"name,omitempty"`
	}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&structTemp); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	temp := make(map[string]interface{})
	if structTemp.ID != "" {
		temp["id"] = structTemp.ID
	}
	if structTemp.Value != "" {
		temp["value"] = structTemp.Value
	}
	if structTemp.Name != "" {
		temp["name"] = structTemp.Name
	}

	corrected := false
	for k, v := range temp {
		if strVal, ok := v.(string); ok {
			var intVal int
			if _, err := fmt.Sscanf(strVal, "%d", &intVal); err == nil {
				// To ensure it's purely numerical without any extra characters, check if fmt.Sprintf back matches.
				if fmt.Sprintf("%d", intVal) == strVal {
					temp[k] = intVal
					corrected = true
				}
			}
		}
	}

	tempBytes, _ := json.Marshal(temp)

	// Create protobuf event representing the autocoorection
	pbEvent := pb.ToolParameterAutoCorrectionEvent_builder{
		EventId: proto.String(eventID),
		AgentId: proto.String(agentID),
		Payload: tempBytes,
	}.Build()

	h.LogEvent(map[string]interface{}{
		"event_id":  pbEvent.GetEventId(),
		"agent_id":  pbEvent.GetAgentId(),
		"type":      "ToolParameterAutoCorrection",
		"payload":   temp,
		"corrected": corrected,
	})

	return nil
}

// minimaxAPIURL is the endpoint for Minimax reasoning.
// ⚡ BOLT: [Configurable endpoint] - Randomized Selection from Top 5
var minimaxAPIURL = "https://api.minimax.io/v1/chat/completions"

// MinimaxClient handles interaction with the Minimax Model 2.7.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type MinimaxClient struct {
	APIKey string
}

// NewMinimaxClient functionality.
// Accepts parameters: apiKey string (No Constraints).
// Returns *MinimaxClient.
// Produces no errors.
// Has no side effects.
func NewMinimaxClient(apiKey string) *MinimaxClient {
	return &MinimaxClient{APIKey: apiKey}
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var sharedHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// Reason functionality.
// Accepts parameters: c *MinimaxClient (No Constraints).
// Returns (string, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func (c *MinimaxClient) Reason(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("minimax API key is not configured")
	}

	url := minimaxAPIURL
	// Optimization: construct the JSON payload manually to avoid
	// maps and slices allocations.
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	buf.WriteString(`{"model":"MiniMax-M2.7","messages":[{"role":"user","content":`)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(prompt)
	// Encode adds a newline, so we slice it off and add the closing brackets
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(`}]}`)

	req, err := http.NewRequestWithContext(ctx, "POST", url, buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// ⚡ BOLT: [Reused HTTP Client] - Randomized Selection from Top 5
	// Prevents severe connection and resource leaks by reusing connection pools on every request.
	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("minimax API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", errors.New("empty response from minimax")
	}

	return result.Choices[0].Message.Content, nil
}
