package itchio

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBuildMetadata(t *testing.T) {
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"build":{"id":1,"uploadId":2}}`))
	}))
	defer server.Close()

	client := ClientWithKey("key")
	client.SetServer(server.URL)

	_, err := client.CreateBuild(context.Background(), CreateBuildParams{
		Target:  "user/game",
		Channel: "linux",
	})
	assert.NoError(t, err)
	values, _ := url.ParseQuery(body)
	assert.False(t, values.Has("metadata"), "metadata omitted when unset")

	_, err = client.CreateBuild(context.Background(), CreateBuildParams{
		Target:  "user/game",
		Channel: "linux",
		Metadata: BuildMetadata{
			"steam": map[string]interface{}{"app_id": 3445480, "branch": "alphatest"},
		},
	})
	assert.NoError(t, err)
	values, _ = url.ParseQuery(body)
	var got map[string]map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(values.Get("metadata")), &got))
	assert.EqualValues(t, 3445480, got["steam"]["app_id"])
	assert.Equal(t, "alphatest", got["steam"]["branch"])
}

func TestCreateBuildLaunchAnalysis(t *testing.T) {
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/wharf/builds", r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"build":{"id":1,"upload_id":2}}`))
	}))
	defer server.Close()
	client := ClientWithKey("key")
	client.SetServer(server.URL)
	for _, tc := range []struct {
		name    string
		targets json.RawMessage
	}{
		{name: "omitted"},
		{name: "empty", targets: json.RawMessage(`[]`)},
		{name: "populated", targets: json.RawMessage(`[{"path":"bin/game","depth":2,"flavor":"linux","linux_info":{"glibcVersion":"2.17"},"engine":{"engine":"godot","details":{"future_field":true}}}]`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := CreateBuildParams{
				Target: "user/game", Channel: "linux", Source: "app",
				Metadata: BuildMetadata{"steam": map[string]interface{}{"app_id": 123}},
			}
			if tc.targets != nil {
				params.LaunchAnalysis = &BuildLaunchAnalysis{SchemaVersion: 1, ScannerVersion: "butler/test", LaunchTargets: tc.targets}
			}
			result, err := client.CreateBuild(context.Background(), params)
			if !assert.NoError(t, err) {
				return
			}
			assert.EqualValues(t, 1, result.Build.ID)
			values, err := url.ParseQuery(body)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, "app", values.Get("source"))
			assert.JSONEq(t, `{"steam":{"app_id":123}}`, values.Get("metadata"))
			if tc.targets == nil {
				assert.False(t, values.Has("launch_analysis"))
				return
			}
			var got struct {
				SchemaVersion  int             `json:"schema_version"`
				ScannerVersion string          `json:"scanner_version"`
				LaunchTargets  json.RawMessage `json:"launch_targets"`
			}
			if !assert.NoError(t, json.Unmarshal([]byte(values.Get("launch_analysis")), &got)) {
				return
			}
			assert.Equal(t, 1, got.SchemaVersion)
			assert.Equal(t, "butler/test", got.ScannerVersion)
			assert.JSONEq(t, string(tc.targets), string(got.LaunchTargets))
		})
	}
}

func TestCreateBuildRejectsInvalidLaunchAnalysis(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("invalid launch analysis must not reach the server")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"build":{"id":1}}`))
	}))
	defer server.Close()
	client := ClientWithKey("key")
	client.SetServer(server.URL)
	for _, tc := range []struct {
		name   string
		report BuildLaunchAnalysis
	}{
		{name: "missing schema", report: BuildLaunchAnalysis{ScannerVersion: "butler/test", LaunchTargets: json.RawMessage(`[]`)}},
		{name: "negative schema", report: BuildLaunchAnalysis{SchemaVersion: -1, ScannerVersion: "butler/test", LaunchTargets: json.RawMessage(`[]`)}},
		{name: "missing scanner", report: BuildLaunchAnalysis{SchemaVersion: 1, LaunchTargets: json.RawMessage(`[]`)}},
		{name: "blank scanner", report: BuildLaunchAnalysis{SchemaVersion: 1, ScannerVersion: " \n", LaunchTargets: json.RawMessage(`[]`)}},
		{name: "missing targets", report: BuildLaunchAnalysis{SchemaVersion: 1, ScannerVersion: "butler/test"}},
		{name: "null targets", report: BuildLaunchAnalysis{SchemaVersion: 1, ScannerVersion: "butler/test", LaunchTargets: json.RawMessage(`null`)}},
		{name: "object targets", report: BuildLaunchAnalysis{SchemaVersion: 1, ScannerVersion: "butler/test", LaunchTargets: json.RawMessage(`{}`)}},
		{name: "encoded string", report: BuildLaunchAnalysis{SchemaVersion: 1, ScannerVersion: "butler/test", LaunchTargets: json.RawMessage(`"[]"`)}},
		{name: "malformed array", report: BuildLaunchAnalysis{SchemaVersion: 1, ScannerVersion: "butler/test", LaunchTargets: json.RawMessage(`[`)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.CreateBuild(context.Background(), CreateBuildParams{Target: "user/game", Channel: "linux", LaunchAnalysis: &tc.report})
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}
