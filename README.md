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

## Response decoding

The itch.io API returns JSON with `snake_case` keys, but Go response types in this package use `camelCase` `json` struct tags (e.g. `json:"coverUrl"`, not `json:"cover_url"`). Decoding happens in two steps:

1. The response body is decoded into a generic `map[string]interface{}`.
2. Every key in that map is recursively rewritten from `snake_case` to `camelCase` by `camelifyMap` (see `camelify.go`), then `mapstructure` populates the target struct using the `json` tag.

When defining new response types, write `json` tags in `camelCase` to match the post-remap keys.

### `upload_headers` blacklist

The remap walks into nested maps and arrays by default, which would also rewrite the keys of arbitrary HTTP header maps returned by the API. To preserve those headers verbatim, `upload_headers` is listed in `camelifyBlacklist` (`camelify.go`): the field name itself is still camelified to `uploadHeaders`, but its inner map keys are left untouched. Add to this blacklist if you introduce another field whose values are user-controlled key/value data rather than a known schema.

## Debugging

Set `GO_ITCHIO_DEBUG` to enable request logging:

```bash
# Log all HTTP requests (method, URL, rate limiting)
GO_ITCHIO_DEBUG=1 ./your-app

# Also dump full API response bodies
GO_ITCHIO_DEBUG=2 ./your-app
```

### Programmatic Logging with slog

You can attach a `slog` logger to emit structured per-attempt HTTP logs. These logs are emitted at `DEBUG` level:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
client := itchio.ClientWithKey("your-api-key")
client.Logger = logger
```

Each log entry includes fields like `method`, `url` (with sensitive query values redacted), `status_code`, `duration_ms`, `attempt`, `retrying`, and `retry_reason`.

## License

Licensed under MIT License, see `LICENSE` for details.
