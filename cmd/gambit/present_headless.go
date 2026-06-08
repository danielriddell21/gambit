//go:build headless

package main

import (
	"context"
	"time"

	"github.com/danielriddell21/gambit/internal/game"
	applog "github.com/danielriddell21/gambit/internal/log"
)

// present runs the game to completion without a GUI, logging every move and the
// final result. The delay is ignored in headless mode.
func present(g *game.Game, logger *applog.Logger, _ time.Duration) error {
	if err := g.Run(context.Background(), logger.Move); err != nil {
		return err
	}
	logger.Result(g.Result(), g.DrawReason())
	return nil
}
