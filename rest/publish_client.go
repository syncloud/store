package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/syncloud/store/model"
)

type PublishClient struct {
	storeUrl string
	token    string
	http     *http.Client
}

func NewPublishClient(storeUrl string) (*PublishClient, error) {
	token, ok := os.LookupEnv(SyncloudToken)
	if !ok {
		return nil, fmt.Errorf("env var is not present: %s", SyncloudToken)
	}
	return &PublishClient{
		storeUrl: storeUrl,
		token:    token,
		http:     &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (c *PublishClient) postJSON(path string, in, out interface{}) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	resp, err := c.http.Post(c.storeUrl+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s -> %d: %s", path, resp.StatusCode, string(raw))
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func (c *PublishClient) Init(name, version, arch, channel string, size int64, sha384 string, partSize int64) (*model.PublishInitResponse, error) {
	req := model.PublishInitRequest{
		Token: c.token, Name: name, Version: version, Arch: arch, Channel: channel,
		Size: size, Sha384: sha384, PartSize: partSize,
	}
	var resp model.PublishInitResponse
	if err := c.postJSON("/syncloud/v1/publish/init", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *PublishClient) PartUrl(key, uploadId string, partNumber int) (string, error) {
	req := model.PublishPartUrlRequest{
		Token: c.token, Key: key, UploadId: uploadId, PartNumber: partNumber,
	}
	var resp model.PublishPartUrlResponse
	if err := c.postJSON("/syncloud/v1/publish/part-url", req, &resp); err != nil {
		return "", err
	}
	return resp.Url, nil
}

func (c *PublishClient) Finalise(req model.PublishFinaliseRequest) error {
	req.Token = c.token
	return c.postJSON("/syncloud/v1/publish/finalise", req, nil)
}
