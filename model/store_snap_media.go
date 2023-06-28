package model

type StoreSnapMedia struct {
	Type   string `json:"type"` // icon/screenshot
	URL    string `json:"url"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}
