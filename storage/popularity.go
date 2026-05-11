package storage

import (
	"sync"
	"time"
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
	defer p.mu.Unlock()
	m, ok := p.seen[snap]
	if !ok {
		m = make(map[string]time.Time)
		p.seen[snap] = m
	}
	m[device] = time.Now()
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
