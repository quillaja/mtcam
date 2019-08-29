// +build ignore

package main

import (
	"net/http"
	"os"

	"github.com/shurcooL/vfsgen"
)

func main() {
	if len(os.Args) < 2 {
		panic("no client directory")
	}

	clientDir := os.Args[1]
	err := vfsgen.Generate(http.Dir(clientDir), vfsgen.Options{
		Filename:     "client.go",
		VariableName: "client",
	})
	if err != nil {
		panic(err)
	}
}
