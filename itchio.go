package itchio

import (
	"net/http"
	"time"

	"github.com/itchio/httpkit/rate"
	"github.com/itchio/httpkit/timeout"
)

type OnRateLimited func(req *http.Request, res *http.Response)

// A Client allows consuming the itch.io API
type Client struct {
	Key              string
	HTTPClient       *http.Client
	BaseURL          string
	RetryPatterns    []time.Duration
	UserAgent        string
	AcceptedLanguage string
	Limiter          rate.Limiter

	onRateLimited OnRateLimited
}

func defaultRetryPatterns() []time.Duration {
	return []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
	}
}

// ClientWithKey creates a new itch.io API client with a given API key
func ClientWithKey(key string) *Client {
	c := &Client{
		Key:              key,
		HTTPClient:       timeout.NewDefaultClient(),
		RetryPatterns:    defaultRetryPatterns(),
		UserAgent:        "go-itchio",
		AcceptedLanguage: "*",
		Limiter:          DefaultRateLimiter(),
	}
	c.SetServer("https://api.itch.io")
	return c
}

func (c *Client) OnRateLimited(cb OnRateLimited) {
	c.onRateLimited = cb
}

// SetServer allows changing the server to which we're making API
// requests (which defaults to the reference itch.io server)
func (c *Client) SetServer(itchioServer string) *Client {
	c.BaseURL = itchioServer
	return c
}
