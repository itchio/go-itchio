package itchio

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"
	"net/url"
)

func testTools(code int, body string) (*httptest.Server, *Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		fmt.Fprintln(w, body)
	}))

	// Make a transport that reroutes all traffic to the example server
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	// Make a http.Client with the transport
	httpClient := &http.Client{Transport: transport}

	client := ClientWithKey("APIKEY")
	client.HTTPClient = httpClient
	client.BaseURL = server.URL

	return server, client
}

func Test_ListProfileGames(t *testing.T) {
	server, client := testTools(200, `{
		"games": [
		  {"url": "https://kenney.itch.io/barb", "id": 123, "min_price": 5000},
		  {"url": "https://leafo.itch.io/x-moon", "id": 456, "min_price": 12000}
		]
	}`)
	defer server.Close()

	games, err := client.ListProfileGames(context.Background())
	assert.NoError(t, err)
	assert.EqualValues(t, len(games.Games), 2)
	assert.EqualValues(t, games.Games[0].ID, 123)
	assert.EqualValues(t, games.Games[0].URL, "https://kenney.itch.io/barb")
	assert.EqualValues(t, games.Games[0].MinPrice, 5000)
}

func Test_ListProfileGamesError(t *testing.T) {
	server, client := testTools(400, `{
		"errors": [
		  "invalid game" 
		]
	}`)
	defer server.Close()

	_, err := client.ListProfileGames(context.Background())
	assert.Error(t, err)
	assert.True(t, IsAPIError(err))
	assert.EqualValues(t, "itch.io API error (400): /profile/games: invalid game", err.Error())
}

func Test_ExchangeOAuthCode(t *testing.T) {
	server, client := testTools(200, `{
		"key": {"id": 123, "userId": 456, "key": "abc123"},
		"cookie": {"itchio_token": "xyz789"}
	}`)
	defer server.Close()

	resp, err := client.ExchangeOAuthCode(context.Background(), ExchangeOAuthCodeParams{
		Code:         "auth_code_123",
		CodeVerifier: "verifier_abc",
		RedirectURI:  "http://localhost:8080/callback",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp.Key)
	assert.EqualValues(t, 123, resp.Key.ID)
	assert.EqualValues(t, "abc123", resp.Key.Key)
	assert.EqualValues(t, "xyz789", resp.Cookie["itchioToken"])
}

func Test_ExchangeOAuthCodeError(t *testing.T) {
	server, client := testTools(400, `{
		"errors": ["invalid_grant"]
	}`)
	defer server.Close()

	_, err := client.ExchangeOAuthCode(context.Background(), ExchangeOAuthCodeParams{
		Code:         "invalid_code",
		CodeVerifier: "verifier",
		RedirectURI:  "http://localhost:8080/callback",
	})
	assert.Error(t, err)
	assert.True(t, IsAPIError(err))
}

func Test_ParseSpec(t *testing.T) {
	var spec *Spec
	var err error

	spec, err = ParseSpec("user/page:channel")
	assert.NoError(t, err)
	assert.Equal(t, spec.Target, "user/page")
	assert.Equal(t, spec.Channel, "channel")

	spec, err = ParseSpec("user/page")
	assert.NoError(t, err)
	assert.Equal(t, spec.Target, "user/page")
	assert.Equal(t, spec.Channel, "")

	err = spec.EnsureChannel()
	assert.Error(t, err)

	_, err = ParseSpec("a:b:c")
	assert.Error(t, err)
}
