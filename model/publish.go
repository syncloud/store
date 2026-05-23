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

type PublishPartUrlRequest struct {
	Token      string `json:"token"`
	Key        string `json:"key"`
	UploadId   string `json:"upload_id"`
	PartNumber int    `json:"part_number"`
}

type PublishPartUrlResponse struct {
	Url string `json:"url"`
}

type PublishPart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

type PublishFinaliseRequest struct {
	Token       string        `json:"token"`
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	Arch        string        `json:"arch"`
	Channel     string        `json:"channel"`
	Key         string        `json:"key"`
	UploadId    string        `json:"upload_id"`
	Parts       []PublishPart `json:"parts"`
	Size        int64         `json:"size"`
	Sha384      string        `json:"sha384"`
	SnapYaml    string        `json:"snap_yaml"`
	IconPngB64  string        `json:"icon_png_b64,omitempty"`
}

type PublishFinaliseResponse struct {
	Ok bool `json:"ok"`
}
