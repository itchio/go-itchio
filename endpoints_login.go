package itchio

import "context"

//-------------------------------------------------------

// LoginWithPasswordParams : params for LoginWithPassword
type LoginWithPasswordParams struct {
	Username          string
	Password          string
	RecaptchaResponse string
	ForceRecaptcha    bool
}

// Cookie represents, well, multiple key=value pairs that
// should be set to obtain a logged-in browser session for
// the user who just logged in.
type Cookie map[string]string

// LoginWithPasswordResponse : response for LoginWithPassword
type LoginWithPasswordResponse struct {
	RecaptchaNeeded bool   `json:"recaptchaNeeded"`
	RecaptchaURL    string `json:"recaptchaUrl"`
	TOTPNeeded      bool   `json:"totpNeeded"`
	Token           string `json:"token"`

	Key    *APIKey `json:"key"`
	Cookie Cookie  `json:"cookie"`
}

// LoginWithPassword attempts to log a user into itch.io with
// their username (or e-mail) and password.
// The response may indicate that a TOTP code is needed (for two-factor auth),
// or a recaptcha challenge is needed (an unfortunate remedy for an unfortunate ailment).
func (c *Client) LoginWithPassword(ctx context.Context, params LoginWithPasswordParams) (*LoginWithPasswordResponse, error) {
	q := NewQuery(c, "/login")
	q.AddString("source", "desktop")
	q.AddString("username", params.Username)
	q.AddString("password", params.Password)
	q.AddStringIfNonEmpty("recaptcha_response", params.RecaptchaResponse)
	q.AddBoolIfTrue("force_recaptcha", params.ForceRecaptcha)

	r := &LoginWithPasswordResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// TOTPVerifyParams : params for TOTPVerify
type TOTPVerifyParams struct {
	Token string
	Code  string
}

// TOTPVerifyResponse : response for TOTPVerify
type TOTPVerifyResponse struct {
	Key    *APIKey `json:"key"`
	Cookie Cookie  `json:"cookie"`
}

// TOTPVerify sends a user-entered TOTP token to the server for
// verification (and to complete login).
func (c *Client) TOTPVerify(ctx context.Context, params TOTPVerifyParams) (*TOTPVerifyResponse, error) {
	q := NewQuery(c, "/totp/verify")
	q.AddString("token", params.Token)
	q.AddString("code", params.Code)

	r := &TOTPVerifyResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// ExchangeOAuthCodeParams : params for ExchangeOAuthCode
type ExchangeOAuthCodeParams struct {
	Code         string
	CodeVerifier string
	RedirectURI  string
}

// ExchangeOAuthCodeResponse : response for ExchangeOAuthCode
type ExchangeOAuthCodeResponse struct {
	Key    *APIKey `json:"key"`
	Cookie Cookie  `json:"cookie"`
}

// ExchangeOAuthCode exchanges an OAuth authorization code (with PKCE) for an API key.
// Used by the desktop app's OAuth login flow.
func (c *Client) ExchangeOAuthCode(ctx context.Context, params ExchangeOAuthCodeParams) (*ExchangeOAuthCodeResponse, error) {
	q := NewQuery(c, "/oauth/token")
	q.AddString("grant_type", "authorization_code")
	q.AddString("code", params.Code)
	q.AddString("code_verifier", params.CodeVerifier)
	q.AddString("redirect_uri", params.RedirectURI)
	q.AddString("client_id", "butler")

	r := &ExchangeOAuthCodeResponse{}
	return r, q.Post(ctx, r)
}

//-------------------------------------------------------

// SubkeyParams : params for Subkey
type SubkeyParams struct {
	GameID int64
	Scope  string
}

// SubkeyResponse : params for Subkey
type SubkeyResponse struct {
	Key       string `json:"key"`
	ExpiresAt string `json:"expiresAt"`
}

// Subkey creates a scoped-down, temporary offspring of the main
// API key this client was created with. It is useful to automatically grant
// some access to games being launched.
func (c *Client) Subkey(ctx context.Context, params SubkeyParams) (*SubkeyResponse, error) {
	q := NewQuery(c, "/credentials/subkey")
	q.AddInt64("game_id", params.GameID)
	q.AddString("scope", params.Scope)

	r := &SubkeyResponse{}
	return r, q.Post(ctx, r)
}
