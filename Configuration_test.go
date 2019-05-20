package aero_test

import (
	"testing"

	"github.com/aerogo/aero"
	qt "github.com/frankban/quicktest"
)

func TestLoadConfig(t *testing.T) {
	config, err := aero.LoadConfig("testdata/config.json")

	// Verify configuration
	c := qt.New(t)
	c.Assert(err, qt.IsNil)
	c.Assert(config, qt.Not(qt.IsNil))
	c.Assert(config.Title, qt.Not(qt.Equals), "")
}
