package api

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const activeDevicesWindow = 48 * time.Hour

type deviceSeen struct {
	version  string
	lastSeen time.Time
}

type SnapdMetrics struct {
	requests           *prometheus.CounterVec
	scheduledRefreshes *prometheus.CounterVec
	activeDevices      *prometheus.Desc
	window             time.Duration
	now                func() time.Time
	mu                 sync.Mutex
	devices            map[string]deviceSeen
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
			"Distinct devices (by client IP) whose last scheduled refresh is within the last 48h, by snapd version. Devices refresh once a day, so the 48h window keeps a device counted between its daily refreshes.",
			[]string{"version"}, nil,
		),
		window:  activeDevicesWindow,
		now:     time.Now,
		devices: map[string]deviceSeen{},
	}
}

func (m *SnapdMetrics) Record(snap, action, arch string, status int) {
	m.requests.WithLabelValues(snap, action, arch, strconv.Itoa(status)).Inc()
}

func (m *SnapdMetrics) RecordScheduledRefresh(deviceId, version string) {
	m.scheduledRefreshes.WithLabelValues(version).Inc()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.devices[deviceId] = deviceSeen{version: version, lastSeen: m.now()}
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
	for id, seen := range m.devices {
		if seen.lastSeen.Before(cutoff) {
			delete(m.devices, id)
			continue
		}
		counts[seen.version]++
	}
	m.mu.Unlock()

	for version, n := range counts {
		ch <- prometheus.MustNewConstMetric(m.activeDevices, prometheus.GaugeValue, float64(n), version)
	}
}
