package aero_test

import (
	"testing"

	"github.com/aerogo/aero"
	qt "github.com/frankban/quicktest"
)

func TestRouterAdd(t *testing.T) {
	c := qt.New(t)
	router := aero.Router{}
	page := func(*aero.Context) error { return nil }
	c.Assert(router.Find("GET", "/"), qt.IsNil)

	router.Add("GET", "/", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog"), qt.IsNil)
	c.Assert(router.Find("GET", "/blog/post"), qt.IsNil)
	c.Assert(router.Find("GET", "/user"), qt.IsNil)

	router.Add("GET", "/blog", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog/post"), qt.IsNil)
	c.Assert(router.Find("GET", "/user"), qt.IsNil)

	router.Add("GET", "/blog/post", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog/post"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user"), qt.IsNil)

	router.Add("GET", "/user", page)
	c.Assert(router.Find("GET", "/"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/blog/post"), qt.Not(qt.IsNil))
	c.Assert(router.Find("GET", "/user"), qt.Not(qt.IsNil))
}
