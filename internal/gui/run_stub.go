//go:build !ebiten

package gui

import (
	"context"
	"fmt"
)

// Available reports whether the Ebiten window is compiled in.
func Available() bool { return false }

// Run plays the game to completion without a GUI, logging every move and the
// final result. The window is built with the "ebiten" tag; this headless path
// is the default build's behaviour. The restart factory and GUI-only knobs in
// cfg are ignored here.
func Run(cfg Config) error {
	if err := cfg.Game.Run(context.Background(), cfg.Logger.Move); err != nil {
		return fmt.Errorf("run game: %w", err)
	}
	cfg.Logger.Result(cfg.Game.Result(), cfg.Game.DrawReason())
	return nil
}
