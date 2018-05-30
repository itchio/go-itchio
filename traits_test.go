package itchio

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GameTraits2(t *testing.T) {
	gt1 := GameTraits2{
		PlatformAndroid: true,
		HasDemo:         true,
	}

	marshalled, err := gt1.MarshalJSON()
	assert.NoError(t, err)
	assert.EqualValues(t, `["p_android","has_demo"]`, string(marshalled))

	var gt2 GameTraits2
	err = gt2.UnmarshalJSON(marshalled)
	assert.NoError(t, err)

	assert.EqualValues(t, gt1, gt2)
	assert.True(t, gt2.PlatformAndroid)
	assert.True(t, gt2.HasDemo)
	assert.False(t, gt2.PlatformWindows)
	assert.False(t, gt2.CanBeBought)
}

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

func Benchmark_GameTraits(b *testing.B) {
	b.Run("map-based", func(b *testing.B) {
		gt1 := GameTraits{
			GameTraitPlatformAndroid: true,
			GameTraitHasDemo:         true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON()
			var gt2 GameTraits
			gt2.UnmarshalJSON(data)
			if !gt2[GameTraitPlatformAndroid] {
				panic("missing platform android")
			}
			if !gt2[GameTraitHasDemo] {
				panic("missing has-demo")
			}
		}
	})

	b.Run("struct-based", func(b *testing.B) {
		gt1 := GameTraits2{
			PlatformAndroid: true,
			HasDemo:         true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON()
			var gt2 GameTraits2
			gt2.UnmarshalJSON(data)
			if !gt2.PlatformAndroid {
				panic("missing platform android")
			}
			if !gt2.HasDemo {
				panic("missing has-demo")
			}
		}
	})
}
