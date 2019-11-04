package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.ListenAndServe(":8000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
	}))
}
