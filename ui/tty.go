package ui

import (
	"os"

	"github.com/mattn/go-isatty"
)

func SupportsANSICodes() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}
