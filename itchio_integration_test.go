package itchio

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Integration(t *testing.T) {
	apiKey := os.Getenv("ITCH_TEST_ACCOUNT_API_KEY")
	if apiKey == "" {
		t.Skipf("Skipping integration tests (no credentials)")
	}

	if testing.Short() {
		t.Skipf("Skipping integration tests (short mode)")
	}

	c := ClientWithKey(apiKey)

	t.Logf("Retrieving profile...")
	p, err := c.GetProfile()
	assert.NoError(t, err)
	assert.NotNil(t, p.User)
	assert.EqualValues(t, "itch-test-account", p.User.Username)

	var testGameID int64 = 141753

	t.Logf("Retrieving game...")
	g, err := c.GetGame(&GetGameParams{
		GameID: testGameID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, g.Game)
	assert.EqualValues(t, "xna4-test", g.Game.Title)

	t.Logf("Listing uploads...")
	lu, err := c.ListGameUploads(&ListGameUploadsParams{
		GameID: testGameID,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(lu.Uploads))

	up := lu.Uploads[0]
	t.Logf("Listing builds for upload %d...", up.ID)
	lb, err := c.ListUploadBuilds(&ListUploadBuildsParams{
		UploadID: up.ID,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, 1, len(lb.Builds))

	b := lb.Builds[0]
	t.Logf("Downloading build %d", b.ID)
	bURL := c.MakeBuildDownloadURL(&MakeBuildDownloadParams{
		BuildID: b.ID,
		Type:    BuildFileTypeArchive,
		SubType: BuildFileSubTypeDefault,
	})

	download := func(url string) []byte {
		req, err := http.NewRequest("GET", bURL, nil)
		assert.NoError(t, err)

		res, err := c.HTTPClient.Do(req)
		assert.NoError(t, err)
		assert.EqualValues(t, 200, res.StatusCode)
		defer res.Body.Close()

		data, err := ioutil.ReadAll(res.Body)
		assert.NoError(t, err)

		return data
	}

	buildBytes := download(bURL)
	assert.EqualValues(t, up.Size, len(buildBytes))

	t.Logf("Listing build files for build %d", b.ID)
	bf, err := c.ListBuildFiles(b.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, bf.Files)

	var af *BuildFile
	for _, f := range bf.Files {
		if f.Type == BuildFileTypeArchive && f.SubType == BuildFileSubTypeDefault {
			af = f
			break
		}
	}
	assert.NotNil(t, af)

	t.Logf("Downloading archive-default build file for build %d", b.ID)
	bfURL := c.MakeBuildFileDownloadURL(&MakeBuildFileDownloadURLParams{
		BuildID: b.ID,
		FileID:  af.ID,
	})
	buildFileBytes := download(bfURL)

	assert.EqualValues(t, af.Size, len(buildFileBytes))
	assert.EqualValues(t, buildBytes, buildFileBytes)
}
