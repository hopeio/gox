package main

import (
	"github.com/hopeio/gox/net/http"
)

// main is the program entry point.
func main() {
	http.DirectorServer(":8080")
}
