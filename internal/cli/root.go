// Package cli wires together the root Cobra command for gambit.
package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/danielriddell21/gambit/internal/agent"
	"github.com/danielriddell21/gambit/internal/game"
	"github.com/danielriddell21/gambit/internal/gui"
	applog "github.com/danielriddell21/gambit/internal/log"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// config holds the resolved command-line options for a run.
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

// Execute builds and runs the root command. Returns non-nil on error.
func Execute(version string) error {
	c := config{}
	root := &cobra.Command{
		Use:           "gambit",
		Short:         "Two chess agents play each other, logged in algebraic notation",
		Long:          "gambit runs a chess game between two agents and logs each move.\n\nBy default it runs headless (no GUI); built with the \"ebiten\" tag it also renders the game in a window.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(c)
		},
	}

	f := root.Flags()
	f.StringVar(&c.white, "white", "minimax", "white agent strategy")
	f.StringVar(&c.black, "black", "random", "black agent strategy")
	f.IntVar(&c.depth, "depth", 4, "search depth/horizon for depth-limited agents")
	f.IntVar(&c.width, "width", 8, "beam width for the beam agent")
	f.IntVar(&c.iterations, "iterations", 20000, "playout budget for the mcts agent")
	f.Int64Var(&c.seed, "seed", time.Now().UnixNano(), "RNG seed for stochastic agents")
	f.StringVar(&c.fen, "fen", "", "starting position FEN (default: standard start)")
	f.DurationVar(&c.delay, "delay", 400*time.Millisecond, "pause between moves in the GUI")
	f.StringVar(&c.record, "record", "", "record the game to this GIF path, then exit (GUI only)")
	f.IntVar(&c.square, "square", 80, "board square size in pixels (GUI only)")

	root.AddCommand(completionCmd())

	if err := root.Execute(); err != nil {
		return fmt.Errorf("gambit: %w\navailable agents: %v", err, agent.Available())
	}
	return nil
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
	cfg := gui.DefaultConfig()
	cfg.Game = g
	cfg.NewGame = newGame
	cfg.Logger = logger
	cfg.MoveDelay = c.delay
	cfg.RecordPath = c.record
	if c.square > 0 {
		cfg.SquareSize = c.square
	}
	if err := gui.Run(cfg); err != nil {
		return fmt.Errorf("run gui: %w", err)
	}
	return nil
}
