package aero_test

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	qt "github.com/frankban/quicktest"
)

func TestRouterStaticRoutes(t *testing.T) {
	c := qt.New(t)
	router := aero.Router{}
	page := func(*aero.Context) error { return nil }
	c.Assert(router.Find("GET", "/"), qt.IsNil)

	router.Add("GET", "/", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog"), qt.IsNil)
	c.Assert(router.Find("GET", "/blog/post"), qt.IsNil)
	c.Assert(router.Find("GET", "/user"), qt.IsNil)
	c.Assert(router.Find("GET", "/日本語"), qt.IsNil)

	router.Add("GET", "/blog", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog/post"), qt.IsNil)
	c.Assert(router.Find("GET", "/user"), qt.IsNil)
	c.Assert(router.Find("GET", "/日本語"), qt.IsNil)

	router.Add("GET", "/blog/post", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog/post"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user"), qt.IsNil)
	c.Assert(router.Find("GET", "/日本語"), qt.IsNil)

	router.Add("GET", "/user", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog/post"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/日本語"), qt.IsNil)

	router.Add("GET", "/日本語", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog/post"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/日本語"), qt.Not(qt.IsNil))
}

type routeDefinition struct {
	method string
	path   string
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

func BenchmarkStaticRoutes(b *testing.B) {
	router := aero.Router{}
	routes := loadRoutes("testdata/router/static.txt")
	page := func(*aero.Context) error { return nil }

	for _, route := range routes {
		router.Add(route.method, route.path, page)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, route := range routes {
				router.Find(route.method, route.path)
			}
		}
	})
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
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, route := range routes {
				router.Find(route.method, route.path)
			}
		}
	})
}
