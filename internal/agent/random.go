package agent

import (
	"context"
	"errors"
	"math/rand/v2"

	"github.com/danielriddell21/crucible/rng"

	"github.com/danielriddell21/gambit/pkg/chess"
)

var errNoMoves = errors.New("agent: no legal moves")

func init() {
	Register("random", func(o Options) (Agent, error) {
		seed := uint64(o.Seed)
		return &randomAgent{rng: rng.Stream(seed, seed^0x9E3779B97F4A7C15)}, nil
	})
}

type randomAgent struct {
	rng *rand.Rand
}

func (a *randomAgent) Name() string { return "random" }

func (a *randomAgent) SelectMove(_ context.Context, b *chess.Board) (chess.Move, error) {
	moves := b.LegalMoves()
	if len(moves) == 0 {
		return chess.NoMove, errNoMoves
	}
	return moves[a.rng.IntN(len(moves))], nil
}
