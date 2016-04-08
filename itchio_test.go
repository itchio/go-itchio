package itchio

import (
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
		w.WriteHeader(200)
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

	client := &Client{
		Key:        "APIKEY",
		HTTPClient: httpClient,
		BaseURL:    server.URL,
	}

	return server, client
}

func Test_MyGames(t *testing.T) {
	server, client := testTools(200, `{
		"games": [
			{"url": "https://kenney.itch.io/barb", "id": 123},
		  {"url": "https://leafo.itch.io/x-moon", "id": 456}
		]
	}`)
	defer server.Close()

	games, err := client.MyGames()
	assert.Nil(t, err)
	assert.Equal(t, len(games.Errors), 0)
	assert.Equal(t, len(games.Games), 2)
	assert.Equal(t, games.Games[0].ID, int64(123))
	assert.Equal(t, games.Games[0].Url, "https://kenney.itch.io/barb")
}
