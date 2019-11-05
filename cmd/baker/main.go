package main

import (
	"net/http"

	"github.com/alinz/baker/gateway"
)

func main() {
	proxy := gateway.NewHandler()
	http.ListenAndServe(":8080", proxy)
}
