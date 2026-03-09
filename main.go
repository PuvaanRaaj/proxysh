package main

import (
	"github.com/PuvaanRaaj/proxysh/cmd"
	"github.com/PuvaanRaaj/proxysh/update"
)

// version is set at build time via -ldflags "-X main.version=x.y.z"
var version = "dev"

func main() {
	go update.Check(version)
	cmd.SetVersion(version)
	cmd.Execute()
}
