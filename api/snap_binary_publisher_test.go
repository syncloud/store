package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syncloud/store/model"
	"go.uber.org/zap"
)

func TestSnapBinaryInit_AuthAndPartCount(t *testing.T) {
	mp := newFakeMP()
	p := NewSnapBinaryPublisher(mp, &fakeCache{}, "secret", zap.NewNop())

	rec, _ := postJSON(t, p.Init, model.PublishInitRequest{Token: "wrong"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	rec, err := postJSON(t, p.Init, model.PublishInitRequest{
		Token: "secret", Name: "app", Version: "1", Arch: "amd64",
		Channel: "master", Size: 33 * 1024 * 1024, Sha384: "deadbeef",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp model.PublishInitResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "apps/app_1_amd64.snap", resp.Key)
	assert.Equal(t, 3, resp.PartCount)
	assert.Len(t, resp.PartUrls, 3)
}

func TestSnapBinaryFinalise_WritesSidecars(t *testing.T) {
	mp := newFakeMP()
	cache := &fakeCache{}
	p := NewSnapBinaryPublisher(mp, cache, "secret", zap.NewNop())

	rec, err := postJSON(t, p.Finalise, model.PublishFinaliseRequest{
		Token: "secret", Name: "app", Version: "1", Arch: "amd64", Channel: "master",
		Key: "apps/app_1_amd64.snap", UploadId: "u1",
		Parts:  []model.PublishPart{{PartNumber: 1, ETag: "etag1"}},
		Size:   0,
		Sha384: "abc",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, mp.objects, "apps/app_1_amd64.snap.sha384")
	assert.Contains(t, mp.objects, "releases/master/app.amd64.version")
	assert.Equal(t, []byte("1"), mp.objects["releases/master/app.amd64.version"])
	assert.True(t, cache.refreshed)
}
