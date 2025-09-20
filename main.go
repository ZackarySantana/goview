package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/zackarysantana/goview/template"
)

func main() {
	http.Handle("/", templ.Handler(template.Index()))

	http.Handle("/assets/",
		http.StripPrefix("/assets",
			http.FileServer(http.Dir("assets"))))

	fmt.Println("Listening on :3000 (the proxy is on :7331)")
	http.ListenAndServe(":3000", nil)
}
