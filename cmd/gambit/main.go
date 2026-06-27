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
	"time"

	"github.com/danielriddell21/gambit/internal/agent"
	"github.com/danielriddell21/gambit/internal/game"
	applog "github.com/danielriddell21/gambit/internal/log"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// version is the build version, overridden at release time via
// -ldflags "-X main.version=...". It defaults to "dev" for local builds.
var version = "dev"

func main() {
	cfg := parseFlags()
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "gambit: %v\n", err)
		fmt.Fprintf(os.Stderr, "available agents: %v\n", agent.Available())
		os.Exit(1)
	}
}

type config struct {
	white, black string
	depth        int
	iterations   int
	width        int
	seed         int64
	fen          string
	delay        time.Duration
	record       string
	square       int
}

func run(c config) error {
	var start *chess.Board
	if c.fen != "" {
		b, err := chess.ParseFEN(c.fen)
		if err != nil {
			return fmt.Errorf("parse FEN: %w", err)
		}
		start = b
	}

	// build constructs a fresh game with newly created agents. The seed is
	// offset per call so restarts of stochastic agents play out differently.
	var restarts int64
	build := func() (*game.Game, error) {
		opts := func(extra int64) agent.Options {
			return agent.Options{Seed: c.seed + restarts*2 + extra, Depth: c.depth, Width: c.width, Iterations: c.iterations}
		}
		white, err := agent.New(c.white, opts(0))
		if err != nil {
			return nil, fmt.Errorf("create white agent: %w", err)
		}
		black, err := agent.New(c.black, opts(1))
		if err != nil {
			return nil, fmt.Errorf("create black agent: %w", err)
		}
		restarts++
		var s *chess.Board
		if start != nil {
			s = start.Clone()
		}
		return game.New(game.Players{White: white, Black: black}, s), nil
	}

	g, err := build() // also validates the agent names
	if err != nil {
		return err
	}

	logger := applog.New()
	fmt.Printf("gambit: white=%s black=%s\n", c.white, c.black)

	newGame := func() *game.Game {
		ng, _ := build() // names already validated above
		return ng
	}
	return present(g, newGame, logger, c)
}
