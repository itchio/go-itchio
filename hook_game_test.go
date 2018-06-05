package itchio

import (
	"encoding/json"
	"testing"

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
}
