package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnapNameFromYaml(t *testing.T) {
	name, err := snapNameFromYaml([]byte("name: testapp1\nsummary: t\n"))
	assert.NoError(t, err)
	assert.Equal(t, "testapp1", name)

	name, err = snapNameFromYaml([]byte("summary: only\n"))
	assert.Error(t, err)
	assert.Empty(t, name)

	name, err = snapNameFromYaml([]byte(`name: "quoted"`))
	assert.NoError(t, err)
	assert.Equal(t, "quoted", name)
}

func TestParseSnapName(t *testing.T) {
	name, version, arch, err := parseSnapName("testapp1_3_amd64.snap")
	assert.NoError(t, err)
	assert.Equal(t, "testapp1", name)
	assert.Equal(t, "3", version)
	assert.Equal(t, "amd64", arch)

	_, _, _, err = parseSnapName("malformed.snap")
	assert.Error(t, err)
}

func TestDebArch(t *testing.T) {
	assert.Equal(t, "amd64", debArch("amd64"))
	assert.Equal(t, "arm64", debArch("arm64"))
	assert.Equal(t, "armhf", debArch("arm"))
}

func TestResolveAppPath(t *testing.T) {
	assert.Equal(t, "test/testapp1/meta/snap.yaml",
		resolveAppPath("test/testapp1", "meta/snap.yaml"))
	assert.Equal(t, "/abs/path",
		resolveAppPath("test/testapp1", "/abs/path"))
}

func TestDeriveSnapFile(t *testing.T) {
	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "version"), []byte("7\n"), 0644))
	yaml := []byte("name: myapp\nsummary: x\n")
	got, err := deriveSnapFile(dir, yaml)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "myapp_7_"+debArch(getRuntimeArch())+".snap"), got)
}

func TestDeriveSnapFile_MissingVersion(t *testing.T) {
	dir := t.TempDir()
	yaml := []byte("name: myapp\n")
	_, err := deriveSnapFile(dir, yaml)
	assert.Error(t, err)
}

func TestValidateSnapYamlMatches(t *testing.T) {
	assert.NoError(t, validateSnapYamlMatches([]byte("name: foo\n"), "foo"))
	assert.Error(t, validateSnapYamlMatches([]byte("name: foo\n"), "bar"))
	assert.Error(t, validateSnapYamlMatches([]byte("summary: only\n"), "foo"))
}

func TestReadIcon_NotExists(t *testing.T) {
	dir := t.TempDir()
	icon, ok, err := readIcon(filepath.Join(dir, "missing.png"))
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, icon)
}

func TestReadIcon_Present(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "icon.png")
	assert.NoError(t, os.WriteFile(path, []byte{0x89, 0x50}, 0644))
	icon, ok, err := readIcon(path)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.NotEmpty(t, icon)
}

// getRuntimeArch is wrapped so tests don't need a runtime import.
func getRuntimeArch() string {
	return runtimeGOARCH()
}
