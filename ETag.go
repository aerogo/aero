package aero

import (
	"strconv"

	"github.com/OneOfOne/xxhash"
)

// ETag produces a hash for the given slice of bytes.
// It is the same hash that Aero uses for its ETag header.
func ETag(b []byte) string {
	h := xxhash.NewS64(0)
	h.Write(b)
	return strconv.FormatUint(h.Sum64(), 16)
}

// ETagString produces a hash for the given string.
// It is the same hash that Aero uses for its ETag header.
func ETagString(b string) string {
	h := xxhash.NewS64(0)
	h.WriteString(b)
	return strconv.FormatUint(h.Sum64(), 16)
}
