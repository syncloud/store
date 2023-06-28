package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSnapId_Name(t *testing.T) {
	assert.Equal(t, "app", SnapId("app.1").Name())
	assert.Equal(t, "app", SnapId("app").Name())
}

func TestSnapId_Version(t *testing.T) {
	assert.Equal(t, "1", SnapId("app.1").Version())
	assert.Equal(t, "", SnapId("app").Version())
}

func TestSnapId_Id(t *testing.T) {
	assert.Equal(t, "app.1", SnapId("app.1").Id())
}
