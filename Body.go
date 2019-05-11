package aero

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/akyoto/stringutils/unsafe"
	jsoniter "github.com/json-iterator/go"
)

// Body represents a request body.
type Body struct {
	reader io.ReadCloser
}

// Reader returns an io.Reader for the request body.
func (body Body) Reader() io.ReadCloser {
	return body.reader
}

// JSON parses the body as a JSON object.
func (body Body) JSON() (interface{}, error) {
	if body.reader == nil {
		return nil, errors.New("Empty body")
	}

	decoder := jsoniter.NewDecoder(body.reader)
	defer body.reader.Close()

	var data interface{}
	err := decoder.Decode(&data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// JSONObject parses the body as a JSON object and returns a map[string]interface{}.
func (body Body) JSONObject() (map[string]interface{}, error) {
	json, err := body.JSON()

	if err != nil {
		return nil, err
	}

	data, ok := json.(map[string]interface{})

	if !ok {
		return nil, errors.New("Invalid format: Expected JSON object")
	}

	return data, nil
}

// Bytes returns a slice of bytes containing the request body.
func (body Body) Bytes() ([]byte, error) {
	data, err := ioutil.ReadAll(body.reader)
	defer body.reader.Close()

	if err != nil {
		return nil, err
	}

	return data, nil
}

// String returns a string containing the request body.
func (body Body) String() (string, error) {
	bytes, err := body.Bytes()

	if err != nil {
		return "", err
	}

	return unsafe.BytesToString(bytes), nil
}
