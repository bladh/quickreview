package main

import (
	"flag"
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"

	"github.com/bladh/quickreview/internal/ui"
)

// Set via -ldflags at release build time, e.g. -X main.version=v0.1.0
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}

	path := ""
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	} else if cwd, err := os.Getwd(); err == nil {
		path = cwd
	}

	a := app.NewWithID("ninja.bladh.quickreview")
	ui.Run(a, path)
}
