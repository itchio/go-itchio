# go-itchio

[![test](https://github.com/itchio/go-itchio/actions/workflows/test.yml/badge.svg)](https://github.com/itchio/go-itchio/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/itchio/go-itchio)](https://goreportcard.com/report/github.com/itchio/go-itchio)
[![Go Reference](https://pkg.go.dev/badge/github.com/itchio/go-itchio.svg)](https://pkg.go.dev/github.com/itchio/go-itchio)
![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)

go-itchio is a set of Go bindings to interact with the itch.io API

## Authentication

### API Key

For static API keys (CI environments, scripts, legacy integrations):

```go
client := itchio.ClientWithKey("your-api-key")

// Use the client
games, err := client.ListProfileGames(ctx)
```

### OAuth with Refresh Tokens

For user-facing applications with automatic token refresh:

```go
// Create an unauthenticated client for the initial exchange
client := itchio.ClientWithKey("")

// Exchange the authorization code for credentials
resp, err := client.ExchangeOAuthCode(ctx, itchio.ExchangeOAuthCodeParams{
    Code:         authCode,
    CodeVerifier: verifier,
    RedirectURI:  redirectURI,
    ClientID:     "your-client-id",
})

// Create an OAuth client with automatic token refresh
client = itchio.NewOAuthClient(
    resp.OAuthCredentials(),
    itchio.OAuthConfig{
        ClientID: "your-client-id",
        OnRefresh: func(creds *itchio.OAuthCredentials) error {
            // Persist refreshed credentials to your storage
            return saveCredentials(creds)
        },
    },
)

// Use normally - tokens refresh automatically
games, err := client.ListProfileGames(ctx)
```

The OAuth client automatically refreshes tokens before they expire and retries requests on 401 responses.

## Debugging

Set `GO_ITCHIO_DEBUG` to enable request logging:

```bash
# Log all HTTP requests (method, URL, rate limiting)
GO_ITCHIO_DEBUG=1 ./your-app

# Also dump full API response bodies
GO_ITCHIO_DEBUG=2 ./your-app
```

## License

Licensed under MIT License, see `LICENSE` for details.
