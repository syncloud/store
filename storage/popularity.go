package storage

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/syncloud/store/metrics"
)

type Popularity struct {
	mu   sync.Mutex
	seen map[string]map[string]time.Time
	ttl  time.Duration
}

func NewPopularity(ttl time.Duration) *Popularity {
	return &Popularity{
		seen: make(map[string]map[string]time.Time),
		ttl:  ttl,
	}
}

func (p *Popularity) Record(snap, device string) {
	if snap == "" || device == "" {
		return
	}
	p.mu.Lock()
	m, ok := p.seen[snap]
	if !ok {
		m = make(map[string]time.Time)
		p.seen[snap] = m
	}
	m[device] = time.Now()
	p.mu.Unlock()
	metrics.PopularityRecord.WithLabelValues(snap).Inc()
}

func (p *Popularity) Count(snap string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	cutoff := time.Now().Add(-p.ttl)
	n := 0
	for _, t := range p.seen[snap] {
		if t.After(cutoff) {
			n++
		}
	}
	return n
}

var (
	popularityDevicesDesc = prometheus.NewDesc(
		"store_popularity_devices_active",
		"Unique devices active within the TTL window, by snap.",
		[]string{"snap"}, nil,
	)
	popularityDevicesUniqueDesc = prometheus.NewDesc(
		"store_popularity_devices_unique",
		"Total unique devices active within the TTL window across all snaps.",
		nil, nil,
	)
)

func (p *Popularity) Describe(ch chan<- *prometheus.Desc) {
	ch <- popularityDevicesDesc
	ch <- popularityDevicesUniqueDesc
}

func (p *Popularity) Collect(ch chan<- prometheus.Metric) {
	p.mu.Lock()
	defer p.mu.Unlock()
	cutoff := time.Now().Add(-p.ttl)
	unique := make(map[string]struct{})
	for snap, devs := range p.seen {
		n := 0
		for dev, t := range devs {
			if t.After(cutoff) {
				n++
				unique[dev] = struct{}{}
			}
		}
		ch <- prometheus.MustNewConstMetric(popularityDevicesDesc, prometheus.GaugeValue, float64(n), snap)
	}
	ch <- prometheus.MustNewConstMetric(popularityDevicesUniqueDesc, prometheus.GaugeValue, float64(len(unique)))
}
