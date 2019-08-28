// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	file, err := os.Create("version.go")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	git := exec.Command("git", "describe")
	output, err := git.Output()
	if err != nil {
		// try something else
		git = exec.Command("git", "rev-parse", "--short", "HEAD")
		output, err = git.Output()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	ver := strings.ReplaceAll(string(output), "\n", "")
	t := time.Now().Format(time.UnixDate)

	fmt.Fprintf(file, contents, ver, t)
}

const contents = `// Code generated by go generate; DO NOT EDIT.

//go:generate go run generate_version.go

package version

// git version info
const Version = "%s"

// time this package was built (format: time.UnixDate)
const BuildTime = "%s"`