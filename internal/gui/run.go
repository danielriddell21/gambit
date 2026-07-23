//go:build ebiten

package gui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/danielriddell21/crucible/window"
)

func Available() bool { return true }

func Run(cfg Config) error {
	u, err := newGameUI(cfg)
	if err != nil {
		return fmt.Errorf("build gui: %w", err)
	}

	w, h := u.WindowSize()
	window.Configure(window.Options{Title: "gambit", Width: w, Height: h, MinWidth: w / 2, MinHeight: h / 2})
	if err := ebiten.RunGame(u); err != nil {
		return fmt.Errorf("run game: %w", err)
	}
	return nil
}
