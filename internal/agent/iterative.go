package agent

import (
	"context"

	"github.com/danielriddell21/gambit/internal/agent/eval"
	"github.com/danielriddell21/gambit/pkg/chess"
)

func init() {
	Register("iterative", func(o Options) (Agent, error) {
		maxDepth := o.Depth
		if maxDepth <= 0 {
			maxDepth = 6
		}
		return &iterativeAgent{maxDepth: maxDepth, eval: eval.Material}, nil
	})
}

type iterativeAgent struct {
	maxDepth int
	eval     eval.Func
}

func (a *iterativeAgent) Name() string { return "iterative" }

func (a *iterativeAgent) SelectMove(ctx context.Context, b *chess.Board) (chess.Move, error) {
	moves := b.GenerateMoves(nil)
	if len(moves) == 0 {
		return chess.NoMove, errNoMoves
	}
	orderMoves(b, moves)

	best := moves[0]
	for depth := 1; depth <= a.maxDepth; depth++ {
		if ctx.Err() != nil {
			break
		}
		m := searchRoot(ctx, b, depth, a.eval, moves)
		if ctx.Err() != nil {
			break // this depth was cut short; keep the previous best
		}
		best = m
		moveToFront(moves, best) // search the best move first next iteration
	}
	return best, nil //nolint:nilerr // on ctx timeout, keep the best move found so far
}

func moveToFront(moves []chess.Move, m chess.Move) {
	for i, mv := range moves {
		if mv == m {
			copy(moves[1:i+1], moves[:i])
			moves[0] = m
			return
		}
	}
}
