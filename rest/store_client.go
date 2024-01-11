package rest

import (
	"fmt"
	"github.com/syncloud/store/model"
	"os"
)

const (
	SyncloudToken = "SYNCLOUD_TOKEN"
)

type StoreClient struct {
	client Client
	token  string
}

func NewStoreClient(client Client) (*StoreClient, error) {
	token, ok := os.LookupEnv(SyncloudToken)
	if !ok {
		return nil, fmt.Errorf("env var is not present: %s", SyncloudToken)
	}
	return &StoreClient{
		client: client,
		token:  token,
	}, nil
}

func (c *StoreClient) RefreshCache() error {
	resp, code, err := c.client.Post(
		"http://api.store.test/syncloud/v1/cache/refresh",
		model.StoreCacheRefreshRequest{Token: c.token},
	)
	if err != nil {
		return err
	}

	if code != 200 {
		return fmt.Errorf("refresh error: %s, code: %d", resp, code)
	}
	return nil
}
