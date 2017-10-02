package itchio

// User represents an itch.io account, with basic profile info
type User struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	CoverURL      string `json:"coverUrl"`
	StillCoverURL string `json:"stillCoverUrl"`
}

// Game represents a page on itch.io, it could be a game,
// a tool, a comic, etc.
type Game struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`

	Title     string `json:"title"`
	ShortText string `json:"shortText"`
	Type      string `json:"type"`

	CoverURL      string `json:"coverUrl"`
	StillCoverURL string `json:"stillCoverUrl"`

	CreatedAt   string `json:"createdAt"`
	PublishedAt string `json:"publishedAt"`

	MinPrice      int64 `json:"minPrice"`
	InPressSystem bool  `json:"inPressSystem"`
	HasDemo       bool  `json:"hasDemo"`

	OSX     bool `json:"pOsx"`
	Linux   bool `json:"pLinux"`
	Windows bool `json:"pWindows"`
	Android bool `json:"pAndroid"`
}

// An Upload is a downloadable file
type Upload struct {
	ID          int64  `json:"id"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ChannelName string `json:"channelName"`
	Build       *BuildInfo

	OSX     bool `json:"pOsx"`
	Linux   bool `json:"pLinux"`
	Windows bool `json:"pWindows"`
	Android bool `json:"pAndroid"`
}

// BuildFileInfo contains information about a build's "file", which could be its
// archive, its signature, its patch, etc.
type BuildFileInfo struct {
	ID      int64
	Size    int64
	State   BuildFileState
	Type    BuildFileType    `json:"type"`
	SubType BuildFileSubType `json:"subType"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// BuildInfo contains information about a specific build
type BuildInfo struct {
	ID            int64
	ParentBuildID int64 `json:"parentBuildId"`
	State         BuildState

	Version     int64
	UserVersion string `json:"userVersion"`

	Files []*BuildFileInfo

	User      User
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ChannelInfo contains information about a channel and its current status
type ChannelInfo struct {
	Name string
	Tags string

	Upload  Upload
	Head    *BuildInfo `json:"head"`
	Pending *BuildInfo `json:"pending"`
}

// A BuildEvent describes something that happened while we were processing a build.
type BuildEvent struct {
	Type    BuildEventType `json:"type"`
	Message string         `json:"message"`
	Data    BuildEventData `json:"data"`
}
