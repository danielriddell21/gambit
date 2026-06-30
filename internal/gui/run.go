//go:build ebiten

package gui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

func Available() bool { return true }

func Run(cfg Config) error {
	u, err := newGameUI(cfg)
	if err != nil {
		return fmt.Errorf("build gui: %w", err)
	}

	w, h := u.WindowSize()
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("gambit")
	if err := ebiten.RunGame(u); err != nil {
		return fmt.Errorf("run game: %w", err)
	}
	return nil
}
