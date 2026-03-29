package minimax

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
)

var APIURL = "https://api.minimax.io/v1/chat/completions"

type Client struct {
	APIKey string
}

func NewClient(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}


type CircuitBreaker struct {
	mu           sync.Mutex
	failures     int
	lastFailure  time.Time
	state        string // CLOSED, OPEN, HALF_OPEN
	threshold    int
	timeout      time.Duration
}

var cb = &CircuitBreaker{
	state:     "CLOSED",
	threshold: 5,
	timeout:   30 * time.Second,
}


func (c *CircuitBreaker) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = "CLOSED"
	c.failures = 0
}

func ResetCircuitBreaker() {
	cb.Reset()
}

func (c *CircuitBreaker) Execute(req func() (*http.Response, error)) (*http.Response, error) {
	c.mu.Lock()
	if c.state == "OPEN" {
		if time.Since(c.lastFailure) > c.timeout {
			c.state = "HALF_OPEN"
		} else {
			c.mu.Unlock()
			return nil, errors.New("circuit breaker is OPEN")
		}
	}
	c.mu.Unlock()

	resp, err := req()

	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil || (resp != nil && resp.StatusCode >= 500) {
		c.failures++
		c.lastFailure = time.Now()
		if c.state == "HALF_OPEN" || c.failures >= c.threshold {
			c.state = "OPEN"
		}
		return resp, err
	}

	c.state = "CLOSED"
	c.failures = 0
	return resp, nil
}

var sharedHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

func (c *Client) Reason(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", errors.New("minimax API key is not configured")
	}

	url := APIURL
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	buf.WriteString(`{"model":"MiniMax-M2.7","messages":[{"role":"user","content":`)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(prompt)
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(`}]}`)

	req, err := http.NewRequestWithContext(ctx, "POST", url, buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := cb.Execute(func() (*http.Response, error) { return sharedHTTPClient.Do(req) })
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
