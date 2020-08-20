package main

import (
	"os"

	"github.com/cicizeo/loran/cmd/loran"
)

func main() {
	cmd := loran.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
