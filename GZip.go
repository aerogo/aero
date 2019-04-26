package aero

import (
	"compress/gzip"
	"io"
	"sync"
)

// gzipWriterPool contains all of our gzip writers.
// We use a pool so that every request can re-use writers.
var gzipWriterPool sync.Pool

// acquireGZipWriter will return a clean gzip writer from the pool.
func acquireGZipWriter(response io.Writer) *gzip.Writer {
	var writer *gzip.Writer
	obj := gzipWriterPool.Get()

	if obj == nil {
		writer, _ = gzip.NewWriterLevel(response, gzip.BestCompression)
		return writer
	}

	writer = obj.(*gzip.Writer)
	writer.Reset(response)
	return writer
}
