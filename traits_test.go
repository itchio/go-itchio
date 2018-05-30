package itchio

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GameTraits(t *testing.T) {
	gt1 := GameTraits{
		GameTraitPlatformAndroid: true,
		GameTraitHasDemo:         true,
	}

	marshalled, err := gt1.MarshalJSON()
	assert.NoError(t, err)
	assert.EqualValues(t, `["p_android","has_demo"]`, string(marshalled))

	gt2 := make(GameTraits)
	err = gt2.UnmarshalJSON(marshalled)
	assert.NoError(t, err)

	assert.EqualValues(t, gt1, gt2)
	assert.True(t, gt2[GameTraitPlatformAndroid])
	assert.True(t, gt2[GameTraitHasDemo])
	assert.False(t, gt2[GameTraitPlatformWindows])
	assert.False(t, gt2[GameTraitCanBeBought])

	gameJson := `{
		"title": "Unreal",
		"traits": ["p_linux","can_be_bought"]
	}`
	game := &Game{}
	err = json.Unmarshal([]byte(gameJson), &game)
	assert.NoError(t, err)

	assert.False(t, game.Traits[GameTraitPlatformWindows])
	assert.True(t, game.Traits[GameTraitPlatformLinux])
	assert.True(t, game.Traits[GameTraitCanBeBought])
	assert.Equal(t, "Unreal", game.Title)

	fakeRes := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(gameJson)),
	}

	gameRes := &Game{}
	err = ParseAPIResponse(gameRes, fakeRes)
	assert.NoError(t, err)
}
