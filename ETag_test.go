package aero_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aerogo/aero"
	"github.com/akyoto/assert"
)

func TestETag(t *testing.T) {
	text1 := bytes.Repeat([]byte("Hello World"), 1000000)
	text2 := bytes.Repeat([]byte("Hello Aero"), 1000000)

	etag1 := aero.ETag(text1)
	etag2 := aero.ETag(text2)

	assert.NotEqual(t, etag1, "")
	assert.NotEqual(t, etag2, "")
	assert.NotEqual(t, etag1, etag2)
}

func TestETagString(t *testing.T) {
	text1 := strings.Repeat("Hello World", 1000000)
	text2 := strings.Repeat("Hello Aero", 1000000)

	etag1 := aero.ETagString(text1)
	etag2 := aero.ETagString(text2)

	assert.NotEqual(t, etag1, "")
	assert.NotEqual(t, etag2, "")
	assert.NotEqual(t, etag1, etag2)
}
