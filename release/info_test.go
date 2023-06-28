package release

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileParse(t *testing.T) {
	info, err := Parse("files_123_armhf.snap", "master")
	assert.Nil(t, err)
	assert.Equal(t, "apps/files_123_armhf.snap", info.StoreSnapPath)
	assert.Equal(t, "apps/files_123_armhf.snap.size", info.StoreSizePath)
	assert.Equal(t, "apps/files_123_armhf.snap.sha384", info.StoreSha384Path)
	assert.Equal(t, "123", info.Version)
	assert.Equal(t, "releases/master/files.armhf.version", info.StoreVersionPath)
}

func TestFileParse_WithPath(t *testing.T) {

	info, err := Parse("/test/files_123_armhf.snap", "master")
	assert.Nil(t, err)
	assert.Equal(t, "apps/files_123_armhf.snap", info.StoreSnapPath)
	assert.Equal(t, "apps/files_123_armhf.snap.size", info.StoreSizePath)
	assert.Equal(t, "apps/files_123_armhf.snap.sha384", info.StoreSha384Path)
	assert.Equal(t, "123", info.Version)
	assert.Equal(t, "releases/master/files.armhf.version", info.StoreVersionPath)
}

func TestFileParse_StableIsRc(t *testing.T) {

	info, err := Parse("files_123_armhf.snap", "stable")
	assert.Nil(t, err)
	assert.Equal(t, "apps/files_123_armhf.snap", info.StoreSnapPath)
	assert.Equal(t, "apps/files_123_armhf.snap.size", info.StoreSizePath)
	assert.Equal(t, "apps/files_123_armhf.snap.sha384", info.StoreSha384Path)
	assert.Equal(t, "123", info.Version)
	assert.Equal(t, "releases/rc/files.armhf.version", info.StoreVersionPath)
}
