package model

type PublishInitRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Arch     string `json:"arch"`
	Channel  string `json:"channel"`
	Size     int64  `json:"size"`
	Sha384   string `json:"sha384"`
	PartSize int64  `json:"part_size"`
}

type PublishInitResponse struct {
	UploadId         string   `json:"upload_id"`
	Key              string   `json:"key"`
	PartCount        int      `json:"part_count"`
	PartUrls         []string `json:"part_urls"`
	ExpiresInSeconds int64    `json:"expires_in_seconds"`
}
