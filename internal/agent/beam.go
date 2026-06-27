package agent

import (
	"context"
	"sort"

	"github.com/danielriddell21/gambit/internal/agent/eval"
	"github.com/danielriddell21/gambit/pkg/chess"
)

func init() {
	Register("beam", func(o Options) (Agent, error) {
		width := o.Width
		if width <= 0 {
			width = 8
		}
		depth := o.Depth
		if depth <= 0 {
			depth = 6
		}
		return &beamAgent{width: width, depth: depth, eval: eval.Material}, nil
	})
}

// beamAgent performs beam search: it keeps the best `width` lines, expands all
// of them each ply for `depth` plies, then plays the root move of the best line
// at the final horizon. Scores are always from the root mover's perspective, so
// the search is optimistic about the opponent — characteristic of beam search
// rather than minimax, which makes it an interesting contrast to study.
type beamAgent struct {
	width int
	depth int
	eval  eval.Func
}

type beamState struct {
	board *chess.Board
	root  chess.Move // the root move this line started with
	score int        // evaluation from the root mover's perspective
}

func (a *beamAgent) Name() string { return "beam" }

func (a *beamAgent) SelectMove(ctx context.Context, b *chess.Board) (chess.Move, error) {
	moves := b.LegalMoves()
	if len(moves) == 0 {
		return chess.NoMove, errNoMoves
	}
	rootColor := b.SideToMove()

	// Seed the beam with each legal root move.
	beam := make([]beamState, 0, len(moves))
	for _, m := range moves {
		nb := b.ApplyMove(m)
		beam = append(beam, beamState{board: nb, root: m, score: rootScore(nb, rootColor, a.eval)})
	}
	beam = prune(beam, a.width)

	for ply := 1; ply < a.depth; ply++ {
		if ctx.Err() != nil {
			break
		}
		next := make([]beamState, 0, len(beam)*8)
		for _, st := range beam {
			children := st.board.LegalMoves()
			if len(children) == 0 {
				next = append(next, st) // terminal line: carry it forward
				continue
			}
			for _, m := range children {
				nb := st.board.ApplyMove(m)
				next = append(next, beamState{board: nb, root: st.root, score: rootScore(nb, rootColor, a.eval)})
			}
		}
		if len(next) == 0 {
			break
		}
		beam = prune(next, a.width)
	}
	return beam[0].root, nil //nolint:nilerr // on ctx timeout, return the best line found so far
}

// rootScore returns the evaluation from rootColor's perspective (eval is from
// the side-to-move's perspective).
func rootScore(b *chess.Board, rootColor chess.Color, e eval.Func) int {
	s := e(b)
	if b.SideToMove() != rootColor {
		return -s
	}
	return s
}

// prune keeps the top-n states by score (descending).
func prune(states []beamState, n int) []beamState {
	sort.SliceStable(states, func(i, j int) bool {
		return states[i].score > states[j].score
	})
	if len(states) > n {
		states = states[:n]
	}
	return states
}
