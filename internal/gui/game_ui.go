//go:build ebiten

package gui

import (
	"context"
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/danielriddell21/crucible/record"

	"github.com/danielriddell21/gambit/internal/agent"
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

	selected    chess.Square
	cancelThink context.CancelFunc

	rec           *record.Recorder
	recPath       string
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
		selected:   chess.NoSquare,
	}
	if cfg.Rec.Recording() {
		delay := cfg.RecordDelay
		if delay <= 0 {
			delay = 70
		}
		// One frame per move at full resolution, quantised to the board's own
		// palette, holding the final position for four seconds before the loop
		// restarts. --record-frames caps the clip; zero records the whole game.
		u.rec = record.NewRecorder(0, 1, cfg.Rec.Frames,
			record.WithPalette(demoPalette),
			record.WithFrameDelay(delay),
			record.WithFinalHold(400),
		)
		u.recPath = cfg.Rec.Path
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
	// Finish when --record-frames is reached, or when the game ends and its
	// final position has been captured.
	if u.rec.Done() || (u.game.Over() && u.recFinalReady) {
		if err := u.rec.Save(u.recPath); err != nil {
			return fmt.Errorf("save recording: %w", err)
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
	case u.currentIsHuman():
		return true // start the worker so it waits for the human's click
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
	// A human has no time budget; agents are bounded by ThinkTimeout.
	ctx, cancel := context.WithCancel(context.Background())
	if !u.currentIsHuman() {
		ctx, cancel = context.WithTimeout(context.Background(), u.cfg.ThinkTimeout)
	}
	u.cancelThink = cancel
	go func() {
		defer cancel()
		ev, ok, err := u.game.Step(ctx)
		u.resultCh <- stepResult{ev: ev, ok: ok, err: err, gen: gen}
	}()
}

func (u *GameUI) currentIsHuman() bool {
	_, ok := u.humanToMove()
	return ok
}

func (u *GameUI) humanToMove() (*agent.HumanAgent, bool) {
	ha, ok := u.game.Agent(u.game.Board().SideToMove()).(*agent.HumanAgent)
	return ha, ok
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
	u.handleMouse()
}

func (u *GameUI) handleMouse() {
	ha, ok := u.humanToMove()
	if !ok || !u.thinking || !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	sq, ok := u.squareAt(ebiten.CursorPosition())
	if !ok {
		return
	}
	if u.selected == chess.NoSquare {
		if u.hasLegalFrom(sq) {
			u.selected = sq
		}
		return
	}
	if m, ok := u.findMove(u.selected, sq); ok {
		ha.Submit(m)
		u.selected = chess.NoSquare
		return
	}
	// A click elsewhere re-selects another friendly piece, or clears.
	if u.hasLegalFrom(sq) {
		u.selected = sq
	} else {
		u.selected = chess.NoSquare
	}
}

func (u *GameUI) squareAt(mx, my int) (chess.Square, bool) {
	size := u.cfg.SquareSize
	if mx < 0 || my < 0 || mx >= size*8 || my >= size*8 {
		return chess.NoSquare, false // outside the board (e.g. the info bar)
	}
	col, row := mx/size, my/size
	file, rank := col, 7-row
	if u.flipped {
		file, rank = 7-col, row
	}
	return chess.NewSquare(file, rank), true
}

func (u *GameUI) hasLegalFrom(sq chess.Square) bool {
	for _, m := range u.game.Board().LegalMoves() {
		if m.From() == sq {
			return true
		}
	}
	return false
}

func (u *GameUI) findMove(from, to chess.Square) (chess.Move, bool) {
	var fallback chess.Move
	found := false
	for _, m := range u.game.Board().LegalMoves() {
		if m.From() != from || m.To() != to {
			continue
		}
		if !m.IsPromotion() {
			return m, true
		}
		if m.Promotion() == chess.Queen {
			return m, true
		}
		fallback, found = m, true
	}
	return fallback, found
}

func (u *GameUI) restart() {
	if u.cancelThink != nil {
		u.cancelThink() // release a worker still waiting on the old game (e.g. a human)
	}
	u.gen++
	u.game = u.newGame()
	u.thinking = false
	u.finished = false
	u.paused = false
	u.stepOnce = false
	u.selected = chess.NoSquare
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
	u.drawHighlights(screen)
	u.drawPieces(screen)
	u.drawInfoBar(screen)
	if u.game.Over() {
		u.drawBanner(screen)
	}

	if u.rec != nil && u.needCapture {
		u.rec.Add(screenRGBA(screen))
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
