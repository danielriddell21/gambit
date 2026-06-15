//go:build ebiten

// Package ui renders an agent-vs-agent game with Ebiten. It reads board state
// from pkg/chess for drawing and runs each agent's search on a worker goroutine
// so the window stays responsive while agents think.
package ui

import (
	"context"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/danielriddell21/gambit/internal/game"
	applog "github.com/danielriddell21/gambit/internal/log"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// Piece tint colors, reused for the info bar and banner so the recorder palette
// stays small.
var (
	pieceWhite = color.RGBA{R: 0xf5, G: 0xf5, B: 0xf0, A: 0xff}
	pieceBlack = color.RGBA{R: 0x20, G: 0x20, B: 0x24, A: 0xff}
)

// Speed bounds for the +/- controls.
const (
	minDelay = 10 * time.Millisecond
	maxDelay = 3 * time.Second
)

// Config controls the look and pacing of the GUI.
type Config struct {
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

// DefaultConfig returns sensible defaults.
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

type stepResult struct {
	ev  game.MoveEvent
	ok  bool
	err error
	gen int // generation this worker was launched under (for restart)
}

// GameUI is the ebiten.Game driving an agent-vs-agent match.
type GameUI struct {
	game      *game.Game
	newGame   func() *game.Game
	log       *applog.Logger
	cfg       Config
	barHeight int

	face       text.Face // piece glyphs
	barFace    text.Face // info bar text
	bannerFace text.Face // game-over banner

	snapshot     [64]chess.Piece // cached board, only mutated on the main goroutine
	thinking     bool
	resultCh     chan stepResult
	lastMoveTime time.Time
	finished     bool

	// Controls.
	paused   bool
	stepOnce bool
	flipped  bool
	gen      int // bumped on restart; stale worker results are discarded

	rec           *recorder // nil unless recording a GIF
	needCapture   bool      // capture a frame on the next Draw
	recFinalReady bool      // the final frame has been captured
	recSaved      bool      // the GIF has been written
}

// New builds a GameUI. newGame rebuilds the game when the user restarts (R).
func New(g *game.Game, newGame func() *game.Game, log *applog.Logger, cfg Config) (*GameUI, error) {
	sq := float64(cfg.SquareSize)
	face, err := newFace(sq * 0.8)
	if err != nil {
		return nil, err
	}
	barHeight := cfg.SquareSize * 7 / 10
	if barHeight < 46 {
		barHeight = 46
	}
	barFace, err := newFace(float64(barHeight) * 0.3)
	if err != nil {
		return nil, err
	}
	bannerFace, err := newFace(sq * 0.42)
	if err != nil {
		return nil, err
	}

	u := &GameUI{
		game:       g,
		newGame:    newGame,
		log:        log,
		cfg:        cfg,
		barHeight:  barHeight,
		face:       face,
		barFace:    barFace,
		bannerFace: bannerFace,
		resultCh:   make(chan stepResult, 1),
	}
	if cfg.RecordPath != "" {
		delay := cfg.RecordDelay
		if delay <= 0 {
			delay = 70
		}
		u.rec = newRecorder(cfg.RecordPath, delay)
		u.needCapture = true // capture the initial position
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
	u.handleInput()

	select {
	case r := <-u.resultCh:
		if r.gen != u.gen {
			break // stale result from a game that was restarted; discard
		}
		u.thinking = false
		if r.err != nil {
			return r.err
		}
		if r.ok {
			u.log.Move(r.ev)
			u.refreshSnapshot()
			u.lastMoveTime = time.Now()
			if u.rec != nil {
				u.needCapture = true
			}
		}
		if u.game.Over() && !u.finished {
			u.finished = true
			u.log.Result(u.game.Result(), u.game.DrawReason())
		}
	default:
	}

	// When recording, finalize and exit once the final frame is captured.
	if u.rec != nil {
		if u.recSaved {
			return ebiten.Termination
		}
		if u.game.Over() && u.recFinalReady {
			if err := u.rec.save(); err != nil {
				return err
			}
			u.recSaved = true
			return ebiten.Termination
		}
	}

	if u.thinking || u.game.Over() {
		return nil
	}
	if u.paused && !u.stepOnce {
		return nil
	}
	if !u.stepOnce && time.Since(u.lastMoveTime) < u.cfg.MoveDelay {
		return nil
	}
	u.stepOnce = false

	// Compute the next move off the main goroutine so drawing keeps ticking.
	u.thinking = true
	gen := u.gen
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), u.cfg.ThinkTimeout)
		defer cancel()
		ev, ok, err := u.game.Step(ctx)
		u.resultCh <- stepResult{ev: ev, ok: ok, err: err, gen: gen}
	}()
	return nil
}

// handleInput processes keyboard controls.
func (u *GameUI) handleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		u.paused = !u.paused
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		u.stepOnce = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		u.flipped = !u.flipped
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		u.restart()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		u.cfg.MoveDelay = clampDelay(u.cfg.MoveDelay * 2 / 3) // faster
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		u.cfg.MoveDelay = clampDelay(u.cfg.MoveDelay * 3 / 2) // slower
	}
}

// restart begins a fresh game. The generation bump makes any in-flight worker's
// result get discarded when it arrives.
func (u *GameUI) restart() {
	u.gen++
	u.game = u.newGame()
	u.thinking = false
	u.finished = false
	u.paused = false
	u.stepOnce = false
	u.lastMoveTime = time.Time{}
	u.refreshSnapshot()
}

func clampDelay(d time.Duration) time.Duration {
	switch {
	case d < minDelay:
		return minDelay
	case d > maxDelay:
		return maxDelay
	default:
		return d
	}
}

// Draw paints the board, pieces, info bar and (when finished) the banner.
func (u *GameUI) Draw(screen *ebiten.Image) {
	u.drawBoard(screen)
	u.drawPieces(screen)
	u.drawInfoBar(screen)
	if u.game.Over() {
		u.drawBanner(screen)
	}

	if u.rec != nil && u.needCapture {
		u.rec.capture(screen)
		u.needCapture = false
		if u.game.Over() {
			u.recFinalReady = true
		}
	}
}

// Layout fixes the logical screen size to the board plus the info bar.
func (u *GameUI) Layout(_, _ int) (int, int) {
	return u.WindowSize()
}

// WindowSize returns the pixel size of the window (board + info bar).
func (u *GameUI) WindowSize() (int, int) {
	return u.boardSize(), u.boardSize() + u.barHeight
}

func (u *GameUI) boardSize() int {
	return u.cfg.SquareSize * 8
}
