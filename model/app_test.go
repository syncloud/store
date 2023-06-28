package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApp_ToInfo(t *testing.T) {
	app := &App{
		Name:     "name",
		Summary:  "summary",
		Icon:     "url",
		Required: false,
	}
	info, err := app.ToInfo("1", 0, "sha", "url")
	assert.NoError(t, err)
	assert.Equal(t, "name.1", info.SnapID)
	assert.Equal(t, int64(0), info.Download.Size)
	assert.Equal(t, "sha", info.Download.Sha3_384)
	assert.Equal(t, "url", info.Download.URL)
	assert.Equal(t, "app", info.Type)
}

func TestApp_ToInfo_Base(t *testing.T) {
	app := &App{
		Name:     "name",
		Summary:  "summary",
		Icon:     "url",
		Required: true,
	}
	info, err := app.ToInfo("1", 0, "sha", "url")
	assert.NoError(t, err)
	assert.Equal(t, "name.1", info.SnapID)
	assert.Equal(t, int64(0), info.Download.Size)
	assert.Equal(t, "sha", info.Download.Sha3_384)
	assert.Equal(t, "url", info.Download.URL)
	assert.Equal(t, "base", info.Type)
}
