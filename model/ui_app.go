package model

type UIApp struct {
	Name        string `json:"name"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	IconUrl     string `json:"iconUrl,omitempty"`
	Version     string `json:"version"`
	SnapID      string `json:"snapId"`
}
