package itchio

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"github.com/pkg/errors"
)

// GeneratePKCE returns a fresh PKCE verifier and its S256 challenge
// (RFC 7636). The challenge goes in the authorization request, the
// verifier in the token exchange.
func GeneratePKCE() (verifier string, challenge string, err error) {
	var bytes [32]byte
	_, err = rand.Read(bytes[:])
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(bytes[:])
	return verifier, PKCEChallenge(verifier), nil
}

// PKCEChallenge is the S256 code challenge for a verifier.
func PKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
