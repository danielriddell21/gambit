//go:build ebiten

package gui

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

var (
	pieceWhite = color.RGBA{R: 0xf5, G: 0xf5, B: 0xf0, A: 0xff}
	pieceBlack = color.RGBA{R: 0x20, G: 0x20, B: 0x24, A: 0xff}
)

const (
	minDelay = 10 * time.Millisecond
	maxDelay = 3 * time.Second
)

type stepResult struct {
	ev  game.MoveEvent
	ok  bool
	err error
	gen int
}

type GameUI struct {
	game      *game.Game
	newGame   func() *game.Game
	log       *applog.Logger
	cfg       Config
	barHeight int

	face       text.Face
	barFace    text.Face
	bannerFace text.Face

	snapshot     [64]chess.Piece
	thinking     bool
	resultCh     chan stepResult
	lastMoveTime time.Time
	finished     bool

	paused   bool
	stepOnce bool
	flipped  bool
	gen      int

	rec           *recorder
	needCapture   bool
	recFinalReady bool
	recSaved      bool
}

func newGameUI(cfg Config) (*GameUI, error) {
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
		game:       cfg.Game,
		newGame:    cfg.NewGame,
		log:        cfg.Logger,
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

func (u *GameUI) refreshSnapshot() {
	u.game.Board().Each(func(s chess.Square, p chess.Piece) {
		u.snapshot[s] = p
	})
}

func (u *GameUI) Update() error {
	u.handleInput()

	if err := u.drainResult(); err != nil {
		return err
	}
	if u.rec != nil {
		if err := u.finalizeRecording(); err != nil {
			return err
		}
	}
	if !u.readyToStep() {
		return nil
	}
	u.stepOnce = false
	u.startThinking()
	return nil
}

func (u *GameUI) drainResult() error {
	select {
	case r := <-u.resultCh:
		if r.gen != u.gen {
			return nil // stale result from a game that was restarted; discard
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
	return nil
}

func (u *GameUI) finalizeRecording() error {
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
	return nil
}

func (u *GameUI) readyToStep() bool {
	switch {
	case u.thinking || u.game.Over():
		return false
	case u.paused && !u.stepOnce:
		return false
	case !u.stepOnce && time.Since(u.lastMoveTime) < u.cfg.MoveDelay:
		return false
	default:
		return true
	}
}

func (u *GameUI) startThinking() {
	u.thinking = true
	gen := u.gen
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), u.cfg.ThinkTimeout)
		defer cancel()
		ev, ok, err := u.game.Step(ctx)
		u.resultCh <- stepResult{ev: ev, ok: ok, err: err, gen: gen}
	}()
}

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

func (u *GameUI) Layout(_, _ int) (int, int) {
	return u.WindowSize()
}

func (u *GameUI) WindowSize() (int, int) {
	return u.boardSize(), u.boardSize() + u.barHeight
}

func (u *GameUI) boardSize() int {
	return u.cfg.SquareSize * 8
}
