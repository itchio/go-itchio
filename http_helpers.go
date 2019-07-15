package itchio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

func getDebugLevel() int64 {
	val, err := strconv.ParseInt(os.Getenv("GO_ITCHIO_DEBUG"), 10, 64)
	if err != nil {
		return 0
	}
	return val
}

var (
	debugLevel   = getDebugLevel()
	logRequests  = debugLevel >= 1
	dumpAPICalls = debugLevel >= 2
)

// Get performs an HTTP GET request to the API
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Do(req)
}

// GetResponse performs an HTTP GET request and parses the API response.
func (c *Client) GetResponse(ctx context.Context, url string, dst interface{}) error {
	resp, err := c.Get(ctx, url)
	if err != nil {
		return errors.WithStack(err)
	}

	err = ParseAPIResponse(dst, resp)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// PostForm performs an HTTP POST request to the API, with url-encoded parameters
func (c *Client) PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.Do(req)
}

// PostFormResponse performs an HTTP POST request to the API *and* parses the API response.
func (c *Client) PostFormResponse(ctx context.Context, url string, data url.Values, dst interface{}) error {
	resp, err := c.PostForm(ctx, url, data)
	if err != nil {
		return errors.WithStack(err)
	}

	err = ParseAPIResponse(dst, resp)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Do performs a request (any method). It takes care of JWT or API key
// authentication, sets the proper user agent, has built-in retry,
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", c.Key)
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept-Language", c.AcceptedLanguage)
	req.Header.Set("Accept", "application/vnd.itch.v2")

	var res *http.Response
	var err error

	if logRequests {
		fmt.Fprintf(os.Stderr, "%s %s [request]\n", req.Method, req.URL)
	}

	if dumpAPICalls {
		for k, vv := range req.Header {
			for _, v := range vv {
				fmt.Fprintf(os.Stderr, "[request] %s: %s\n", k, v)
			}
		}
	}

	retryPatterns := append(c.RetryPatterns, time.Millisecond)

	for _, sleepTime := range retryPatterns {
		c.Limiter.Wait(req.Context())
		if c.onOutgoingRequest != nil {
			c.onOutgoingRequest(req)
		}
		res, err = c.HTTPClient.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "TLS handshake timeout") {
				time.Sleep(sleepTime + time.Duration(rand.Int()%1000)*time.Millisecond)
				continue
			}
			return nil, err
		}

		if res.StatusCode == 503 {
			res.Body.Close()

			// Rate limited, try again according to patterns.
			// following https://cloud.google.com/storage/docs/json_api/v1/how-tos/upload#exp-backoff to the letter
			actualSleepTime := sleepTime + time.Duration(rand.Int()%1000)*time.Millisecond
			if c.onRateLimited != nil {
				c.onRateLimited(req, res)
			}
			if logRequests {
				fmt.Fprintf(os.Stderr, "%s %s [rate limited, sleeping %v]\n", req.Method, req.URL, actualSleepTime)
			}
			time.Sleep(actualSleepTime)
			continue
		}

		break
	}

	return res, err
}

// MakePath crafts an API url from our configured base URL
func (c *Client) MakePath(format string, a ...interface{}) string {
	return c.MakeValuesPath(nil, format, a...)
}

// MakeValuesPath crafts an API url from our configured base URL
func (c *Client) MakeValuesPath(values url.Values, format string, a ...interface{}) string {
	base := strings.Trim(c.BaseURL, "/")
	subPath := strings.Trim(fmt.Sprintf(format, a...), "/")
	path := fmt.Sprintf("%s/%s", base, subPath)
	if len(values) == 0 {
		return path
	}
	return fmt.Sprintf("%s?%s", path, values.Encode())
}

func asHTTPCodeError(res *http.Response) error {
	if res.StatusCode/100 != 2 {
		err := fmt.Errorf("Server error: HTTP %s for %s", res.Status, res.Request.URL.Path)
		return err
	}
	return nil
}

// ParseAPIResponse unmarshals an HTTP response into one of out response
// data structures
func ParseAPIResponse(dst interface{}, res *http.Response) error {
	if res == nil || res.Body == nil {
		return fmt.Errorf("No response from server")
	}

	bodyReader := res.Body
	defer bodyReader.Close()

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return errors.WithStack(err)
	}

	if dumpAPICalls {
		fmt.Fprintf(os.Stderr, "[response] %s\n", string(body))
	}

	intermediate := make(map[string]interface{})

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&intermediate)
	if err != nil {
		if he := asHTTPCodeError(res); he != nil {
			return he
		}

		msg := fmt.Sprintf("JSON decode error: %s\n\nBody: %s\n\n", err.Error(), string(body))
		return errors.New(msg)
	}

	if errorsField, ok := intermediate["errors"]; ok {
		if errorsList, ok := errorsField.([]interface{}); ok {
			var messages []string
			for _, el := range errorsList {
				if errorMessage, ok := el.(string); ok {
					messages = append(messages, errorMessage)
				}
			}
			if len(messages) > 0 {
				return &APIError{Messages: messages, StatusCode: res.StatusCode, Path: res.Request.URL.Path}
			}
		}
	}

	if he := asHTTPCodeError(res); he != nil {
		return he
	}

	intermediate = camelifyMap(intermediate)

	if dumpAPICalls {
		enc := json.NewEncoder(os.Stderr)
		enc.SetIndent("[intermediate] ", "  ")
		enc.Encode(intermediate)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  dst,
		// see https://github.com/itchio/itch/issues/1549
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
			GameHookFunc,
			UploadHookFunc,
		),
	})
	if err != nil {
		return errors.WithStack(err)
	}

	err = decoder.Decode(intermediate)
	if err != nil {
		msg := fmt.Sprintf("mapstructure decode error: %s\n\nBody: %#v\n\n", err.Error(), intermediate)
		return errors.New(msg)
	}

	if res.StatusCode != 200 {
		return errors.Errorf("HTTP %v", res.StatusCode)
	}

	return nil
}

// FindBuildFile looks for an uploaded file of the right type
// in a list of file. Returns nil if it can't find one.
func FindBuildFile(fileType BuildFileType, files []*BuildFile) *BuildFile {
	for _, f := range files {
		if f.Type == fileType && f.State == BuildFileStateUploaded {
			return f
		}
	}

	return nil
}

// FindBuildFileEx looks for an uploaded file of the right type
// and subtype in a list of file. Returns nil if it can't find one.
func FindBuildFileEx(fileType BuildFileType, fileSubType BuildFileSubType, files []*BuildFile) *BuildFile {
	for _, f := range files {
		if f.Type == fileType && f.SubType == fileSubType && f.State == BuildFileStateUploaded {
			return f
		}
	}

	return nil
}
