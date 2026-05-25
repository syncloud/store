package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syncloud/store/model"
	"go.uber.org/zap"
)

type fakeMP struct {
	mu       sync.Mutex
	objects  map[string][]byte
	uploadId string
	parts    [][]*s3.CompletedPart
	getErr   error
}

func newFakeMP() *fakeMP { return &fakeMP{objects: map[string][]byte{}} }

func (f *fakeMP) Create(_ string) (string, error)                { return "upload-1", nil }
func (f *fakeMP) PresignPart(k, u string, n int) (string, error) { return "https://s3.example/?p=" + u + "&n=" + itoa(n), nil }
func (f *fakeMP) Complete(_, _ string, parts []*s3.CompletedPart) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.parts = append(f.parts, parts)
	return nil
}
func (f *fakeMP) Abort(_, _ string) error          { return nil }
func (f *fakeMP) HeadSize(k string) (int64, error) { return int64(len(f.objects[k])), nil }
func (f *fakeMP) Put(k string, b []byte, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[k] = b
	return nil
}
func (f *fakeMP) Get(k string) ([]byte, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	b, ok := f.objects[k]
	if !ok {
		return nil, errors.New("nosuchkey")
	}
	return b, nil
}

func itoa(i int) string {
	const d = "0123456789"
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(d[i%10]) + s
		i /= 10
	}
	return s
}

type fakeCache struct{ refreshed bool }

func (f *fakeCache) Refresh() error { f.refreshed = true; return nil }

func postJSON(t *testing.T, h echo.HandlerFunc, body interface{}) (*httptest.ResponseRecorder, error) {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(req, rec)
	err := h(ctx)
	return rec, err
}

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

func TestSnapYamlPublisher_FirstWrite(t *testing.T) {
	mp := newFakeMP()
	p := NewSnapYamlPublisher(mp, "secret", zap.NewNop())

	rec, err := postJSON(t, p.Publish, model.PublishSnapYamlRequest{
		Token: "secret", Name: "app", Channel: "master",
		SnapYaml: "name: app\nsummary: A\ndescription: B\n",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, mp.objects, "v2/apps/master/app/snap.yaml")
}

func TestSnapYamlPublisher_OverwritesExisting(t *testing.T) {
	mp := newFakeMP()
	mp.objects["v2/apps/master/app/snap.yaml"] = []byte("name: app\nsummary: Old\ndescription: O\n")
	p := NewSnapYamlPublisher(mp, "secret", zap.NewNop())

	newYaml := "name: app\nsummary: New\ndescription: N\n"
	rec, _ := postJSON(t, p.Publish, model.PublishSnapYamlRequest{
		Token: "secret", Name: "app", Channel: "master", SnapYaml: newYaml,
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []byte(newYaml), mp.objects["v2/apps/master/app/snap.yaml"])
}

func TestIconPublisher_WritesObject(t *testing.T) {
	mp := newFakeMP()
	p := NewIconPublisher(mp, "secret", zap.NewNop())

	icon := []byte{0x89, 0x50, 0x4e, 0x47}
	rec, err := postJSON(t, p.Publish, model.PublishIconRequest{
		Token: "secret", Name: "app", Channel: "master",
		IconPngB64: base64.StdEncoding.EncodeToString(icon),
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, icon, mp.objects["v2/apps/master/app/icon.png"])
}

func TestIconPublisher_BadAuth(t *testing.T) {
	p := NewIconPublisher(newFakeMP(), "secret", zap.NewNop())
	rec, _ := postJSON(t, p.Publish, model.PublishIconRequest{Token: "wrong"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
