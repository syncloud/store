package api

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type SnapdMetrics struct {
	requests           *prometheus.CounterVec
	versions           *prometheus.CounterVec
	scheduledRefreshes *prometheus.CounterVec
}

func NewSnapdMetrics() *SnapdMetrics {
	return &SnapdMetrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "store_snapd_request_total",
				Help: "Number of snapd API requests, by snap, action, arch and HTTP status.",
			},
			[]string{"snap", "action", "arch", "status"},
		),
		versions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "store_snapd_client_version_total",
				Help: "Number of snapd API requests, by the snapd client version reported in the User-Agent and the endpoint.",
			},
			[]string{"version", "endpoint"},
		),
		scheduledRefreshes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "store_snapd_scheduled_refresh_total",
				Help: "Number of scheduled (daily auto-refresh) snapd refresh requests, by the snapd client version. One bulk request per device per day.",
			},
			[]string{"version"},
		),
	}
}

func (m *SnapdMetrics) Record(snap, action, arch string, status int) {
	m.requests.WithLabelValues(snap, action, arch, strconv.Itoa(status)).Inc()
}

func (m *SnapdMetrics) RecordVersion(version, endpoint string) {
	m.versions.WithLabelValues(version, endpoint).Inc()
}

func (m *SnapdMetrics) RecordScheduledRefresh(version string) {
	m.scheduledRefreshes.WithLabelValues(version).Inc()
}

func (m *SnapdMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.requests.Describe(ch)
	m.versions.Describe(ch)
	m.scheduledRefreshes.Describe(ch)
}

func (m *SnapdMetrics) Collect(ch chan<- prometheus.Metric) {
	m.requests.Collect(ch)
	m.versions.Collect(ch)
	m.scheduledRefreshes.Collect(ch)
}
