package api

import (
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
	f.mu.Lock(); defer f.mu.Unlock()
	f.parts = append(f.parts, parts)
	return nil
}
func (f *fakeMP) Abort(_, _ string) error             { return nil }
func (f *fakeMP) HeadSize(k string) (int64, error)    { return int64(len(f.objects[k])), nil }
func (f *fakeMP) Put(k string, b []byte, _ string) error {
	f.mu.Lock(); defer f.mu.Unlock()
	f.objects[k] = b
	return nil
}
func (f *fakeMP) Get(k string) ([]byte, error) {
	if f.getErr != nil { return nil, f.getErr }
	b, ok := f.objects[k]
	if !ok { return nil, errors.New("nosuchkey") }
	return b, nil
}

func itoa(i int) string {
	const d = "0123456789"
	if i == 0 { return "0" }
	s := ""
	for i > 0 { s = string(d[i%10]) + s; i /= 10 }
	return s
}

type fakeCache struct{ refreshed bool }
func (f *fakeCache) Refresh() error { f.refreshed = true; return nil }

func newHandler(t *testing.T, mp MultipartStore) (*Publish, *fakeCache) {
	c := &fakeCache{}
	p := NewPublish(mp, c, "secret", zap.NewNop())
	return p, c
}

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

func TestPublishInit_AuthAndPartCount(t *testing.T) {
	mp := newFakeMP()
	p, _ := newHandler(t, mp)

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

func TestPublishFinalise_WritesAllSidecars(t *testing.T) {
	mp := newFakeMP()
	p, cache := newHandler(t, mp)

	rec, err := postJSON(t, p.Finalise, model.PublishFinaliseRequest{
		Token: "secret", Name: "app", Version: "1", Arch: "amd64", Channel: "master",
		Key: "apps/app_1_amd64.snap", UploadId: "u1",
		Parts:    []model.PublishPart{{PartNumber: 1, ETag: "etag1"}},
		Size:     0,
		Sha384:   "abc",
		SnapYaml: "name: app\nsummary: App\ndescription: A test\n",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, mp.objects, "v2/apps/master/app/snap.yaml")
	assert.Contains(t, mp.objects, "v2/apps/master/apps.json")
	assert.Contains(t, mp.objects, "apps/app_1_amd64.snap.sha384")
	assert.Contains(t, mp.objects, "releases/master/app.amd64.version")
	assert.Equal(t, []byte("1"), mp.objects["releases/master/app.amd64.version"])
	assert.True(t, cache.refreshed)
}

func TestPublishFinalise_DriftRejected(t *testing.T) {
	mp := newFakeMP()
	mp.objects["v2/apps/master/app/snap.yaml"] = []byte("name: app\nsummary: Old\ndescription: O\n")
	p, _ := newHandler(t, mp)

	rec, _ := postJSON(t, p.Finalise, model.PublishFinaliseRequest{
		Token: "secret", Name: "app", Version: "1", Arch: "amd64", Channel: "master",
		Key: "apps/app_1_amd64.snap", UploadId: "u1",
		Parts:    []model.PublishPart{{PartNumber: 1, ETag: "etag1"}},
		SnapYaml: "name: app\nsummary: New\ndescription: N\n",
	})
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "metadata drift")
}

func TestPublishFinalise_IdenticalSnapYamlAccepted(t *testing.T) {
	mp := newFakeMP()
	y := "name: app\nsummary: App\ndescription: D\n"
	mp.objects["v2/apps/master/app/snap.yaml"] = []byte(y)
	p, _ := newHandler(t, mp)

	rec, _ := postJSON(t, p.Finalise, model.PublishFinaliseRequest{
		Token: "secret", Name: "app", Version: "1", Arch: "amd64", Channel: "master",
		Key: "apps/app_1_amd64.snap", UploadId: "u1",
		Parts:    []model.PublishPart{{PartNumber: 1, ETag: "etag1"}},
		SnapYaml: y,
	})
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPublishFinalise_AppsIndexUpdated(t *testing.T) {
	mp := newFakeMP()
	mp.objects["v2/apps/master/apps.json"] = []byte(`{"apps":["existing"]}`)
	p, _ := newHandler(t, mp)

	_, _ = postJSON(t, p.Finalise, model.PublishFinaliseRequest{
		Token: "secret", Name: "app", Version: "1", Arch: "amd64", Channel: "master",
		Key: "apps/app_1_amd64.snap", UploadId: "u1",
		Parts: []model.PublishPart{{PartNumber: 1, ETag: "etag1"}},
	})
	var idx model.AppsIndex
	require.NoError(t, json.Unmarshal(mp.objects["v2/apps/master/apps.json"], &idx))
	assert.ElementsMatch(t, []string{"existing", "app"}, idx.Apps)
}
