package api

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/syncloud/store/model"
	"github.com/syncloud/store/rest"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

type Signer interface {
	AccountKey(key string) (string, error)
	SnapDeclaration(series, snapId string) (string, error)
	SnapRevision(key, revision string) (string, error)
}

type ApiCache interface {
	Refresh() error
	Find(channel string, query string, architecture string) *model.SearchResults
	Info(name, arch string) *model.StoreInfo
	InfoById(channel, snapId, action, actionName, arch string) (*model.StoreResult, error)
}

type Popularity interface {
	Record(snap string)
	Count(snap string) int
}

type SyncloudStore struct {
	client     rest.Client
	echo       *echo.Echo
	address    string
	apiCache   ApiCache
	signer     Signer
	token      string
	logger     *zap.Logger
	web        *Web
	iconProxy  *IconProxy
	popularity Popularity
	metrics    *SnapdMetrics
	publish    *Publish
}

func NewSyncloudStore(
	address string,
	apiCache ApiCache,
	client rest.Client,
	signer Signer,
	token string,
	web *Web,
	iconProxy *IconProxy,
	popularity Popularity,
	metrics *SnapdMetrics,
	publish *Publish,
	logger *zap.Logger,
) *SyncloudStore {
	return &SyncloudStore{
		client:     client,
		echo:       echo.New(),
		signer:     signer,
		apiCache:   apiCache,
		address:    address,
		token:      token,
		web:        web,
		iconProxy:  iconProxy,
		popularity: popularity,
		metrics:    metrics,
		publish:    publish,
		logger:     logger,
	}
}

func (s *SyncloudStore) Start() <-chan error {

	s.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	s.echo.Use(middleware.Recover())

	s.echo.GET("/api/v1/snaps/sections", s.Sections)
	s.echo.GET("/api/v1/snaps/names", s.Names)
	s.echo.POST("/v2/snaps/refresh", s.Refresh)
	s.echo.GET("/v2/assertions/snap-revision/:key", s.SnapRevision)
	s.echo.GET("/v2/assertions/snap-declaration/:series/:snap-id", s.SnapDeclaration)
	s.echo.GET("/v2/assertions/account-key/:key", s.AccountKey)
	s.echo.GET("/v2/snaps/find", s.Find)
	s.echo.GET("/v2/snaps/info/:name", s.Info)
	s.echo.POST("/syncloud/v1/cache/refresh", s.SyncloudCacheRefresh)
	if s.publish != nil {
		s.echo.POST("/syncloud/v1/publish/init", s.publish.Init)
		s.echo.POST("/syncloud/v1/publish/part-url", s.publish.PartUrl)
		s.echo.POST("/syncloud/v1/publish/finalise", s.publish.Finalise)
	}
	s.echo.GET("/api/ui/v1/apps", s.web.Apps)
	s.echo.GET("/api/ui/v1/version", s.web.Version)
	s.echo.GET("/api/ui/v1/icons/*", echo.WrapHandler(s.iconProxy))
	s.echo.GET("/*", echo.WrapHandler(http.HandlerFunc(s.web.Serve)))

	s.logger.Info("listening on", zap.String("address", s.address))
	errs := make(chan error, 1)
	go func() {
		if s.IsUnixSocket() {
			_ = os.RemoveAll(s.address)
			l, err := net.Listen("unix", s.address)
			if err != nil {
				errs <- err
				return
			}
			if err := os.Chmod(s.address, 0777); err != nil {
				errs <- err
				return
			}
			s.echo.Listener = l
			errs <- s.echo.Start("")
			return
		}
		errs <- s.echo.Start(s.address)
	}()
	return errs
}

func (s *SyncloudStore) IsUnixSocket() bool {
	return strings.HasPrefix(s.address, "/")
}

func (s *SyncloudStore) Sections(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/hal+json")
	return c.String(http.StatusOK, `{
  "_embedded": {
    "clickindex:sections": [
      {
        "name": "apps"
      }
    ]
  }
}
`)
}

func (s *SyncloudStore) Names(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/hal+json")
	return c.String(http.StatusOK, `
{
  "_embedded": {
    "clickindex:package": [
      
    ]
  }
}`)
}

