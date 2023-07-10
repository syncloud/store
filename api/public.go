package api

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/syncloud/store/model"
	"github.com/syncloud/store/rest"
	"github.com/syncloud/store/storage"
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

type SyncloudStore struct {
	client  rest.Client
	echo    *echo.Echo
	address string
	index   storage.Index
	signer  Signer
	logger  *zap.Logger
}

func NewSyncloudStore(
	address string,
	index storage.Index,
	client rest.Client,
	signer Signer,
	logger *zap.Logger,
) *SyncloudStore {
	return &SyncloudStore{
		client:  client,
		echo:    echo.New(),
		signer:  signer,
		index:   index,
		address: address,
		logger:  logger,
	}
}

func (s *SyncloudStore) Start() error {

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

	s.logger.Info("listening on", zap.String("address", s.address))
	if s.IsUnixSocket() {
		_ = os.RemoveAll(s.address)
		l, err := net.Listen("unix", s.address)
		if err != nil {
			s.logger.Error("error", zap.Error(err))
			return err
		}
		if err := os.Chmod(s.address, 0777); err != nil {
			return err
		}

		s.echo.Listener = l
		return s.echo.Start("")
	} else {
		return s.echo.Start(s.address)
	}
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
	arch := c.Request().Heqder.Get("Syncloud-Architecture")
	s.logger.Info("refresh", zap.String("arch", arch))

	var request model.SnapActionRequest
	err = json.Unmarshal(req, &request)
	if err != nil {
		c.Error(err)
		return nil
	}
	s.logger.Info(fmt.Sprintf("refresh request: %s", string(req)))
	result := &model.StoreResults{}
	for _, action := range request.Actions {
		if action.Action == "fetch-assertions" {
			info := &model.StoreResult{
				Result: action.Action,
				Key:    action.Key,
			}
			result.Results = append(result.Results, info)
		} else {
			info, err := s.index.InfoById(action.Channel, action.SnapID, action.Action, action.Name, arch)
			if err != nil {
				return err
			}
			info.InstanceKey = action.InstanceKey
			result.Results = append(result.Results, info)
		}
	}
	return c.JSON(http.StatusOK, result)
}

func (s *SyncloudStore) Info(c echo.Context) error {
	name := c.Param("name")
	arch := c.QueryParam("architecture")
	result := s.index.Info(name, arch)
	if result == nil {
		return c.String(http.StatusNotFound, "not found")
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
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
	results := s.index.Find(channel, query, architecture)
	if results == nil {
		c.Error(fmt.Errorf("no channel: %s in the index", channel))
		return nil
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
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
