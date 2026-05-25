package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/syncloud/store/model"
	"github.com/syncloud/store/release"
	"go.uber.org/zap"
)

type SnapYamlPublisher struct {
	store  release.ObjectPutter
	token  string
	logger *zap.Logger
}

func NewSnapYamlPublisher(store release.ObjectPutter, token string, logger *zap.Logger) *SnapYamlPublisher {
	return &SnapYamlPublisher{store: store, token: token, logger: logger}
}

func snapYamlKey(channel, app string) string {
	return fmt.Sprintf("v2/apps/%s/%s/snap.yaml", channel, app)
}

func (p *SnapYamlPublisher) Publish(c echo.Context) error {
	var req model.PublishSnapYamlRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if req.Token != p.token {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}
	if req.Name == "" || req.Channel == "" || req.SnapYaml == "" {
		return c.String(http.StatusBadRequest, "name, channel, snap_yaml are required")
	}
	if err := p.store.Put(snapYamlKey(req.Channel, req.Name), []byte(req.SnapYaml), "application/x-yaml"); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, &model.PublishSnapYamlResponse{Ok: true})
}
