// Command gambit runs a chess game between two agents and logs each move to the
// terminal. By default it runs headless (no GUI, Ebiten not linked), which suits
// batch strategy experiments and display-less CI; built with the "ebiten" tag it
// also renders the game in an Ebiten window.
//
//	go run ./cmd/gambit                # terminal only (default)
//	go run -tags ebiten ./cmd/gambit   # GUI window
package main

import (
	"fmt"
	"os"

	"github.com/danielriddell21/gambit/internal/cli"
)

// version is the build version, overridden at release time via
// -ldflags "-X main.version=...". It defaults to "dev" for local builds.
var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
