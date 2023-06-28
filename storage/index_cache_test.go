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

	cache := New(client, "http://localhost", "amd64", log.Default())
	err := cache.Refresh()
	assert.NoError(t, err)

	index, ok := cache.Read("master")
	assert.True(t, ok)
	assert.Equal(t, 1, len(index))
	assert.Equal(t, "app", index["app"].Name)
	assert.Equal(t, "http://localhost/apps/app_123_amd64.snap", index["app"].Download.URL)

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

	cache := New(client, "http://localhost", "amd64", log.Default())
	err := cache.Refresh()
	assert.NoError(t, err)

	cache.Read("test")

}

func TestIndexCache_Find(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"channel1": {
				"app1": &model.Snap{
					Name: "app1",
				},
			},
			"channel2": {
				"app2": &model.Snap{
					Name: "app2",
				},
			},
		},
		arch:   "amd64",
		logger: log.Default(),
	}
	results := cache.Find("channel1", "")
	assert.Equal(t, 1, len(results.Results))
	assert.Equal(t, "app1", results.Results[0].Name)
}
func TestIndexCache_Find_Sorted(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"channel": {
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
		arch:   "amd64",
		logger: log.Default(),
	}
	results := cache.Find("channel", "*")
	assert.Equal(t, 3, len(results.Results))
	assert.Equal(t, "app1", results.Results[0].Name)
	assert.Equal(t, "app2", results.Results[1].Name)
	assert.Equal(t, "app3", results.Results[2].Name)
}

func TestIndexCache_Find_PopulateChannel(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"channel": {
				"": &model.Snap{
					Name: "app",
				},
			},
		},
		arch:   "amd64",
		logger: log.Default(),
	}
	results := cache.Find("channel", "")
	assert.Equal(t, "channel", results.Results[0].Revision.Channel)
}

func TestIndexCache_Info(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"stable": {
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
		arch:   "amd64",
		logger: log.Default(),
	}
	result := cache.Info("app", "amd64")
	assert.Equal(t, "app", result.Name)
	assert.Equal(t, "stable", result.ChannelMap[0].Channel.Name)
}

func TestIndexCache_Info_NotFound(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"stable": {
				"app": &model.Snap{
					Name: "app",
				},
			},
		},
		arch:   "amd64",
		logger: log.Default(),
	}
	result := cache.Info("app1", "amd64")
	assert.Nil(t, result)
}

func TestIndexCache_Info_FirstOneIsASpecial(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"master": {
				"app": &model.Snap{
					Name: "app",
				},
			},
			"stable": {
				"app": &model.Snap{
					Name: "app",
				},
			},
		},
		arch:   "amd64",
		logger: log.Default(),
	}
	result := cache.Info("app", "amd64")
	assert.Equal(t, "app", result.Name)
	assert.Equal(t, "stable", result.ChannelMap[0].Channel.Name)
}

func TestIndexCache_InfoById(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"stable": {
				"app": &model.Snap{
					Name: "app",
				},
			},
		},
		arch:   "amd64",
		logger: log.Default(),
	}
	result, err := cache.InfoById("stable", "app.1", "action", "actionName")
	assert.NoError(t, err)
	assert.Equal(t, "action", result.Result)
	assert.Equal(t, "stable", result.EffectiveChannel)
}

func TestIndexCache_InfoById_NotFound(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"stable": {},
		},
		arch:   "amd64",
		logger: log.Default(),
	}
	result, err := cache.InfoById("stable", "app.1", "action", "actionName")
	assert.NoError(t, err)
	assert.Equal(t, "error", result.Result)
}

func TestIndexCache_InfoById_SnapIdEmpty(t *testing.T) {

	cache := &IndexCache{
		indexByChannel: map[string]map[string]*model.Snap{
			"stable": {},
		},
		arch:   "amd64",
		logger: log.Default(),
	}
	result, err := cache.InfoById("stable", "", "action", "actionName")
	assert.NoError(t, err)
	assert.Equal(t, "error", result.Result)
}
