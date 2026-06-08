//go:build !headless

package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/danielriddell21/gambit/internal/game"
	applog "github.com/danielriddell21/gambit/internal/log"
	"github.com/danielriddell21/gambit/internal/ui"
)

// present renders the game in an Ebiten window.
func present(g *game.Game, logger *applog.Logger, delay time.Duration) error {
	cfg := ui.DefaultConfig()
	cfg.MoveDelay = delay

	gui, err := ui.New(g, logger, cfg)
	if err != nil {
		return err
	}

	ebiten.SetWindowSize(gui.BoardPixels(), gui.BoardPixels())
	ebiten.SetWindowTitle("gambit")
	return ebiten.RunGame(gui)
}
