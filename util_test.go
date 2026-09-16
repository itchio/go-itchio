package itchio

import "testing"

func TestParseSpec(t *testing.T) {
	cases := []struct {
		in      string
		target  string
		channel string
		err     bool
	}{
		{in: "leafo/x-moon", target: "leafo/x-moon"},
		{in: "leafo/x-moon:win-64", target: "leafo/x-moon", channel: "win-64"},
		{in: "12345:linux", target: "12345", channel: "linux"},
		{in: "Leafo/X-Moon:Win-64", target: "leafo/x-moon", channel: "win-64"},

		{in: "https://leafo.itch.io/test-steam-sync", target: "leafo/test-steam-sync"},
		{in: "http://leafo.itch.io/test-steam-sync", target: "leafo/test-steam-sync"},
		{in: "leafo.itch.io/test-steam-sync", target: "leafo/test-steam-sync"},
		{in: "https://leafo.itch.io/test-steam-sync:win-64", target: "leafo/test-steam-sync", channel: "win-64"},
		{in: "leafo.itch.io/test-steam-sync:win-64", target: "leafo/test-steam-sync", channel: "win-64"},
		{in: "https://leafo.itch.io/test-steam-sync/", target: "leafo/test-steam-sync"},
		{in: "https://leafo.itch.io/test-steam-sync/purchase", target: "leafo/test-steam-sync"},
		{in: "https://leafo.itch.io/test-steam-sync/devlog/1/hello", target: "leafo/test-steam-sync"},
		{in: "https://leafo.itch.io/test-steam-sync?secret=abc", target: "leafo/test-steam-sync"},
		{in: "https://leafo.itch.io/test-steam-sync#comments", target: "leafo/test-steam-sync"},

		{in: "", err: true},
		{in: "https://", err: true},
		{in: "leafo/x-moon:win-64:extra", err: true},
		// host without a game page is left alone and fails server-side
		{in: "https://leafo.itch.io", target: "leafo.itch.io"},
		{in: "https://leafo.itch.io/", target: "leafo.itch.io/"},
		// custom domains and non-user hosts can't be resolved client-side
		{in: "https://mygame.com/page", target: "mygame.com/page"},
		{in: "https://itch.io/games/x", target: "itch.io/games/x"},
		{in: "https://a.b.itch.io/page", target: "a.b.itch.io/page"},
	}

	for _, c := range cases {
		spec, err := ParseSpec(c.in)
		if c.err {
			if err == nil {
				t.Errorf("%q: expected error, got %+v", c.in, spec)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: unexpected error: %v", c.in, err)
			continue
		}
		if spec.Target != c.target || spec.Channel != c.channel {
			t.Errorf("%q: got target=%q channel=%q, want target=%q channel=%q", c.in, spec.Target, spec.Channel, c.target, c.channel)
		}
	}
}
