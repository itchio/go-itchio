package itchio

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/itchio/httpkit/timeout"
)

// contextKey is used for context values to avoid collisions
type contextKey string

// skipOAuthRefreshKey marks a context as being part of a token refresh operation.
// Requests with this context skip OAuth refresh logic to avoid deadlocks.
const skipOAuthRefreshKey contextKey = "skipOAuthRefresh"

// OAuthCredentials holds the current OAuth token state
type OAuthCredentials struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Copy returns a copy of the credentials (not a reference to internal state)
func (c *OAuthCredentials) Copy() *OAuthCredentials {
	return &OAuthCredentials{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		ExpiresAt:    c.ExpiresAt,
	}
}

// IsExpired returns true if the token has expired
func (c *OAuthCredentials) IsExpired() bool {
	if c == nil || c.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(c.ExpiresAt)
}

// ExpiresWithin returns true if the token expires within the given duration
func (c *OAuthCredentials) ExpiresWithin(d time.Duration) bool {
	if c == nil || c.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().Add(d).After(c.ExpiresAt)
}

// OnTokenRefresh is called after successful token refresh.
// Implementations should persist the new credentials.
// Errors are logged but do not prevent the API call from proceeding.
type OnTokenRefresh func(creds *OAuthCredentials) error

// OAuthConfig holds configuration for OAuth client behavior
type OAuthConfig struct {
	// ClientID is required for token refresh requests
	ClientID string

	// OnRefresh is called after each successful token refresh
	OnRefresh OnTokenRefresh

	// RefreshBuffer is how early before expiry to proactively refresh.
	// Defaults to 60 seconds if zero.
	RefreshBuffer time.Duration
}

// oauthState holds the internal OAuth state for a Client
type oauthState struct {
	creds     *OAuthCredentials
	config    OAuthConfig
	credsMu   sync.RWMutex
	refreshMu sync.Mutex
}

// DefaultRefreshBuffer is the default duration before expiry to refresh tokens
const DefaultRefreshBuffer = 60 * time.Second

// NewOAuthClient creates a client using OAuth credentials and configuration.
// The client will automatically refresh tokens when they are near expiry or
// when a 401 response is received.
//
// Panics if creds is nil or config.ClientID is empty.
func NewOAuthClient(creds *OAuthCredentials, config OAuthConfig) *Client {
	if creds == nil {
		panic("itchio: NewOAuthClient called with nil credentials")
	}
	if config.ClientID == "" {
		panic("itchio: NewOAuthClient called with empty ClientID")
	}
	if config.RefreshBuffer == 0 {
		config.RefreshBuffer = DefaultRefreshBuffer
	}

	c := &Client{
		HTTPClient:       timeout.NewDefaultClient(),
		RetryPatterns:    defaultRetryPatterns(),
		UserAgent:        "go-itchio",
		AcceptedLanguage: "*",
		Limiter:          DefaultRateLimiter(),
		oauth: &oauthState{
			creds:  creds.Copy(),
			config: config,
		},
	}
	c.SetServer("https://api.itch.io")
	return c
}

// expiresAtFromExpiresIn converts an expiresIn value to a time.Time.
// Non-positive values are treated as non-expiring (zero time).
func expiresAtFromExpiresIn(expiresIn int64) time.Time {
	if expiresIn <= 0 {
		return time.Time{}
	}
	return time.Now().Add(time.Duration(expiresIn) * time.Second)
}

// tokenNeedsRefresh checks if the token needs to be refreshed
func (c *Client) tokenNeedsRefresh() bool {
	if c.oauth == nil {
		return false
	}

	c.oauth.credsMu.RLock()
	defer c.oauth.credsMu.RUnlock()

	if c.oauth.creds == nil {
		return false
	}

	return c.oauth.creds.ExpiresWithin(c.oauth.config.RefreshBuffer)
}

// refreshTokenIfNeeded refreshes the token if it's near expiry.
// Uses double-check pattern to prevent thundering herd.
func (c *Client) refreshTokenIfNeeded(ctx context.Context) error {
	if c.oauth == nil {
		return nil
	}

	// Quick check without lock
	if !c.tokenNeedsRefresh() {
		return nil
	}

	// Acquire refresh lock - only one goroutine refreshes at a time
	c.oauth.refreshMu.Lock()
	defer c.oauth.refreshMu.Unlock()

	// Re-check after acquiring lock (another goroutine may have refreshed)
	if !c.tokenNeedsRefresh() {
		return nil
	}

	return c.doTokenRefresh(ctx)
}

// forceTokenRefresh forces a token refresh regardless of expiry.
// Used for reactive refresh on 401 responses.
func (c *Client) forceTokenRefresh(ctx context.Context) error {
	if c.oauth == nil {
		return nil
	}

	c.oauth.refreshMu.Lock()
	defer c.oauth.refreshMu.Unlock()

	return c.doTokenRefresh(ctx)
}

// doTokenRefresh performs the actual token refresh.
// Must be called with refreshMu held.
func (c *Client) doTokenRefresh(ctx context.Context) error {
	c.oauth.credsMu.RLock()
	refreshToken := c.oauth.creds.RefreshToken
	c.oauth.credsMu.RUnlock()

	// Use a context that skips OAuth refresh logic for the refresh request itself.
	// This prevents deadlock (we're holding refreshMu) and ensures the refresh
	// endpoint doesn't use the expired bearer token.
	refreshCtx := context.WithValue(ctx, skipOAuthRefreshKey, true)

	resp, err := c.RefreshOAuthToken(refreshCtx, RefreshOAuthTokenParams{
		RefreshToken: refreshToken,
		ClientID:     c.oauth.config.ClientID,
	})
	if err != nil {
		return err
	}

	newRefreshToken := resp.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = refreshToken
	}

	newCreds := &OAuthCredentials{
		AccessToken:  resp.AccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAtFromExpiresIn(resp.ExpiresIn),
	}

	// Update internal state
	c.oauth.credsMu.Lock()
	c.oauth.creds = newCreds
	c.oauth.credsMu.Unlock()

	// Notify callback (errors logged, not propagated)
	if c.oauth.config.OnRefresh != nil {
		if err := c.oauth.config.OnRefresh(newCreds.Copy()); err != nil {
			log.Printf("go-itchio: token refresh callback error: %v", err)
		}
	}

	return nil
}

// getAuthHeader returns the appropriate Authorization header value
func (c *Client) getAuthHeader() string {
	if c.oauth != nil {
		c.oauth.credsMu.RLock()
		defer c.oauth.credsMu.RUnlock()
		if c.oauth.creds != nil {
			return "Bearer " + c.oauth.creds.AccessToken
		}
	}
	return c.Key
}

// isOAuthClient returns true if this client uses OAuth authentication
func (c *Client) isOAuthClient() bool {
	return c.oauth != nil
}

// shouldSkipOAuthRefresh returns true if the context indicates we're in a refresh operation
func shouldSkipOAuthRefresh(ctx context.Context) bool {
	return ctx.Value(skipOAuthRefreshKey) != nil
}
