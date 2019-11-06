package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("running")

	http.ListenAndServe(":8000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)

		if r.URL.Path == "/config" {
			fmt.Fprint(w, `{"domain": "localhost", "path": "/service1"}`)
			return
		} else if r.URL.Path == "/service1" {
			fmt.Fprint(w, "service 1")
			return
		}

		fmt.Fprintf(w, "hello world")
	}))
}
