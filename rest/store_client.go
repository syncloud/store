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
	client   Client
	storeUrl string
	token    string
}

func NewStoreClient(client Client, storeUrl string) (*StoreClient, error) {
	token, ok := os.LookupEnv(SyncloudToken)
	if !ok {
		return nil, fmt.Errorf("env var is not present: %s", SyncloudToken)
	}
	return &StoreClient{
		client:   client,
		token:    token,
		storeUrl: storeUrl,
	}, nil
}

func (c *StoreClient) RefreshCache() error {
	resp, code, err := c.client.Post(
		fmt.Sprintf("%s/syncloud/v1/cache/refresh", c.storeUrl),
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
