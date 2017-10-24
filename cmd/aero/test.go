package main

const mainTestCode = `package main

import (
	"testing"

	"github.com/aerogo/aero"
)

func TestApp(t *testing.T) {
	app := configure(aero.New())
	println(app.Config.Title)
}
`
