package main

import (
	"github.com/dk-a-dev/termf1/cmd"
)

// version is set at build time via -ldflags "-X main.version=v1.2.3"
var version = "dev"

func main() {
	cmd.Execute(version)
}
