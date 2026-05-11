package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPopularity_RecordCountsUniqueDevices(t *testing.T) {
	p := NewPopularity(time.Hour)

	p.Record("a", "d1")
	p.Record("a", "d2")
	p.Record("a", "d1")
	p.Record("b", "d1")

	assert.Equal(t, 2, p.Count("a"))
	assert.Equal(t, 1, p.Count("b"))
	assert.Equal(t, 0, p.Count("c"))
}

func TestPopularity_IgnoresEmpty(t *testing.T) {
	p := NewPopularity(time.Hour)

	p.Record("", "d1")
	p.Record("a", "")

	assert.Equal(t, 0, p.Count("a"))
	assert.Equal(t, 0, p.Count(""))
}
