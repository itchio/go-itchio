package itchio

// Response is what the itch.io API replies with. It may
// include one or several errors
type Response struct {
	Errors []string
}

// WharfStatusResponse is what the API responds with when we ask for
// the status of the wharf infrastructure
type WharfStatusResponse struct {
	Response

	Success bool
}

// ListMyGamesResponse is what the API server answers when we ask for what games
// an account develops.
type ListMyGamesResponse struct {
	Response

	Games []Game
}
