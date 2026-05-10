package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/syncloud/store/model"
	"github.com/syncloud/store/rest"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"strconv"
	"sync"
	"time"
)

type Cache struct {
	snapCache SnapCache
	appCache  AppCache
	lock      sync.RWMutex
	client    rest.Client
	baseUrl   string
	logger    *zap.Logger
}

type SnapByName map[string]*model.Snap
type SnapByArch map[string]SnapByName
type SnapByChannel map[string]SnapByArch
type SnapCache SnapByChannel

type AppByName map[string]*model.App
type AppCache map[string]AppByName

const (
	ChannelMaster = "master"
	ChannelRc     = "rc"
	ChannelStable = "stable"
	ArchAmd64     = "amd64"
	ArchArm64     = "arm64"
	ArchArm32     = "armhf"
)

var AvailableChannels = []string{ChannelStable, ChannelMaster, ChannelRc}
var AvailableArchitectures = []string{ArchAmd64, ArchArm64, ArchArm32}

func New(client rest.Client, baseUrl string, logger *zap.Logger) *Cache {
	return &Cache{
		client:    client,
		baseUrl:   baseUrl,
		logger:    logger,
		snapCache: make(SnapCache),
		appCache:  make(AppCache),
	}
}

func (i *Cache) InfoById(channelFull, snapId, action, actionName, arch string) (*model.StoreResult, error) {
	channel := parseChannel(channelFull)
	snapName := actionName
	if snapId != "" {
		id := model.SnapId(snapId)
		snapName = id.Name()
		//arch = id.Arch()
	}
	architectures, ok := i.Read(channel)
	if !ok {
		return nil, fmt.Errorf("no channel: %s in the index", channel)
	}
	i.logger.Info("lookup", zap.String("app", snapName))
	app, ok := architectures[arch][snapName]
	if !ok {
		return &model.StoreResult{
			Result: "error",
			Name:   snapName,
			Error: &model.StoreError{
				Code:    "name-not-found",
				Message: "name-not-found",
			},
			SnapID: snapId,
		}, nil
	}
	//app.SnapID = snapId
	return &model.StoreResult{
		Result:           action,
		Snap:             app,
		SnapID:           snapId,
		EffectiveChannel: channel,
	}, nil
}

func parseChannel(channel string) string {
	switch channel {
	case "master":
		return "master"
	case "master/stable":
		return "master"
	case "rc/stable":
		return "rc"
	case "rc":
		return "rc"
	case "latest/stable":
		return "stable"
	default:
		return "stable"
	}
}

func (i *Cache) Info(name string, architecture string) *model.StoreInfo {
	found := false
	info := &model.StoreInfo{}
	var stableApp *model.Snap
	for _, channel := range AvailableChannels {
		architectures, ok := i.Read(channel)
		if !ok {
			i.logger.Warn("no channel in the index", zap.String("channel", channel))
			continue
		}
		app, ok := architectures[architecture][name]
		if !ok {
			i.logger.Info("app is not found", zap.String("channel", channel), zap.String("name", name))
			continue
		}
		info.Name = app.Name
		info.SnapID = app.SnapID
		channelInfo := &model.StoreInfoChannelSnap{
			Snap: *app,
			Channel: model.StoreInfoChannel{
				Name:         channel,
				Architecture: architecture,
				Risk:         channel,
				Track:        "",
			},
		}
		info.ChannelMap = append(info.ChannelMap, channelInfo)
		info.Snap = *app
		if channel == "stable" {
			stableApp = app
		}
		found = true
	}
	if found {
		if stableApp != nil {
			info.Snap = *stableApp
			info.Name = stableApp.Name
			info.SnapID = stableApp.SnapID
		}
		return info
	}
	return nil
}

func (i *Cache) Find(channel string, query string, architecture string) *model.SearchResults {
	architectures, ok := i.Read(channel)
	if !ok {
		i.logger.Warn("no channel in the index", zap.String("channel", channel))
		return nil
	}
	results := &model.SearchResults{}
	for name, app := range architectures[architecture] {
		if query == "*" || query == "" || query == name {
			result := &model.SearchResult{
				Revision: model.SearchRevision{Channel: channel},
				Snap:     *app,
				Name:     app.Name,
				SnapID:   app.SnapID,
			}
			results.Results = append(results.Results, result)
		}
	}
	slices.SortFunc(results.Results, func(a, b *model.SearchResult) bool {
		return a.Name < b.Name
	})
	return results
}

func (i *Cache) Refresh() error {
	i.logger.Info("refresh cache")
	for _, channel := range AvailableChannels {
		snaps, apps, err := i.downloadIndex(channel)
		if err != nil {
			return err
		}
		if snaps == nil {
			i.logger.Warn("index not found", zap.String("channel", channel))
			continue
		}
		i.WriteIndex(channel, snaps, apps)
	}
	i.logger.Info("refresh cache finished")
	return nil
}

