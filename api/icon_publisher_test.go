package api

import (
	"encoding/base64"
	"errors"
	"testing"

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

func TestIconPublisher_WritesObject(t *testing.T) {
	store := &fakeIconStore{objects: map[string][]byte{}}
	p := NewIconPublisher(store, "secret", zap.NewNop())

	icon := []byte{0x89, 0x50, 0x4e, 0x47}
	resp, err := p.Publish(model.PublishIconRequest{
		Token: "secret", Name: "app", Channel: "master",
		IconPngB64: base64.StdEncoding.EncodeToString(icon),
	})
	require.NoError(t, err)
	assert.True(t, resp.Ok)
	assert.Equal(t, icon, store.objects["v2/apps/master/app/icon.png"])
}

func TestIconPublisher_BadAuth(t *testing.T) {
	p := NewIconPublisher(&fakeIconStore{objects: map[string][]byte{}}, "secret", zap.NewNop())
	_, err := p.Publish(model.PublishIconRequest{Token: "wrong"})
	var ae *apiError
	require.True(t, errors.As(err, &ae))
	assert.Equal(t, 401, ae.Status)
}
