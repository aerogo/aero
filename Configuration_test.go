package aero_test

import (
	"testing"

	"github.com/aerogo/aero"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := aero.LoadConfig("test/config.json")

	// Verify configuration
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.Title)
}
