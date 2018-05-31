package itchio

import (
	"encoding/json"
	"testing"

	"github.com/mitchellh/mapstructure"

	"github.com/stretchr/testify/assert"
)

func Test_GameTraits(t *testing.T) {
	gt1 := GameTraits{
		PlatformAndroid: true,
		HasDemo:         true,
	}

	{
		marshalled, err := json.Marshal(gt1)
		assert.NoError(t, err)
		assert.EqualValues(t, `["p_android","has_demo"]`, string(marshalled))

		var gt2 GameTraits
		err = json.Unmarshal(marshalled, &gt2)
		assert.NoError(t, err)

		assert.EqualValues(t, gt1, gt2)
	}

	g1 := Game{
		Title:  "Unreal",
		Traits: gt1,
	}

	{
		marshalled, err := json.Marshal(g1)
		assert.NoError(t, err)

		var g2 Game
		err = json.Unmarshal(marshalled, &g2)
		assert.NoError(t, err)

		assert.EqualValues(t, g1, g2)
	}

	{
		marshalled, err := json.Marshal(g1)
		assert.NoError(t, err)

		intermediate := make(map[string]interface{})
		err = json.Unmarshal(marshalled, &intermediate)
		assert.NoError(t, err)

		var g2 Game
		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName:    "json",
			DecodeHook: GameTraitHookFunc,
			Result:     &g2,
		})
		assert.NoError(t, err)
		err = dec.Decode(intermediate)
		assert.NoError(t, err)

		assert.EqualValues(t, g1, g2)
	}

	{
		g1 := Game{
			ID: 123,
		}
		marshalled := []byte(`{
			"id": 123,
			"traits": {}
		}`)

		intermediate := make(map[string]interface{})
		err := json.Unmarshal(marshalled, &intermediate)
		assert.NoError(t, err)

		var g2 Game
		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName:          "json",
			DecodeHook:       GameTraitHookFunc,
			WeaklyTypedInput: true,
			Result:           &g2,
		})
		assert.NoError(t, err)
		err = dec.Decode(intermediate)
		assert.NoError(t, err)

		assert.EqualValues(t, g1, g2)
	}
}
