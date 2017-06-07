package aero

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"

	"encoding/base64"
	"encoding/json"

	"io/ioutil"

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
	Sessions SessionManager
	Security ApplicationSecurity

	router    *httprouter.Router
	routes    []string
	gzipCache *cache.Cache
	start     time.Time
	rewrite   func(*RewriteContext)

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
	app.gzipCache = cache.New(gzipCacheDuration, gzipCacheCleanup)
	app.Layout = func(ctx *Context, content string) string {
		return content
	}
	app.Config = new(Configuration)
	app.Config.Reset()
	app.Load()

	app.Sessions.Store = NewMemoryStore()

	return app
}

// Get registers your function to be called when a certain path has been requested.
func (app *Application) Get(path string, handle Handle) {
	app.routes = append(app.routes, path)

	app.router.GET(path, func(response http.ResponseWriter, request *http.Request, params httprouter.Params) {
		// Create context.
		ctx := Context{
			App:        app,
			StatusCode: http.StatusOK,
			request:    request,
			response:   response,
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
	})
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

// Run starts your application.
func (app *Application) Run() {
	app.TestManifest()
	app.TestRoutes()
	app.Listen()
}

// Use adds middleware to your middleware chain.
func (app *Application) Use(middlewares ...Middleware) {
	app.middleware = append(app.middleware, middlewares...)
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

	if app.Config.Manifest.Name == "" {
		app.Config.Manifest.Name = app.Config.Title
	}

	if app.Config.Manifest.ShortName == "" {
		app.Config.Manifest.ShortName = app.Config.Title
	}

	if app.Config.Manifest.Lang == "" {
		app.Config.Manifest.Lang = "en"
	}

	if app.Config.Manifest.Display == "" {
		app.Config.Manifest.Display = "standalone"
	}

	if app.Config.Manifest.StartURL == "" {
		app.Config.Manifest.StartURL = "/"
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

	app.serveHTTP(":" + strconv.Itoa(app.Config.Ports.HTTP))
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
	serveError := http.ListenAndServe(address, app.Handler())

	if serveError != nil {
		panic(serveError)
	}
}

// serveHTTPS serves requests from the given listener.
func (app *Application) serveHTTPS(address string) {
	serveError := http.ListenAndServeTLS(address, app.Security.Certificate, app.Security.Key, app.Handler())

	if serveError != nil {
		panic(serveError)
	}
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
