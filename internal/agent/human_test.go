package agent_test

import (
	"context"
	"testing"
	"time"

	"github.com/danielriddell21/gambit/internal/agent"
	"github.com/danielriddell21/gambit/pkg/chess"
)

func TestHumanAgentSubmit(t *testing.T) {
	h := agent.NewHuman()
	b := chess.NewStartingBoard()
	want := b.LegalMoves()[0]

	got := make(chan chess.Move, 1)
	go func() {
		m, err := h.SelectMove(context.Background(), b)
		if err != nil {
			t.Errorf("SelectMove: %v", err)
		}
		got <- m
	}()

	deadline := time.After(time.Second)
	for !h.Submit(want) { // retry until SelectMove is waiting
		select {
		case <-deadline:
			t.Fatal("Submit was never accepted")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	if m := <-got; m != want {
		t.Errorf("SelectMove returned %s, want %s", m, want)
	}
}

func TestHumanAgentContextCancel(t *testing.T) {
	h := agent.NewHuman()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := h.SelectMove(ctx, chess.NewStartingBoard()); err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func TestHumanRegistered(t *testing.T) {
	a, err := agent.New("human", agent.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if a.Name() != "human" {
		t.Errorf("Name = %q, want human", a.Name())
	}
}
