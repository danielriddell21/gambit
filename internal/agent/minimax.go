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

// alphaBetaAgent searches with negamax + alpha-beta pruning to a fixed depth and
// evaluates leaves with a static evaluation function. It is the baseline
// "real" player and the template future depth-limited strategies follow.
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

	best := moves[0]
	alpha := -infinity
	for _, m := range moves {
		if ctx.Err() != nil {
			break // honor cancellation; keep best move found so far
		}
		u := b.MakeMove(m)
		score := -a.negamax(ctx, b, a.depth-1, -infinity, -alpha)
		b.UnmakeMove(m, u)
		if score > alpha {
			alpha = score
			best = m
		}
	}
	return best, nil
}

// negamax returns the value of the position from the side-to-move's perspective.
func (a *alphaBetaAgent) negamax(ctx context.Context, b *chess.Board, depth, alpha, beta int) int {
	if ctx.Err() != nil {
		return a.eval(b)
	}
	if depth == 0 {
		return a.eval(b)
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
		score := -a.negamax(ctx, b, depth-1, -beta, -alpha)
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

// orderMoves sorts captures and promotions first (MVV-LVA style) to improve
// alpha-beta pruning.
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
