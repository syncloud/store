package api

import (
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

type fakeYamlStore struct {
	objects map[string][]byte
}

func (f *fakeYamlStore) Put(key string, body []byte, _ string) error {
	f.objects[key] = body
	return nil
}

func yamlPost(t *testing.T, h echo.HandlerFunc, body interface{}) (*httptest.ResponseRecorder, error) {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	return rec, h(echo.New().NewContext(req, rec))
}

func TestSnapYamlPublisher_FirstWrite(t *testing.T) {
	store := &fakeYamlStore{objects: map[string][]byte{}}
	p := NewSnapYamlPublisher(store, "secret", zap.NewNop())

	rec, err := yamlPost(t, p.Publish, model.PublishSnapYamlRequest{
		Token: "secret", Name: "app", Channel: "master",
		SnapYaml: "name: app\nsummary: A\ndescription: B\n",
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, store.objects, "v2/apps/master/app/snap.yaml")
}

func TestSnapYamlPublisher_OverwritesExisting(t *testing.T) {
	store := &fakeYamlStore{objects: map[string][]byte{}}
	store.objects["v2/apps/master/app/snap.yaml"] = []byte("name: app\nsummary: Old\ndescription: O\n")
	p := NewSnapYamlPublisher(store, "secret", zap.NewNop())

	newYaml := "name: app\nsummary: New\ndescription: N\n"
	rec, _ := yamlPost(t, p.Publish, model.PublishSnapYamlRequest{
		Token: "secret", Name: "app", Channel: "master", SnapYaml: newYaml,
	})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []byte(newYaml), store.objects["v2/apps/master/app/snap.yaml"])
}
