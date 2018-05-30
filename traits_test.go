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
	b.Run("map simplest", func(b *testing.B) {
		gt1 := GameTraits{
			GameTraitPlatformLinux:   true,
			GameTraitPlatformWindows: true,
			GameTraitPlatformOSX:     true,
			GameTraitHasDemo:         true,
			GameTraitCanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON()
			var gt2 GameTraits
			gt2.UnmarshalJSON(data)
			if !(gt2[GameTraitPlatformWindows] && gt2[GameTraitPlatformOSX] && gt2[GameTraitPlatformLinux] && gt2[GameTraitHasDemo] && gt2[GameTraitCanBeBought]) {
				panic("missing fields")
			}
		}
	})

	b.Run("struct simplest", func(b *testing.B) {
		gt1 := GameTraits2{
			PlatformLinux:   true,
			PlatformWindows: true,
			PlatformOSX:     true,
			HasDemo:         true,
			CanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON()
			var gt2 GameTraits2
			gt2.UnmarshalJSON(data)
			if !(gt2.PlatformWindows && gt2.PlatformOSX && gt2.PlatformLinux && gt2.HasDemo && gt2.CanBeBought) {
				panic("missing fields")
			}
		}
	})

	b.Run("struct cachereflect", func(b *testing.B) {
		gt1 := GameTraits2{
			PlatformLinux:   true,
			PlatformWindows: true,
			PlatformOSX:     true,
			HasDemo:         true,
			CanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON2()
			var gt2 GameTraits2
			gt2.UnmarshalJSON2(data)
			if !(gt2.PlatformWindows && gt2.PlatformOSX && gt2.PlatformLinux && gt2.HasDemo && gt2.CanBeBought) {
				panic("missing fields")
			}
		}
	})

	b.Run("struct handrolled", func(b *testing.B) {
		gt1 := GameTraits2{
			PlatformLinux:   true,
			PlatformWindows: true,
			PlatformOSX:     true,
			HasDemo:         true,
			CanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON3()
			var gt2 GameTraits2
			gt2.UnmarshalJSON3(data)
			if !(gt2.PlatformWindows && gt2.PlatformOSX && gt2.PlatformLinux && gt2.HasDemo && gt2.CanBeBought) {
				panic("missing fields")
			}
		}
	})

	b.Run("unreasonably custom", func(b *testing.B) {
		gt1 := GameTraits2{
			PlatformLinux:   true,
			PlatformWindows: true,
			PlatformOSX:     true,
			HasDemo:         true,
			CanBeBought:     true,
		}

		for n := 0; n < b.N; n++ {
			data, _ := gt1.MarshalJSON4()
			var gt2 GameTraits2
			gt2.UnmarshalJSON4(data)
			if !(gt2.PlatformWindows && gt2.PlatformOSX && gt2.PlatformLinux && gt2.HasDemo && gt2.CanBeBought) {
				panic("missing fields")
			}
		}
	})
}
