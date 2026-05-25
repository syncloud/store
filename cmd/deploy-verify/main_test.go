package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestVerifier(srv *httptest.Server, token string) *verifier {
	return &verifier{
		deployUrl: srv.URL,
		token:     token,
		http:      srv.Client(),
	}
}

func TestCacheRefresh_TokenAndAwsValid(t *testing.T) {
	var received string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/syncloud/v1/cache/refresh", r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		received = string(b)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	v := newTestVerifier(srv, "secret")
	assert.NoError(t, v.cacheRefresh())
	assert.Contains(t, received, `"token":"secret"`)
}

func TestCacheRefresh_BadToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer srv.Close()

	v := newTestVerifier(srv, "wrong")
	err := v.cacheRefresh()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestCacheRefresh_AwsFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("AuthorizationHeaderMalformed"))
	}))
	defer srv.Close()

	v := newTestVerifier(srv, "secret")
	err := v.cacheRefresh()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "AuthorizationHeaderMalformed")
}

func TestAssertApps_NonEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"snapId":"testapp1.1"},{"snapId":"testapp2.1"}]`))
	}))
	defer srv.Close()

	v := newTestVerifier(srv, "secret")
	assert.NoError(t, v.assertApps())
}

func TestAssertApps_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	v := newTestVerifier(srv, "secret")
	err := v.assertApps()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestAssertFind_NonEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{},{}]}`))
	}))
	defer srv.Close()

	v := newTestVerifier(srv, "secret")
	assert.NoError(t, v.assertFind())
}

func TestAssertFind_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()

	v := newTestVerifier(srv, "secret")
	err := v.assertFind()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no results")
}

func TestAssertWebUI_200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	v := newTestVerifier(srv, "secret")
	assert.NoError(t, v.assertWebUI())
}

func TestAssertWebUI_500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	v := newTestVerifier(srv, "secret")
	err := v.assertWebUI()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestCountJSONList(t *testing.T) {
	n, err := countJSONList(strings.NewReader(`[{"a":1},{"b":2},{"c":3}]`))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	n, err = countJSONList(strings.NewReader(`[]`))
	assert.NoError(t, err)
	assert.Equal(t, 0, n)

	_, err = countJSONList(strings.NewReader(`{"oops": true}`))
	assert.Error(t, err)
}

func TestCountResults(t *testing.T) {
	n, err := countResults(strings.NewReader(`{"results":[{},{},{}]}`))
	assert.NoError(t, err)
	assert.Equal(t, 3, n)

	n, err = countResults(strings.NewReader(`{"results":[]}`))
	assert.NoError(t, err)
	assert.Equal(t, 0, n)

	n, err = countResults(strings.NewReader(`{"other":"thing"}`))
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}
