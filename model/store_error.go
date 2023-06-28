package model

type StoreError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Extra   struct {
		Releases []SnapRelease `json:"releases"`
	} `json:"extra"`
}
