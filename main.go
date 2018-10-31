package main

import "net/http"
import "./pkg"

func main() {
	i := pkg.GetInstance()
	http.HandleFunc("/", i.Handler)
	http.ListenAndServe(":8080", nil)
}
