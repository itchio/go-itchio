package itchio

import "context"

//-------------------------------------------------------

// GetBundleGamesParams : params for GetBundleGames
type GetBundleGamesParams struct {
	BundleID int64 `json:"bundleId"`
	Page     int64 `json:"page"`
}

// GetBundleGamesResponse : response for GetBundleGames
type GetBundleGamesResponse struct {
	Page        int64         `json:"page"`
	PerPage     int64         `json:"perPage"`
	BundleGames []*BundleGame `json:"bundleGames"`
}

// GetBundleGames retrieves a page of a bundle's games. The current user
// must own the bundle.
func (c *Client) GetBundleGames(ctx context.Context, params GetBundleGamesParams) (*GetBundleGamesResponse, error) {
	q := NewQuery(c, "/bundles/%d/bundle-games", params.BundleID)
	q.AddInt64IfNonZero("page", params.Page)
	r := &GetBundleGamesResponse{}
	return r, q.Get(ctx, r)
}

//-------------------------------------------------------

// ClaimBundleGameParams : params for ClaimBundleGame
type ClaimBundleGameParams struct {
	BundleID int64 `json:"bundleId"`
	GameID   int64 `json:"gameId"`
}

// ClaimBundleGameResponse : response for ClaimBundleGame
type ClaimBundleGameResponse struct {
	DownloadKey *DownloadKey `json:"downloadKey"`
}

// ClaimBundleGame returns a DownloadKey for a game owned via a bundle
// purchase, materializing one if the bundle defers download key creation.
// It's atomic and idempotent: safe to call multiple times, and always
// returns the existing key if one is already present.
func (c *Client) ClaimBundleGame(ctx context.Context, params ClaimBundleGameParams) (*ClaimBundleGameResponse, error) {
	q := NewQuery(c, "/bundles/%d/claim-game", params.BundleID)
	q.AddInt64("game_id", params.GameID)
	r := &ClaimBundleGameResponse{}
	return r, q.Post(ctx, r)
}
