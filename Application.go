package aero

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"crypto/sha256"

	"encoding/base64"
	"encoding/json"

	"io/ioutil"

	"github.com/buaazp/fasthttprouter"
	"github.com/fatih/color"
	cache "github.com/patrickmn/go-cache"
	"github.com/valyala/fasthttp"
)

const (
	gzipCacheDuration = 5 * time.Minute
	gzipCacheCleanup  = 1 * time.Minute
)

// Application represents a single web service.
type Application struct {
	Config   *Configuration
	Layout   func(*Context, string) string
	Security struct {
		Key         []byte
		Certificate []byte
	}

	css            string
	cssHash        string
	cssReplacement string
	root           string
	router         *fasthttprouter.Router
	routes         []string
	gzipCache      *cache.Cache
	start          time.Time
	requestCount   uint64
}

// New creates a new application.
func New() *Application {
	app := new(Application)
	app.root = ""
	app.router = fasthttprouter.New()
	app.gzipCache = cache.New(gzipCacheDuration, gzipCacheCleanup)
	app.start = time.Now()
	app.showStatistics("/__/")
	app.Layout = func(ctx *Context, content string) string {
		return content
	}
	app.Config = new(Configuration)
	app.Config.Reset()
	app.Load()

	return app
}

// Get registers your function to be called when a certain path has been requested.
func (app *Application) Get(path string, handle Handle) {
	app.router.GET(path, func(fasthttpContext *fasthttp.RequestCtx) {
		ctx := Context{
			App:        app,
			requestCtx: fasthttpContext,
			start:      time.Now(),
		}

		response := handle(&ctx)
		ctx.Respond(response)

		atomic.AddUint64(&app.requestCount, 1)
	})

	app.routes = append(app.routes, path)
}

// Ajax calls app.Get for both /route and /_/route
func (app *Application) Ajax(path string, handle Handle) {
	app.Get("/_"+path, handle)
	app.Get(path, func(ctx *Context) string {
		page := handle(ctx)
		html := app.Layout(ctx, page)
		return strings.Replace(html, "</head><body", app.cssReplacement, 1)
	})
}

// SetStyle ...
func (app *Application) SetStyle(css string) {
	app.css = css

	hash := sha256.Sum256([]byte(css))
	app.cssHash = base64.StdEncoding.EncodeToString(hash[:])
	app.cssReplacement = "<style>" + app.css + "</style></head><body"
}

// Test tests your application's routes.
func (app *Application) Test() {
	go func() {
		for _, route := range app.routes {
			if strings.HasPrefix(route, "/_") {
				continue
			}

			body, _ := Get("http://localhost:" + strconv.Itoa(app.Config.Ports.HTTP) + route).Send()
			faint := color.New(color.Faint).SprintFunc()
			fmt.Println(color.BlueString(route), len(body)/1024, faint("KB"))
		}

		// json, _ := Post("https://html5.validator.nu/?out=json").Header("Content-Type", "text/html; charset=utf-8").Header("Content-Encoding", "gzip").Body(body).Send()
		// fmt.Println(json)
	}()
}

// Run starts your application.
func (app *Application) Run() {
	app.Test()
	app.Listen()
}

// Load loads the application data from the file system.
func (app *Application) Load() {
	config, readError := ioutil.ReadFile("config.json")

	if readError == nil {
		jsonDecodeError := json.Unmarshal(config, app.Config)

		if jsonDecodeError != nil {
			color.Red(jsonDecodeError.Error())
		}
	}
}

// Listen starts the server.
func (app *Application) Listen() {
	if app.Security.Key != nil && app.Security.Certificate != nil {
		go func() {
			httpsListener := app.listen(app.Config.Ports.HTTPS)
			app.serveHTTPS(httpsListener)
		}()

		fmt.Println("Server running on:", color.GreenString("https://localhost:"+strconv.Itoa(app.Config.Ports.HTTPS)))
	} else {
		fmt.Println("Server running on:", color.GreenString("http://localhost:"+strconv.Itoa(app.Config.Ports.HTTP)))
	}

	httpListener := app.listen(app.Config.Ports.HTTP)
	app.serveHTTP(httpListener)
}

// listen listens on the specified host and port.
func (app *Application) listen(port int) net.Listener {
	address := ":" + strconv.Itoa(port)

	listener, bindError := net.Listen("tcp", address)

	if bindError != nil {
		panic(bindError)
	}

	return listener
}

// serveHTTP serves requests from the given listener.
func (app *Application) serveHTTP(listener net.Listener) {
	server := &fasthttp.Server{
		Handler: app.router.Handler,
	}

	serveError := server.Serve(listener)

	if serveError != nil {
		panic(serveError)
	}
}

// serveHTTPS serves requests from the given listener.
func (app *Application) serveHTTPS(listener net.Listener) {
	server := &fasthttp.Server{
		Handler: app.router.Handler,
	}

	serveError := server.ServeTLSEmbed(listener, app.Security.Certificate, app.Security.Key)

	if serveError != nil {
		panic(serveError)
	}
}
