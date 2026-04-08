package main

import (
	"os"

	"dogear/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
