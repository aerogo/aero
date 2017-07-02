package aero

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"crypto/sha256"

	"encoding/base64"

	"github.com/aerogo/session"
	memstore "github.com/aerogo/session-store-memory"
	"github.com/fatih/color"
	"github.com/julienschmidt/httprouter"
	cache "github.com/patrickmn/go-cache"
)

const (
	gzipCacheDuration = 5 * time.Minute
	gzipCacheCleanup  = 1 * time.Minute
)

// Middleware ...
type Middleware func(*Context, func())

// Application represents a single web service.
type Application struct {
	root string

	Config   *Configuration
	Layout   func(*Context, string) string
	Sessions session.Manager
	Security ApplicationSecurity
	Servers  [2]*http.Server

	router *httprouter.Router
	routes struct {
		GET  []string
		POST []string
	}
	routeTests map[string][]string
	gzipCache  *cache.Cache
	start      time.Time
	rewrite    func(*RewriteContext)

	middleware []Middleware

	css            string
	cssHash        string
	cssReplacement string

	contentSecurityPolicy string
}

// New creates a new application.
func New() *Application {
	app := new(Application)
	app.start = time.Now()
	app.router = httprouter.New()
	app.routeTests = make(map[string][]string)
	app.gzipCache = cache.New(gzipCacheDuration, gzipCacheCleanup)
	app.Layout = func(ctx *Context, content string) string {
		return content
	}
	app.Config = new(Configuration)
	app.Config.Reset()
	app.Load()

	// Default session store: Memory
	app.Sessions.Store = memstore.New()

	return app
}

// Get registers your function to be called when a certain GET path has been requested.
func (app *Application) Get(path string, handle Handle) {
	app.routes.GET = append(app.routes.GET, path)
	app.router.GET(path, app.createRouteHandler(path, handle))
}

// Post registers your function to be called when a certain POST path has been requested.
func (app *Application) Post(path string, handle Handle) {
	app.routes.POST = append(app.routes.POST, path)
	app.router.POST(path, app.createRouteHandler(path, handle))
}

// createRouteHandler creates a handler function for httprouter.
func (app *Application) createRouteHandler(path string, handle Handle) httprouter.Handle {
	return func(response http.ResponseWriter, request *http.Request, params httprouter.Params) {
		// Create context.
		ctx := Context{
			App:        app,
			StatusCode: http.StatusOK,
			Request:    request,
			Response:   response,
			params:     params,
			start:      time.Now(),
		}

		// The last part of the call chain will send the actual response.
		lastPartOfCallChain := func() {
			data := handle(&ctx)
			ctx.Respond(data)
		}

		// Declare the type of generateNext so that we can define it recursively in the next part.
		var generateNext func(index int) func()

		// Create a function that returns a bound function next()
		// which can be used as the 2nd parameter in the call chain.
		generateNext = func(index int) func() {
			if index == len(app.middleware) {
				return lastPartOfCallChain
			}

			return func() {
				app.middleware[index](&ctx, generateNext(index+1))
			}
		}

		generateNext(0)()
	}
}

// Ajax calls app.Get for both /route and /_/route
func (app *Application) Ajax(path string, handle Handle) {
	app.Get("/_"+path, handle)
	app.Get(path, func(ctx *Context) string {
		page := handle(ctx)
		html := app.Layout(ctx, page)
		html = strings.Replace(html, "</head><body", app.cssReplacement, 1)
		return html
	})
}

// Run starts your application.
func (app *Application) Run() {
	app.TestManifest()
	app.TestRoutes()
	app.Listen()
	app.Wait()
	app.Shutdown()
}

// Use adds middleware to your middleware chain.
func (app *Application) Use(middlewares ...Middleware) {
	app.middleware = append(app.middleware, middlewares...)
}

// Load loads the application configuration from config.json.
func (app *Application) Load() {
	var err error
	app.Config, err = LoadConfig("config.json")

	if err != nil {
		color.Red(err.Error())
	}
}

// Listen starts the server.
func (app *Application) Listen() {
	if app.Security.Key != "" && app.Security.Certificate != "" {
		go func() {
			app.serveHTTPS(":" + strconv.Itoa(app.Config.Ports.HTTPS))
		}()

		fmt.Println("Server running on:", color.GreenString("https://localhost:"+strconv.Itoa(app.Config.Ports.HTTPS)))
	} else {
		fmt.Println("Server running on:", color.GreenString("http://localhost:"+strconv.Itoa(app.Config.Ports.HTTP)))
	}

	go func() {
		app.serveHTTP(":" + strconv.Itoa(app.Config.Ports.HTTP))
	}()
}

// Wait will make the process wait until it is killed.
func (app *Application) Wait() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}

// Shutdown will gracefully shut down the server.
func (app *Application) Shutdown() {
	for _, server := range app.Servers {
		if server == nil {
			continue
		}

		server.Shutdown(context.Background())
	}
}

