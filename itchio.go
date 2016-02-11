package itchio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

type MyGamesResponse struct {
	Response

	Games []Game
}

type GameUploadsResponse struct {
	Response

	Uploads []Upload `json:"uploads"`
}

type UploadDownloadResponse struct {
	Response

	Url string
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

func (c *Client) MyGames() (r MyGamesResponse, err error) {
	path := c.makePath("my-games")

	resp, err := http.Get(path)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return
}

func (c *Client) GameUploads(gameID int64) (r GameUploadsResponse, err error) {
	path := c.makePath("game/%d/uploads", gameID)

	resp, err := http.Get(path)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return
}

func (c *Client) UploadDownload(uploadID int64) (r UploadDownloadResponse, err error) {
	path := c.makePath("upload/%d/download", uploadID)

	resp, err := http.Get(path)
	if err != nil {
		return
	}

	err = parseAPIResponse(&r, resp.Body)
	return r, err
}

func (c *Client) makePath(format string, a ...interface{}) string {
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
