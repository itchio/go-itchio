package itchio

import (
	"encoding/json"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func Test_UploadTraits(t *testing.T) {
	ut1 := UploadTraits{
		PlatformLinux: true,
		Demo:          true,
	}

	{
		marshalled, err := json.Marshal(ut1)
		assert.NoError(t, err)
		assert.EqualValues(t, `["p_linux","demo"]`, string(marshalled))

		var ut2 UploadTraits
		err = json.Unmarshal(marshalled, &ut2)
		assert.NoError(t, err)

		assert.EqualValues(t, ut1, ut2)
	}

	u1 := Upload{
		DisplayName: "Unreal for macOS",
		Traits:      ut1,
	}

	{
		marshalled, err := json.Marshal(u1)
		assert.NoError(t, err)

		var u2 Upload
		err = json.Unmarshal(marshalled, &u2)
		assert.NoError(t, err)

		assert.EqualValues(t, u1, u2)
	}

	{
		marshalled, err := json.Marshal(u1)
		assert.NoError(t, err)

		intermediate := make(map[string]interface{})
		err = json.Unmarshal(marshalled, &intermediate)
		assert.NoError(t, err)

		var u2 Upload
		dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName:    "json",
			DecodeHook: UploadTraitHookFunc,
			Result:     &u2,
		})
		assert.NoError(t, err)
		err = dec.Decode(intermediate)
		assert.NoError(t, err)

		assert.EqualValues(t, u1, u2)
	}
}
