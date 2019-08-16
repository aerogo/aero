package aero_test

import (
	"testing"

	"github.com/aerogo/aero"
	"github.com/akyoto/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := aero.LoadConfig("testdata/config.json")

	// Verify configuration
	assert.Nil(t, err)
	assert.NotNil(t, config)
}
