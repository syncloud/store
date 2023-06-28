package model

type StoreSnapDelta struct {
	Format   string `json:"format"`
	Sha3_384 string `json:"sha3-384"`
	Size     int64  `json:"size"`
	Source   int    `json:"source"`
	Target   int    `json:"target"`
	URL      string `json:"url"`
}
