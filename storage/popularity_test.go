package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPopularity_Counts(t *testing.T) {
	p := NewPopularity()

	p.Record("a")
	p.Record("a")
	p.Record("b")

	assert.Equal(t, 2, p.Count("a"))
	assert.Equal(t, 1, p.Count("b"))
	assert.Equal(t, 0, p.Count("c"))
}

func TestPopularity_IgnoresEmpty(t *testing.T) {
	p := NewPopularity()
	p.Record("")
	assert.Equal(t, 0, p.Count(""))
}
