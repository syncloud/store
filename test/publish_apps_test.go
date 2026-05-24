package test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestPublishedApps(t *testing.T) {
	type uiApp struct {
		SnapID  string `json:"snapId"`
		Version string `json:"version"`
	}

	resp, err := resty.New().R().Get("http://api.store/api/ui/v1/apps?channel=stable")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode(), string(resp.Body()))

	var apps []uiApp
	assert.NoError(t, json.Unmarshal(resp.Body(), &apps), string(resp.Body()))

	got := map[string]uiApp{}
	for _, a := range apps {
		got[strings.SplitN(a.SnapID, ".", 2)[0]] = a
	}
	a1, ok1 := got["testapp1"]
	assert.True(t, ok1, "testapp1 not in /api/ui/v1/apps?channel=stable: %s", string(resp.Body()))
	a2, ok2 := got["testapp2"]
	assert.True(t, ok2, "testapp2 not in /api/ui/v1/apps?channel=stable: %s", string(resp.Body()))

	assert.Equal(t, "3", a1.Version, "testapp1 should be at version 3 after publish")
	assert.Equal(t, "2", a2.Version, "testapp2 should be at version 2 after publish")
}
