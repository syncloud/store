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
	base := rp.Director
	rp.Director = func(r *http.Request) {
		base(r)
		r.Host = upstream.Host
		icon := strings.TrimPrefix(r.URL.Path, iconRoutePrefix)
		r.URL.Path = strings.TrimSuffix(upstream.Path, "/") + "/releases/stable/images/" + icon
		r.URL.RawPath = ""
	}
	return &IconProxy{proxy: rp}
}

func (p *IconProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
