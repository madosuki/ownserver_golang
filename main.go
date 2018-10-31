package main

import "net/http"
import "./pkg"

func main() {
	i := new(pkg.Server)
	http.HandleFunc("/", i.Handler)
	http.ListenAndServe(":8080", nil)
}
