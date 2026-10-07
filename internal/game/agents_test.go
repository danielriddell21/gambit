package game

import (
	"context"
	"testing"

	"github.com/danielriddell21/gambit/internal/agent"
)

func TestAgentsPlayLegalMoves(t *testing.T) {
	cases := []struct {
		name string
		opts agent.Options
	}{
		{"beam", agent.Options{Seed: 1, Depth: 4, Width: 6}},
		{"greedy", agent.Options{Seed: 1}},
		{"iterative", agent.Options{Seed: 1, Depth: 3}},
		{"mcts", agent.Options{Seed: 1, Iterations: 300}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			white, err := agent.New(tc.name, tc.opts)
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
					t.Fatalf("%s: %v", tc.name, err)
				}
			}
		})
	}
}
