package main

import (
	"flag"
	"fmt"
	"net/http"
)

var (
	version = flag.Int("version", 0, "version of the application")
)

func versionFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "version=%d\n", *version)
}

func headersFunc(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/version", versionFunc)
	http.HandleFunc("/headers", headersFunc)

	http.ListenAndServe(":8000", nil)
}
