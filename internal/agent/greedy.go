package agent

import (
	"context"
	"math/rand/v2"

	"github.com/danielriddell21/gambit/internal/agent/eval"
	"github.com/danielriddell21/gambit/pkg/chess"
)

func init() {
	Register("greedy", func(o Options) (Agent, error) {
		seed := uint64(o.Seed)
		return &greedyAgent{rng: rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))}, nil
	})
}

type greedyAgent struct {
	rng *rand.Rand
}

func (a *greedyAgent) Name() string { return "greedy" }

func (a *greedyAgent) SelectMove(_ context.Context, b *chess.Board) (chess.Move, error) {
	moves := b.LegalMoves()
	if len(moves) == 0 {
		return chess.NoMove, errNoMoves
	}

	best := moves[0]
	bestScore := -infinity
	ties := 0
	for _, m := range moves {
		// eval.Material is from the side-to-move's perspective; after our move
		// that is the opponent, so negate to get the score for us.
		score := -eval.Material(b.ApplyMove(m))
		switch {
		case score > bestScore:
			bestScore, best, ties = score, m, 1
		case score == bestScore:
			ties++
			if a.rng.IntN(ties) == 0 { // reservoir tie-break
				best = m
			}
		}
	}
	return best, nil
}
