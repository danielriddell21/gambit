//go:build ebiten

package cli

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/danielriddell21/gambit/internal/game"
	applog "github.com/danielriddell21/gambit/internal/log"
	"github.com/danielriddell21/gambit/internal/ui"
)

// present renders the game in an Ebiten window. newGame rebuilds the game when
// the user restarts.
func present(g *game.Game, newGame func() *game.Game, logger *applog.Logger, c config) error {
	cfg := ui.DefaultConfig()
	cfg.MoveDelay = c.delay
	cfg.RecordPath = c.record
	if c.square > 0 {
		cfg.SquareSize = c.square
	}

	gui, err := ui.New(g, newGame, logger, cfg)
	if err != nil {
		return fmt.Errorf("build ui: %w", err)
	}

	w, h := gui.WindowSize()
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("gambit")
	if err := ebiten.RunGame(gui); err != nil {
		return fmt.Errorf("run game: %w", err)
	}
	return nil
}
