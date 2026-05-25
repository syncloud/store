package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syncloud/store/model"
	"go.uber.org/zap"
)

type fakeIconStore struct {
	objects map[string][]byte
}

func (f *fakeIconStore) Put(key string, body []byte, _ string) error {
	f.objects[key] = body
	return nil
}

func iconPost(t *testing.T, h echo.HandlerFunc, body interface{}) (*httptest.ResponseRecorder, error) {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	return rec, h(echo.New().NewContext(req, rec))
}

func TestIconPublisher_WritesObject(t *testing.T) {
	store := &fakeIconStore{objects: map[string][]byte{}}
	p := NewIconPublisher(store, "secret", zap.NewNop())

	icon := []byte{0x89, 0x50, 0x4e, 0x47}
	rec, err := iconPost(t, p.Publish, model.PublishIconRequest{
		Token: "secret", Name: "app", Channel: "master",
		IconPngB64: base64.StdEncoding.EncodeToString(icon),
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, icon, store.objects["v2/apps/master/app/icon.png"])
}

func TestIconPublisher_BadAuth(t *testing.T) {
	store := &fakeIconStore{objects: map[string][]byte{}}
	p := NewIconPublisher(store, "secret", zap.NewNop())
	rec, _ := iconPost(t, p.Publish, model.PublishIconRequest{Token: "wrong"})
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
