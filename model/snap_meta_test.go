package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSnapMeta_AllFields(t *testing.T) {
	m, err := ParseSnapMeta([]byte(`name: bitwarden
summary: Password manager
description: Vaultwarden snap for syncloud
type: app
`))
	assert.NoError(t, err)
	assert.Equal(t, "bitwarden", m.Name)
	assert.Equal(t, "Password manager", m.Summary)
	assert.Equal(t, "Vaultwarden snap for syncloud", m.Description)
	assert.Equal(t, "app", m.Type)
}

func TestParseSnapMeta_OnlyName(t *testing.T) {
	m, err := ParseSnapMeta([]byte("name: minimal\n"))
	assert.NoError(t, err)
	assert.Equal(t, "minimal", m.Name)
	assert.Empty(t, m.Summary)
	assert.Empty(t, m.Description)
	assert.Empty(t, m.Type)
}

func TestParseSnapMeta_TypeBase(t *testing.T) {
	m, err := ParseSnapMeta([]byte("name: platform\ntype: base\n"))
	assert.NoError(t, err)
	assert.Equal(t, "base", m.Type)
}

func TestParseSnapMeta_IgnoresUnknownKeys(t *testing.T) {
	m, err := ParseSnapMeta([]byte(`name: x
unknown_key: ignored
apps:
  test:
    command: bin/test.sh
`))
	assert.NoError(t, err)
	assert.Equal(t, "x", m.Name)
}

func TestParseSnapMeta_BadYaml(t *testing.T) {
	_, err := ParseSnapMeta([]byte("name: [unterminated"))
	assert.Error(t, err)
}

func TestParseSnapMeta_Empty(t *testing.T) {
	m, err := ParseSnapMeta([]byte(""))
	assert.NoError(t, err)
	assert.Equal(t, SnapMeta{}, m)
}
