package aero

import (
	"fmt"
	"net"
	"sort"
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

var sidBytes = []byte("sid")

// Application represents a single web service.
type Application struct {
	Config   *Configuration
	Layout   func(*Context, string) string
	Sessions SessionManager
	Security struct {
		Key         []byte
		Certificate []byte
	}

	css            string
	cssHash        string
	cssReplacement string
	root           string

	router          *fasthttprouter.Router
	routes          []string
	gzipCache       *cache.Cache
	start           time.Time
	rewrite         func(*RewriteContext)
	routeStatistics map[string]*RouteStatistics
}

// New creates a new application.
func New() *Application {
	app := new(Application)
	app.root = ""
	app.router = fasthttprouter.New()
	app.gzipCache = cache.New(gzipCacheDuration, gzipCacheCleanup)
	app.start = time.Now()
	app.routeStatistics = make(map[string]*RouteStatistics)
	app.showStatistics("/__/")
	app.Layout = func(ctx *Context, content string) string {
		return content
	}
	app.Config = new(Configuration)
	app.Config.Reset()
	app.Load()

	// app.Sessions.Store = NewMemoryStore()

	return app
}

// Get registers your function to be called when a certain path has been requested.
func (app *Application) Get(path string, handle Handle) {
	statistics := new(RouteStatistics)
	app.routeStatistics[path] = statistics

	app.router.GET(path, func(fasthttpContext *fasthttp.RequestCtx) {
		ctx := Context{
			App:        app,
			requestCtx: fasthttpContext,
			start:      time.Now(),
		}

		// Session cookie
		if app.Sessions.Store != nil {
			sid := fasthttpContext.Request.Header.CookieBytes(sidBytes)

			if sid != nil {
				ctx.Session = app.Sessions.Store.Get(BytesToStringUnsafe(sid))
			}

			if app.Sessions.AutoCreate && ctx.Session == nil {
				ctx.Session = app.Sessions.NewSession()

				sessionCookie := fasthttp.AcquireCookie()
				sessionCookie.SetKeyBytes(sidBytes)
				sessionCookie.SetValueBytes(ctx.Session.id)
				sessionCookie.SetHTTPOnly(true)
				sessionCookie.SetSecure(true)

				fasthttpContext.Response.Header.SetCookie(sessionCookie)
			}
		}

		// Response
		response := handle(&ctx)
		ctx.Respond(response)

		// Statistics
		responseTime := uint64(time.Since(ctx.start).Nanoseconds() / 1000000)
		atomic.AddUint64(&statistics.requestCount, 1)
		atomic.AddUint64(&statistics.responseTime, responseTime)
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

// RequestCount calculates the total number of requests made to the application.
func (app *Application) RequestCount() uint64 {
	total := uint64(0)

	for _, stats := range app.routeStatistics {
		total += atomic.LoadUint64(&stats.requestCount)
	}

	return total
}

// Test tests your application's routes.
func (app *Application) Test() {
	fmt.Println(strings.Repeat("-", 80))

	go func() {
		sort.Strings(app.routes)

		for _, route := range app.routes {
			if strings.HasPrefix(route, "/_") {
				continue
			}

			start := time.Now()
			body, _ := Get("http://localhost:" + strconv.Itoa(app.Config.Ports.HTTP) + route).Send()
			responseTime := time.Since(start).Nanoseconds() / 1000000
			responseSize := float64(len(body)) / 1024

			faint := color.New(color.Faint).SprintFunc()

			// Response size color
			var responseSizeColor func(a ...interface{}) string

			if responseSize < 15 {
				responseSizeColor = color.New(color.FgGreen).SprintFunc()
			} else if responseSize < 100 {
				responseSizeColor = color.New(color.FgYellow).SprintFunc()
			} else {
				responseSizeColor = color.New(color.FgRed).SprintFunc()
			}

			// Response time color
			var responseTimeColor func(a ...interface{}) string

			if responseTime < 10 {
				responseTimeColor = color.New(color.FgGreen).SprintFunc()
			} else if responseTime < 100 {
				responseTimeColor = color.New(color.FgYellow).SprintFunc()
			} else {
				responseTimeColor = color.New(color.FgRed).SprintFunc()
			}

			fmt.Printf("%-67s %s %s %s %s\n", color.BlueString(route), responseSizeColor(fmt.Sprintf("%6.0f", responseSize)), faint("KB"), responseTimeColor(fmt.Sprintf("%7d", responseTime)), faint("ms"))
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

// Rewrite sets the URL rewrite function.
func (app *Application) Rewrite(rewrite func(*RewriteContext)) {
	app.rewrite = rewrite
}

// Handler returns the request handler.
func (app *Application) Handler() func(*fasthttp.RequestCtx) {
	router := app.router.Handler
	rewrite := app.rewrite

	if rewrite != nil {
		return func(ctx *fasthttp.RequestCtx) {
			rewrite(&RewriteContext{ctx})
			router(ctx)
		}
	}

	return router
}

// serveHTTP serves requests from the given listener.
func (app *Application) serveHTTP(listener net.Listener) {
	server := &fasthttp.Server{
		Handler: app.Handler(),
	}

	serveError := server.Serve(listener)

	if serveError != nil {
		panic(serveError)
	}
}

// serveHTTPS serves requests from the given listener.
func (app *Application) serveHTTPS(listener net.Listener) {
	server := &fasthttp.Server{
		Handler: app.Handler(),
	}

	serveError := server.ServeTLSEmbed(listener, app.Security.Certificate, app.Security.Key)

	if serveError != nil {
		panic(serveError)
	}
}
