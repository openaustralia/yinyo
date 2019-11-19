package main

import (
	"log"

	"github.com/openaustralia/morph-ng/cmd/clay/cmd"
)

func main() {
	// Show the source of the error with the standard logger. Don't show date & time
	log.SetFlags(log.Lshortfile)

	cmd.Execute()
}
