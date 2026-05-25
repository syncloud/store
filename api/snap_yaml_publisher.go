package api

import (
	"fmt"

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

func (p *SnapYamlPublisher) Publish(req model.PublishSnapYamlRequest) (*model.PublishSnapYamlResponse, error) {
	if req.Token != p.token {
		return nil, unauthorized()
	}
	if req.Name == "" || req.Channel == "" || req.SnapYaml == "" {
		return nil, badRequest("name, channel, snap_yaml are required")
	}
	if err := p.store.Put(snapYamlKey(req.Channel, req.Name), []byte(req.SnapYaml), "application/x-yaml"); err != nil {
		return nil, err
	}
	return &model.PublishSnapYamlResponse{Ok: true}, nil
}
