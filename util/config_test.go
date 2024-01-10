package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadConfig(t *testing.T) {

	config, err := LoadConfig("../config/test/secret.yaml")

	assert.NoError(t, err)
	assert.Equal(t, "123", config.Token)

}
