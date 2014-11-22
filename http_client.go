package main

import (
	"net/http"
)

type loggingRoundTripper struct {
	http.RoundTripper
}

func (rt *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	logf(req.Method, "[%p] ---> %s", req, req.URL)

	resp, err := rt.RoundTripper.RoundTrip(req)
	if err != nil {
		logf("error", "[%p] xxxx %s", req, err)
		return nil, err
	}

	logf(req.Method, "[%p] <--- %s", req, resp.Status)
	return resp, err
}

func init() {
	http.DefaultClient.Transport = &loggingRoundTripper{http.DefaultTransport}
}
