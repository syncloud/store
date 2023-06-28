package rest

import (
	"github.com/go-resty/resty/v2"
	"time"
)

type Client interface {
	Get(url string) (string, int, error)
}

type RestyClient struct {
	client *resty.Client
}

func New() *RestyClient {
	client := resty.New()
	client.SetRetryCount(3)
	client.SetRetryWaitTime(5 * time.Second)
	return &RestyClient{
		client: client,
	}
}

func (c *RestyClient) Get(url string) (string, int, error) {
	response, err := c.client.R().Get(url)
	if err != nil {
		return "", 0, err
	}
	return response.String(), response.StatusCode(), err
}
