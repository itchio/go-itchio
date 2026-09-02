package itchio

import "context"

//-------------------------------------------------------

// GetCollectionParams : params for GetCollection
type GetCollectionParams struct {
	CollectionID int64 `json:"collectionId"`
}

// GetCollectionResponse : response for GetCollection
type GetCollectionResponse struct {
	Collection *Collection `json:"collection"`
}

// GetCollection retrieves a single collection by ID.
func (c *Client) GetCollection(ctx context.Context, params GetCollectionParams) (*GetCollectionResponse, error) {
	q := NewQuery(c, "/collections/%d", params.CollectionID)
	r := &GetCollectionResponse{}
	return r, q.Get(ctx, r)
}

//-------------------------------------------------------

// GetCollectionGamesParams : params for GetCollectionGames
type GetCollectionGamesParams struct {
	CollectionID int64
	Page         int64
}

// GetCollectionGamesResponse : response for GetCollectionGames
type GetCollectionGamesResponse struct {
	Page            int64             `json:"page"`
	PerPage         int64             `json:"perPage"`
	CollectionGames []*CollectionGame `json:"collectionGames"`
}

// GetCollectionGames retrieves a page of a collection's games.
func (c *Client) GetCollectionGames(ctx context.Context, params GetCollectionGamesParams) (*GetCollectionGamesResponse, error) {
	q := NewQuery(c, "/collections/%d/collection-games", params.CollectionID)
	q.AddInt64IfNonZero("page", params.Page)
	r := &GetCollectionGamesResponse{}
	return r, q.Get(ctx, r)
}

//-------------------------------------------------------

// CreateCollectionParams : params for CreateCollection
type CreateCollectionParams struct {
	// Optional, defaults to "<username>'s Collection"
	Title   string
	Private bool
	// Optional HTML description
	Description string
	// Optional, defaults to list when a blurb is given, grid otherwise
	Layout CollectionLayout
	// Optional game to add right away
	GameID int64
	// Optional HTML blurb for that game
	Blurb string
}

// CreateCollectionResponse : response for CreateCollection
type CreateCollectionResponse struct {
	Collection *Collection `json:"collection"`
}

