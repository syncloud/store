package api

import (
	"github.com/labstack/echo/v4"
	"github.com/syncloud/store/internal/version"
	"github.com/syncloud/store/model"
	"github.com/syncloud/store/storage"
	"io/fs"
	"net/http"
	"strings"
)

type Web struct {
	fs         fs.FS
	index      storage.Index
	fileServer http.Handler
}

func NewWeb(webFS fs.FS, index storage.Index) *Web {
	return &Web{
		fs:         webFS,
		index:      index,
		fileServer: http.FileServer(http.FS(webFS)),
	}
}

func (w *Web) Apps(c echo.Context) error {
	channel := c.QueryParam("channel")
	if channel == "" {
		channel = "stable"
	}
	apps := w.index.UIApps(channel)
	if apps == nil {
		apps = []*model.UIApp{}
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
