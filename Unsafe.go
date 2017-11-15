package aero

import (
	"reflect"
	"unsafe"
)

// BytesToStringUnsafe converts a byte slice to a string.
// It's fast, but not safe. Use it only if you know what you're doing.
func BytesToStringUnsafe(b []byte) string {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	strHeader := reflect.StringHeader{
		Data: bytesHeader.Data,
		Len:  bytesHeader.Len,
	}
	return *(*string)(unsafe.Pointer(&strHeader))
}

// StringToBytesUnsafe converts a string to a byte slice.
// It's fast, but not safe. Use it only if you know what you're doing.
func StringToBytesUnsafe(s string) []byte {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bytesHeader := reflect.SliceHeader{
		Data: strHeader.Data,
		Len:  strHeader.Len,
		Cap:  strHeader.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bytesHeader))
}
