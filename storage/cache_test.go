package storage

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/store/log"
	"github.com/syncloud/store/model"
)

type Response struct {
	body string
	code int
	err  error
}

func OK(body string) Response { return Response{body: body, code: 200} }

type ClientStub struct {
	response map[string]Response
}

func (c *ClientStub) Post(url string, body interface{}) (string, int, error) {
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

type ListerStub struct {
	apps map[string][]string
}

func (l *ListerStub) ListAppIds(channel string) ([]string, error) {
	return l.apps[channel], nil
}

func TestCache_Refresh_LoadsFromV2(t *testing.T) {
	sha := base64.RawURLEncoding.EncodeToString([]byte("sha384"))
	client := &ClientStub{
		response: map[string]Response{
			"http://localhost/v2/apps/master/app/snap.yaml":      OK("name: app\nsummary: My App\ndescription: D\n"),
			"http://localhost/releases/master/app.amd64.version": OK("123"),
			"http://localhost/apps/app_123_amd64.snap.size":      OK("1"),
			"http://localhost/apps/app_123_amd64.snap.sha384":    OK(sha),
		},
	}
	lister := &ListerStub{apps: map[string][]string{"master": {"app"}}}
	cache := New(client, lister, "http://localhost", log.Default())
	assert.NoError(t, cache.Refresh())

	index, ok := cache.Read("master")
	assert.True(t, ok)
	assert.Equal(t, "app", index["amd64"]["app"].Name)
	assert.Equal(t, "My App", index["amd64"]["app"].Summary)
	assert.Equal(t, "http://localhost/apps/app_123_amd64.snap", index["amd64"]["app"].Download.URL)
	assert.Equal(t, "app", index["amd64"]["app"].Type)
	assert.Equal(t, "/api/ui/v1/icons/master/app", index["amd64"]["app"].Media[0].URL)
}

func TestCache_Refresh_TypeBaseFromSnapYaml(t *testing.T) {
	sha := base64.RawURLEncoding.EncodeToString([]byte("sha384"))
	client := &ClientStub{
		response: map[string]Response{
			"http://localhost/v2/apps/master/platform/snap.yaml":      OK("name: platform\nsummary: Platform\ndescription: P\ntype: base\n"),
			"http://localhost/releases/master/platform.amd64.version": OK("1"),
			"http://localhost/apps/platform_1_amd64.snap.size":        OK("1"),
			"http://localhost/apps/platform_1_amd64.snap.sha384":      OK(sha),
		},
	}
	lister := &ListerStub{apps: map[string][]string{"master": {"platform"}}}
	cache := New(client, lister, "http://localhost", log.Default())
	assert.NoError(t, cache.Refresh())

	index, _ := cache.Read("master")
	assert.Equal(t, "base", index["amd64"]["platform"].Type)
}

func TestCache_Refresh_EmptyChannelIsSkipped(t *testing.T) {
	client := &ClientStub{response: map[string]Response{}}
	lister := &ListerStub{apps: map[string][]string{}}
	cache := New(client, lister, "http://localhost", log.Default())
	assert.NoError(t, cache.Refresh())
	_, ok := cache.Read("master")
	assert.False(t, ok)
}

func TestCache_Find(t *testing.T) {
	cache := &Cache{
		snapCache: SnapCache{
			"channel1": {"amd64": {"app1": &model.Snap{Name: "app1"}}},
			"channel2": {"amd64": {"app2": &model.Snap{Name: "app2"}}},
		},
		logger: log.Default(),
	}
	results := cache.Find("channel1", "", "amd64")
	assert.Equal(t, 1, len(results.Results))
	assert.Equal(t, "app1", results.Results[0].Name)
}

func TestCache_Find_Sorted(t *testing.T) {
	cache := &Cache{
		snapCache: SnapCache{"channel": {"amd64": {
			"app1": &model.Snap{Name: "app1"},
			"app3": &model.Snap{Name: "app3"},
			"app2": &model.Snap{Name: "app2"},
		}}},
		logger: log.Default(),
	}
	results := cache.Find("channel", "*", "amd64")
	assert.Equal(t, "app1", results.Results[0].Name)
	assert.Equal(t, "app2", results.Results[1].Name)
	assert.Equal(t, "app3", results.Results[2].Name)
}

func TestCache_Info_PreferStable(t *testing.T) {
	cache := &Cache{
		snapCache: SnapCache{
			"master": {"amd64": {"app": &model.Snap{Name: "app", Revision: 2}}},
			"stable": {"amd64": {"app": &model.Snap{Name: "app", Revision: 1}}},
		},
		logger: log.Default(),
	}
	result := cache.Info("app", "amd64")
	assert.Equal(t, 1, result.Snap.Revision)
	assert.Equal(t, "stable", result.ChannelMap[0].Channel.Name)
}

func TestCache_InfoById_NotFound(t *testing.T) {
	cache := &Cache{snapCache: SnapCache{"stable": {}}, logger: log.Default()}
	result, err := cache.InfoById("stable", "app.1", "action", "actionName", "amd64")
	assert.NoError(t, err)
	assert.Equal(t, "error", result.Result)
}

func TestCache_UIApps_EmptyChannel(t *testing.T) {
	cache := New(nil, nil, "http://localhost", log.Default())
	apps := cache.UIApps("stable")
	assert.NotNil(t, apps)
	assert.Equal(t, 0, len(apps))
}

func TestCache_UIApps_BaseHidden(t *testing.T) {
	cache := &Cache{
		baseUrl: "http://apps.syncloud.org",
		snapCache: SnapCache{"stable": {"amd64": {
			"platform":  &model.Snap{Version: "1", SnapID: "platform.1"},
			"bitwarden": &model.Snap{Version: "2", SnapID: "bitwarden.2"},
		}}},
		appCache: AppCache{"stable": {
			"platform":  &model.App{Name: "platform", Summary: "Platform", Type: "base"},
			"bitwarden": &model.App{Name: "bitwarden", Summary: "Bitwarden"},
		}},
		logger: log.Default(),
	}
	apps := cache.UIApps("stable")
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, "Bitwarden", apps[0].Name)
}

func TestCache_UIApps_IconUrlIsChannelScoped(t *testing.T) {
	cache := &Cache{
		baseUrl: "http://apps.syncloud.org",
		snapCache: SnapCache{"stable": {"amd64": {
			"nextcloud": &model.Snap{Version: "1", SnapID: "nextcloud.1"},
		}}},
		appCache: AppCache{"stable": {
			"nextcloud": &model.App{Name: "nextcloud", Summary: "Nextcloud"},
		}},
		logger: log.Default(),
	}
	apps := cache.UIApps("stable")
	assert.Equal(t, "/api/ui/v1/icons/stable/nextcloud", apps[0].IconUrl)
	assert.False(t, strings.HasPrefix(apps[0].IconUrl, "http"),
		"icon URL must be same-origin so it inherits HTTPS")
}
