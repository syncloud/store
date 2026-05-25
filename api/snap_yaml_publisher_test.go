package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syncloud/store/model"
	"go.uber.org/zap"
)

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
