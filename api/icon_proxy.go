package api

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const iconRoutePrefix = "/api/ui/v1/icons/"

type IconProxy struct {
	proxy *httputil.ReverseProxy
}

func NewIconProxy(upstream *url.URL) *IconProxy {
	rp := httputil.NewSingleHostReverseProxy(upstream)
	rp.Director = func(r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, iconRoutePrefix)
		r.URL.Scheme = upstream.Scheme
		r.URL.Host = upstream.Host
		r.URL.Path = strings.TrimSuffix(upstream.Path, "/") + "/v2/apps/" + rest + "/icon.png"
		r.URL.RawPath = ""
		r.Host = upstream.Host
		if _, ok := r.Header["User-Agent"]; !ok {
			r.Header.Set("User-Agent", "")
		}
	}
	return &IconProxy{proxy: rp}
}

func (p *IconProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
