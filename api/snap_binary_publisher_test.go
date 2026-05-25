package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syncloud/store/model"
	"go.uber.org/zap"
)

type fakeBinaryStore struct {
	objects map[string][]byte
}

func (f *fakeBinaryStore) Create(_ string) (string, error) { return "upload-1", nil }
func (f *fakeBinaryStore) PresignPart(_, u string, n int) (string, error) {
	return fmt.Sprintf("https://s3.example/?p=%s&n=%d", u, n), nil
}
func (f *fakeBinaryStore) Complete(_, _ string, _ []*s3.CompletedPart) error { return nil }
func (f *fakeBinaryStore) Abort(_, _ string) error                           { return nil }
func (f *fakeBinaryStore) HeadSize(k string) (int64, error)                  { return int64(len(f.objects[k])), nil }
func (f *fakeBinaryStore) Put(k string, b []byte, _ string) error {
	f.objects[k] = b
	return nil
}

type fakeRefresher struct{ refreshed bool }

func (f *fakeRefresher) Refresh() error { f.refreshed = true; return nil }

func TestSnapBinaryInit_AuthAndPartCount(t *testing.T) {
	store := &fakeBinaryStore{objects: map[string][]byte{}}
	p := NewSnapBinaryPublisher(store, &fakeRefresher{}, "secret", zap.NewNop())

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
	store := &fakeBinaryStore{objects: map[string][]byte{}}
	cache := &fakeRefresher{}
	p := NewSnapBinaryPublisher(store, cache, "secret", zap.NewNop())

	rec, err := postJSON(t, p.Finalise, model.PublishFinaliseRequest{
		Token: "secret", Name: "app", Version: "1", Arch: "amd64", Channel: "master",
		Key: "apps/app_1_amd64.snap", UploadId: "u1",
		Parts:  []model.PublishPart{{PartNumber: 1, ETag: "etag1"}},
		Size:   0,
		Sha384: "abc",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, store.objects, "apps/app_1_amd64.snap.sha384")
	assert.Contains(t, store.objects, "releases/master/app.amd64.version")
	assert.Equal(t, []byte("1"), store.objects["releases/master/app.amd64.version"])
	assert.True(t, cache.refreshed)
}
