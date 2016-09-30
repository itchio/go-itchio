package itchfs

import (
	"fmt"
	"net/url"

	itchio "github.com/itchio/go-itchio"
)

type ItchFS struct {
	ItchServer string
}

func (ifs *ItchFS) Scheme() string {
	return "itchfs"
}

type GetURLFunc func() (string, error)

type itchfsResource struct {
	getURL GetURLFunc
}

func (ir *itchfsResource) GetURL() (string, error) {
	return ir.getURL()
}

func (ir *itchfsResource) NeedsRenewal() bool {
	// FIXME: stub
	return true
}

func (ifs *ItchFS) MakeResource(u *url.URL) (*itchfsResource, error) {
	if u.Host != "" {
		return nil, fmt.Errorf("invalid itchfs URL (must start with itchfs:///): %s", u.String())
	}

	vals := u.Query()

	apiKey := vals.Get("api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("missing API key")
	}

	itchClient := itchio.ClientWithKey(apiKey)
	if ifs.ItchServer != "" {
		itchClient.SetServer(ifs.ItchServer)
	}

	source, err := ObtainSource(itchClient, u.Path)
	if err != nil {
		return nil, err
	}

	getURL, err := source.makeGetURL()
	if err != nil {
		return nil, err
	}

	return &itchfsResource{getURL}, nil
}
