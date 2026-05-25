package api

import (
	"testing"

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

func TestSnapYamlPublisher_FirstWrite(t *testing.T) {
	store := &fakeYamlStore{objects: map[string][]byte{}}
	p := NewSnapYamlPublisher(store, "secret", zap.NewNop())

	resp, err := p.Publish(model.PublishSnapYamlRequest{
		Token: "secret", Name: "app", Channel: "master",
		SnapYaml: "name: app\nsummary: A\ndescription: B\n",
	})
	require.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Contains(t, store.objects, "v2/apps/master/app/snap.yaml")
}

func TestSnapYamlPublisher_OverwritesExisting(t *testing.T) {
	store := &fakeYamlStore{objects: map[string][]byte{}}
	store.objects["v2/apps/master/app/snap.yaml"] = []byte("name: app\nsummary: Old\ndescription: O\n")
	p := NewSnapYamlPublisher(store, "secret", zap.NewNop())

	newYaml := "name: app\nsummary: New\ndescription: N\n"
	resp, err := p.Publish(model.PublishSnapYamlRequest{
		Token: "secret", Name: "app", Channel: "master", SnapYaml: newYaml,
	})
	require.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Equal(t, []byte(newYaml), store.objects["v2/apps/master/app/snap.yaml"])
}
