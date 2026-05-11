package api

import (
	"github.com/labstack/echo/v4"
	"github.com/syncloud/store/internal/version"
	"github.com/syncloud/store/model"
	"golang.org/x/exp/slices"
	"io/fs"
	"net/http"
	"strings"
)

type WebCache interface {
	UIApps(channel string) []*model.UIApp
}

type Web struct {
	fs         fs.FS
	webCache   WebCache
	popularity Popularity
	fileServer http.Handler
}

func NewWeb(webFS fs.FS, webCache WebCache, popularity Popularity) *Web {
	return &Web{
		fs:         webFS,
		webCache:   webCache,
		popularity: popularity,
		fileServer: http.FileServer(http.FS(webFS)),
	}
}

func (w *Web) Apps(c echo.Context) error {
	channel := c.QueryParam("channel")
	if channel == "" {
		channel = "stable"
	}
	apps := w.webCache.UIApps(channel)
	if w.popularity != nil {
		for _, a := range apps {
			a.Popularity = w.popularity.Count(model.SnapId(a.SnapID).Name())
		}
		slices.SortFunc(apps, func(a, b *model.UIApp) bool {
			if a.Popularity != b.Popularity {
				return a.Popularity > b.Popularity
			}
			return a.Name < b.Name
		})
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	return c.JSON(http.StatusOK, apps)
}

func (w *Web) Version(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"gitSha":      version.GitSha,
		"buildNumber": version.BuildNumber,
		"buildTime":   version.BuildTime,
	})
}

func (w *Web) Serve(rw http.ResponseWriter, r *http.Request) {
	requested := strings.TrimPrefix(r.URL.Path, "/")
	if requested == "" {
		w.fileServer.ServeHTTP(rw, r)
		return
	}
	if f, err := w.fs.Open(requested); err == nil {
		_ = f.Close()
		w.fileServer.ServeHTTP(rw, r)
		return
	}
	r2 := r.Clone(r.Context())
	r2.URL.Path = "/"
	w.fileServer.ServeHTTP(rw, r2)
}
