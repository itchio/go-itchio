package itchio

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/itchio/wharf/wtest"
	"github.com/mitchellh/mapstructure"

	"github.com/stretchr/testify/assert"
)

func Test_GameHook(t *testing.T) {
	ref := Game{
		ID:            123,
		InPressSystem: true,
		Platforms: Platforms{
			Linux:   ArchitecturesAll,
			Windows: ArchitecturesAll,
		},
	}
	marshalledTraits := []byte(`{
			"id": 123,
			"traits": ["in_press_system", "p_linux", "p_windows"]
		}`)

	marshalledSane := []byte(`{
			"id": 123,
			"inPressSystem": true,
			"platforms": {"linux": "all", "windows": "all"}
		}`)

	{
		intermediateTraits := make(map[string]interface{})
		err := json.Unmarshal(marshalledTraits, &intermediateTraits)
		assert.NoError(t, err)

		var decodedTraits Game
		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName:          "json",
			DecodeHook:       GameHookFunc,
			WeaklyTypedInput: true,
			Result:           &decodedTraits,
		})
		assert.NoError(t, err)
		err = dec.Decode(intermediateTraits)
		assert.NoError(t, err)
		assert.EqualValues(t, ref, decodedTraits)
	}

	{
		intermediateSane := make(map[string]interface{})
		err := json.Unmarshal(marshalledSane, &intermediateSane)
		assert.NoError(t, err)

		var decodedSane Game
		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName:          "json",
			DecodeHook:       GameHookFunc,
			WeaklyTypedInput: true,
			Result:           &decodedSane,
		})
		assert.NoError(t, err)
		err = dec.Decode(intermediateSane)
		assert.NoError(t, err)
		assert.EqualValues(t, ref, decodedSane)
	}

	// -------------

	bs, err := json.Marshal(ref)
	assert.NoError(t, err)

	var unmarshalled Game
	err = json.Unmarshal(bs, &unmarshalled)
	assert.NoError(t, err)

	assert.EqualValues(t, ref, unmarshalled)

	// ------------
}

func Test_GameHookNested(t *testing.T) {
	type Res struct {
		Games []*Game `json:"games"`
	}

	marshalledSane := []byte(`{"games": [{
			"id": 123,
			"inPressSystem": true,
			"platforms": {"linux": "all", "windows": "all"}
		}]}`)

	intermediate := make(map[string]interface{})
	err := json.Unmarshal(marshalledSane, &intermediate)
	wtest.Must(t, err)

	var res Res
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: &res,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
			GameHookFunc,
		),
		WeaklyTypedInput: true,
	})
	wtest.Must(t, err)

	err = dec.Decode(intermediate)
	wtest.Must(t, err)

	assert.EqualValues(t, Res{
		Games: []*Game{
			&Game{
				ID:            123,
				InPressSystem: true,
				Platforms: Platforms{
					Linux:   "all",
					Windows: "all",
				},
			},
		},
	}, res)
}