func (i *Cache) downloadIndex(channel string) (SnapByArch, AppByName, error) {
	resp, code, err := i.client.Get(fmt.Sprintf("%s/releases/%s/index-v2", i.baseUrl, channel))
	if err != nil {
		return nil, nil, err
	}

	if code != 200 {
		return nil, nil, nil
	}

	index, err := i.parseIndex(resp)
	if err != nil {
		return nil, nil, err
	}
	snaps := make(SnapByArch)
	for _, indexApp := range index {
		for _, arch := range AvailableArchitectures {
			snap, err := i.downloadAppInfo(indexApp, channel, arch)
			if err != nil {
				return nil, nil, err
			}
			if snap == nil {
				i.logger.Info("not found", zap.String("app", indexApp.Name), zap.String("channel", channel))
				continue
			}
			_, found := snaps[arch]
			if !found {
				snaps[arch] = make(SnapByName)
			}
			snaps[arch][indexApp.Name] = snap
		}
	}

	return snaps, index, nil
}

func (i *Cache) downloadAppInfo(app *model.App, channel string, arch string) (*model.Snap, error) {
	versionUrl := fmt.Sprintf("%s/releases/%s/%s.%s.version", i.baseUrl, channel, app.Name, arch)
	i.logger.Info("version", zap.String("url", versionUrl))
	resp, code, err := i.client.Get(versionUrl)
	if err != nil {
		return nil, err
	}
	if code == 404 {
		return nil, nil
	}
	version := resp
	downloadUrl := fmt.Sprintf("%s/apps/%s_%s_%s.snap", i.baseUrl, app.Name, version, arch)

	resp, _, err = i.client.Get(fmt.Sprintf("%s/apps/%s_%s_%s.snap.size", i.baseUrl, app.Name, version, arch))
	if err != nil {
		return nil, err
	}
	size, err := strconv.ParseInt(resp, 10, 0)
	if err != nil {
		if channel == "stable" {
			i.logger.Warn("not valid size", zap.String("app", app.Name), zap.Error(err))
		}
		return nil, nil
	}

	resp, _, err = i.client.Get(fmt.Sprintf("%s/apps/%s_%s_%s.snap.sha384", i.baseUrl, app.Name, version, arch))
	if err != nil {
		return nil, err
	}
	sha384Encoded := resp
	sha384, err := base64.RawURLEncoding.DecodeString(sha384Encoded)
	if err != nil {
		return nil, err
	}
	return app.ToInfo(version, size, fmt.Sprintf("%x", sha384), downloadUrl, arch)
}

func (i *Cache) WriteIndex(channel string, snaps SnapByArch, apps AppByName) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.snapCache[channel] = snaps
	i.appCache[channel] = apps
}

func (i *Cache) UIApps(channel string) []*model.UIApp {
	i.lock.RLock()
	defer i.lock.RUnlock()

	apps := i.appCache[channel]
	archs := i.snapCache[channel]
	results := make([]*model.UIApp, 0, len(apps))
	for name, app := range apps {
		if app.Required {
			continue
		}
		var version, snapId string
		for _, arch := range AvailableArchitectures {
			if snap, ok := archs[arch][name]; ok {
				version = snap.Version
				snapId = snap.SnapID
				break
			}
		}
		if version == "" {
			continue
		}
		results = append(results, &model.UIApp{
			Name:        app.Summary,
			Summary:     app.Summary,
			Description: app.Description,
			IconUrl:     i.iconUrl(channel, app.Icon),
			Version:     version,
			SnapID:      snapId,
		})
	}
	slices.SortFunc(results, func(a, b *model.UIApp) bool {
		return a.Name < b.Name
	})
	return results
}

func (i *Cache) iconUrl(channel, icon string) string {
	if icon == "" {
		return ""
	}
	return fmt.Sprintf("%s/releases/%s/images/%s", i.baseUrl, channel, icon)
}

func (i *Cache) Read(channel string) (SnapByArch, bool) {
	i.lock.RLock()
	defer i.lock.RUnlock()
	snaps, ok := i.snapCache[channel]
	return snaps, ok
}

func (i *Cache) Start() error {
	go func() {
		if err := i.Refresh(); err != nil {
			i.logger.Error("initial refresh failed", zap.Error(err))
		}
		for range time.Tick(time.Minute * 60) {
			if err := i.Refresh(); err != nil {
				i.logger.Error("error", zap.Error(err))
			}
		}
	}()
	return nil
}

func (i *Cache) parseIndex(resp string) (map[string]*model.App, error) {
	var index model.Index
	err := json.Unmarshal([]byte(resp), &index)
	if err != nil {
		i.logger.Error("cannot parse index response", zap.Error(err))
		return nil, err
	}

	apps := make(map[string]*model.App)

	for ind := range index.Apps {
		app := &model.App{
			Enabled: true,
		}
		err := json.Unmarshal(index.Apps[ind], app)
		if err != nil {
			return nil, err
		}
		if !app.Enabled {
			continue
		}
		i.logger.Info("index", zap.String("app", app.Name))
		apps[app.Name] = app

	}

	return apps, nil

}
