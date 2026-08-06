// Command demogen renders gambit's documentation media headlessly: short,
// deterministic games between agents, one frame per move. It plays the game
// and composes frames on a software canvas — the same code the window draws
// with — so it needs no display.
//
// Regenerate every asset under docs/demos with:
//
//	just demos      // or: go run ./tools/demogen
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/gambit/internal/agent"
	"github.com/danielriddell21/gambit/internal/game"
	"github.com/danielriddell21/gambit/internal/gui"
	applog "github.com/danielriddell21/gambit/internal/log"
)

const outDir = "docs/demos"

// squareSize keeps the clips small enough for a README while the pieces stay
// legible.
const squareSize = 48

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demogen:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	for _, c := range clips() {
		if err := c.record(); err != nil {
			return fmt.Errorf("%s: %w", c.name, err)
		}
	}
	return nil
}

// clip is one recorded game: who plays, how deeply, and from what seed.
type clip struct {
	name string
	// ext is the output extension; an .mp4 records video instead of a GIF.
	ext          string
	white, black string
	depth        int
	seed         int64
}

// clips is the documentation set: a search agent dismantling a random one, and
// two searchers grinding against each other.
func clips() []clip {
	return []clip{
		{name: "minimax-vs-random", ext: ".gif", white: "minimax", black: "random", depth: 4, seed: 7},
		{name: "minimax-vs-minimax", ext: ".gif", white: "minimax", black: "minimax", depth: 3, seed: 1},
	}
}

func (c clip) record() error {
	white, err := agent.New(c.white, agent.Options{Seed: c.seed, Depth: c.depth})
	if err != nil {
		return fmt.Errorf("create white agent: %w", err)
	}
	black, err := agent.New(c.black, agent.Options{Seed: c.seed + 1, Depth: c.depth})
	if err != nil {
		return fmt.Errorf("create black agent: %w", err)
	}

	cfg := gui.DefaultConfig()
	cfg.Game = game.New(game.Players{White: white, Black: black}, nil)
	cfg.Logger = applog.New()
	cfg.SquareSize = squareSize
	cfg.Rec = record.Options{Path: filepath.Join(outDir, c.name+c.ext)}
	if err := gui.Render(cfg); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	return nil
}
