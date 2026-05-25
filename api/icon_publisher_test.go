package api

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syncloud/store/model"
	"go.uber.org/zap"
)

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
