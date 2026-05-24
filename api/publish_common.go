package api

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
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
