package chatwoot

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Setup Intent: Setup auto-configures the Chatwoot instance for OHC by: 1. Signing in with admin credentials (or signing up on first run). 2. Ensuring a default "OHC" API inbox exists (creates it if absent).  It retries up to maxAttempts times to accommodate Chatwoot's startup time. The configured Client is ready to use after a successful Setup call.
//
// Params: None.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (c *Client) Setup() error {
	email := os.Getenv("CHATWOOT_ADMIN_EMAIL")
	if email == "" {
		email = "admin@ohc.local"
	}
	password := os.Getenv("CHATWOOT_ADMIN_PASSWORD")
	if password == "" {
		password = "changeme"
	}

	const maxAttempts = 20
	const retryDelay = 5 * time.Second

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
			log.Printf("chatwoot setup: attempt %d/%d failed: %v; retrying in %s", i+1, maxAttempts, lastErr, retryDelay)
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

	log.Printf("chatwoot setup: ready (account_id=%d)", c.AccountID)
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

// IsEnabled Intent: IsEnabled reports whether Chatwoot is configured in the environment.
//
// Params: None.
//
// Returns:
//   - bool: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func IsEnabled() bool {
	return os.Getenv("CHATWOOT_URL") != "" || os.Getenv("CHATWOOT_ENABLED") == "true"
}
