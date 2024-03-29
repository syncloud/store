package storage

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/store/log"
	"github.com/syncloud/store/model"
	"testing"
)

type Response struct {
	body string
	code int
	err  error
}

func OK(body string) Response {
	return Response{body: body, code: 200}
}

type ClientStub struct {
	response map[string]Response
}

func (c *ClientStub) Post(url string, body interface{}) (string, int, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ClientStub) Get(url string) (string, int, error) {
	fmt.Println(url)
	response, ok := c.response[url]
	if !ok {
		return "", 404, nil
	}
	return response.body, response.code, response.err
}

func TestIndexCache_Refresh(t *testing.T) {

	client := &ClientStub{
		response: map[string]Response{
			"http://localhost/releases/master/index-v2": OK(`
{
  "apps" : [
    {
      "name" : "App",
      "id" : "app",
      "required" : false,
      "ui": true
    }
  ]
}
`),
			"http://localhost/releases/master/app.amd64.version": OK("123"),
			"http://localhost/apps/app_123_amd64.snap.size":      OK("1"),
			"http://localhost/apps/app_123_amd64.snap.sha384":    OK(base64.RawURLEncoding.EncodeToString([]byte("sha384"))),
		},
	}

	cache := New(client, "http://localhost", log.Default())
	err := cache.Refresh()
	assert.NoError(t, err)

	index, ok := cache.Read("master")
	assert.True(t, ok)
	assert.Equal(t, 1, len(index))
	assert.Equal(t, "app", index["amd64"]["app"].Name)
	assert.Equal(t, "http://localhost/apps/app_123_amd64.snap", index["amd64"]["app"].Download.URL)

}

func TestIndexCache_Refresh_EmptySize(t *testing.T) {

	client := &ClientStub{
		response: map[string]Response{
			"http://localhost/releases/master/index-v2": OK(`
{
  "apps" : [
    {
      "name" : "Platform",
      "id" : "platform",
      "required" : true,
      "ui": false
    }
  ]
}
`),
			"http://localhost/releases/master/platform.amd64.version": OK("123"),
			"http://localhost/apps/platform__amd64.snap.size":         OK(""),
		},
	}

	cache := New(client, "http://localhost", log.Default())
	err := cache.Refresh()
	assert.NoError(t, err)

	cache.Read("test")

}

func TestIndexCache_Find(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"channel1": {
				"amd64": {
					"app1": &model.Snap{
						Name: "app1",
					},
				},
			},
			"channel2": {
				"amd64": {
					"app2": &model.Snap{
						Name: "app2",
					},
				},
			},
		},
		logger: log.Default(),
	}
	results := cache.Find("channel1", "", "amd64")
	assert.Equal(t, 1, len(results.Results))
	assert.Equal(t, "app1", results.Results[0].Name)
}
func TestIndexCache_Find_Sorted(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"channel": {
				"amd64": {
					"app1": &model.Snap{
						Name: "app1",
					},
					"app3": &model.Snap{
						Name: "app3",
					},
					"app2": &model.Snap{
						Name: "app2",
					},
				},
			},
		},
		logger: log.Default(),
	}
	results := cache.Find("channel", "*", "amd64")
	assert.Equal(t, 3, len(results.Results))
	assert.Equal(t, "app1", results.Results[0].Name)
	assert.Equal(t, "app2", results.Results[1].Name)
	assert.Equal(t, "app3", results.Results[2].Name)
}

func TestIndexCache_Find_PopulateChannel(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"channel": {
				"amd64": {
					"": &model.Snap{
						Name: "app",
					},
				},
			},
		},
		logger: log.Default(),
	}
	results := cache.Find("channel", "", "amd64")
	assert.Equal(t, "channel", results.Results[0].Revision.Channel)
}

func TestIndexCache_Info(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"stable": {
				"amd64": {
					"app": &model.Snap{
						SnapID:        "snap-id",
						Name:          "app",
						Summary:       "summary",
						Version:       "1",
						Type:          "app",
						Architectures: nil,
						Revision:      2,
						Download: model.StoreSnapDownload{
							Sha3_384: "sha",
							Size:     1,
							URL:      "http://donload",
							Deltas:   nil,
						},
						Media: nil,
					},
				},
			},
		},
		logger: log.Default(),
	}
	result := cache.Info("app", "amd64")
	assert.Equal(t, "app", result.Name)
	assert.Equal(t, "stable", result.ChannelMap[0].Channel.Name)
}

func TestIndexCache_Info_NotFound(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"amd64": {
				"stable": {
					"app": &model.Snap{
						Name: "app",
					},
				},
			},
		},
		logger: log.Default(),
	}
	result := cache.Info("app1", "amd64")
	assert.Nil(t, result)
}

func TestIndexCache_Info_FirstOneIsASpecial(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"master": {
				"amd64": {
					"app": &model.Snap{
						Name: "app",
					},
				},
			},
			"stable": {
				"amd64": {
					"app": &model.Snap{
						Name: "app",
					},
				},
			},
		},
		logger: log.Default(),
	}
	result := cache.Info("app", "amd64")
	assert.Equal(t, "app", result.Name)
	assert.Equal(t, "stable", result.ChannelMap[0].Channel.Name)
}

func TestIndexCache_Info_PreferStable(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"master": {
				"amd64": {
					"app": &model.Snap{
						Name:     "app",
						Revision: 2,
					},
				},
			},
			"stable": {
				"amd64": {
					"app": &model.Snap{
						Name:     "app",
						Revision: 1,
					},
				},
			},
		},
		logger: log.Default(),
	}
	result := cache.Info("app", "amd64")
	assert.Equal(t, "app", result.Name)
	assert.Equal(t, 1, result.Snap.Revision)
	assert.Equal(t, "stable", result.ChannelMap[0].Channel.Name)
}

func TestIndexCache_InfoById(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"stable": {
				"amd64": {
					"app": &model.Snap{
						SnapID: "app.1",
						Name:   "app",
					},
				},
			},
		},
		logger: log.Default(),
	}
	result, err := cache.InfoById("stable", "app.1", "action", "actionName", "amd64")
	assert.NoError(t, err)
	assert.Equal(t, "action", result.Result)
	assert.Equal(t, "stable", result.EffectiveChannel)
	assert.Equal(t, "app.1", result.SnapID)
	//assert.Equal(t, "app.1", result.Snap.SnapID)
}

/*
func TestIndexCache_InfoById_OldSnapId_DefaultArch(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"stable": {
				"amd64": {
					"app": &model.Snap{
						SnapID: "app.1.arm64",
						Name:   "app",
					},
				},
			},
		},
		logger: log.Default(),
	}
	result, err := cache.InfoById("stable", "app.1", "action", "actionName")
	assert.NoError(t, err)
	assert.Equal(t, "action", result.Result)
	assert.Equal(t, "stable", result.EffectiveChannel)
	assert.Equal(t, "app.1", result.SnapID)
	assert.Equal(t, "app.1", result.Snap.SnapID)

}
*/
func TestIndexCache_InfoById_NotFound(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"stable": {},
		},
		logger: log.Default(),
	}
	result, err := cache.InfoById("stable", "app.1", "action", "actionName", "amd64")
	assert.NoError(t, err)
	assert.Equal(t, "error", result.Result)
}

func TestIndexCache_InfoById_SnapIdEmpty(t *testing.T) {

	cache := &CachedIndex{
		cache: Cache{
			"stable": {},
		},
		logger: log.Default(),
	}
	result, err := cache.InfoById("stable", "", "action", "actionName", "amd64")
	assert.NoError(t, err)
	assert.Equal(t, "error", result.Result)
}