// // listen listens on the specified host and port.
// func (app *Application) listen(port int) net.Listener {
// 	address := ":" + strconv.Itoa(port)

// 	listener, bindError := net.Listen("tcp", address)

// 	if bindError != nil {
// 		panic(bindError)
// 	}

// 	return listener
// }

// Rewrite sets the URL rewrite function.
func (app *Application) Rewrite(rewrite func(*RewriteContext)) {
	app.rewrite = rewrite
}

// SetStyle ...
func (app *Application) SetStyle(css string) {
	app.css = css

	hash := sha256.Sum256([]byte(css))
	app.cssHash = base64.StdEncoding.EncodeToString(hash[:])
	app.cssReplacement = "<style>" + app.css + "</style></head><body"
	app.contentSecurityPolicy = "default-src 'none'; img-src https:; script-src 'self'; style-src 'sha256-" + app.cssHash + "'; font-src https:; manifest-src 'self'; child-src https:; connect-src https: wss:"
}

// StartTime ...
func (app *Application) StartTime() time.Time {
	return app.start
}

// Handler returns the request handler.
func (app *Application) Handler() http.Handler {
	router := app.router
	rewrite := app.rewrite

	if rewrite != nil {
		return &rewriteHandler{
			rewrite: rewrite,
			router:  router,
		}
	}

	return router
}

// serveHTTP serves requests from the given listener.
func (app *Application) serveHTTP(address string) {
	server := &http.Server{
		Addr:    address,
		Handler: app.Handler(),
	}

	app.Servers[0] = server

	// This will block the calling goroutine until the server shuts down.
	serveError := server.ListenAndServe()

	if serveError != nil && strings.Index(serveError.Error(), "closed") == -1 {
		panic(serveError)
	}
}

// serveHTTPS serves requests from the given listener.
func (app *Application) serveHTTPS(address string) {
	server := &http.Server{
		Addr:    address,
		Handler: app.Handler(),
	}

	app.Servers[1] = server

	// This will block the calling goroutine until the server shuts down.
	serveError := server.ListenAndServeTLS(app.Security.Certificate, app.Security.Key)

	if serveError != nil && strings.Index(serveError.Error(), "closed") == -1 {
		panic(serveError)
	}
}

// Test tests the given URI paths when the application starts.
func (app *Application) Test(route string, paths []string) {
	app.routeTests[route] = paths
}

// TestManifest tests your application's manifest.
func (app *Application) TestManifest() {
	manifest := app.Config.Manifest

	if len(manifest.ShortName) >= 12 {
		color.Yellow("The short name of your application should have less than 12 characters")
	}
}

// TestRoutes tests your application's routes.
func (app *Application) TestRoutes() {
	fmt.Println(strings.Repeat("-", 80))

	go func() {
		sort.Strings(app.routes.GET)

		for _, route := range app.routes.GET {
			// Ajax routes
			if strings.HasPrefix(route, "/_") {
				continue
			}

			testRoutes, exists := app.routeTests[route]

			if exists {
				for _, testRoute := range testRoutes {
					app.TestRoute(route, testRoute)
				}

				continue
			}

			// Routes with parameters
			if strings.Contains(route, ":") {
				color.Yellow(route)
				continue
			}

			app.TestRoute(route, route)
		}

		// json, _ := Post("https://html5.validator.nu/?out=json").Header("Content-Type", "text/html; charset=utf-8").Header("Content-Encoding", "gzip").Body(body).Send()
		// fmt.Println(json)
	}()
}

// TestRoute tests the given route.
func (app *Application) TestRoute(label string, route string) {
	start := time.Now()
	body, _ := Get("http://localhost:" + strconv.Itoa(app.Config.Ports.HTTP) + route).Send()
	responseTime := time.Since(start).Nanoseconds() / 1000000
	responseSize := float64(len(body)) / 1024

	faint := color.New(color.Faint).SprintFunc()

	// Response size color
	var responseSizeColor func(a ...interface{}) string

	switch {
	case responseSize < 15:
		responseSizeColor = color.New(color.FgGreen).SprintFunc()
	case responseSize < 100:
		responseSizeColor = color.New(color.FgYellow).SprintFunc()
	default:
		responseSizeColor = color.New(color.FgRed).SprintFunc()
	}

	// Response time color
	var responseTimeColor func(a ...interface{}) string

	switch {
	case responseTime < 10:
		responseTimeColor = color.New(color.FgGreen).SprintFunc()
	case responseTime < 100:
		responseTimeColor = color.New(color.FgYellow).SprintFunc()
	default:
		responseTimeColor = color.New(color.FgRed).SprintFunc()
	}

	fmt.Printf("%-67s %s %s %s %s\n", color.BlueString(label), responseSizeColor(fmt.Sprintf("%6.0f", responseSize)), faint("KB"), responseTimeColor(fmt.Sprintf("%7d", responseTime)), faint("ms"))
}
