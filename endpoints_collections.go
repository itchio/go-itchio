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
