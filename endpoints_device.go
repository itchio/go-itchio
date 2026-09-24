package itchio

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

//-------------------------------------------------------

// DeviceRedirectURI is the redirect URI sent to ExchangeOAuthCode for a
// code obtained through the device grant. The server holds the code for
// the poll instead of redirecting anywhere.
const DeviceRedirectURI = "urn:itchio:poll"

// DeviceAuthParams : params for DeviceAuth
type DeviceAuthParams struct {
	ClientID string
	// Defaults to "itch"
	Scope string
	// S256 challenge of the PKCE verifier, see GeneratePKCE
	CodeChallenge string
}

// DeviceAuthResponse : response for DeviceAuth
type DeviceAuthResponse struct {
	// Polled with; never shown to the user
	DeviceCode string `json:"deviceCode"`
	// Short code shown to the user, which the consent page repeats
	UserCode string `json:"userCode"`
	// Consent page without the code
	VerificationURI string `json:"verificationUri"`
	// Consent page with the code, to render as a QR code
	VerificationURIComplete string `json:"verificationUriComplete"`
	// Seconds until the request expires
	ExpiresIn int64 `json:"expiresIn"`
	// Seconds to wait between polls
	Interval int64 `json:"interval"`
}

// DeviceAuth starts an OAuth device authorization grant (RFC 8628) with
// PKCE, for signing in from a device with no browser. Poll for the
// outcome with PollDeviceAuth, then exchange the code with
// ExchangeOAuthCode using DeviceRedirectURI.
func (c *Client) DeviceAuth(ctx context.Context, params DeviceAuthParams) (*DeviceAuthResponse, error) {
	scope := params.Scope
	if scope == "" {
		scope = "itch"
	}

	q := NewQuery(c, "/oauth/device")
	q.AddString("client_id", params.ClientID)
	q.AddString("scope", scope)
	q.AddString("code_challenge", params.CodeChallenge)
	q.AddString("code_challenge_method", "S256")

	r := &DeviceAuthResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// PollDeviceAuthParams : params for PollDeviceAuth
type PollDeviceAuthParams struct {
	ClientID   string
	DeviceCode string
}

// DeviceAuthStatus is the outcome of one poll
type DeviceAuthStatus string

const (
	// Not decided yet; poll again after Interval
	DeviceAuthPending DeviceAuthStatus = "pending"
	// Approved; exchange Code
	DeviceAuthApproved DeviceAuthStatus = "approved"
	// The user pressed deny
	DeviceAuthDenied DeviceAuthStatus = "denied"
	// The request timed out, or the code was already exchanged
	DeviceAuthExpired DeviceAuthStatus = "expired"
	// Too many polls from this address; wait longer before the next
	DeviceAuthSlowDown DeviceAuthStatus = "slow_down"
)

// PollDeviceAuthResponse : response for PollDeviceAuth
type PollDeviceAuthResponse struct {
	Status DeviceAuthStatus `json:"status"`
	// Authorization code, when approved
	Code string `json:"code"`
	// Seconds to wait before the next poll, when pending
	Interval int64 `json:"interval"`
}

// PollDeviceAuth asks whether a DeviceAuth request has been approved.
// An HTTP 429 comes back as DeviceAuthSlowDown rather than an error.
func (c *Client) PollDeviceAuth(ctx context.Context, params PollDeviceAuthParams) (*PollDeviceAuthResponse, error) {
	q := NewQuery(c, "/oauth/device/poll")
	q.AddString("client_id", params.ClientID)
	q.AddString("device_code", params.DeviceCode)

	res, err := c.PostForm(ctx, c.MakePath("%s", q.Path), q.Values)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if res.StatusCode == http.StatusTooManyRequests {
		res.Body.Close()
		return &PollDeviceAuthResponse{Status: DeviceAuthSlowDown}, nil
	}

	r := &PollDeviceAuthResponse{}
	err = ParseAPIResponse(r, res)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return r, nil
}
