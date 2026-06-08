// Command gambit runs a chess game between two agents and logs each move to the
// terminal. By default it renders the game with an Ebiten window; built with
// the "headless" tag it runs without any GUI (and without linking Ebiten),
// which suits batch strategy experiments and display-less CI.
//
//	go run ./cmd/gambit                 # GUI
//	go run -tags headless ./cmd/gambit  # terminal only
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
			return err
		}
		start = b
	}

	white, err := agent.New(c.white, agent.Options{Seed: c.seed, Depth: c.depth})
	if err != nil {
		return err
	}
	black, err := agent.New(c.black, agent.Options{Seed: c.seed + 1, Depth: c.depth})
	if err != nil {
		return err
	}

	g := game.New(game.Players{White: white, Black: black}, start)
	logger := applog.New()
	fmt.Printf("gambit: white=%s black=%s\n", white.Name(), black.Name())

	return present(g, logger, c)
}
