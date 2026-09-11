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
