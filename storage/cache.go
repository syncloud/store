package storage

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/syncloud/store/model"
	"github.com/syncloud/store/rest"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type AppLister interface {
	ListAppIds(channel string) ([]string, error)
}

type Cache struct {
	snapCache SnapCache
	appCache  AppCache
	lock      sync.RWMutex
	client    rest.Client
	lister    AppLister
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

func New(client rest.Client, lister AppLister, baseUrl string, logger *zap.Logger) *Cache {
	return &Cache{
		client:    client,
		lister:    lister,
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
		snaps, apps, err := i.loadChannel(channel)
		if err != nil {
			return err
		}
		if apps == nil {
			i.logger.Warn("apps.json missing for channel", zap.String("channel", channel))
			continue
		}
		i.WriteIndex(channel, snaps, apps)
	}
	i.logger.Info("refresh cache finished")
	return nil
}

func (i *Cache) loadChannel(channel string) (SnapByArch, AppByName, error) {
	ids, err := i.lister.ListAppIds(channel)
	if err != nil {
		return nil, nil, err
	}
	if len(ids) == 0 {
		return nil, nil, nil
	}
	apps := make(AppByName)
	snaps := make(SnapByArch)
	for _, arch := range AvailableArchitectures {
		snaps[arch] = make(SnapByName)
	}
	for _, appId := range ids {
		app, ferr := i.fetchAppMetadata(channel, appId)
		if ferr != nil {
			i.logger.Warn("snap.yaml fetch failed", zap.String("app", appId), zap.Error(ferr))
			continue
		}
		if app == nil {
			i.logger.Info("snap.yaml missing", zap.String("channel", channel), zap.String("app", appId))
			continue
		}
		apps[appId] = app
		for _, arch := range AvailableArchitectures {
			snap, serr := i.resolveSnap(channel, app, arch)
			if serr != nil {
				i.logger.Warn("snap resolve failed",
					zap.String("app", appId), zap.String("arch", arch), zap.Error(serr))
				continue
			}
			if snap == nil {
				continue
			}
			snaps[arch][appId] = snap
		}
	}
	return snaps, apps, nil
}

func (i *Cache) fetchAppMetadata(channel, appId string) (*model.App, error) {
	url := fmt.Sprintf("%s/v2/apps/%s/%s/snap.yaml", i.baseUrl, channel, appId)
	resp, code, err := i.client.Get(url)
	if err != nil {
		return nil, err
	}
	if code == 404 {
		return nil, nil
	}
	if code != 200 {
		return nil, fmt.Errorf("snap.yaml %s -> %d", url, code)
	}
	m, err := model.ParseSnapMeta([]byte(resp))
	if err != nil {
		return nil, err
	}
	name := m.Name
	if name == "" {
		name = appId
	}
	return &model.App{
		Name:        name,
		Summary:     m.Summary,
		Description: m.Description,
		Type:        m.Type,
		Enabled:     true,
	}, nil
}

func (i *Cache) resolveSnap(channel string, app *model.App, arch string) (*model.Snap, error) {
	versionUrl := fmt.Sprintf("%s/releases/%s/%s.%s.version", i.baseUrl, channel, app.Name, arch)
	resp, code, err := i.client.Get(versionUrl)
	if err != nil {
		return nil, err
	}
	if code == 404 {
		return nil, nil
	}
	if code != 200 {
		return nil, fmt.Errorf("%s -> %d", versionUrl, code)
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
	sha384, err := base64.RawURLEncoding.DecodeString(resp)
	if err != nil {
		return nil, err
	}
	return app.ToInfo(version, size, fmt.Sprintf("%x", sha384), downloadUrl, arch, i.iconUrlAbsolute(channel, app.Name))
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
		if app.Type == "base" {
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
			Name:    app.Summary,
			Summary: app.Description,
			IconUrl: i.iconUrl(channel, name),
			Version: version,
			SnapID:  snapId,
		})
	}
	slices.SortFunc(results, func(a, b *model.UIApp) bool {
		return a.Name < b.Name
	})
	return results
}

func (i *Cache) iconUrl(channel, appId string) string {
	if appId == "" {
		return ""
	}
	return fmt.Sprintf("/api/ui/v1/icons/%s/%s", channel, appId)
}

func (i *Cache) iconUrlAbsolute(channel, appId string) string {
	if appId == "" {
		return ""
	}
	return fmt.Sprintf("%s/v2/apps/%s/%s/icon.png", i.baseUrl, channel, appId)
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
