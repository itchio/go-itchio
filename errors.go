package itchio

import (
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// APIError represents an itch.io API error. Some errors
// are just HTTP status codes, others have more detailed messages.
type APIError struct {
	Messages   []string `json:"messages"`
	StatusCode int      `json:"statusCode"`
	Path       string   `json:"path"`
}

var _ error = (*APIError)(nil)

func (ae *APIError) Error() string {
	return fmt.Sprintf("itch.io API error (%d): %s: %s", ae.StatusCode, ae.Path, strings.Join(ae.Messages, ", "))
}

// IsAPIError returns true if an error is an itch.io API error,
// even if it's wrapped with github.com/pkg/errors
func IsAPIError(err error) bool {
	_, ok := AsAPIError(err)
	return ok
}

// AsAPIError returns an *APIError and true if the
// passed error (no matter how deeply wrapped it is)
// is an *APIError. Otherwise it returns nil, false.
func AsAPIError(err error) (*APIError, bool) {
	rootErr := errors.Cause(err)
	apiError, ok := rootErr.(*APIError)
	return apiError, ok
}

// HTTPError is returned for a response that carried no structured API
// error but was not a success, such as a proxy or origin failure.
type HTTPError struct {
	StatusCode int
	Status     string
	Path       string
}

var _ error = (*HTTPError)(nil)

func (he *HTTPError) Error() string {
	return fmt.Sprintf("Server error: HTTP %s for %s", he.Status, he.Path)
}

// AsHTTPError returns the *HTTPError in err's chain, if any.
func AsHTTPError(err error) (*HTTPError, bool) {
	var he *HTTPError
	if stderrors.As(err, &he) {
		return he, true
	}
	return nil, false
}
