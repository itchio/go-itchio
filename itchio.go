package itchio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type Client struct {
	Key        string
	HTTPClient *http.Client
	BaseURL    string
}

type Response struct {
	Errors []string
}

type User struct {
	ID       int64
	Username string
	CoverUrl string `json:"cover_url"`
}

type Game struct {
	ID  int64
	Url string
}

type Upload struct {
	ID       int64
	Filename string
	Size     int64

	OSX     bool `json:"p_osx"`
	Linux   bool `json:"p_linux"`
	Windows bool `json:"p_windows"`
	Android bool `json:"p_android"`
}

func ClientWithKey(key string) *Client {
	return &Client{
		Key:        key,
		HTTPClient: &http.Client{},
		BaseURL:    "https://itch.io/api/1",
	}
}

type MyGamesResponse struct {
	Response

	Games []Game
}

func (c *Client) MyGames() (r MyGamesResponse, err error) {
	path := c.MakePath("my-games")

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

type GameUploadsResponse struct {
	Response

	Uploads []Upload `json:"uploads"`
}

func (c *Client) GameUploads(gameID int64) (r GameUploadsResponse, err error) {
	path := c.MakePath("game/%d/uploads", gameID)

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return
}

type UploadDownloadResponse struct {
	Response

	Url string
}

func (c *Client) UploadDownload(uploadID int64) (r UploadDownloadResponse, err error) {
	path := c.MakePath("upload/%d/download", uploadID)

	resp, err := c.Get(path)
	if err != nil {
		return
	}

	err = ParseAPIResponse(&r, resp)
	return r, err
}

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

func (c *Client) CreateBuild(target string, channel string) (r NewBuildResponse, err error) {
	path := c.MakePath("wharf/builds")

	form := url.Values{}
	form.Add("target", target)
	form.Add("channel", channel)

	resp, err := c.PostForm(path, form)
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
)

type BuildFile struct {
	ID   int64
	Type BuildFileType
	Size int64
}

type ListBuildFilesResponse struct {
	Response

	Files []BuildFile
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
		ID           int64
		UploadURL    string            `json:"upload_url"`
		UploadParams map[string]string `json:"upload_params"`
	}
}

func (c *Client) CreateBuildFile(buildID int64, fileType BuildFileType) (r NewBuildFileResponse, err error) {
	path := c.MakePath("wharf/builds/%d/files", buildID)

	form := url.Values{}
	form.Add("type", string(fileType))

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

	reader = dlResp.Body
	return
}

type BuildEventType string

const (
	BuildEvent_JOB_STARTED   BuildEventType = "job_started"
	BuildEvent_JOB_FAILED                   = "job_failed"
	BuildEvent_JOB_COMPLETED                = "job_completed"
)

type BuildEventData map[string]interface{}

type NewBuildEventResponse struct {
	Response
}

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

type BuildEvent struct {
	Type    BuildEventType
	Message string
	Data    BuildEventData
}

type ListBuildEventsResponse struct {
	Response

	Events []BuildEvent
}

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

func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) PostForm(url string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(c.Key, "jwt:") {
		req.Header.Add("Authorization", strings.Split(c.Key, ":")[1])
	}
	return http.DefaultClient.Do(req)
}

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

func ParseAPIResponse(dst interface{}, res *http.Response) error {
	bodyReader := res.Body
	defer bodyReader.Close()

	if res.StatusCode/100 != 2 {
		return fmt.Errorf("Server returned %s for %s", res.Status, res.Request.URL.String())
	}

	err := json.NewDecoder(bodyReader).Decode(dst)
	if err != nil {
		return err
	}

	errs := reflect.Indirect(reflect.ValueOf(dst)).FieldByName("Errors")
	if errs.Len() > 0 {
		// TODO: handle other errors too
		return errors.New(errs.Index(0).String())
	}

	return nil
}
