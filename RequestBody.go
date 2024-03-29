package aero

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/akyoto/stringutils/unsafe"
)

// RequestBody represents a request body.
type RequestBody struct {
	reader io.ReadCloser
}

// Reader returns an io.Reader for the request body.
func (body RequestBody) Reader() io.ReadCloser {
	return body.reader
}

// JSON parses the body as a JSON object.
func (body RequestBody) JSON() (interface{}, error) {
	if body.reader == nil {
		return nil, ErrEmptyBody
	}

	decoder := json.NewDecoder(body.reader)
	defer body.reader.Close()

	var data interface{}
	err := decoder.Decode(&data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// JSONObject parses the body as a JSON object and returns a map[string]interface{}.
func (body RequestBody) JSONObject() (map[string]interface{}, error) {
	json, err := body.JSON()

	if err != nil {
		return nil, err
	}

	data, ok := json.(map[string]interface{})

	if !ok {
		return nil, ErrExpectedJSONObject
	}

	return data, nil
}

// Bytes returns a slice of bytes containing the request body.
func (body RequestBody) Bytes() ([]byte, error) {
	data, err := ioutil.ReadAll(body.reader)
	defer body.reader.Close()

	if err != nil {
		return nil, err
	}

	return data, nil
}

// String returns a string containing the request body.
func (body RequestBody) String() (string, error) {
	bytes, err := body.Bytes()

	if err != nil {
		return "", err
	}

	return unsafe.BytesToString(bytes), nil
}
