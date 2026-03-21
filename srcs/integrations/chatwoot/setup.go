package chatwoot

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

// Summary: Setup functionality.
// Parameters: None
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (c *Client) Setup() error {
	email := os.Getenv("CHATWOOT_ADMIN_EMAIL")
	if email == "" {
		email = "admin@ohc.local"
	}
	password := os.Getenv("CHATWOOT_ADMIN_PASSWORD")
	if password == "" {
		password = "changeme"
	}

	maxAttempts := 20
	retryDelay := 5 * time.Second
	if os.Getenv("CI") != "" || os.Getenv("GO_TEST") == "true" {
		maxAttempts = 1
		retryDelay = 0
	}

	var lastErr error
	for i := range maxAttempts {
		// Try sign-in first; on a fresh instance fall back to sign-up.
		lastErr = c.SignIn(email, password)
		if lastErr != nil {
			if signupErr := c.signUp(email, password); signupErr == nil {
				// sign-up succeeded; now sign in.
				lastErr = c.SignIn(email, password)
			}
		}
		if lastErr == nil {
			break
		}
		if i < maxAttempts-1 {
			slog.Warn("chatwoot setup attempt failed", "attempt", i+1, "maxAttempts", maxAttempts, "error", lastErr, "retryDelay", retryDelay)
			time.Sleep(retryDelay)
		}
	}
	if lastErr != nil {
		return fmt.Errorf("chatwoot setup: could not authenticate after %d attempts: %w", maxAttempts, lastErr)
	}

	// Ensure default inbox exists.
	if _, err := c.ensureInbox("OHC"); err != nil {
		return fmt.Errorf("chatwoot setup: ensure inbox: %w", err)
	}

	slog.Info("chatwoot setup: ready", "account_id", c.AccountID)
	return nil
}

// signUp attempts to create the initial admin account on a fresh Chatwoot instance.
func (c *Client) signUp(email, password string) error {
	body := map[string]string{
		"email":                 email,
		"password":              password,
		"password_confirmation": password,
		"account_name":          "OHC",
	}
	var result map[string]interface{}
	return c.post("/auth/sign_up", body, &result)
}

// ensureInbox returns an inbox with the given name, creating it if necessary.
func (c *Client) ensureInbox(name string) (Inbox, error) {
	inboxes, err := c.ListInboxes()
	if err != nil {
		return Inbox{}, err
	}
	for _, ib := range inboxes {
		if ib.Name == name {
			return ib, nil
		}
	}
	return c.CreateAPIInbox(name)
}

// Summary: IsEnabled functionality.
// Parameters: None
// Returns: bool
// Errors: None
// Side Effects: None
func IsEnabled() bool {
	return os.Getenv("CHATWOOT_URL") != "" || os.Getenv("CHATWOOT_ENABLED") == "true"
}
