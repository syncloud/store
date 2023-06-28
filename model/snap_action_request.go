package model

type SnapActionRequest struct {
	Actions             []*SnapAction  `json:"actions"`
	Fields              []string       `json:"fields"`
	AssertionMaxFormats map[string]int `json:"assertion-max-formats,omitempty"`
}
