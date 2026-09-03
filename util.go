package itchio

import (
	"fmt"
	"strings"
)

// A Spec points to a given itch.io game, optionally to a specific channel
type Spec struct {
	Target  string
	Channel string
}

// ParseSpec parses something of the form `user/page:channel` and returns
// `user/page` and `channel` separately.
//
// The target may also be given as an itch.io game page URL, with or without
// a scheme: `https://user.itch.io/page` or `user.itch.io/page:channel` are
// normalized to `user/page`.
func ParseSpec(specIn string) (*Spec, error) {
	specStr := strings.ToLower(strings.TrimSpace(specIn))

	// strip the scheme first, otherwise its colon reads as the channel separator
	specStr = strings.TrimPrefix(specStr, "https://")
	specStr = strings.TrimPrefix(specStr, "http://")

	tokens := strings.Split(specStr, ":")

	spec := &Spec{}

	switch len(tokens) {
	case 1:
		// no channel
		spec.Target = tokens[0]
	case 2:
		spec.Target = tokens[0]
		spec.Channel = tokens[1]
	default:
		return nil, fmt.Errorf("invalid spec: %s, expected something of the form user/page:channel", specIn)
	}

	if target, ok := targetFromURL(spec.Target); ok {
		spec.Target = target
	}

	if spec.Target == "" {
		return nil, fmt.Errorf("invalid spec: %s, expected something of the form user/page:channel", specIn)
	}

	return spec, nil
}

// targetFromURL turns a scheme-less itch.io game page URL like
// `user.itch.io/page/anything?x#y` into `user/page`. It only matches hosts
// of the form `<user>.itch.io`; custom domains can't be resolved client-side.
func targetFromURL(s string) (string, bool) {
	host, path, hasPath := strings.Cut(s, "/")
	user, ok := strings.CutSuffix(host, ".itch.io")
	if !ok || user == "" || strings.Contains(user, ".") || !hasPath {
		return "", false
	}

	page := path
	if i := strings.IndexAny(page, "/?#"); i >= 0 {
		page = page[:i]
	}
	if page == "" {
		return "", false
	}

	return user + "/" + page, true
}

func (spec *Spec) String() string {
	if spec.Channel != "" {
		return fmt.Sprintf("%s/%s", spec.Target, spec.Channel)
	}
	return spec.Target
}

// EnsureChannel returns an error if this spec is missing a channel
func (spec *Spec) EnsureChannel() error {
	if spec.Channel == "" {
		specStr := spec.String()
		return fmt.Errorf("invalid spec: %s, missing channel (examples: %s:windows-32-beta, %s:linux-64)", specStr, specStr, specStr)
	}

	return nil
}
