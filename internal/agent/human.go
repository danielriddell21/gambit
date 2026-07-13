package agent

import (
	"context"
	"fmt"

	"github.com/danielriddell21/gambit/pkg/chess"
)

func init() {
	Register("human", func(_ Options) (Agent, error) { return NewHuman(), nil })
}

// HumanAgent is a human-controlled player. Its move is supplied externally (by
// the GUI, from mouse input) via Submit; SelectMove blocks until a move arrives
// or the context is cancelled. It is only useful in the GUI build.
type HumanAgent struct {
	moves chan chess.Move
}

// NewHuman returns a ready-to-use human agent.
func NewHuman() *HumanAgent {
	return &HumanAgent{moves: make(chan chess.Move)}
}

// Name identifies the strategy.
func (a *HumanAgent) Name() string { return "human" }

// SelectMove blocks until Submit provides a move or ctx is cancelled.
func (a *HumanAgent) SelectMove(ctx context.Context, b *chess.Board) (chess.Move, error) {
	if len(b.LegalMoves()) == 0 {
		return chess.NoMove, errNoMoves
	}
	select {
	case m := <-a.moves:
		return m, nil
	case <-ctx.Done():
		return chess.NoMove, fmt.Errorf("human: %w", ctx.Err())
	}
}

// Submit hands the human's chosen move to a pending SelectMove. It is
// non-blocking and returns whether the move was accepted: a move offered while
// no SelectMove is waiting is dropped rather than queued for a later turn.
func (a *HumanAgent) Submit(m chess.Move) bool {
	select {
	case a.moves <- m:
		return true
	default:
		return false
	}
}
