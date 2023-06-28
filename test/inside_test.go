package test

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"testing"
)

const (
	SnapdSocket = "/var/run/snapd.socket"
)

func TestInside(t *testing.T) {

	client := resty.New()
	transport := http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", SnapdSocket)
		},
	}

	client.SetTransport(&transport).SetScheme("http").SetBaseURL("unix")

	resp, err := client.R().Get("v2/find?name=testapp1")
	assert.NoError(t, err, resp.String())
	assert.Equal(t, 200, resp.StatusCode())
	assert.Contains(t, string(resp.Body()), `"id":"testapp1.3"`)
	assert.Contains(t, string(resp.Body()), `"channel":"stable"`)

	resp, err = client.R().Get("v2/snaps/testapp1")
	assert.NoError(t, err, resp.String())
	assert.Equal(t, 200, resp.StatusCode())
	assert.Contains(t, string(resp.Body()), `"id":"testapp1.3"`)
	assert.Contains(t, string(resp.Body()), `"channel":"stable"`)
}
