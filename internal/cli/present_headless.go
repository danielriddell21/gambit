//go:build !ebiten

package cli

import (
	"context"
	"fmt"

	"github.com/danielriddell21/gambit/internal/game"
	applog "github.com/danielriddell21/gambit/internal/log"
)

// present runs the game to completion without a GUI, logging every move and the
// final result. The restart factory and GUI-only config are ignored here.
func present(g *game.Game, _ func() *game.Game, logger *applog.Logger, _ config) error {
	if err := g.Run(context.Background(), logger.Move); err != nil {
		return fmt.Errorf("run game: %w", err)
	}
	logger.Result(g.Result(), g.DrawReason())
	return nil
}
