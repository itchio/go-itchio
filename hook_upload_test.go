package itchio

import (
	"encoding/json"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func Test_UploadHook(t *testing.T) {
	ref := Upload{
		ID:   123,
		Demo: true,
		Platforms: Platforms{
			OSX:     ArchitecturesAll,
			Windows: ArchitecturesAll,
		},
	}
	marshalledTraits := []byte(`{
			"id": 123,
			"traits": ["demo", "p_osx", "p_windows"]
		}`)

	marshalledSane := []byte(`{
			"id": 123,
			"demo": true,
			"platforms": {"osx": "all", "windows": "all"}
		}`)

	{
		intermediateTraits := make(map[string]interface{})
		err := json.Unmarshal(marshalledTraits, &intermediateTraits)
		assert.NoError(t, err)

		var decodedTraits Upload
		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName:          "json",
			DecodeHook:       UploadHookFunc,
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

		var decodedSane Upload
		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName:          "json",
			DecodeHook:       UploadHookFunc,
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

	var unmarshalled Upload
	err = json.Unmarshal(bs, &unmarshalled)
	assert.NoError(t, err)

	assert.EqualValues(t, ref, unmarshalled)
}
