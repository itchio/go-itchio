package itchio

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestOAuthClient(t *testing.T, server *httptest.Server, creds *OAuthCredentials, config OAuthConfig) *Client {
	t.Helper()
	client := NewOAuthClient(creds, config)
	client.HTTPClient = server.Client()
	client.BaseURL = server.URL
	return client
}

func TestRefreshRetainsRefreshTokenWhenOmitted(t *testing.T) {
	var refreshCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		values, err := url.ParseQuery(string(body))
		assert.NoError(t, err)
		assert.Equal(t, "refresh_token", values.Get("grant_type"))
		assert.Equal(t, "old-refresh", values.Get("refresh_token"))
		assert.Equal(t, "client-123", values.Get("client_id"))

		atomic.AddInt32(&refreshCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		// Server does not rotate refresh token (field empty)
		_, _ = w.Write([]byte(`{"accessToken":"new-access","expiresIn":120}`))
	}))
	defer server.Close()

	client := newTestOAuthClient(t, server, &OAuthCredentials{
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		// Expired to force refresh
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}, OAuthConfig{
		ClientID: "client-123",
	})

	err := client.refreshTokenIfNeeded(context.Background())
	assert.NoError(t, err)
	assert.EqualValues(t, 1, atomic.LoadInt32(&refreshCalls))

	client.oauth.credsMu.RLock()
	defer client.oauth.credsMu.RUnlock()
	assert.Equal(t, "new-access", client.oauth.creds.AccessToken)
	// Should retain previous refresh token when server omits it
	assert.Equal(t, "old-refresh", client.oauth.creds.RefreshToken)
	assert.False(t, client.oauth.creds.ExpiresAt.IsZero())
}

func TestNonPositiveExpiryDoesNotTriggerRefresh(t *testing.T) {
	var refreshCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCalls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestOAuthClient(t, server, &OAuthCredentials{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    expiresAtFromExpiresIn(0),
	}, OAuthConfig{
		ClientID: "client-123",
	})

	err := client.refreshTokenIfNeeded(context.Background())
	assert.NoError(t, err)
	assert.EqualValues(t, 0, atomic.LoadInt32(&refreshCalls), "refresh should not be attempted for non-expiring tokens")
}

func TestProactiveRefreshWithinBuffer(t *testing.T) {
	var refreshCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		atomic.AddInt32(&refreshCalls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accessToken":"refreshed","refreshToken":"rotated","expiresIn":300}`))
	}))
	defer server.Close()

	client := newTestOAuthClient(t, server, &OAuthCredentials{
		AccessToken:  "access",
		RefreshToken: "refresh",
		// Expiring soon (within buffer)
		ExpiresAt: time.Now().Add(30 * time.Second),
	}, OAuthConfig{
		ClientID:      "client-123",
		RefreshBuffer: 60 * time.Second,
	})

	err := client.refreshTokenIfNeeded(context.Background())
	assert.NoError(t, err)
	assert.EqualValues(t, 1, atomic.LoadInt32(&refreshCalls))

	client.oauth.credsMu.RLock()
	defer client.oauth.credsMu.RUnlock()
	assert.Equal(t, "refreshed", client.oauth.creds.AccessToken)
	assert.Equal(t, "rotated", client.oauth.creds.RefreshToken)
	assert.False(t, client.oauth.creds.ExpiresAt.IsZero())
}

func TestRetryOn401WithTokenRefresh(t *testing.T) {
	var (
		apiCalls     int32
		refreshCalls int32
		authHeaders  []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			atomic.AddInt32(&refreshCalls, 1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accessToken":"refreshed-token","refreshToken":"new-refresh","expiresIn":300}`))

		case "/profile":
			callNum := atomic.AddInt32(&apiCalls, 1)
			authHeaders = append(authHeaders, r.Header.Get("Authorization"))

			if callNum == 1 {
				// First call: return 401 to trigger refresh
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"errors":["invalid token"]}`))
				return
			}
			// Second call (after refresh): succeed
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"user":{"id":1}}`))

		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestOAuthClient(t, server, &OAuthCredentials{
		AccessToken:  "expired-token",
		RefreshToken: "my-refresh",
		// Not expired (so proactive refresh won't trigger)
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}, OAuthConfig{
		ClientID: "client-123",
	})

	// Make an API call - should get 401, refresh, then retry
	_, err := client.GetProfile(context.Background())
	assert.NoError(t, err)

	// Verify the flow
	assert.EqualValues(t, 2, atomic.LoadInt32(&apiCalls), "expected 2 API calls (initial + retry)")
	assert.EqualValues(t, 1, atomic.LoadInt32(&refreshCalls), "expected 1 refresh call")

	// Verify auth headers: first with old token, second with refreshed token
	assert.Len(t, authHeaders, 2)
	assert.Equal(t, "Bearer expired-token", authHeaders[0])
	assert.Equal(t, "Bearer refreshed-token", authHeaders[1])

	// Verify credentials were updated
	client.oauth.credsMu.RLock()
	defer client.oauth.credsMu.RUnlock()
	assert.Equal(t, "refreshed-token", client.oauth.creds.AccessToken)
	assert.Equal(t, "new-refresh", client.oauth.creds.RefreshToken)
}

func TestRetryOn401WithPOSTBody(t *testing.T) {
	var (
		apiCalls    int32
		requestBody string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accessToken":"new-token","expiresIn":300}`))

		case "/wharf/builds":
			callNum := atomic.AddInt32(&apiCalls, 1)

			// Read body on retry to verify it was preserved
			if callNum == 2 {
				body, _ := io.ReadAll(r.Body)
				requestBody = string(body)
			}

			if callNum == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"build":{"id":456,"uploadId":789}}`))

		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestOAuthClient(t, server, &OAuthCredentials{
		AccessToken:  "old-token",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}, OAuthConfig{
		ClientID: "client-123",
	})

	// CreateBuild is a POST request
	_, err := client.CreateBuild(context.Background(), CreateBuildParams{
		Target:  "user/game",
		Channel: "stable",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 2, atomic.LoadInt32(&apiCalls))

	// Verify POST body was preserved on retry
	values, err := url.ParseQuery(requestBody)
	assert.NoError(t, err)
	assert.Equal(t, "user/game", values.Get("target"))
	assert.Equal(t, "stable", values.Get("channel"))
}

func TestNo401RetryLoopOnPersistentFailure(t *testing.T) {
	var (
		apiCalls     int32
		refreshCalls int32
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			atomic.AddInt32(&refreshCalls, 1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accessToken":"still-bad","expiresIn":300}`))

		case "/profile":
			atomic.AddInt32(&apiCalls, 1)
			// Always return 401 (simulates revoked access)
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"errors":["invalid token"]}`))

		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestOAuthClient(t, server, &OAuthCredentials{
		AccessToken:  "bad-token",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}, OAuthConfig{
		ClientID: "client-123",
	})

	_, err := client.GetProfile(context.Background())

	// Should fail after one retry attempt (not infinite loop)
	assert.Error(t, err)
	assert.EqualValues(t, 2, atomic.LoadInt32(&apiCalls), "expected exactly 2 API calls")
	assert.EqualValues(t, 1, atomic.LoadInt32(&refreshCalls), "expected exactly 1 refresh attempt")
}
