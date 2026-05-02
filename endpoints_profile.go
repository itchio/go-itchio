package itchio

import "context"

//-------------------------------------------------------

// GetProfileResponse is what the API server responds when we ask for the user's profile
type GetProfileResponse struct {
	User *User `json:"user"`
}

// GetProfile returns information about the user the current credentials belong to
func (c *Client) GetProfile(ctx context.Context) (*GetProfileResponse, error) {
	q := NewQuery(c, "/profile")
	r := &GetProfileResponse{}
	return r, q.Get(ctx, r)
}

//-------------------------------------------------------

// ListProfileGamesResponse is what the API server answers when we ask for what games
// an account develops.
type ListProfileGamesResponse struct {
	Games []*Game `json:"games"`
}

// ListProfileGames lists the games one develops (ie. can edit)
func (c *Client) ListProfileGames(ctx context.Context) (*ListProfileGamesResponse, error) {
	q := NewQuery(c, "/profile/games")
	r := &ListProfileGamesResponse{}
	return r, q.Get(ctx, r)
}

//-------------------------------------------------------

// ListProfileOwnedKeysParams : params for ListProfileOwnedKeys
type ListProfileOwnedKeysParams struct {
	Page int64
}

// ListProfileOwnedKeysResponse : response for ListProfileOwnedKeys
type ListProfileOwnedKeysResponse struct {
	Page      int64          `json:"page"`
	PerPage   int64          `json:"perPage"`
	OwnedKeys []*DownloadKey `json:"ownedKeys"`
}

// ListProfileOwnedKeys lists the download keys the account with
// the current API key owns.
func (c *Client) ListProfileOwnedKeys(ctx context.Context, p ListProfileOwnedKeysParams) (*ListProfileOwnedKeysResponse, error) {
	q := NewQuery(c, "/profile/owned-keys")
	q.AddInt64IfNonZero("page", p.Page)
	r := &ListProfileOwnedKeysResponse{}
	return r, q.Get(ctx, r)
}

//-------------------------------------------------------

// ListProfileCollectionsResponse : response for ListProfileCollections
type ListProfileCollectionsResponse struct {
	Collections []*Collection `json:"collections"`
}

// ListProfileCollections lists the collections associated to a profile.
func (c *Client) ListProfileCollections(ctx context.Context) (*ListProfileCollectionsResponse, error) {
	q := NewQuery(c, "/profile/collections")
	r := &ListProfileCollectionsResponse{}
	return r, q.Get(ctx, r)
}

//-------------------------------------------------------

// ProfileBuildsTotals : per-state counts and project count for the current page set
type ProfileBuildsTotals struct {
	All          int64 `json:"all"`
	Live         int64 `json:"live"`
	Processing   int64 `json:"processing"`
	Failed       int64 `json:"failed"`
	ProjectCount int64 `json:"projectCount"`
}

// ListProfileBuildsParams : params for ListProfileBuilds
type ListProfileBuildsParams struct {
	Page    int64
	PerPage int64
	// Optional. One of "live", "processing", "failed". Empty for all.
	State string
	// Optional. If set, include aggregate totals in the response.
	IncludeTotals bool
}

// ListProfileBuildsResponse : response for ListProfileBuilds
//
// Builds are denormalized: each build carries scalar GameID/UploadID, and the
// referenced games/uploads are returned once each in Games/Uploads. This avoids
// large response inflation when many builds share the same game or upload.
type ListProfileBuildsResponse struct {
	Builds  []*Build             `json:"builds"`
	Games   []*Game              `json:"games"`
	Uploads []*Upload            `json:"uploads"`
	Page    int64                `json:"page"`
	PerPage int64                `json:"perPage"`
	Totals  *ProfileBuildsTotals `json:"totals,omitempty"`
}

// ListProfileBuilds lists builds across all games the current user develops.
// The response is denormalized — see ListProfileBuildsResponse.
func (c *Client) ListProfileBuilds(ctx context.Context, p ListProfileBuildsParams) (*ListProfileBuildsResponse, error) {
	q := NewQuery(c, "/profile/builds")
	q.AddInt64IfNonZero("page", p.Page)
	q.AddInt64IfNonZero("per_page", p.PerPage)
	q.AddStringIfNonEmpty("state", p.State)
	if p.IncludeTotals {
		q.AddStringIfNonEmpty("include_totals", "1")
	}
	r := &ListProfileBuildsResponse{}
	return r, q.Get(ctx, r)
}
