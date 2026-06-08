// Package ui renders an agent-vs-agent game with Ebiten. It reads board state
// from pkg/chess for drawing and runs each agent's search on a worker goroutine
// so the window stays responsive while agents think.
package ui

import (
	"context"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/danielriddell21/gambit/internal/game"
	applog "github.com/danielriddell21/gambit/internal/log"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// Piece tint colors.
var (
	pieceWhite = color.RGBA{R: 0xf5, G: 0xf5, B: 0xf0, A: 0xff}
	pieceBlack = color.RGBA{R: 0x20, G: 0x20, B: 0x24, A: 0xff}
)

// Config controls the look and pacing of the GUI.
type Config struct {
	SquareSize   int
	LightSquare  color.RGBA
	DarkSquare   color.RGBA
	MoveDelay    time.Duration // pause between moves so the game is watchable
	ThinkTimeout time.Duration // hard cap on an agent's thinking time
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		SquareSize:   80,
		LightSquare:  color.RGBA{R: 0xec, G: 0xd9, B: 0xb6, A: 0xff},
		DarkSquare:   color.RGBA{R: 0xa9, G: 0x7a, B: 0x55, A: 0xff},
		MoveDelay:    400 * time.Millisecond,
		ThinkTimeout: 10 * time.Second,
	}
}

type stepResult struct {
	ev  game.MoveEvent
	ok  bool
	err error
}

// GameUI is the ebiten.Game driving an agent-vs-agent match.
type GameUI struct {
	game *game.Game
	log  *applog.Logger
	cfg  Config
	face text.Face

	snapshot     [64]chess.Piece // cached board, only mutated on the main goroutine
	thinking     bool
	resultCh     chan stepResult
	lastMoveTime time.Time
	finished     bool
}

// New builds a GameUI for the given game.
func New(g *game.Game, log *applog.Logger, cfg Config) (*GameUI, error) {
	face, err := newFace(float64(cfg.SquareSize) * 0.8)
	if err != nil {
		return nil, err
	}
	u := &GameUI{
		game:     g,
		log:      log,
		cfg:      cfg,
		face:     face,
		resultCh: make(chan stepResult, 1),
	}
	u.refreshSnapshot()
	return u, nil
}

// refreshSnapshot copies the live board into the draw snapshot. Must be called
// only on the main (Update) goroutine while no worker is running.
func (u *GameUI) refreshSnapshot() {
	u.game.Board().Each(func(s chess.Square, p chess.Piece) {
		u.snapshot[s] = p
	})
}

// Update advances the game without ever blocking the main loop.
func (u *GameUI) Update() error {
	select {
	case r := <-u.resultCh:
		u.thinking = false
		if r.err != nil {
			return r.err
		}
		if r.ok {
			u.log.Move(r.ev)
			u.refreshSnapshot()
			u.lastMoveTime = time.Now()
		}
		if u.game.Over() && !u.finished {
			u.finished = true
			u.log.Result(u.game.Result(), u.game.DrawReason())
		}
	default:
	}

	if u.thinking || u.game.Over() {
		return nil
	}
	if time.Since(u.lastMoveTime) < u.cfg.MoveDelay {
		return nil
	}

	// Compute the next move off the main goroutine so drawing keeps ticking.
	u.thinking = true
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), u.cfg.ThinkTimeout)
		defer cancel()
		ev, ok, err := u.game.Step(ctx)
		u.resultCh <- stepResult{ev: ev, ok: ok, err: err}
	}()
	return nil
}

// Draw paints the board and pieces.
func (u *GameUI) Draw(screen *ebiten.Image) {
	u.drawBoard(screen)
	u.drawPieces(screen)
}

// Layout fixes the logical screen size to the board dimensions.
func (u *GameUI) Layout(_, _ int) (int, int) {
	side := u.cfg.SquareSize * 8
	return side, side
}

// BoardPixels returns the window's pixel size.
func (u *GameUI) BoardPixels() int {
	return u.cfg.SquareSize * 8
}
