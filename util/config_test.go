package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`token: test
base_url: http://apps.s3:3902
bucket: apps.s3
aws_access_key_id: key
aws_secret_access_key: secret
aws_s3_endpoint: http://apps.s3
aws_region: garage
`), 0o600))

	config, err := LoadConfig(path)

	assert.NoError(t, err)
	assert.Equal(t, "test", config.Token)
	assert.Equal(t, "http://apps.s3:3902", config.BaseUrl)
	assert.Equal(t, "apps.s3", config.Bucket)
	assert.Equal(t, "key", config.AwsAccessKeyId)
	assert.Equal(t, "secret", config.AwsSecretAccessKey)
	assert.Equal(t, "http://apps.s3", config.AwsS3Endpoint)
	assert.Equal(t, "garage", config.AwsRegion)
}
