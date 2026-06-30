package game

import (
	"context"
	"fmt"

	"github.com/danielriddell21/gambit/internal/agent"
	"github.com/danielriddell21/gambit/pkg/chess"
)

type Players struct {
	White agent.Agent
	Black agent.Agent
}

type MoveEvent struct {
	Ply       int
	Mover     chess.Color
	AgentName string
	Move      chess.Move
	SAN       string
	FEN       string
}

type Game struct {
	board   *chess.Board
	players Players
	history map[uint64]int
	moves   []chess.Move
	result  chess.Result
	reason  chess.DrawReason
}

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

func (g *Game) Board() *chess.Board { return g.board }

func (g *Game) Result() chess.Result { return g.result }

func (g *Game) DrawReason() chess.DrawReason { return g.reason }

func (g *Game) Moves() []chess.Move { return g.moves }

func (g *Game) AgentName(c chess.Color) string {
	if c == chess.White {
		return g.players.White.Name()
	}
	return g.players.Black.Name()
}

func (g *Game) Over() bool { return g.result != chess.InProgress }

func (g *Game) current() agent.Agent {
	if g.board.SideToMove() == chess.White {
		return g.players.White
	}
	return g.players.Black
}

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
