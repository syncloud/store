package rest

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestClient(url string, retries int) *PublishClient {
	return &PublishClient{
		storeUrl:  url,
		token:     "token",
		http:      &http.Client{Transport: &http.Transport{DisableKeepAlives: true}},
		retries:   retries,
		retryWait: 0,
	}
}

func closeConn(w http.ResponseWriter) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err == nil {
		conn.Close()
	}
}

func TestPostJSONRetriesOnTransientError(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			closeConn(w)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	err := newTestClient(server.URL, 3).SnapYaml("app", "stable", "name: app")

	assert.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
}

func TestPostJSONReturnsErrorWhenRetriesExhausted(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		closeConn(w)
	}))
	defer server.Close()

	err := newTestClient(server.URL, 2).SnapYaml("app", "stable", "name: app")

	assert.Error(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls))
}

func TestPostJSONDoesNotRetryOnHttpError(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	err := newTestClient(server.URL, 3).SnapYaml("app", "stable", "name: app")

	assert.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}
