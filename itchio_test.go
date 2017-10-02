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

func Test_ListMyGames(t *testing.T) {
	server, client := testTools(200, `{
		"games": [
			{"url": "https://kenney.itch.io/barb", "id": 123, "min_price": 5000},
		  {"url": "https://leafo.itch.io/x-moon", "id": 456, "min_price": 12000}
		]
	}`)
	defer server.Close()

	games, err := client.ListMyGames()
	assert.Nil(t, err)
	assert.EqualValues(t, len(games.Errors), 0)
	assert.EqualValues(t, len(games.Games), 2)
	assert.EqualValues(t, games.Games[0].ID, 123)
	assert.EqualValues(t, games.Games[0].URL, "https://kenney.itch.io/barb")
	assert.EqualValues(t, games.Games[0].MinPrice, 5000)
}

func Test_ParseSpec(t *testing.T) {
	var spec *Spec
	var err error

	spec, err = ParseSpec("user/page:channel")
	assert.Nil(t, err)
	assert.Equal(t, spec.Target, "user/page")
	assert.Equal(t, spec.Channel, "channel")

	spec, err = ParseSpec("user/page")
	assert.Nil(t, err)
	assert.Equal(t, spec.Target, "user/page")
	assert.Equal(t, spec.Channel, "")

	err = spec.EnsureChannel()
	assert.NotNil(t, err)

	spec, err = ParseSpec("a:b:c")
	assert.NotNil(t, err)
}
