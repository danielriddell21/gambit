package game

import (
	"context"
	"testing"

	"github.com/danielriddell21/gambit/internal/agent"
)

// TestBeamPlaysLegalMoves drives the beam-search agent; Step errors on any
// illegal move, so reaching the ply cap without error asserts correctness.
func TestBeamPlaysLegalMoves(t *testing.T) {
	white, err := agent.New("beam", agent.Options{Seed: 1, Depth: 4, Width: 6})
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
			t.Fatalf("beam: %v", err)
		}
	}
}
