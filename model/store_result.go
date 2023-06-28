package model

type StoreResult struct {
	Result              string      `json:"result"`
	Name                string      `json:"name,omitempty"`
	Snap                *Snap       `json:"snap,omitempty"`
	InstanceKey         string      `json:"instance-key"`
	SnapID              string      `json:"snap-id,omitempty"`
	Error               *StoreError `json:"error,omitempty"`
	Key                 string      `json:"key"`
	AssertionStreamURLs []string    `json:"assertion-stream-urls"`
	EffectiveChannel    string      `json:"effective-channel,omitempty"`
}
