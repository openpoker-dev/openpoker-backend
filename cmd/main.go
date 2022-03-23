package main

import (
	"log"
	"net/http"
)

var (
	htmlContent = []byte(`hello world`)
	version     = "v0.0.1"
)

func index(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Header.Get("User-Agent"))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(htmlContent)
}

func main() {
	log.Println("running openpoker: " + version)

	http.HandleFunc("/", index)
	http.ListenAndServe(":8080", nil)
}
