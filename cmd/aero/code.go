package main

const mainCode = `package main

import (
	"github.com/aerogo/aero"
)

func main() {
	app := aero.New()
	configure(app).Run()
}

func configure(app *aero.Application) *aero.Application {
	app.Get("/", func(ctx *aero.Context) string {
		return ctx.Text("Hello World")
	})

	return app
}
`
