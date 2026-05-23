package api

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/syncloud/store/model"
	"github.com/syncloud/store/release"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type MultipartStore interface {
	Create(key string) (string, error)
	PresignPart(key, uploadId string, partNumber int) (string, error)
	Complete(key, uploadId string, parts []*s3.CompletedPart) error
	Abort(key, uploadId string) error
	HeadSize(key string) (int64, error)
	Put(key string, body []byte, contentType string) error
	Get(key string) ([]byte, error)
}

type CacheRefresher interface {
	Refresh() error
}

type Publish struct {
	mp     MultipartStore
	cache  CacheRefresher
	token  string
	logger *zap.Logger
}

func NewPublish(mp MultipartStore, cache CacheRefresher, token string, logger *zap.Logger) *Publish {
	return &Publish{mp: mp, cache: cache, token: token, logger: logger}
}

func snapKey(app, version, arch string) string {
	return fmt.Sprintf("apps/%s_%s_%s.snap", app, version, arch)
}

func sha384Key(app, version, arch string) string {
	return fmt.Sprintf("apps/%s_%s_%s.snap.sha384", app, version, arch)
}

func sizeKey(app, version, arch string) string {
	return fmt.Sprintf("apps/%s_%s_%s.snap.size", app, version, arch)
}

func versionKey(channel, app, arch string) string {
	return fmt.Sprintf("releases/%s/%s.%s.version", channel, app, arch)
}

func snapYamlKey(channel, app string) string {
	return fmt.Sprintf("v2/apps/%s/%s/snap.yaml", channel, app)
}

func iconKey(channel, app string) string {
	return fmt.Sprintf("v2/apps/%s/%s/icon.png", channel, app)
}

func (p *Publish) Init(c echo.Context) error {
	var req model.PublishInitRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if req.Token != p.token {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}
	if req.Name == "" || req.Version == "" || req.Arch == "" || req.Channel == "" {
		return c.String(http.StatusBadRequest, "name, version, arch, channel are required")
	}
	if req.Size <= 0 {
		return c.String(http.StatusBadRequest, "size must be > 0")
	}
	partSize := req.PartSize
	if partSize <= 0 {
		partSize = release.DefaultPartSize
	}
	key := snapKey(req.Name, req.Version, req.Arch)
	uploadId, err := p.mp.Create(key)
	if err != nil {
		p.logger.Error("multipart create failed", zap.Error(err))
		return c.String(http.StatusInternalServerError, err.Error())
	}
	partCount := int((req.Size + partSize - 1) / partSize)
	urls := make([]string, 0, partCount)
	for i := 1; i <= partCount; i++ {
		u, err := p.mp.PresignPart(key, uploadId, i)
		if err != nil {
			_ = p.mp.Abort(key, uploadId)
			return c.String(http.StatusInternalServerError, err.Error())
		}
		urls = append(urls, u)
	}
	return c.JSON(http.StatusOK, &model.PublishInitResponse{
		UploadId:         uploadId,
		Key:              key,
		PartCount:        partCount,
		PartUrls:         urls,
		ExpiresInSeconds: int64(release.PresignedUrlTTL.Seconds()),
	})
}

func (p *Publish) PartUrl(c echo.Context) error {
	var req model.PublishPartUrlRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if req.Token != p.token {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}
	if req.Key == "" || req.UploadId == "" || req.PartNumber <= 0 {
		return c.String(http.StatusBadRequest, "key, upload_id, part_number are required")
	}
	u, err := p.mp.PresignPart(req.Key, req.UploadId, req.PartNumber)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, &model.PublishPartUrlResponse{Url: u})
}

func (p *Publish) Finalise(c echo.Context) error {
	var req model.PublishFinaliseRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if req.Token != p.token {
		return c.String(http.StatusUnauthorized, "unauthorized")
	}
	if req.Key == "" || req.UploadId == "" || len(req.Parts) == 0 {
		return c.String(http.StatusBadRequest, "key, upload_id, parts are required")
	}

	parts := make([]*s3.CompletedPart, 0, len(req.Parts))
	for _, pt := range req.Parts {
		parts = append(parts, &s3.CompletedPart{
			ETag:       aws.String(pt.ETag),
			PartNumber: aws.Int64(int64(pt.PartNumber)),
		})
	}
	if err := p.mp.Complete(req.Key, req.UploadId, parts); err != nil {
		p.logger.Error("multipart complete failed", zap.Error(err))
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if req.Size > 0 {
		size, err := p.mp.HeadSize(req.Key)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		if size != req.Size {
			return c.String(http.StatusConflict,
				fmt.Sprintf("uploaded size %d does not match declared %d", size, req.Size))
		}
	}

	if req.SnapYaml != "" {
		if err := p.writeSnapYaml(req.Channel, req.Name, []byte(req.SnapYaml)); err != nil {
			return c.String(http.StatusConflict, err.Error())
		}
	}

	if req.IconPngB64 != "" {
		icon, err := base64.StdEncoding.DecodeString(req.IconPngB64)
		if err != nil {
			return c.String(http.StatusBadRequest, "icon_png_b64 not valid base64: "+err.Error())
		}
		if err := p.mp.Put(iconKey(req.Channel, req.Name), icon, "image/png"); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	if req.Sha384 != "" {
		if err := p.mp.Put(sha384Key(req.Name, req.Version, req.Arch),
			[]byte(req.Sha384), "text/plain"); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	if req.Size > 0 {
		if err := p.mp.Put(sizeKey(req.Name, req.Version, req.Arch),
			[]byte(fmt.Sprintf("%d", req.Size)), "text/plain"); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	if err := p.mp.Put(versionKey(req.Channel, req.Name, req.Arch),
		[]byte(req.Version), "text/plain"); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if err := p.cache.Refresh(); err != nil {
		p.logger.Warn("cache refresh after publish failed", zap.Error(err))
	}
	return c.JSON(http.StatusOK, &model.PublishFinaliseResponse{Ok: true})
}

func (p *Publish) writeSnapYaml(channel, app string, newYaml []byte) error {
	key := snapYamlKey(channel, app)
	existing, err := p.mp.Get(key)
	if err == nil && len(existing) > 0 {
		ex, errA := parseMeta(existing)
		nx, errB := parseMeta(newYaml)
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

func parseMeta(b []byte) (snapMeta, error) {
	var m snapMeta
	err := yaml.Unmarshal(b, &m)
	return m, err
}
