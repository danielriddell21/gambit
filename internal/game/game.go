// Package game orchestrates a chess game between two agents: it alternates
// turns, validates and applies moves, tracks game-history draw conditions
// (threefold repetition and the fifty-move rule), and emits a move event per
// ply. It is deliberately decoupled from any UI so the same logic drives both
// the headless runner and the Ebiten GUI.
package game

import (
	"context"
	"fmt"

	"github.com/danielriddell21/gambit/internal/agent"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// Players holds the agent controlling each color.
type Players struct {
	White agent.Agent
	Black agent.Agent
}

// MoveEvent describes a single completed ply.
type MoveEvent struct {
	Ply       int         // 1-based half-move number
	Mover     chess.Color // side that made the move
	AgentName string      // strategy that chose it
	Move      chess.Move  // the move played
	SAN       string      // human-readable notation
	FEN       string      // resulting position
}

// Game tracks the state of a single agent-vs-agent game.
type Game struct {
	board   *chess.Board
	players Players
	history map[uint64]int // Zobrist hash -> times reached, for threefold
	moves   []chess.Move
	result  chess.Result
	reason  chess.DrawReason
}

// New creates a game. If start is nil the standard initial position is used.
func New(p Players, start *chess.Board) *Game {
	if start == nil {
		start = chess.NewStartingBoard()
	}
	g := &Game{
		board:   start,
		players: p,
		history: map[uint64]int{},
	}
	g.history[start.Hash()]++
	return g
}

// Board returns the live board. Callers (e.g. the GUI) must only read it.
func (g *Game) Board() *chess.Board { return g.board }

// Result returns the current result (InProgress until the game ends).
func (g *Game) Result() chess.Result { return g.result }

// DrawReason returns why the game was drawn, if it was.
func (g *Game) DrawReason() chess.DrawReason { return g.reason }

// Moves returns the moves played so far.
func (g *Game) Moves() []chess.Move { return g.moves }

// AgentName returns the name of the agent playing the given color.
func (g *Game) AgentName(c chess.Color) string {
	if c == chess.White {
		return g.players.White.Name()
	}
	return g.players.Black.Name()
}

// Over reports whether the game has finished.
func (g *Game) Over() bool { return g.result != chess.InProgress }

func (g *Game) current() agent.Agent {
	if g.board.SideToMove() == chess.White {
		return g.players.White
	}
	return g.players.Black
}

// Step plays exactly one ply: it asks the side-to-move's agent for a move,
// validates it, applies it and updates the result. The boolean is false when
// the game was already over. It is safe to drive from a GUI tick or a headless
// loop.
func (g *Game) Step(ctx context.Context) (MoveEvent, bool, error) {
	if g.Over() {
		return MoveEvent{}, false, nil
	}

	ag := g.current()
	mover := g.board.SideToMove()

	m, err := ag.SelectMove(ctx, g.board)
	if err != nil {
		return MoveEvent{}, false, fmt.Errorf("agent %s: %w", ag.Name(), err)
	}
	if !g.isLegal(m) {
		return MoveEvent{}, false, fmt.Errorf("agent %s returned illegal move %s", ag.Name(), m)
	}

	san := SAN(g.board, m)
	g.board.MakeMove(m)
	g.moves = append(g.moves, m)

	ev := MoveEvent{
		Ply:       len(g.moves),
		Mover:     mover,
		AgentName: ag.Name(),
		Move:      m,
		SAN:       san,
		FEN:       g.board.FEN(),
	}

	g.updateResult()
	return ev, true, nil
}

// Run plays the game to completion, invoking onMove (if non-nil) after each
// ply. It returns the first error encountered, if any.
func (g *Game) Run(ctx context.Context, onMove func(MoveEvent)) error {
	for !g.Over() {
		ev, ok, err := g.Step(ctx)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		if onMove != nil {
			onMove(ev)
		}
	}
	return nil
}

func (g *Game) isLegal(m chess.Move) bool {
	for _, lm := range g.board.LegalMoves() {
		if lm == m {
			return true
		}
	}
	return false
}

// updateResult sets the game result after a move, checking position-derivable
// outcomes first, then the history-dependent draw rules.
func (g *Game) updateResult() {
	if res, reason := g.board.Status(); res != chess.InProgress {
		g.result = res
		g.reason = reason
		return
	}
	if g.board.HalfMoveClock() >= 100 {
		g.result = chess.Draw
		g.reason = chess.FiftyMoveRule
		return
	}
	h := g.board.Hash()
	g.history[h]++
	if g.history[h] >= 3 {
		g.result = chess.Draw
		g.reason = chess.ThreefoldRepetition
	}
}
