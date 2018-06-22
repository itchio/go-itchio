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

	{
		t.Logf("Failing login")
		c := ClientWithKey("<nope>")
		res, err := c.LoginWithPassword(LoginWithPasswordParams{
			Username: "itch-test-account",
			Password: "nope",
		})
		if err == nil && res.RecaptchaNeeded {
			t.Logf("Failing recaptcha...")
			_, err = c.LoginWithPassword(LoginWithPasswordParams{
				Username:          "itch-test-account",
				Password:          "nope",
				RecaptchaResponse: "oooh nope.",
			})
		}

		assert.Error(t, err)
	}

	c := ClientWithKey(apiKey)

	t.Logf("Retrieving profile...")
	p, err := c.GetProfile()
	assert.NoError(t, err)
	assert.NotNil(t, p.User)
	assert.EqualValues(t, "itch-test-account", p.User.Username)

	t.Logf("Retrieving owned keys...")
	ownedKeysRes, err := c.ListProfileOwnedKeys(ListProfileOwnedKeysParams{})
	assert.NoError(t, err)
	assert.NotEmpty(t, ownedKeysRes.OwnedKeys)

	ownedKeysRes, err = c.ListProfileOwnedKeys(ListProfileOwnedKeysParams{
		// if this tests ever breaks, we're doing well
		Page: 200,
	})
	assert.NoError(t, err)
	assert.Empty(t, ownedKeysRes.OwnedKeys)

	var testCollID int64 = 105002
	var testGameID int64 = 141753

	t.Logf("Retrieving collection...")
	collRes, err := c.GetCollection(GetCollectionParams{
		CollectionID: testCollID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, collRes.Collection)
	assert.NotEmpty(t, collRes.Collection.Title)

	t.Logf("Retrieving collection games...")
	collGamesRes, err := c.GetCollectionGames(GetCollectionGamesParams{
		CollectionID: testCollID,
		Page:         0,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collGamesRes.CollectionGames)

	t.Logf("Retrieving game...")
	g, err := c.GetGame(GetGameParams{
		GameID: testGameID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, g.Game)
	assert.EqualValues(t, "xna4-test", g.Game.Title)
	assert.EqualValues(t, ArchitecturesAll, g.Game.Platforms.Windows)
	assert.EqualValues(t, "", g.Game.Platforms.Linux)

	t.Logf("Listing uploads...")
	lu, err := c.ListGameUploads(ListGameUploadsParams{
		GameID: testGameID,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, len(lu.Uploads))

	up := lu.Uploads[0]
	assert.EqualValues(t, ArchitecturesAll, up.Platforms.Windows)
	t.Logf("Listing builds for upload %d...", up.ID)
	lb, err := c.ListUploadBuilds(ListUploadBuildsParams{
		UploadID: up.ID,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, 1, len(lb.Builds))

	b := lb.Builds[0]
	t.Logf("Downloading build %d", b.ID)
	bURL := c.MakeBuildDownloadURL(MakeBuildDownloadParams{
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
	bfURL := c.MakeBuildFileDownloadURL(MakeBuildFileDownloadURLParams{
		BuildID: b.ID,
		FileID:  af.ID,
	})
	buildFileBytes := download(bfURL)

	assert.EqualValues(t, af.Size, len(buildFileBytes))
	assert.EqualValues(t, buildBytes, buildFileBytes)

	t.Logf("Looking for upgrade paths")
	var oldBuild int64 = 64011
	var newBuild int64 = 64020
	pathRes, err := c.GetBuildUpgradePath(GetBuildUpgradePathParams{
		CurrentBuildID: oldBuild,
		TargetBuildID:  newBuild,
	})
	assert.NoError(t, err)
	assert.NotNil(t, pathRes.UpgradePath)
	assert.NotEmpty(t, pathRes.UpgradePath.Builds)

	var foundOld, foundNew bool
	for _, b := range pathRes.UpgradePath.Builds {
		if b.ID == oldBuild {
			foundOld = true
		}
		if b.ID == newBuild {
			foundNew = true
		}
	}
	assert.True(t, foundOld)
	assert.True(t, foundNew)
}
