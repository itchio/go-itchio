package itchfs

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/itchio/httpkit/httpfile"
	"github.com/itchio/wharf/eos"
)

func Test_Register(t *testing.T) {
	ifs := &ItchFS{}
	assert.NoError(t, eos.RegisterHandler(ifs))
	defer eos.DeregisterHandler(ifs)
	assert.Error(t, eos.RegisterHandler(ifs))
}

func Test_Renewal(t *testing.T) {
	res := &http.Response{
		StatusCode: 400,
	}
	assert.True(t, needsRenewal(res, nil))

	res.StatusCode = 200
	assert.False(t, needsRenewal(res, nil))
}

func Test_Resource(t *testing.T) {
	var responses = map[string]string{
		"/api/1/x/upload/13/download": `{
			"url": "http://localhost/upload-url"
		}`,
		"/api/1/x/upload/13/download?download_key_id=key": `{
			"url": "http://localhost/upload-url-with-key"
		}`,
		"/api/1/x/upload/13/download/builds/57": `{
			"archive": {
				"url": "http://localhost/upload-builds-url-archive"
			},
			"patch": {
				"url": "http://localhost/upload-builds-url-patch"
			}
		}`,
		"/api/1/x/upload/13/download/builds/57?download_key_id=key": `{
			"signature": {
				"url": "http://localhost/upload-builds-url-with-key-signature"
			},
			"manifest": {
				"url": "http://localhost/upload-builds-url-with-key-manifest"
			}
		}`,
		"/api/1/x/wharf/builds/57/files/90/download": `{
			"url": "http://localhost/wharf-build-url"
		}`,
	}

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			path = fmt.Sprintf("%s?%s", path, req.URL.RawQuery)
		}
		log.Printf("path = %s, URL = %#v\n", path, req.URL)

		res, ok := responses[path]
		if !ok {
			http.Error(w, "Not Found", 404)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(res))
	}))
	defer fakeServer.CloseClientConnections()

	ifs := &ItchFS{
		ItchServer: fakeServer.URL,
	}

	parseURL := func(rawurl string) *url.URL {
		u, err := url.Parse(rawurl)
		assert.NoError(t, err)
		return u
	}

	testFail := func(itchfsURL string) {
		_, _, err := ifs.MakeResource(parseURL(itchfsURL))
		assert.Error(t, err)
	}

	testFail("itchfs://lacking-slash?api_key=x")
	testFail("itchfs:///nonsense?api_key=x")
	testFail("itchfs:///upload/13/download?no_api_key=true")
	testFail("itchfs:///upload/NaN/download?api_key=x")
	testFail("itchfs:///upload/NaN/download/builds/13/archive?api_key=x")
	testFail("itchfs:///download-key/key/download/NaN?api_key=x")
	testFail("itchfs:///upload/13/download/builds/NaN/archive?api_key=x")
	testFail("itchfs:///download-key/key/download/NaN/builds/57/archive?api_key=x")
	testFail("itchfs:///download-key/key/download/13/builds/NaN/archive?api_key=x")
	testFail("itchfs:///wharf/builds/NaN/files/90/download?api_key=x")
	testFail("itchfs:///wharf/builds/57/files/NaN/download?api_key=x")

	testFailGetter := func(itchfsURL string) {
		getURL, _, err := ifs.MakeResource(parseURL(itchfsURL))
		assert.NoError(t, err)

		_, err = getURL()
		assert.Error(t, err)
	}

	testFailGetter("itchfs:///upload/13/download/builds/57/butt?api_key=x")

	testPair := func(itchfsURL string, downloadURL string) {
		var getURL httpfile.GetURLFunc
		var url string
		var err error

		getURL, _, err = ifs.MakeResource(parseURL(itchfsURL))
		assert.NoError(t, err)

		url, err = getURL()
		assert.NoError(t, err)
		assert.Equal(t, downloadURL, url)
	}

	testPair("itchfs:///upload/13/download?api_key=x", "http://localhost/upload-url")
	testPair("itchfs:///download-key/key/download/13?api_key=x", "http://localhost/upload-url-with-key")
	testPair("itchfs:///upload/13/download?api_key=x&download_key_id=key", "http://localhost/upload-url-with-key")

	testPair("itchfs:///upload/13/download/builds/57/archive?api_key=x", "http://localhost/upload-builds-url-archive")
	testPair("itchfs:///upload/13/download/builds/57/patch?api_key=x", "http://localhost/upload-builds-url-patch")

	testPair("itchfs:///download-key/key/download/13/builds/57/signature?api_key=x", "http://localhost/upload-builds-url-with-key-signature")
	testPair("itchfs:///download-key/key/download/13/builds/57/manifest?api_key=x", "http://localhost/upload-builds-url-with-key-manifest")

	testPair("itchfs:///upload/13/download/builds/57/signature?download_key_id=key&api_key=x", "http://localhost/upload-builds-url-with-key-signature")
	testPair("itchfs:///upload/13/download/builds/57/manifest?api_key=x&download_key_id=key", "http://localhost/upload-builds-url-with-key-manifest")

	testPair("itchfs:///wharf/builds/57/files/90/download?api_key=x", "http://localhost/wharf-build-url")
}
