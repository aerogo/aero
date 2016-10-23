package aero

import (
	"fmt"

	"github.com/buaazp/fasthttprouter"
	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
)

// Application ...
type Application struct {
	root   string
	router *fasthttprouter.Router
	Config Configuration
}

// New ...
func New() *Application {
	app := new(Application)
	app.root = ""
	app.router = fasthttprouter.New()

	// Default configuration
	app.Config.GZip = true
	app.Config.GZipCache = true

	return app
}

// Get ...
func (app *Application) Get(route string, handle Handle) {
	app.router.GET(route, func(fasthttpContext *fasthttp.RequestCtx, params fasthttprouter.Params) {
		ctx := Context{
			App:        app,
			Params:     params,
			requestCtx: fasthttpContext,
		}

		handle(&ctx)
	})
}

// Run ...
func (app *Application) Run() {
	fmt.Println("Server running on:", color.GreenString("http://localhost:4000/"))
	err := fasthttp.ListenAndServe(":4000", app.router.Handler)

	if err != nil {
		panic(err)
	}
}
