package itchio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
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

	resp, err := http.Get(path)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return
}

type GameUploadsResponse struct {
	Response

	Uploads []Upload `json:"uploads"`
}

func (c *Client) GameUploads(gameID int64) (r GameUploadsResponse, err error) {
	path := c.MakePath("game/%d/uploads", gameID)

	resp, err := http.Get(path)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return
}

type UploadDownloadResponse struct {
	Response

	Url string
}

func (c *Client) UploadDownload(uploadID int64) (r UploadDownloadResponse, err error) {
	path := c.MakePath("upload/%d/download", uploadID)

	resp, err := http.Get(path)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return r, err
}

type NewBuildResponse struct {
	ID       int64 `json:"id"`
	ParentID int64 `json:"parent_id"`
}

func (c *Client) CreateBuild(target string, channel string) (r NewBuildResponse, err error) {
	path := c.MakePath("builds")

	form := url.Values{}
	form.Add("target", target)
	form.Add("channel", channel)

	resp, err := http.PostForm(path, form)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
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
	Files []BuildFile
}

func (c *Client) ListBuildFiles(buildID int64) (r ListBuildFilesResponse, err error) {
	path := c.MakePath("builds/%d/files")

	resp, err := http.Get(path)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return
}

type NewBuildFileResponse struct {
	ID        int64
	UploadURL string
}

func (c *Client) CreateBuildFile(buildID int64, fileType BuildFileType) (r NewBuildFileResponse, err error) {
	path := c.MakePath("builds/%d/files", buildID)

	form := url.Values{}
	form.Add("type", string(fileType))

	resp, err := http.PostForm(path, form)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return
}

type FinalizeBuildFileResponse struct{}

func (c *Client) FinalizeBuildFile(buildID int64, fileID int64, size int64) (r FinalizeBuildFileResponse, err error) {
	path := c.MakePath("builds/%d/files/%d", buildID, fileID)

	form := url.Values{}
	form.Add("size", fmt.Sprintf("%d", size))

	resp, err := http.PostForm(path, form)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return
}

func (c *Client) DownloadBuildFile(buildID int64, fileID int64) (reader io.ReadCloser, err error) {
	path := c.MakePath("builds/%d/files/%d/download", buildID, fileID)

	resp, err := http.Get(path)
	if err != nil {
		return
	}

	if resp.StatusCode/100 != 3 {
		err = fmt.Errorf("expected redirection")
		return
	}

	dlPath := resp.Header.Get("Location")

	req, err := http.NewRequest("GET", dlPath, nil)
	if err != nil {
		return
	}

	reader = req.Body
	return
}

type BuildEventType string

const (
	BuildEvent_JOB_STARTED   BuildEventType = "job_started"
	BuildEvent_JOB_FAILED                   = "job_failed"
	BuildEvent_JOB_COMPLETED                = "job_completed"
)

type BuildEventData map[string]interface{}

type NewBuildEventResponse struct{}

func (c *Client) CreateBuildEvent(buildID int64, eventType BuildEventType, message string, data BuildEventData) (r NewBuildEventResponse, err error) {
	path := c.MakePath("builds/%d/events", buildID)

	form := url.Values{}
	form.Add("type", string(eventType))
	form.Add("message", message)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	form.Add("data", string(jsonData))

	resp, err := http.PostForm(path, form)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return
}

type BuildEvent struct {
	Type    BuildEventType
	Message string
	Data    BuildEventData
}

type ListBuildEventsResponse struct {
	Events []BuildEvent
}

func (c *Client) ListBuildEvents(buildID int64) (r ListBuildEventsResponse, err error) {
	path := c.MakePath("builds/%d/events", buildID)

	resp, err := http.Get(path)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return
}

func (c *Client) MakePath(format string, a ...interface{}) string {
	subPath := fmt.Sprintf(format, a...)
	return fmt.Sprintf("%s/%s/%s", c.BaseURL, c.Key, subPath)
}

func parseAPIResponse(dst interface{}, bodyReader io.ReadCloser) error {
	defer bodyReader.Close()

	err := json.NewDecoder(bodyReader).Decode(dst)
	if err != nil {
		return err
	}

	errs := reflect.Indirect(reflect.ValueOf(dst)).FieldByName("Errors")
	if errs.Len() > 0 {
		return errors.New(errs.Index(0).String())
	}

	return nil
}
