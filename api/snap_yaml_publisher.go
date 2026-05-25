package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/syncloud/store/model"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type SnapYamlPublisher struct {
	mp     MultipartStore
	token  string
	logger *zap.Logger
}

func NewSnapYamlPublisher(mp MultipartStore, token string, logger *zap.Logger) *SnapYamlPublisher {
	return &SnapYamlPublisher{mp: mp, token: token, logger: logger}
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
	if err := p.write(req.Channel, req.Name, []byte(req.SnapYaml)); err != nil {
		return c.String(http.StatusConflict, err.Error())
	}
	return c.JSON(http.StatusOK, &model.PublishSnapYamlResponse{Ok: true})
}

func (p *SnapYamlPublisher) write(channel, app string, newYaml []byte) error {
	key := snapYamlKey(channel, app)
	existing, err := p.mp.Get(key)
	if err == nil && len(existing) > 0 {
		ex, errA := parseSnapMeta(existing)
		nx, errB := parseSnapMeta(newYaml)
		if errA == nil && errB == nil {
			if ex.Name != nx.Name || ex.Summary != nx.Summary ||
				ex.Description != nx.Description || ex.Type != nx.Type {
				return fmt.Errorf("snap.yaml metadata drift for %s/%s: "+
					"existing=(name=%q summary=%q description=%q type=%q) "+
					"new=(name=%q summary=%q description=%q type=%q)",
					channel, app,
					ex.Name, ex.Summary, ex.Description, ex.Type,
					nx.Name, nx.Summary, nx.Description, nx.Type)
			}
		}
	}
	return p.mp.Put(key, newYaml, "application/x-yaml")
}

type snapMeta struct {
	Name        string `yaml:"name"`
	Summary     string `yaml:"summary"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
}

func parseSnapMeta(b []byte) (snapMeta, error) {
	var m snapMeta
	err := yaml.Unmarshal(b, &m)
	return m, err
}
