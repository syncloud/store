package api

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const activeDevicesWindow = 24 * time.Hour

type SnapdMetrics struct {
	requests           *prometheus.CounterVec
	scheduledRefreshes *prometheus.CounterVec
	activeDevices      *prometheus.Desc
	window             time.Duration
	now                func() time.Time
	mu                 sync.Mutex
	calls              map[string][]time.Time
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
		scheduledRefreshes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "store_snapd_scheduled_refresh_total",
				Help: "Number of scheduled (daily auto-refresh) snapd refresh requests, by the snapd client version. One bulk request per device per day.",
			},
			[]string{"version"},
		),
		activeDevices: prometheus.NewDesc(
			"store_snapd_active_devices",
			"Scheduled refreshes in the last 24h by snapd version. One request per device per day, so this approximates active devices; a version with no refresh in the window drops out.",
			[]string{"version"}, nil,
		),
		window: activeDevicesWindow,
		now:    time.Now,
		calls:  map[string][]time.Time{},
	}
}

func (m *SnapdMetrics) Record(snap, action, arch string, status int) {
	m.requests.WithLabelValues(snap, action, arch, strconv.Itoa(status)).Inc()
}

func (m *SnapdMetrics) RecordScheduledRefresh(version string) {
	m.scheduledRefreshes.WithLabelValues(version).Inc()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls[version] = append(m.calls[version], m.now())
}

func (m *SnapdMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.requests.Describe(ch)
	m.scheduledRefreshes.Describe(ch)
	ch <- m.activeDevices
}

func (m *SnapdMetrics) Collect(ch chan<- prometheus.Metric) {
	m.requests.Collect(ch)
	m.scheduledRefreshes.Collect(ch)

	m.mu.Lock()
	cutoff := m.now().Add(-m.window)
	counts := map[string]int{}
	for version, times := range m.calls {
		kept := times[:0]
		for _, t := range times {
			if t.Before(cutoff) {
				continue
			}
			kept = append(kept, t)
		}
		if len(kept) == 0 {
			delete(m.calls, version)
			continue
		}
		m.calls[version] = kept
		counts[version] = len(kept)
	}
	m.mu.Unlock()

	for version, n := range counts {
		ch <- prometheus.MustNewConstMetric(m.activeDevices, prometheus.GaugeValue, float64(n), version)
	}
}
