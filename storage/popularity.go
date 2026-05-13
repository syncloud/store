package storage

import "sync"

type Popularity struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewPopularity() *Popularity {
	return &Popularity{counts: map[string]int{}}
}

func (p *Popularity) Record(snap string) {
	if snap == "" {
		return
	}
	p.mu.Lock()
	p.counts[snap]++
	p.mu.Unlock()
}

func (p *Popularity) Count(snap string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.counts[snap]
}
