package aero_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	qt "github.com/frankban/quicktest"
)

func TestETag(t *testing.T) {
	text1 := bytes.Repeat([]byte("Hello World"), 1000000)
	text2 := bytes.Repeat([]byte("Hello Aero"), 1000000)

	etag1 := aero.ETag(text1)
	etag2 := aero.ETag(text2)

	c := qt.New(t)
	c.Assert(etag1, qt.Not(qt.Equals), "")
	c.Assert(etag2, qt.Not(qt.Equals), "")
	c.Assert(etag1, qt.Not(qt.Equals), etag2)
}

func TestETagString(t *testing.T) {
	text1 := strings.Repeat("Hello World", 1000000)
	text2 := strings.Repeat("Hello Aero", 1000000)

	etag1 := aero.ETagString(text1)
	etag2 := aero.ETagString(text2)

	c := qt.New(t)
	c.Assert(etag1, qt.Not(qt.Equals), "")
	c.Assert(etag2, qt.Not(qt.Equals), "")
	c.Assert(etag1, qt.Not(qt.Equals), etag2)
}
