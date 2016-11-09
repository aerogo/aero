package aero

import "github.com/valyala/fasthttp"

// Headers is a synonym for map[string]string.
type Headers map[string]string

// HTTPClientRequest ...
type HTTPClientRequest struct {
	client   fasthttp.Client
	request  *fasthttp.Request
	response *fasthttp.Response
}

// Get builds a GET request.
func Get(url string) *HTTPClientRequest {
	http := new(HTTPClientRequest)
	http.request = fasthttp.AcquireRequest()
	http.response = fasthttp.AcquireResponse()

	http.request.SetRequestURI(url)
	return http
}

// Post builds a POST request.
func Post(url string) *HTTPClientRequest {
	http := new(HTTPClientRequest)
	http.request = fasthttp.AcquireRequest()
	http.response = fasthttp.AcquireResponse()

	http.request.SetRequestURI(url)
	http.request.Header.SetMethod("POST")
	return http
}

// Header sets one HTTP header for the request.
func (http *HTTPClientRequest) Header(key string, value string) *HTTPClientRequest {
	http.request.Header.Set(key, value)
	return http
}

// Headers sets the HTTP headers for the request.
func (http *HTTPClientRequest) Headers(headers Headers) *HTTPClientRequest {
	for key, value := range headers {
		http.request.Header.Set(key, value)
	}
	return http
}

// Body sets the request body.
func (http *HTTPClientRequest) Body(raw string) *HTTPClientRequest {
	http.request.SetBodyString(raw)
	return http
}

// Send executes the request and returns the response body.
func (http *HTTPClientRequest) Send() (string, error) {
	err := http.client.Do(http.request, http.response)
	return string(http.response.Body()), err
}
