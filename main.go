package main

import (
	"flag"
	"fmt"

	"github.com/cthtrifork/kglance/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("kglance", version.Version)
		return
	}
}
