package model

type PublishPart struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
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
