package aero_test

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	qt "github.com/frankban/quicktest"
)

type routeDefinition struct {
	method string
	path   string
}

func TestRouterStatic(t *testing.T) {
	c := qt.New(t)
	router := aero.Router{}
	page := func(*aero.Context) error { return nil }
	c.Assert(router.Find("GET", "/"), qt.IsNil)

	router.Add("GET", "/", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/日本語"), qt.IsNil)

	router.Add("GET", "/日本語", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/日本語"), qt.Not(qt.IsNil))
}

func TestRouterParameters(t *testing.T) {
	c := qt.New(t)
	router := aero.Router{}
	page := func(*aero.Context) error { return nil }

	router.Add("GET", "/", page)
	router.Add("GET", "/user", page)
	router.Add("GET", "/user/:id", page)
	router.Add("GET", "/user/:id/profile", page)
	router.Add("GET", "/user/:id/profile/:theme", page)
	router.Add("GET", "/user/:id/:something", page)
	router.Add("GET", "/admin", page)

	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user/123"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user/123/profile"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user/123/profile/456"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user/123/456"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/admin"), qt.Not(qt.IsNil))
}

func TestRouterStaticData(t *testing.T) {
	router := aero.Router{}
	routes := loadRoutes("testdata/router/static.txt")
	page := func(*aero.Context) error { return nil }

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
	page := func(*aero.Context) error { return nil }

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
	page := func(*aero.Context) error { return nil }

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
	page := func(*aero.Context) error { return nil }

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
