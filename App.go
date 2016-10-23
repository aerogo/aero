package aero

import (
	"fmt"
	"strconv"

	"github.com/buaazp/fasthttprouter"
	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
)

// Application represents a single web service.
type Application struct {
	Config Configuration

	root   string
	router *fasthttprouter.Router
}

// New ...
func New() *Application {
	app := new(Application)
	app.root = ""
	app.router = fasthttprouter.New()

	// Default configuration
	app.Config.GZip = true
	app.Config.GZipCache = true
	app.Config.Ports.HTTP = 4000

	return app
}

// Get registers your function to be called when a certain path has been requested.
func (app *Application) Get(path string, handle Handle) {
	app.router.GET(path, func(fasthttpContext *fasthttp.RequestCtx, params fasthttprouter.Params) {
		ctx := Context{
			App:        app,
			Params:     params,
			requestCtx: fasthttpContext,
		}

		handle(&ctx)
	})
}

// Run calls app.Load() and app.Listen().
func (app *Application) Run() {
	app.Listen()
}

// Listen starts the server.
func (app *Application) Listen() {
	fmt.Println("Server running on:", color.GreenString("http://localhost:"+strconv.Itoa(app.Config.Ports.HTTP)))
	err := fasthttp.ListenAndServe(":"+strconv.Itoa(app.Config.Ports.HTTP), app.router.Handler)

	if err != nil {
		panic(err)
	}
}
