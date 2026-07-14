package agent

import (
	"context"
	"fmt"

	"github.com/danielriddell21/gambit/pkg/chess"
)

func init() {
	Register("human", func(_ Options) (Agent, error) { return NewHuman(), nil })
}

type HumanAgent struct {
	moves chan chess.Move
}

func NewHuman() *HumanAgent {
	return &HumanAgent{moves: make(chan chess.Move)}
}

func (a *HumanAgent) Name() string { return "human" }

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

func (a *HumanAgent) Submit(m chess.Move) bool {
	// Non-blocking: a move offered while no SelectMove is waiting is dropped
	// rather than queued for a later turn.
	select {
	case a.moves <- m:
		return true
	default:
		return false
	}
}
