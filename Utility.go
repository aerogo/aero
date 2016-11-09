package aero

import "github.com/valyala/fasthttp"

// Headers is a synonym for map[string]string.
type Headers map[string]string

// Get peforms a GET request and returns the response body.
func Get(url string) (string, error) {
	var client fasthttp.Client
	var req fasthttp.Request
	var res fasthttp.Response

	req.SetRequestURI(url)
	httpError := client.Do(&req, &res)

	return string(res.Body()), httpError
}

// GetWithHeaders peforms a GET request with the specified headers and returns the response body.
func GetWithHeaders(url string, headers Headers) (string, error) {
	var client fasthttp.Client
	var req fasthttp.Request
	var res fasthttp.Response

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	req.SetRequestURI(url)
	httpError := client.Do(&req, &res)

	return string(res.Body()), httpError
}

// Post peforms a POST request and returns the response body.
// func Post(url string, raw string) (string, error) {
// 	var client fasthttp.Client
// 	var req fasthttp.Request
// 	var res fasthttp.Response

// 	req.SetRequestURI(url)
// 	httpError := client.Do(&req, &res)
// 	client.Post()

// 	return string(res.Body()), httpError
// }
