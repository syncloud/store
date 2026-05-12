package api

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type SnapdMetrics struct {
	requests *prometheus.CounterVec
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
	}
}

func (m *SnapdMetrics) Record(snap, action, arch string, status int) {
	m.requests.WithLabelValues(snap, action, arch, strconv.Itoa(status)).Inc()
}

func (m *SnapdMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.requests.Describe(ch)
}

func (m *SnapdMetrics) Collect(ch chan<- prometheus.Metric) {
	m.requests.Collect(ch)
}
