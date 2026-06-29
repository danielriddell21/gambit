// Package gui renders an agent-vs-agent game with Ebiten. It reads board state
// from pkg/chess for drawing and runs each agent's search on a worker goroutine
// so the window stays responsive while agents think. The window is gated behind
// the "ebiten" build tag; without it Run plays the game out headlessly.
package gui

import (
	"image/color"
	"time"

	"github.com/danielriddell21/gambit/internal/game"
	applog "github.com/danielriddell21/gambit/internal/log"
)

// Config is the single argument to Run. It carries the game to present and the
// look/pacing knobs. It holds no Ebiten types, so the CLI constructs it in
// either build.
type Config struct {
	// Game is the match to present; NewGame rebuilds it when the user restarts
	// (R), and Logger receives move and result events (used by the headless path).
	Game    *game.Game
	NewGame func() *game.Game
	Logger  *applog.Logger

	SquareSize   int
	LightSquare  color.RGBA
	DarkSquare   color.RGBA
	MoveDelay    time.Duration // pause between moves so the game is watchable
	ThinkTimeout time.Duration // hard cap on an agent's thinking time

	// RecordPath, when set, records the rendered game to an animated GIF at that
	// path and exits when the game ends. RecordDelay is the per-frame delay in
	// hundredths of a second.
	RecordPath  string
	RecordDelay int
}

// DefaultConfig returns sensible defaults for the look and pacing knobs.
func DefaultConfig() Config {
	return Config{
		SquareSize:   80,
		LightSquare:  color.RGBA{R: 0xec, G: 0xd9, B: 0xb6, A: 0xff},
		DarkSquare:   color.RGBA{R: 0xa9, G: 0x7a, B: 0x55, A: 0xff},
		MoveDelay:    400 * time.Millisecond,
		ThinkTimeout: 10 * time.Second,
		RecordDelay:  70,
	}
}
