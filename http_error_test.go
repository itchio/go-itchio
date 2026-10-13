package itchio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestHTTPErrorForNonJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>bad gateway</html>"))
	}))
	defer server.Close()

	client := ClientWithKey("APIKEY")
	client.HTTPClient = server.Client()
	client.BaseURL = server.URL
	client.Limiter = rate.NewLimiter(rate.Inf, 1)

	_, err := client.ListProfileGames(context.Background())
	he, ok := AsHTTPError(err)
	if assert.True(t, ok, "%v", err) {
		assert.Equal(t, 502, he.StatusCode)
		assert.Equal(t, "/profile/games", he.Path)
	}
	assert.Contains(t, err.Error(), "HTTP 502")
}

func TestHTTPErrorForDecodableNon200(t *testing.T) {
	for _, code := range []int{201, 500} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			_, _ = w.Write([]byte(`{"games":[]}`))
		}))
		client := ClientWithKey("APIKEY")
		client.HTTPClient = server.Client()
		client.BaseURL = server.URL
		client.RetryPatterns = []time.Duration{0}
		client.Limiter = rate.NewLimiter(rate.Inf, 1)

		_, err := client.ListProfileGames(context.Background())
		server.Close()
		he, ok := AsHTTPError(err)
		if assert.True(t, ok, "code %d: %v", code, err) {
			assert.Equal(t, code, he.StatusCode)
		}
	}
}
