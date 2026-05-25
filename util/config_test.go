package util

import (
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("../config/test/secret.yaml")

	assert.NoError(t, err)
	assert.Equal(t, "@token@", config.Token)
	assert.Equal(t, "http://apps.s3:3902", config.BaseUrl)
	assert.Equal(t, "apps.s3", config.Bucket)
	assert.Equal(t, "@aws_access_key_id@", config.AwsAccessKeyId)
	assert.Equal(t, "@aws_secret_access_key@", config.AwsSecretAccessKey)
	assert.Equal(t, "http://apps.s3", config.AwsS3Endpoint)
	assert.Equal(t, "garage", config.AwsRegion)
}

func TestSecretYamlSchemaMatches(t *testing.T) {
	envs := []string{"test", "uat", "prod"}
	keys := make(map[string][]string, len(envs))
	for _, env := range envs {
		raw, err := os.ReadFile("../config/" + env + "/secret.yaml")
		require.NoError(t, err)
		var m map[string]interface{}
		require.NoError(t, yaml.Unmarshal(raw, &m))
		k := make([]string, 0, len(m))
		for name := range m {
			k = append(k, name)
		}
		sort.Strings(k)
		keys[env] = k
	}
	assert.Equal(t, keys["test"], keys["uat"], "test vs uat secret.yaml keys diverge")
	assert.Equal(t, keys["test"], keys["prod"], "test vs prod secret.yaml keys diverge")
}
