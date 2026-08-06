//go:build !ebiten

package gui

import (
	"context"
	"fmt"
)

func Available() bool { return false }

func Run(cfg Config) error {
	// Recording needs no window: the board is composed in software, so demo
	// media builds anywhere, with no display.
	if cfg.Rec.Recording() {
		return Render(cfg)
	}
	if err := cfg.Game.Run(context.Background(), cfg.Logger.Move); err != nil {
		return fmt.Errorf("run game: %w", err)
	}
	cfg.Logger.Result(cfg.Game.Result(), cfg.Game.DrawReason())
	return nil
}
