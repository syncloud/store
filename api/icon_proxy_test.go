package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIconProxy_RewritesPathToReleasesImages(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte("PNG"))
	}))
	defer server.Close()

	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	proxy := NewIconProxy(target)

	req := httptest.NewRequest(http.MethodGet, "/api/ui/v1/icons/bitwarden-128.png", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	assert.Equal(t, "/releases/stable/images/bitwarden-128.png", gotPath)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestIconProxy_ForwardsUpstreamHostHeader(t *testing.T) {
	var gotHost string
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		if r.Host != server.Listener.Addr().String() {
			http.Error(w, "WebsiteRedirect: Request does not contain a bucket name.", http.StatusMovedPermanently)
			return
		}
		_, _ = w.Write([]byte("PNG"))
	}))
	defer server.Close()

	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	proxy := NewIconProxy(target)

	req := httptest.NewRequest(http.MethodGet, "/api/ui/v1/icons/bitwarden-128.png", nil)
	req.Host = "uatstore.syncloud.org"
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)

	expectedHost := strings.TrimPrefix(server.URL, "http://")
	assert.Equal(t, expectedHost, gotHost,
		"upstream Host header must be the proxy target, not the incoming browser host (S3 routes by Host)")

	body, _ := io.ReadAll(rec.Body)
	assert.Equal(t, http.StatusOK, rec.Code, "body=%q", string(body))
	assert.Equal(t, "PNG", string(body))
}
