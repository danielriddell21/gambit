package gui

import (
	"image/color"
	"time"

	"github.com/danielriddell21/gambit/internal/game"
	applog "github.com/danielriddell21/gambit/internal/log"
)

type Config struct {
	Game    *game.Game
	NewGame func() *game.Game
	Logger  *applog.Logger

	SquareSize   int
	LightSquare  color.RGBA
	DarkSquare   color.RGBA
	MoveDelay    time.Duration
	ThinkTimeout time.Duration

	RecordPath  string
	RecordDelay int
}

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
