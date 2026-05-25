package model

type PublishFinaliseRequest struct {
	Token    string        `json:"token"`
	Name     string        `json:"name"`
	Version  string        `json:"version"`
	Arch     string        `json:"arch"`
	Channel  string        `json:"channel"`
	Key      string        `json:"key"`
	UploadId string        `json:"upload_id"`
	Parts    []PublishPart `json:"parts"`
	Size     int64         `json:"size"`
	Sha384   string        `json:"sha384"`
}

type PublishFinaliseResponse struct {
	Ok bool `json:"ok"`
}
