package main

import (
	"fmt"
	"net/http"
	"strings"
)

func main() {
	fmt.Println("running")

	http.ListenAndServe(":8000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)

		if r.URL.Path == "/config" {
			fmt.Fprint(w, `
{
	"domain": "localhost", 
	"path": "/service1", 
	"ready": true,
	"rules": { 
		"request_updaters": [
			{
				"name": "replace_path",
				"search": "/service1",
				"replace": "/",
				"times": -1
			}
		] 
	}
}`)
			return
		} else if strings.HasPrefix(r.URL.Path, "/service1") {
			fmt.Fprint(w, "service 1")
			return
		}

		fmt.Fprintf(w, "hello world")
	}))
}
