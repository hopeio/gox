package main

import (
	"github.com/hopeio/gox/net/http"
)

func main() {
	http.DirectorServer(":8080")
}
