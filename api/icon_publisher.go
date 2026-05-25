package api

import (
	"encoding/base64"
	"fmt"

	"github.com/syncloud/store/model"
	"github.com/syncloud/store/release"
	"go.uber.org/zap"
)

type IconPublisher struct {
	store  release.ObjectPutter
	token  string
	logger *zap.Logger
}

func NewIconPublisher(store release.ObjectPutter, token string, logger *zap.Logger) *IconPublisher {
	return &IconPublisher{store: store, token: token, logger: logger}
}

func iconKey(channel, app string) string {
	return fmt.Sprintf("v2/apps/%s/%s/icon.png", channel, app)
}

func (p *IconPublisher) Publish(req model.PublishIconRequest) (*model.PublishIconResponse, error) {
	if req.Token != p.token {
		return nil, unauthorized()
	}
	if req.Name == "" || req.Channel == "" || req.IconPngB64 == "" {
		return nil, badRequest("name, channel, icon_png_b64 are required")
	}
	icon, err := base64.StdEncoding.DecodeString(req.IconPngB64)
	if err != nil {
		return nil, badRequest("icon_png_b64 not valid base64: " + err.Error())
	}
	if err := p.store.Put(iconKey(req.Channel, req.Name), icon, "image/png"); err != nil {
		return nil, err
	}
	return &model.PublishIconResponse{Ok: true}, nil
}