// CreateCollection creates a collection owned by the current user.
// Requires the collection:edit scope.
func (c *Client) CreateCollection(ctx context.Context, p CreateCollectionParams) (*CreateCollectionResponse, error) {
	q := NewQuery(c, "/collections")
	q.AddStringIfNonEmpty("title", p.Title)
	q.AddBool("private", p.Private)
	q.AddStringIfNonEmpty("description", p.Description)
	q.AddStringIfNonEmpty("layout", string(p.Layout))
	q.AddInt64IfNonZero("game_id", p.GameID)
	q.AddStringIfNonEmpty("blurb", p.Blurb)
	r := &CreateCollectionResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// UpdateCollectionParams : params for UpdateCollection. Nil fields
// are left unchanged.
type UpdateCollectionParams struct {
	CollectionID int64

	Title *string
	// HTML, an empty string clears it
	Description *string
	Private     *bool
	Layout      *CollectionLayout
	// Whether the collection is shown on the current user's profile
	OnProfile *bool
}

// UpdateCollectionResponse : response for UpdateCollection
type UpdateCollectionResponse struct {
	Collection *Collection `json:"collection"`
}

// UpdateCollection changes the fields of a collection that are set in params.
// Requires the collection:edit scope.
func (c *Client) UpdateCollection(ctx context.Context, p UpdateCollectionParams) (*UpdateCollectionResponse, error) {
	q := NewQuery(c, "/collections/%d", p.CollectionID)
	q.AddStringPtr("title", p.Title)
	q.AddStringPtr("description", p.Description)
	q.AddBoolPtr("private", p.Private)
	if p.Layout != nil {
		q.AddString("layout", string(*p.Layout))
	}
	q.AddBoolPtr("on_profile", p.OnProfile)
	r := &UpdateCollectionResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// DeleteCollectionParams : params for DeleteCollection
type DeleteCollectionParams struct {
	CollectionID int64
}

// DeleteCollectionResponse : response for DeleteCollection
type DeleteCollectionResponse struct {
	Success bool `json:"success"`
}

// DeleteCollection deletes a collection and everything in it.
// Requires the collection:edit scope.
func (c *Client) DeleteCollection(ctx context.Context, p DeleteCollectionParams) (*DeleteCollectionResponse, error) {
	q := NewQuery(c, "/collections/%d/delete", p.CollectionID)
	r := &DeleteCollectionResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// AddCollectionGameParams : params for AddCollectionGame
type AddCollectionGameParams struct {
	CollectionID int64
	GameID       int64
	// Optional HTML blurb for the game
	Blurb string
}

// AddCollectionGameResponse : response for AddCollectionGame
type AddCollectionGameResponse struct {
	CollectionGame *CollectionGame `json:"collectionGame"`
}

// AddCollectionGame adds a game to the end of a collection. Adding a game
// that is already in the collection returns the existing entry.
// Requires the collection:edit scope.
func (c *Client) AddCollectionGame(ctx context.Context, p AddCollectionGameParams) (*AddCollectionGameResponse, error) {
	q := NewQuery(c, "/collections/%d/add-game", p.CollectionID)
	q.AddInt64("game_id", p.GameID)
	q.AddStringIfNonEmpty("blurb", p.Blurb)
	r := &AddCollectionGameResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// RemoveCollectionGameParams : params for RemoveCollectionGame
type RemoveCollectionGameParams struct {
	CollectionID int64
	GameID       int64
}

// RemoveCollectionGameResponse : response for RemoveCollectionGame
type RemoveCollectionGameResponse struct {
	Success bool `json:"success"`
	// False if the game was not in the collection
	Removed bool `json:"removed"`
}

// RemoveCollectionGame removes a game from a collection.
// Requires the collection:edit scope.
func (c *Client) RemoveCollectionGame(ctx context.Context, p RemoveCollectionGameParams) (*RemoveCollectionGameResponse, error) {
	q := NewQuery(c, "/collections/%d/remove-game", p.CollectionID)
	q.AddInt64("game_id", p.GameID)
	r := &RemoveCollectionGameResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// UpdateCollectionGameParams : params for UpdateCollectionGame
type UpdateCollectionGameParams struct {
	CollectionID int64
	GameID       int64
	// HTML blurb, an empty string clears it. Nil leaves it unchanged.
	Blurb *string
}

// UpdateCollectionGameResponse : response for UpdateCollectionGame
type UpdateCollectionGameResponse struct {
	CollectionGame *CollectionGame `json:"collectionGame"`
}

// UpdateCollectionGame edits a game's entry in a collection.
// Requires the collection:edit scope.
func (c *Client) UpdateCollectionGame(ctx context.Context, p UpdateCollectionGameParams) (*UpdateCollectionGameResponse, error) {
	q := NewQuery(c, "/collections/%d/collection-games/%d", p.CollectionID, p.GameID)
	q.AddStringPtr("blurb", p.Blurb)
	r := &UpdateCollectionGameResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// OrderCollectionGamesParams : params for OrderCollectionGames
type OrderCollectionGamesParams struct {
	CollectionID int64
	// Game ids in the desired order, first is shown first. Up to 500.
	GameIDs []int64
	// Optional games to remove before ordering
	RemoveGameIDs []int64
}

// OrderCollectionGamesResponse : response for OrderCollectionGames
type OrderCollectionGamesResponse struct {
	Success bool `json:"success"`
}

// OrderCollectionGames sets the order of games in a collection, optionally
// removing games at the same time.
// Requires the collection:edit scope.
func (c *Client) OrderCollectionGames(ctx context.Context, p OrderCollectionGamesParams) (*OrderCollectionGamesResponse, error) {
	q := NewQuery(c, "/collections/%d/order", p.CollectionID)
	if p.GameIDs == nil {
		p.GameIDs = []int64{}
	}
	q.AddInt64List("game_ids", p.GameIDs)
	if len(p.RemoveGameIDs) > 0 {
		q.AddInt64List("remove_game_ids", p.RemoveGameIDs)
	}
	r := &OrderCollectionGamesResponse{}
	return r, q.Post(ctx, r)
}
