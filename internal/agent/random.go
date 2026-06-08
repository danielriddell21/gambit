package agent

import (
	"context"
	"errors"
	"math/rand/v2"

	"github.com/danielriddell21/gambit/pkg/chess"
)

// errNoMoves is returned when an agent is asked to move in a terminal position.
var errNoMoves = errors.New("agent: no legal moves")

func init() {
	Register("random", func(o Options) (Agent, error) {
		seed := uint64(o.Seed)
		return &randomAgent{rng: rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))}, nil
	})
}

// randomAgent plays a uniformly random legal move. It is the baseline opponent
// for evaluating every other strategy.
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
