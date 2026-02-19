package itchio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func decodeJSONLogRecords(t *testing.T, b *bytes.Buffer) []map[string]interface{} {
	t.Helper()

	var records []map[string]interface{}
	for _, line := range strings.Split(strings.TrimSpace(b.String()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var record map[string]interface{}
		err := json.Unmarshal([]byte(line), &record)
		assert.NoError(t, err)
		records = append(records, record)
	}

	return records
}

func TestSlogHTTP_SingleSuccess(t *testing.T) {
	server, client := testTools(200, `{"games":[]}`)
	defer server.Close()

	client.RetryPatterns = nil
	client.Limiter = rate.NewLimiter(rate.Inf, 1)

	var logs bytes.Buffer
	client.Logger = slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	_, err := client.ListProfileGames(context.Background())
	assert.NoError(t, err)

	records := decodeJSONLogRecords(t, &logs)
	assert.Len(t, records, 1)

	record := records[0]
	assert.EqualValues(t, "http request", record["msg"])
	assert.EqualValues(t, "GET", record["method"])
	assert.EqualValues(t, 200, record["status_code"])
	assert.EqualValues(t, 1, record["attempt"])
	assert.EqualValues(t, false, record["retrying"])
	_, hasRetryReason := record["retry_reason"]
	assert.False(t, hasRetryReason)

	loggedURL, ok := record["url"].(string)
	assert.True(t, ok)
	assert.Contains(t, loggedURL, "/profile/games")
	assert.NotContains(t, loggedURL, "APIKEY")
}

func TestSlogHTTP_503ThenSuccess(t *testing.T) {
	var calls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"errors":["rate limited"]}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"games":[]}`))
	}))
	defer server.Close()

	client := ClientWithKey("APIKEY")
	client.HTTPClient = server.Client()
	client.BaseURL = server.URL
	client.RetryPatterns = []time.Duration{0}
	client.Limiter = rate.NewLimiter(rate.Inf, 1)

	var logs bytes.Buffer
	client.Logger = slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	_, err := client.ListProfileGames(context.Background())
	assert.NoError(t, err)

	records := decodeJSONLogRecords(t, &logs)
	assert.Len(t, records, 2)

	assert.EqualValues(t, 503, records[0]["status_code"])
	assert.EqualValues(t, true, records[0]["retrying"])
	assert.EqualValues(t, "rate_limited_503", records[0]["retry_reason"])
	assert.EqualValues(t, 1, records[0]["attempt"])

	assert.EqualValues(t, 200, records[1]["status_code"])
	assert.EqualValues(t, false, records[1]["retrying"])
	_, hasRetryReason := records[1]["retry_reason"]
	assert.False(t, hasRetryReason)
	assert.EqualValues(t, 2, records[1]["attempt"])
}

func TestSlogHTTP_OAuth401RefreshRetry(t *testing.T) {
	var apiCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"accessToken":"refreshed-token","refreshToken":"new-refresh","expiresIn":300}`))
		case "/profile":
			if atomic.AddInt32(&apiCalls, 1) == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"errors":["invalid token"]}`))
				return
			}

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
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}, OAuthConfig{
		ClientID: "client-123",
	})
	client.RetryPatterns = nil
	client.Limiter = rate.NewLimiter(rate.Inf, 1)

	var logs bytes.Buffer
	client.Logger = slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	_, err := client.GetProfile(context.Background())
	assert.NoError(t, err)

	records := decodeJSONLogRecords(t, &logs)

	var profileLogs []map[string]interface{}
	for _, record := range records {
		loggedURL, ok := record["url"].(string)
		if ok && strings.Contains(loggedURL, "/profile") {
			profileLogs = append(profileLogs, record)
		}
	}

	assert.Len(t, profileLogs, 2)
	assert.EqualValues(t, 401, profileLogs[0]["status_code"])
	assert.EqualValues(t, true, profileLogs[0]["retrying"])
	assert.EqualValues(t, "oauth_401_refresh", profileLogs[0]["retry_reason"])
	assert.EqualValues(t, 1, profileLogs[0]["attempt"])

	assert.EqualValues(t, 200, profileLogs[1]["status_code"])
	assert.EqualValues(t, false, profileLogs[1]["retrying"])
	_, hasRetryReason := profileLogs[1]["retry_reason"]
	assert.False(t, hasRetryReason)
	assert.EqualValues(t, 2, profileLogs[1]["attempt"])
}

func TestSlogHTTP_URLRedaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := ClientWithKey("APIKEY")
	client.HTTPClient = server.Client()
	client.RetryPatterns = nil
	client.Limiter = rate.NewLimiter(rate.Inf, 1)

	var logs bytes.Buffer
	client.Logger = slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	req, err := http.NewRequest("GET", server.URL+"/example?api_key=ak&password=pw&secret=sc&refresh_token=rt&code=cd&code_verifier=cv&token=tk&safe=value", nil)
	assert.NoError(t, err)

	res, err := client.Do(req)
	assert.NoError(t, err)
	_, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()

	records := decodeJSONLogRecords(t, &logs)
	assert.Len(t, records, 1)

	loggedURL, ok := records[0]["url"].(string)
	assert.True(t, ok)
	assert.Contains(t, loggedURL, "api_key=REDACTED")
	assert.Contains(t, loggedURL, "password=REDACTED")
	assert.Contains(t, loggedURL, "secret=REDACTED")
	assert.Contains(t, loggedURL, "refresh_token=REDACTED")
	assert.Contains(t, loggedURL, "code=REDACTED")
	assert.Contains(t, loggedURL, "code_verifier=REDACTED")
	assert.Contains(t, loggedURL, "token=REDACTED")
	assert.Contains(t, loggedURL, "safe=value")
	assert.NotContains(t, loggedURL, "ak")
	assert.NotContains(t, loggedURL, "pw")
	assert.NotContains(t, loggedURL, "sc")
	assert.NotContains(t, loggedURL, "rt")
	assert.NotContains(t, loggedURL, "cd")
	assert.NotContains(t, loggedURL, "cv")
	assert.NotContains(t, loggedURL, "tk")
}

func TestSlogHTTP_NonRetriableTransportError(t *testing.T) {
	client := ClientWithKey("APIKEY")
	client.HTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("boom")
		}),
	}
	client.RetryPatterns = nil
	client.Limiter = rate.NewLimiter(rate.Inf, 1)

	var logs bytes.Buffer
	client.Logger = slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	req, err := http.NewRequest("GET", "https://api.itch.io/profile", nil)
	assert.NoError(t, err)

	_, err = client.Do(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boom")

	records := decodeJSONLogRecords(t, &logs)
	assert.Len(t, records, 1)
	assert.EqualValues(t, 0, records[0]["status_code"])
	assert.EqualValues(t, false, records[0]["retrying"])
	_, hasRetryReason := records[0]["retry_reason"]
	assert.False(t, hasRetryReason)
	assert.Contains(t, records[0]["error"], "boom")
}
