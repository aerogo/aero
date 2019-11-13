package aero_test

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	"github.com/akyoto/assert"
)

type routeDefinition struct {
	method string
	path   string
}

func TestRouterStatic(t *testing.T) {
	router := aero.Router{}
	page := func(aero.Context) error { return nil }
	assert.Nil(t, router.Find("GET", "/"))
	assert.Nil(t, router.Find("GET", "/日"))
	assert.Nil(t, router.Find("GET", "/日本"))
	assert.Nil(t, router.Find("GET", "/日本語"))

	router.Add("GET", "/", page)
	assert.NotNil(t, router.Find("GET", "/"))
	assert.Nil(t, router.Find("GET", "/日"))
	assert.Nil(t, router.Find("GET", "/日本"))
	assert.Nil(t, router.Find("GET", "/日本語"))

	router.Add("GET", "/日本語", page)
	assert.NotNil(t, router.Find("GET", "/"))
	assert.Nil(t, router.Find("GET", "/日"))
	assert.Nil(t, router.Find("GET", "/日本"))
	assert.NotNil(t, router.Find("GET", "/日本語"))
}

func TestRouterParameters(t *testing.T) {
	router := aero.Router{}
	page := func(aero.Context) error { return nil }

	router.Add("GET", "/", page)
	router.Add("GET", "/user", page)
	router.Add("GET", "/user/:id", page)
	router.Add("GET", "/user/:id/profile", page)
	router.Add("GET", "/user/:id/profile/:theme", page)
	router.Add("GET", "/user/:id/:something", page)
	router.Add("GET", "/admin", page)

	assert.NotNil(t, router.Find("GET", "/"))
	assert.NotNil(t, router.Find("GET", "/user"))
	assert.NotNil(t, router.Find("GET", "/user/123"))
	assert.NotNil(t, router.Find("GET", "/user/123/profile"))
	assert.NotNil(t, router.Find("GET", "/user/123/profile/456"))
	assert.NotNil(t, router.Find("GET", "/user/123/456"))
	assert.NotNil(t, router.Find("GET", "/admin"))

	assert.Nil(t, router.Find("GET", "/x"))
	assert.Nil(t, router.Find("GET", "/user/123/456/x"))
	assert.Nil(t, router.Find("GET", "/admin/x"))
}

func TestRouterWildcards(t *testing.T) {
	router := aero.Router{}
	page := func(aero.Context) error { return nil }

	router.Add("GET", "/", page)
	router.Add("GET", "/images", page)
	router.Add("GET", "/images/*file", page)
	router.Add("GET", "/videos/*file", page)
	router.Add("GET", "/*anything", page)

	assert.NotNil(t, router.Find("GET", "/"))
	assert.NotNil(t, router.Find("GET", "/images"))
	assert.NotNil(t, router.Find("GET", "/images/hello.webp"))
	assert.NotNil(t, router.Find("GET", "/videos/hello.webm"))
	assert.NotNil(t, router.Find("GET", "/documents/hello.txt"))
}

func TestRouterStaticData(t *testing.T) {
	router := aero.Router{}
	routes := loadRoutes("testdata/router/static.txt")
	page := func(aero.Context) error { return nil }

	for _, route := range routes {
		router.Add(route.method, route.path, page)
	}

	for _, route := range routes {
		if router.Find(route.method, route.path) == nil {
			t.Fatal(route.method + " " + route.path)
		}
	}
}

func TestRouterGitHubData(t *testing.T) {
	router := aero.Router{}
	routes := loadRoutes("testdata/router/github.txt")
	page := func(aero.Context) error { return nil }

	for _, route := range routes {
		router.Add(route.method, route.path, page)
	}

	for _, route := range routes {
		if router.Find(route.method, route.path) == nil {
			t.Fatal(route.method + " " + route.path)
		}
	}
}

func BenchmarkStaticRoutes(b *testing.B) {
	router := aero.Router{}
	routes := loadRoutes("testdata/router/static.txt")
	page := func(aero.Context) error { return nil }

	for _, route := range routes {
		router.Add(route.method, route.path, page)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			if router.Find(route.method, route.path) == nil {
				b.Fatal(route.method + " " + route.path)
			}
		}
	}

	// b.RunParallel(func(pb *testing.PB) {
	// 	for pb.Next() {
	// 		for _, route := range routes {
	// 			if router.Find(route.method, route.path) == nil {
	// 				b.Fatal(route.method + " " + route.path)
	// 			}
	// 		}
	// 	}
	// })
}

func BenchmarkGitHubRoutes(b *testing.B) {
	router := aero.Router{}
	routes := loadRoutes("testdata/router/github.txt")
	page := func(aero.Context) error { return nil }

	for _, route := range routes {
		router.Add(route.method, route.path, page)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			if router.Find(route.method, route.path) == nil {
				b.Fatal(route.method + " " + route.path)
			}
		}
	}

	// b.RunParallel(func(pb *testing.PB) {
	// 	for pb.Next() {
	// 		for _, route := range routes {
	// 			if router.Find(route.method, route.path) == nil {
	// 				b.Fatal(route.method + " " + route.path)
	// 			}
	// 		}
	// 	}
	// })
}

func loadRoutes(filePath string) []routeDefinition {
	routes := []routeDefinition{}
	f, err := os.Open(filePath)

	if err != nil {
		panic(err)
	}

	bufferedReader := bufio.NewReader(f)

	for {
		line, err := bufferedReader.ReadString('\n')

		if line != "" {
			line = strings.TrimSpace(line)
			parts := strings.Split(line, " ")
			routes = append(routes, routeDefinition{
				method: parts[0],
				path:   parts[1],
			})
		}

		if err != nil {
			break
		}
	}

	f.Close()
	return routes
}
