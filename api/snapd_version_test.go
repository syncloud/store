package api

import "testing"

func TestSnapdVersion(t *testing.T) {
	cases := map[string]string{
		"snapd/639 (series 16; classic) debian/10 (amd64) linux/4.19.0": "639",
		"snapd/639":                    "639",
		"snapd/2.59.5 (series 16)":     "2.59.5",
		"snapd/ (series 16)":           "other",
		"snapd/injected label} (x)":    "other",
		"Mozilla/5.0 (X11; Linux x86_64)": "unknown",
		"":                             "unknown",
	}
	for ua, want := range cases {
		if got := snapdVersion(ua); got != want {
			t.Errorf("snapdVersion(%q) = %q, want %q", ua, got, want)
		}
	}
}
