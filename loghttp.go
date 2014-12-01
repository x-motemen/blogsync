package main

import (
	"fmt"
	"github.com/motemen/go-loghttp"
	_ "github.com/motemen/go-loghttp/global"
	"net/http"
)

func init() {
	loghttp.DefaultLogRequest = func(req *http.Request) {
		logf(req.Method, "---> %s", req.URL)
	}

	loghttp.DefaultLogResponse = func(resp *http.Response) {
		logf(fmt.Sprintf("%d", resp.StatusCode), "<--- %s", resp.Request.URL)
	}
}
