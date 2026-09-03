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

// captureTools is like testTools but also records the body of the last
// request so tests can check how params were encoded
func captureTools(code int, body string, lastBody *url.Values) (*httptest.Server, *Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err == nil {
			*lastBody = r.PostForm
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		fmt.Fprintln(w, body)
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	client := ClientWithKey("APIKEY")
	client.HTTPClient = &http.Client{Transport: transport}
	client.BaseURL = server.URL

	return server, client
}

func Test_CreateCollection(t *testing.T) {
	var form url.Values
	server, client := captureTools(200, `{
		"collection": {
			"id": 12,
			"title": "Favorites",
			"url": "https://itch.io/c/12/favorites",
			"private": true,
			"layout": "list",
			"games_count": 1
		}
	}`, &form)
	defer server.Close()

	res, err := client.CreateCollection(context.Background(), CreateCollectionParams{
		Title:   "Favorites",
		Private: true,
		GameID:  99,
		Blurb:   "<p>good</p>",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 12, res.Collection.ID)
	assert.EqualValues(t, CollectionLayoutList, res.Collection.Layout)
	assert.True(t, res.Collection.Private)

	assert.EqualValues(t, "Favorites", form.Get("title"))
	assert.EqualValues(t, "true", form.Get("private"))
	assert.EqualValues(t, "99", form.Get("game_id"))
	assert.EqualValues(t, "<p>good</p>", form.Get("blurb"))
	assert.False(t, form.Has("description"))
	assert.False(t, form.Has("layout"))
}

func Test_UpdateCollectionPartial(t *testing.T) {
	var form url.Values
	server, client := captureTools(200, `{"collection": {"id": 12, "title": "Old"}}`, &form)
	defer server.Close()

	private := false
	description := ""
	_, err := client.UpdateCollection(context.Background(), UpdateCollectionParams{
		CollectionID: 12,
		Private:      &private,
		Description:  &description,
	})
	assert.NoError(t, err)

	assert.EqualValues(t, "false", form.Get("private"))
	assert.True(t, form.Has("description"))
	assert.EqualValues(t, "", form.Get("description"))
	assert.False(t, form.Has("title"))
	assert.False(t, form.Has("layout"))
	assert.False(t, form.Has("on_profile"))
}

func Test_OrderCollectionGames(t *testing.T) {
	var form url.Values
	server, client := captureTools(200, `{"success": true}`, &form)
	defer server.Close()

	res, err := client.OrderCollectionGames(context.Background(), OrderCollectionGamesParams{
		CollectionID:  12,
		GameIDs:       []int64{3, 1, 2},
		RemoveGameIDs: []int64{7},
	})
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.EqualValues(t, "[3,1,2]", form.Get("game_ids"))
	assert.EqualValues(t, "[7]", form.Get("remove_game_ids"))

	_, err = client.OrderCollectionGames(context.Background(), OrderCollectionGamesParams{
		CollectionID: 12,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, "[]", form.Get("game_ids"))
	assert.False(t, form.Has("remove_game_ids"))
}

func Test_ListProfileCollectionsHasGame(t *testing.T) {
	server, client := testTools(200, `{
		"collections": [
			{"id": 1, "title": "A", "has_game": true},
			{"id": 2, "title": "B", "has_game": false}
		]
	}`)
	defer server.Close()

	res, err := client.ListProfileCollections(context.Background(), ListProfileCollectionsParams{GameID: 5})
	assert.NoError(t, err)
	assert.Len(t, res.Collections, 2)
	if assert.NotNil(t, res.Collections[0].HasGame) {
		assert.True(t, *res.Collections[0].HasGame)
	}
	if assert.NotNil(t, res.Collections[1].HasGame) {
		assert.False(t, *res.Collections[1].HasGame)
	}
}

func Test_ListProfileCollectionsWithoutGameOmitsHasGame(t *testing.T) {
	server, client := testTools(200, `{"collections": [{"id": 1, "title": "A"}]}`)
	defer server.Close()

	res, err := client.ListProfileCollections(context.Background(), ListProfileCollectionsParams{})
	assert.NoError(t, err)
	assert.Len(t, res.Collections, 1)
	assert.Nil(t, res.Collections[0].HasGame)
}
