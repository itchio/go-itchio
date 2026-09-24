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
		intermediateTraits := make(map[string]any)
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
		intermediateSane := make(map[string]any)
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

// decodeResponse mirrors the camelify + mapstructure path used for API
// responses in http_helpers.go
func decodeResponse(t *testing.T, body string, dst any) {
	intermediate := make(map[string]any)
	assert.NoError(t, json.Unmarshal([]byte(body), &intermediate))
	intermediate = camelifyMap(intermediate)

	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "json",
		WeaklyTypedInput: true,
		Result:           dst,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			RawJSONHookFunc,
			GameHookFunc,
			UploadHookFunc,
		),
	})
	assert.NoError(t, err)
	assert.NoError(t, dec.Decode(intermediate))
}

func Test_UploadLaunchTargets(t *testing.T) {
	{
		var upload Upload
		decodeResponse(t, `{
			"id": 123,
			"traits": ["p_linux"],
			"launch_targets": [
				{"path": "bin/game", "depth": 2, "flavor": "linux", "arch": "amd64", "linux_info": {"arch": "amd64", "imports": ["libc.so.6"]}}
			],
			"launch_targets_source": "client",
			"launch_targets_scanner_version": "butler/15.26.0",
			"launch_targets_extracted_size": 5000000000
		}`, &upload)

		assert.EqualValues(t, LaunchTargetsSourceClient, upload.LaunchTargetsSource)
		assert.EqualValues(t, "butler/15.26.0", upload.LaunchTargetsScannerVersion)
		assert.EqualValues(t, 5000000000, upload.LaunchTargetsExtractedSize)
		assert.EqualValues(t, ArchitecturesAll, upload.Platforms.Linux)

		// keys inside targets keep dash's snake_case
		var targets []map[string]any
		assert.NoError(t, json.Unmarshal(upload.LaunchTargets, &targets))
		assert.Len(t, targets, 1)
		assert.EqualValues(t, "bin/game", targets[0]["path"])
		assert.EqualValues(t, "amd64", targets[0]["linux_info"].(map[string]any)["arch"])
	}

	{
		var upload Upload
		decodeResponse(t, `{"id": 123, "launch_targets": [], "launch_targets_source": "server"}`, &upload)
		assert.EqualValues(t, "[]", string(upload.LaunchTargets))
		assert.EqualValues(t, LaunchTargetsSourceServer, upload.LaunchTargetsSource)
	}

	{
		var upload Upload
		decodeResponse(t, `{"id": 123}`, &upload)
		assert.Nil(t, upload.LaunchTargets)
		assert.EqualValues(t, "", upload.LaunchTargetsSource)
	}
}
