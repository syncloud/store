package api

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/syncloud/store/model"
	"go.uber.org/zap"
)

type IconPublisher struct {
	store  ObjectPutter
	token  string
	logger *zap.Logger
}

func NewIconPublisher(store ObjectPutter, token string, logger *zap.Logger) *IconPublisher {
	return &IconPublisher{store: store, token: token, logger: logger}
}

func iconKey(channel, app string) string {
	return fmt.Sprintf("v2/apps/%s/%s/icon.png", channel, app)
}

func (p *IconPublisher) Publish(c echo.Context) error {
	var req model.PublishIconRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if req.Token != p.token {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}
	if req.Name == "" || req.Channel == "" || req.IconPngB64 == "" {
		return c.String(http.StatusBadRequest, "name, channel, icon_png_b64 are required")
	}
	icon, err := base64.StdEncoding.DecodeString(req.IconPngB64)
	if err != nil {
		return c.String(http.StatusBadRequest, "icon_png_b64 not valid base64: "+err.Error())
	}
	if err := p.store.Put(iconKey(req.Channel, req.Name), icon, "image/png"); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, &model.PublishIconResponse{Ok: true})
}
