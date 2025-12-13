package itchio

import (
	"net/http"
	"time"

	"github.com/itchio/httpkit/timeout"
	"golang.org/x/time/rate"
)

// OnRateLimited is the callback type for rate limiting events
type OnRateLimited func(req *http.Request, res *http.Response)

// OnRateLimited is the callback type for rate limiting events
type OnOutgoingRequest func(req *http.Request)

// A Client allows consuming the itch.io API
type Client struct {
	Key              string
	HTTPClient       *http.Client
	BaseURL          string
	RetryPatterns    []time.Duration
	UserAgent        string
	AcceptedLanguage string
	Limiter          *rate.Limiter

	onRateLimited     OnRateLimited
	onOutgoingRequest OnOutgoingRequest

	// OAuth state (nil for API key auth)
	oauth *oauthState
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

// OnRateLimited allows registering a function that gets called
// every time the server responds with 503
func (c *Client) OnRateLimited(cb OnRateLimited) {
	c.onRateLimited = cb
}

// OnOutgoingRequest allows registering a function that gets called
// every time the client makes an HTTP request
func (c *Client) OnOutgoingRequest(cb OnOutgoingRequest) {
	c.onOutgoingRequest = cb
}

// SetServer allows changing the server to which we're making API
// requests (which defaults to the reference itch.io server)
func (c *Client) SetServer(itchioServer string) *Client {
	c.BaseURL = itchioServer
	return c
}
