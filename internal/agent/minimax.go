package agent

import (
	"context"
	"sort"

	"github.com/danielriddell21/gambit/internal/agent/eval"
	"github.com/danielriddell21/gambit/pkg/chess"
)

const (
	mateScore = 1_000_000
	infinity  = 1 << 30
)

func init() {
	Register("minimax", func(o Options) (Agent, error) {
		depth := o.Depth
		if depth <= 0 {
			depth = 4
		}
		return &alphaBetaAgent{depth: depth, eval: eval.Material}, nil
	})
}

type alphaBetaAgent struct {
	depth int
	eval  eval.Func
}

func (a *alphaBetaAgent) Name() string { return "minimax" }

func (a *alphaBetaAgent) SelectMove(ctx context.Context, b *chess.Board) (chess.Move, error) {
	moves := b.GenerateMoves(nil)
	if len(moves) == 0 {
		return chess.NoMove, errNoMoves
	}
	orderMoves(b, moves)

	best := searchRoot(ctx, b, a.depth, a.eval, moves)
	return best, nil
}

func searchRoot(ctx context.Context, b *chess.Board, depth int, e eval.Func, moves []chess.Move) chess.Move {
	best := moves[0]
	alpha := -infinity
	for _, m := range moves {
		if ctx.Err() != nil {
			break // honor cancellation; keep best move found so far
		}
		u := b.MakeMove(m)
		score := -negamax(ctx, b, depth-1, -infinity, -alpha, e)
		b.UnmakeMove(m, u)
		if score > alpha {
			alpha = score
			best = m
		}
	}
	return best
}

func negamax(ctx context.Context, b *chess.Board, depth, alpha, beta int, e eval.Func) int {
	if ctx.Err() != nil || depth == 0 {
		return e(b)
	}

	moves := b.GenerateMoves(nil)
	if len(moves) == 0 {
		if b.InCheck() {
			// Prefer faster mates: deeper remaining depth means a sooner mate.
			return -(mateScore + depth)
		}
		return 0 // stalemate
	}
	orderMoves(b, moves)

	best := -infinity
	for _, m := range moves {
		u := b.MakeMove(m)
		score := -negamax(ctx, b, depth-1, -beta, -alpha, e)
		b.UnmakeMove(m, u)
		if score > best {
			best = score
		}
		if best > alpha {
			alpha = best
		}
		if alpha >= beta {
			break // beta cutoff
		}
	}
	return best
}

func orderMoves(b *chess.Board, moves []chess.Move) {
	sort.SliceStable(moves, func(i, j int) bool {
		return moveScore(b, moves[i]) > moveScore(b, moves[j])
	})
}

func moveScore(b *chess.Board, m chess.Move) int {
	score := 0
	if victim := b.PieceAt(m.To()); !victim.IsEmpty() {
		score += 10*pieceWorth(victim.Type()) - pieceWorth(b.PieceAt(m.From()).Type())
	}
	if m.IsPromotion() {
		score += pieceWorth(m.Promotion())
	}
	return score
}

func pieceWorth(t chess.PieceType) int {
	switch t {
	case chess.Pawn:
		return 1
	case chess.Knight, chess.Bishop:
		return 3
	case chess.Rook:
		return 5
	case chess.Queen:
		return 9
	default:
		return 0
	}
}
