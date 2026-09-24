package itchio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DeviceAuth(t *testing.T) {
	var body url.Values
	server, client := captureTools(200, `{
		"device_code": "lIGtYIzInNWBdGJILLwrCmdVIJPa5Cs2",
		"user_code": "LVTD-HGCC",
		"verification_uri": "https://itch.io/user/oauth/device",
		"verification_uri_complete": "https://itch.io/user/oauth/device?code=41iCR8LYdZpF6cZNBzc9cM",
		"expires_in": 600,
		"interval": 5
	}`, &body)
	defer server.Close()

	resp, err := client.DeviceAuth(context.Background(), DeviceAuthParams{
		ClientID:      "client-123",
		CodeChallenge: "challenge",
	})
	assert.NoError(t, err)
	assert.Equal(t, "client-123", body.Get("client_id"))
	assert.Equal(t, "itch", body.Get("scope"))
	assert.Equal(t, "challenge", body.Get("code_challenge"))
	assert.Equal(t, "S256", body.Get("code_challenge_method"))

	assert.Equal(t, "lIGtYIzInNWBdGJILLwrCmdVIJPa5Cs2", resp.DeviceCode)
	assert.Equal(t, "LVTD-HGCC", resp.UserCode)
	assert.Equal(t, "https://itch.io/user/oauth/device", resp.VerificationURI)
	assert.Equal(t, "https://itch.io/user/oauth/device?code=41iCR8LYdZpF6cZNBzc9cM", resp.VerificationURIComplete)
	assert.EqualValues(t, 600, resp.ExpiresIn)
	assert.EqualValues(t, 5, resp.Interval)
}

func Test_DeviceAuthScope(t *testing.T) {
	var body url.Values
	server, client := captureTools(200, `{"device_code": "x", "user_code": "y", "expires_in": 1, "interval": 1}`, &body)
	defer server.Close()

	_, err := client.DeviceAuth(context.Background(), DeviceAuthParams{
		ClientID:      "client-123",
		Scope:         "profile:me",
		CodeChallenge: "challenge",
	})
	assert.NoError(t, err)
	assert.Equal(t, "profile:me", body.Get("scope"))
}

func Test_DeviceAuthError(t *testing.T) {
	server, client := testTools(400, `{"errors": ["missing code_challenge"]}`)
	defer server.Close()

	_, err := client.DeviceAuth(context.Background(), DeviceAuthParams{ClientID: "client-123"})
	assert.Error(t, err)
	apiErr, ok := AsAPIError(err)
	assert.True(t, ok)
	assert.Equal(t, []string{"missing code_challenge"}, apiErr.Messages)
}

func Test_PollDeviceAuth(t *testing.T) {
	cases := []struct {
		body     string
		expected PollDeviceAuthResponse
	}{
		{`{"status": "pending", "interval": 7}`, PollDeviceAuthResponse{Status: DeviceAuthPending, Interval: 7}},
		{`{"status": "approved", "code": "auth-code"}`, PollDeviceAuthResponse{Status: DeviceAuthApproved, Code: "auth-code"}},
		{`{"status": "denied"}`, PollDeviceAuthResponse{Status: DeviceAuthDenied}},
		{`{"status": "expired"}`, PollDeviceAuthResponse{Status: DeviceAuthExpired}},
	}

	for _, c := range cases {
		var body url.Values
		server, client := captureTools(200, c.body, &body)

		resp, err := client.PollDeviceAuth(context.Background(), PollDeviceAuthParams{
			ClientID:   "client-123",
			DeviceCode: "device-code",
		})
		assert.NoError(t, err, c.body)
		assert.Equal(t, "client-123", body.Get("client_id"))
		assert.Equal(t, "device-code", body.Get("device_code"))
		assert.Equal(t, c.expected, *resp, c.body)

		server.Close()
	}
}

func Test_PollDeviceAuthSlowDown(t *testing.T) {
	// Plain 429 with no JSON body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := ClientWithKey("APIKEY")
	client.HTTPClient = server.Client()
	client.BaseURL = server.URL

	resp, err := client.PollDeviceAuth(context.Background(), PollDeviceAuthParams{
		ClientID:   "client-123",
		DeviceCode: "device-code",
	})
	assert.NoError(t, err)
	assert.Equal(t, DeviceAuthSlowDown, resp.Status)
}

func Test_PollDeviceAuthSlowDownWithErrors(t *testing.T) {
	server, client := testTools(429, `{"errors": ["slow_down"]}`)
	defer server.Close()

	resp, err := client.PollDeviceAuth(context.Background(), PollDeviceAuthParams{
		ClientID:   "client-123",
		DeviceCode: "device-code",
	})
	assert.NoError(t, err)
	assert.Equal(t, DeviceAuthSlowDown, resp.Status)
}

func Test_PollDeviceAuthError(t *testing.T) {
	server, client := testTools(400, `{"errors": ["invalid_grant"]}`)
	defer server.Close()

	_, err := client.PollDeviceAuth(context.Background(), PollDeviceAuthParams{
		ClientID:   "client-123",
		DeviceCode: "unknown",
	})
	assert.Error(t, err)
	assert.True(t, IsAPIError(err))
}

func Test_GeneratePKCE(t *testing.T) {
	verifier, challenge, err := GeneratePKCE()
	assert.NoError(t, err)
	// 32 bytes, unpadded base64url
	assert.Len(t, verifier, 43)
	assert.Equal(t, PKCEChallenge(verifier), challenge)

	other, _, err := GeneratePKCE()
	assert.NoError(t, err)
	assert.NotEqual(t, verifier, other)
}

func Test_PKCEChallenge(t *testing.T) {
	// RFC 7636 appendix B
	assert.Equal(t,
		"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		PKCEChallenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"),
	)
}
