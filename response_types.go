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

// ListChannelsResponse is what the API responds with when we ask for all the
// channels of a particular game
type ListChannelsResponse struct {
	Response

	Channels map[string]ChannelInfo
}

// GetChannelResponse is what the API responds with when we ask info about a channel
type GetChannelResponse struct {
	Response

	Channel ChannelInfo
}

// ListBuildFilesResponse is what the API responds with when we ask for the files
// in a specific build
type ListBuildFilesResponse struct {
	Response

	Files []*BuildFileInfo
}

// CreateBuildFileResponse is what the API responds when we create a new build file
type CreateBuildFileResponse struct {
	Response

	File struct {
		ID            int64
		UploadURL     string            `json:"uploadUrl"`
		UploadParams  map[string]string `json:"uploadParams"`
		UploadHeaders map[string]string `json:"uploadHeaders"`
	}
}

// FinalizeBuildFileResponse is what the API responds when we finalize a build file
type FinalizeBuildFileResponse struct {
	Response
}

// DownloadUploadBuildResponseItem contains download information for a specific
// build file
type DownloadUploadBuildResponseItem struct {
	URL string
}

// DownloadUploadBuildResponse is what the API responds when we want to download
// a build
type DownloadUploadBuildResponse struct {
	Response

	// Patch is the download info for the wharf patch, if any
	Patch *DownloadUploadBuildResponseItem
	// Signature is the download info for the wharf signature, if any
	Signature *DownloadUploadBuildResponseItem
	// Manifest is reserved
	Manifest *DownloadUploadBuildResponseItem
	// Archive is the download info for the .zip archive, if any
	Archive *DownloadUploadBuildResponseItem
}

// CreateBuildEventResponse is what the API responds with when you create a new build event
type CreateBuildEventResponse struct {
	Response
}

// CreateBuildFailureResponse is what the API responds with when we mark a build as failed
type CreateBuildFailureResponse struct {
	Response
}

// ListBuildEventsResponse is what the API responds with when we ask for the list of events for a build
type ListBuildEventsResponse struct {
	Response

	Events []BuildEvent
}
