package itchio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/go-errors/errors"
)

// A Client allows consuming the itch.io API
type Client struct {
	Key           string
	HTTPClient    *http.Client
	BaseURL       string
	RetryPatterns []time.Duration
	UserAgent     string
}

func defaultRetryPatterns() []time.Duration {
	return []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
	}
}

// ClientWithKey creates a new itch.io API client with a given API key
func ClientWithKey(key string) *Client {
	c := &Client{
		Key:           key,
		HTTPClient:    http.DefaultClient,
		RetryPatterns: defaultRetryPatterns(),
		UserAgent:     "go-itchio",
	}
	c.SetServer("https://itch.io")
	return c
}

// SetServer allows changing the server to which we're making API
// requests (which defaults to the reference itch.io server)
func (c *Client) SetServer(itchioServer string) *Client {
	c.BaseURL = fmt.Sprintf("%s/api/1", itchioServer)
	return c
}

// WharfStatus requests the status of the wharf infrastructure
func (c *Client) WharfStatus() (r WharfStatusResponse, err error) {
	path := c.MakePath("wharf/status")

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

// ListMyGames lists the games one develops (ie. can edit)
func (c *Client) ListMyGames() (r ListMyGamesResponse, err error) {
	path := c.MakePath("my-games")
	log.Printf("Requesting %s\n", path)

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

// GameUploadsResponse is what the server replies with when asked for a game's uploads
type GameUploadsResponse struct {
	Response

	Uploads []*Upload `json:"uploads"`
}

// GameUploads lists the uploads for a game that we have access to with our API key
func (c *Client) GameUploads(gameID int64) (r GameUploadsResponse, err error) {
	path := c.MakePath("game/%d/uploads", gameID)

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

// UploadDownloadResponse is what the API replies to when we ask to download an upload
type UploadDownloadResponse struct {
	Response

	URL string
}

// UploadDownload attempts to download an upload without a download key
func (c *Client) UploadDownload(uploadID int64) (r UploadDownloadResponse, err error) {
	return c.UploadDownloadWithKey("", uploadID)
}

// UploadDownloadWithKey attempts to download an upload with a download key
func (c *Client) UploadDownloadWithKey(downloadKey string, uploadID int64) (r UploadDownloadResponse, err error) {
	return c.UploadDownloadWithKeyAndValues(downloadKey, uploadID, nil)
}

// UploadDownloadWithKeyAndValues attempts to download an upload with a download key and additional parameters
func (c *Client) UploadDownloadWithKeyAndValues(downloadKey string, uploadID int64, values url.Values) (r UploadDownloadResponse, err error) {
	if values == nil {
		values = url.Values{}
	}

	if downloadKey != "" {
		values.Add("download_key_id", downloadKey)
	}

	path := c.MakePath("upload/%d/download", uploadID)
	if len(values) > 0 {
		path = fmt.Sprintf("%s?%s", path, values.Encode())
	}

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return r, err
}

// NewBuildResponse is what the API replies with when we create a new build
type NewBuildResponse struct {
	Response

	Build struct {
		ID          int64 `json:"id"`
		UploadID    int64 `json:"upload_id"`
		ParentBuild struct {
			ID int64 `json:"id"`
		} `json:"parent_build"`
	}
}

// CreateBuild creates a new build for a given user/game:channel, with
// an optional user version
func (c *Client) CreateBuild(target string, channel string, userVersion string) (r NewBuildResponse, err error) {
	path := c.MakePath("wharf/builds")

	form := url.Values{}
	form.Add("target", target)
	form.Add("channel", channel)
	if userVersion != "" {
		form.Add("user_version", userVersion)
	}

	resp, err := c.PostForm(path, form)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

// ListChannelsResponse is what the API responds with when we ask for all the
// channels of a particular game
type ListChannelsResponse struct {
	Response

	Channels map[string]ChannelInfo
}

// GetChannelResponse
type GetChannelResponse struct {
	Response

	Channel ChannelInfo
}

func (c *Client) ListChannels(target string) (r ListChannelsResponse, err error) {
	form := url.Values{}
	form.Add("target", target)
	path := c.MakePath("wharf/channels?%s", form.Encode())

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

func (c *Client) GetChannel(target string, channel string) (r GetChannelResponse, err error) {
	form := url.Values{}
	form.Add("target", target)
	path := c.MakePath("wharf/channels/%s?%s", channel, form.Encode())

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

type BuildFileType string

const (
	BuildFileType_PATCH     BuildFileType = "patch"
	BuildFileType_ARCHIVE                 = "archive"
	BuildFileType_SIGNATURE               = "signature"
	BuildFileType_MANIFEST                = "manifest"
	BuildFileType_UNPACKED                = "unpacked"
)

type BuildFileSubType string

const (
	BuildFileSubType_DEFAULT   BuildFileSubType = "default"
	BuildFileSubType_GZIP                       = "gzip"
	BuildFileSubType_OPTIMIZED                  = "optimized"
)

type UploadType string

const (
	UploadType_MULTIPART          UploadType = "multipart"
	UploadType_RESUMABLE                     = "resumable"
	UploadType_DEFERRED_RESUMABLE            = "deferred_resumable"
)

type BuildState string

const (
	BuildState_STARTED    BuildState = "started"
	BuildState_PROCESSING            = "processing"
	BuildState_COMPLETED             = "completed"
	BuildState_FAILED                = "failed"
)

type BuildFileState string

const (
	BuildFileState_CREATED   BuildFileState = "created"
	BuildFileState_UPLOADING                = "uploading"
	BuildFileState_UPLOADED                 = "uploaded"
	BuildFileState_FAILED                   = "failed"
)

type ListBuildFilesResponse struct {
	Response

	Files []*BuildFileInfo
}

func (c *Client) ListBuildFiles(buildID int64) (r ListBuildFilesResponse, err error) {
	path := c.MakePath("wharf/builds/%d/files", buildID)

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

type NewBuildFileResponse struct {
	Response

	File struct {
		ID            int64
		UploadURL     string            `json:"upload_url"`
		UploadParams  map[string]string `json:"upload_params"`
		UploadHeaders map[string]string `json:"upload_headers"`
	}
}

func (c *Client) CreateBuildFile(buildID int64, fileType BuildFileType, subType BuildFileSubType, uploadType UploadType) (r NewBuildFileResponse, err error) {
	path := c.MakePath("wharf/builds/%d/files", buildID)

	form := url.Values{}
	form.Add("type", string(fileType))
	if subType != "" {
		form.Add("sub_type", string(subType))
	}
	if uploadType != "" {
		form.Add("upload_type", string(uploadType))
	}

	resp, err := c.PostForm(path, form)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

func (c *Client) CreateBuildFileWithName(buildID int64, fileType BuildFileType, subType BuildFileSubType, uploadType UploadType, name string) (r NewBuildFileResponse, err error) {
	path := c.MakePath("wharf/builds/%d/files", buildID)

	form := url.Values{}
	form.Add("type", string(fileType))
	if subType != "" {
		form.Add("sub_type", string(subType))
	}
	if uploadType != "" {
		form.Add("upload_type", string(uploadType))
	}
	if name != "" {
		form.Add("filename", name)
	}

	resp, err := c.PostForm(path, form)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

type FinalizeBuildFileResponse struct {
	Response
}

func (c *Client) FinalizeBuildFile(buildID int64, fileID int64, size int64) (r FinalizeBuildFileResponse, err error) {
	path := c.MakePath("wharf/builds/%d/files/%d", buildID, fileID)

	form := url.Values{}
	form.Add("size", fmt.Sprintf("%d", size))

	resp, err := c.PostForm(path, form)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

type DownloadBuildFileResponse struct {
	Response

	URL string
}

var (
	BuildFileNotFound = errors.New("build file not found in storage")
)

func (c *Client) GetBuildFileDownloadURL(buildID int64, fileID int64) (r DownloadBuildFileResponse, err error) {
	return c.GetBuildFileDownloadURLWithValues(buildID, fileID, nil)
}

func (c *Client) GetBuildFileDownloadURLWithValues(buildID int64, fileID int64, values url.Values) (r DownloadBuildFileResponse, err error) {
	path := c.MakePath("wharf/builds/%d/files/%d/download", buildID, fileID)
	if values != nil {
		path = path + "?" + values.Encode()
	}

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	if err != nil {
		return
	}

	return
}

func (c *Client) DownloadBuildFile(buildID int64, fileID int64) (reader io.ReadCloser, err error) {
	path := c.MakePath("wharf/builds/%d/files/%d/download", buildID, fileID)

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	var r DownloadBuildFileResponse
	err = ParseAPIResponse(&r, resp)
	if err != nil {
		return
	}

	req, err := http.NewRequest("GET", r.URL, nil)
	if err != nil {
		return
	}

	// not an API request, going directly with http's DefaultClient
	dlResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if dlResp.StatusCode == 200 {
		reader = dlResp.Body
		return
	}

	dlResp.Body.Close()

	if dlResp.StatusCode == 404 {
		err = BuildFileNotFound
	} else {
		err = fmt.Errorf("Can't download: %s", dlResp.Status)
	}
	return
}

type DownloadUploadBuildResponseItem struct {
	URL string
}

type DownloadUploadBuildResponse struct {
	Response

	Patch     *DownloadUploadBuildResponseItem
	Signature *DownloadUploadBuildResponseItem
	Manifest  *DownloadUploadBuildResponseItem
	Archive   *DownloadUploadBuildResponseItem
}

func (c *Client) DownloadUploadBuild(uploadID int64, buildID int64) (r DownloadUploadBuildResponse, err error) {
	return c.DownloadUploadBuildWithKey("", uploadID, buildID)
}

func (c *Client) DownloadUploadBuildWithKey(downloadKey string, uploadID int64, buildID int64) (r DownloadUploadBuildResponse, err error) {
	return c.DownloadUploadBuildWithKeyAndValues("", uploadID, buildID, nil)
}

func (c *Client) DownloadUploadBuildWithKeyAndValues(downloadKey string, uploadID int64, buildID int64, values url.Values) (r DownloadUploadBuildResponse, err error) {
	if values == nil {
		values = url.Values{}
	}

	if downloadKey != "" {
		values.Add("download_key_id", downloadKey)
	}

	path := c.MakePath("upload/%d/download/builds/%d", uploadID, buildID)
	if len(values) > 0 {
		path = fmt.Sprintf("%s?%s", path, values.Encode())
	}

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	if err != nil {
		return
	}

	return
}

// BuildEventType specifies what kind of event a build event is - could be a log message, etc.
type BuildEventType string

const (
	// BuildEvent_LOG is for build events of type log message
	BuildEvent_LOG BuildEventType = "log"
)

// BuildEventData is a JSON object associated with a build event
type BuildEventData map[string]interface{}

// NewBuildEventResponse is what the API responds with when you create a new build event
type NewBuildEventResponse struct {
	Response
}

// CreateBuildEvent associates a new build event to a build
func (c *Client) CreateBuildEvent(buildID int64, eventType BuildEventType, message string, data BuildEventData) (r NewBuildEventResponse, err error) {
	path := c.MakePath("wharf/builds/%d/events", buildID)

	form := url.Values{}
	form.Add("type", string(eventType))
	form.Add("message", message)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	form.Add("data", string(jsonData))

	resp, err := c.PostForm(path, form)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

// CreateBuildFailureResponse is what the API responds with when we mark a build as failed
type CreateBuildFailureResponse struct {
	Response
}

// CreateBuildFailure marks a given build as failed. We get to specify an error message and
// if it's a fatal error (if not, the build can be retried after a bit)
func (c *Client) CreateBuildFailure(buildID int64, message string, fatal bool) (r CreateBuildFailureResponse, err error) {
	path := c.MakePath("wharf/builds/%d/failures", buildID)

	form := url.Values{}
	form.Add("message", message)
	if fatal {
		form.Add("fatal", fmt.Sprintf("%v", fatal))
	}

	resp, err := c.PostForm(path, form)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

// CreateRediffBuildFailure marks a given build as having failed to rediff (optimize)
func (c *Client) CreateRediffBuildFailure(buildID int64, message string) (r CreateBuildFailureResponse, err error) {
	path := c.MakePath("wharf/builds/%d/failures/rediff", buildID)

	form := url.Values{}
	form.Add("message", message)

	resp, err := c.PostForm(path, form)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

// A BuildEvent describes something that happened while we were processing a build.
type BuildEvent struct {
	Type    BuildEventType
	Message string
	Data    BuildEventData
}

// ListBuildEventsResponse is what the API responds with when we ask for the list of events for a build
type ListBuildEventsResponse struct {
	Response

	Events []BuildEvent
}

// ListBuildEvents returns a series of events associated with a given build
func (c *Client) ListBuildEvents(buildID int64) (r ListBuildEventsResponse, err error) {
	path := c.MakePath("wharf/builds/%d/events", buildID)

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

// Helpers

// Get performs an HTTP GET request to the API
func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// PostForm performs an HTTP POST request to the API, with url-encoded parameters
func (c *Client) PostForm(url string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.Do(req)
}

// Do performs a request (any method). It takes care of JWT or API key
// authentication, sets the propre user agent, has built-in retry,
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(c.Key, "jwt:") {
		req.Header.Add("Authorization", strings.Split(c.Key, ":")[1])
	}
	req.Header.Set("User-Agent", c.UserAgent)

	var res *http.Response
	var err error

	retryPatterns := append(c.RetryPatterns, time.Millisecond)

	for _, sleepTime := range retryPatterns {
		res, err = c.HTTPClient.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "TLS handshake timeout") {
				time.Sleep(sleepTime + time.Duration(rand.Int()%1000)*time.Millisecond)
				continue
			}
			return nil, err
		}

		if res.StatusCode == 503 {
			// Rate limited, try again according to patterns.
			// following https://cloud.google.com/storage/docs/json_api/v1/how-tos/upload#exp-backoff to the letter
			res.Body.Close()
			time.Sleep(sleepTime + time.Duration(rand.Int()%1000)*time.Millisecond)
			continue
		}

		break
	}

	return res, err
}

// MakePath crafts an API url from our configured base URL
func (c *Client) MakePath(format string, a ...interface{}) string {
	base := strings.Trim(c.BaseURL, "/")
	subPath := strings.Trim(fmt.Sprintf(format, a...), "/")

	var key string
	if strings.HasPrefix(c.Key, "jwt:") {
		key = "jwt"
	} else {
		key = c.Key
	}
	return fmt.Sprintf("%s/%s/%s", base, key, subPath)
}

// ParseAPIResponse unmarshals an HTTP response into one of out response
// data structures
func ParseAPIResponse(dst interface{}, res *http.Response) error {
	if res == nil || res.Body == nil {
		return fmt.Errorf("No response from server")
	}

	bodyReader := res.Body
	defer bodyReader.Close()

	if res.StatusCode/100 != 2 {
		return fmt.Errorf("Server returned %s for %s", res.Status, res.Request.URL.String())
	}

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return errors.Wrap(err, 1)
	}

	err = json.NewDecoder(bytes.NewReader(body)).Decode(dst)
	if err != nil {
		msg := fmt.Sprintf("JSON decode error: %s\n\nBody: %s\n\n", err.Error(), string(body))
		return errors.Wrap(errors.New(msg), 1)
	}

	errs := reflect.Indirect(reflect.ValueOf(dst)).FieldByName("Errors")
	if errs.Len() > 0 {
		// TODO: handle other errors too
		return fmt.Errorf("itch.io API error: %s", errs.Index(0).String())
	}

	return nil
}

// FindBuildFile looks for an uploaded file of the right type
// in a list of file. Returns nil if it can't find one.
func FindBuildFile(fileType BuildFileType, files []*BuildFileInfo) *BuildFileInfo {
	for _, f := range files {
		if f.Type == fileType && f.State == BuildFileState_UPLOADED {
			return f
		}
	}

	return nil
}

// ItchfsURL returns the itchfs:/// url usable to download a given file
// from a given build
func (build BuildInfo) ItchfsURL(file *BuildFileInfo, apiKey string) string {
	return ItchfsURL(build.ID, file.ID, apiKey)
}

// ItchfsURL returns the itchfs:/// url usable to download a given file
// from a given build
func ItchfsURL(buildID int64, fileID int64, apiKey string) string {
	values := url.Values{}
	values.Set("api_key", apiKey)
	return fmt.Sprintf("itchfs:///wharf/builds/%d/files/%d/download?%s",
		buildID, fileID, values.Encode())
}