func (s *SyncloudStore) Refresh(c echo.Context) error {
	req, err := io.ReadAll(c.Request().Body)
	if err != nil {
		c.Error(err)
		return nil
	}
	arch := c.Request().Header.Get("Syncloud-Architecture")
	s.logger.Info("refresh",
		zap.String("arch", arch),
		zap.String("remote_addr", c.RealIP()),
		zap.String("body", string(req)),
	)

	var request model.SnapActionRequest
	err = json.Unmarshal(req, &request)
	if err != nil {
		c.Error(err)
		return nil
	}
	result := &model.StoreResults{}
	for _, action := range request.Actions {
		if action.Action == "fetch-assertions" {
			info := &model.StoreResult{
				Result: action.Action,
				Key:    action.Key,
			}
			result.Results = append(result.Results, info)
			s.metrics.Record("", action.Action, arch, http.StatusOK)
			continue
		}
		info, err := s.apiCache.InfoById(action.Channel, action.SnapID, action.Action, action.Name, arch)
		if err != nil {
			return err
		}
		info.InstanceKey = action.InstanceKey
		result.Results = append(result.Results, info)
		snap := snapName(action)
		s.metrics.Record(snap, action.Action, arch, http.StatusOK)
		if action.Action == "refresh" {
			s.popularity.Record(snap)
		}
	}
	return c.JSON(http.StatusOK, result)
}

func snapName(action *model.SnapAction) string {
	if action.Name != "" {
		return action.Name
	}
	if action.SnapID != "" {
		return model.SnapId(action.SnapID).Name()
	}
	return ""
}

func (s *SyncloudStore) Info(c echo.Context) error {
	name := c.Param("name")
	arch := c.QueryParam("architecture")
	result := s.apiCache.Info(name, arch)
	if result == nil {
		s.metrics.Record(name, "info", arch, http.StatusNotFound)
		return c.String(http.StatusNotFound, "not found")
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	s.metrics.Record(name, "info", arch, http.StatusOK)
	return c.JSON(http.StatusOK, result)
}

func (s *SyncloudStore) Find(c echo.Context) error {
	channel := c.QueryParam("channel")
	query := c.QueryParam("q")
	architecture := c.QueryParam("architecture")
	s.logger.Info("find",
		zap.String("channel", channel),
		zap.String("query", query),
		zap.String("architecture", architecture),
	)

	if channel == "" {
		channel = "stable"
	}
	results := s.apiCache.Find(channel, query, architecture)
	if results == nil {
		s.metrics.Record("", "find", architecture, http.StatusInternalServerError)
		c.Error(fmt.Errorf("no channel: %s in the index", channel))
		return nil
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	s.metrics.Record("", "find", architecture, http.StatusOK)
	return c.JSON(http.StatusOK, results)
}

func (s *SyncloudStore) AccountKey(c echo.Context) error {
	content, err := s.signer.AccountKey(c.Param("key"))
	if err != nil {
		c.Error(err)
		return nil
	}
	return c.String(http.StatusOK, content)
}

func (s *SyncloudStore) SnapDeclaration(c echo.Context) error {
	content, err := s.signer.SnapDeclaration(c.Param("series"), c.Param("snap-id"))
	if err != nil {
		c.Error(err)
		return nil
	}
	return c.String(http.StatusOK, content)
}

func (s *SyncloudStore) SnapRevision(c echo.Context) error {
	key := c.Param("key")
	s.logger.Info("snap revision", zap.String("key", key))

	revision, _, err := s.client.Get(fmt.Sprintf("%s/revisions/%s.revision", Url, key))
	if err != nil {
		c.Error(err)
		return nil
	}
	content, err := s.signer.SnapRevision(key, revision)
	if err != nil {
		c.Error(err)
		return nil
	}
	return c.String(http.StatusOK, content)
}

func (s *SyncloudStore) SyncloudCacheRefresh(c echo.Context) error {
	req, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	var request model.StoreCacheRefreshRequest
	err = json.Unmarshal(req, &request)
	if err != nil {
		return err
	}

	if request.Token != s.token {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}

	return s.apiCache.Refresh()
}
