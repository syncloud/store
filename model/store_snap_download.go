package model

type StoreSnapDownload struct {
	Sha3_384 string           `json:"sha3-384"`
	Size     int64            `json:"size"`
	URL      string           `json:"url"`
	Deltas   []StoreSnapDelta `json:"deltas,omitempty"`
}
