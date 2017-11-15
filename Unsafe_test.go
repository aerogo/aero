package aero_test

import (
	"testing"

	"github.com/aerogo/aero"
	"github.com/stretchr/testify/assert"
)

func TestBytesToStringUnsafe(t *testing.T) {
	out := "hello"
	in := []byte(out)

	assert.Equal(t, out, aero.BytesToStringUnsafe(in))
}

func TestStringToBytesUnsafe(t *testing.T) {
	in := "hello"
	out := []byte(in)

	assert.Equal(t, out, aero.StringToBytesUnsafe(in))
}
