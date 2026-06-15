package game

import (
	"context"
	"testing"

	"github.com/danielriddell21/gambit/internal/agent"
)

// TestMCTSPlaysLegalMoves drives the MCTS agent; Step errors on any illegal
// move, so reaching the ply cap without error asserts correctness.
func TestMCTSPlaysLegalMoves(t *testing.T) {
	white, err := agent.New("mcts", agent.Options{Seed: 1, Iterations: 300})
	if err != nil {
		t.Fatal(err)
	}
	black, err := agent.New("random", agent.Options{Seed: 2})
	if err != nil {
		t.Fatal(err)
	}
	g := New(Players{White: white, Black: black}, nil)
	for i := 0; i < 40 && !g.Over(); i++ {
		if _, _, err := g.Step(context.Background()); err != nil {
			t.Fatalf("mcts: %v", err)
		}
	}
}
