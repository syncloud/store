package model

type UIApp struct {
	Name       string `json:"name"`
	Summary    string `json:"summary"`
	IconUrl    string `json:"iconUrl,omitempty"`
	Version    string `json:"version"`
	SnapID     string `json:"snapId"`
	Popularity int    `json:"popularity"`
}
