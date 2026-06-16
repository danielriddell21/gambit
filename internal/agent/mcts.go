package agent

import (
	"context"
	"math"
	"math/rand/v2"

	"github.com/danielriddell21/gambit/internal/agent/eval"
	"github.com/danielriddell21/gambit/pkg/chess"
)

// maxRolloutPlies caps a random playout so simulations always terminate; at the
// cap the position is estimated with the static evaluation.
const maxRolloutPlies = 40

func init() {
	Register("mcts", func(o Options) (Agent, error) {
		iters := o.Iterations
		if iters <= 0 {
			iters = 20000
		}
		seed := uint64(o.Seed)
		return &mctsAgent{
			iterations: iters,
			rng:        rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15)),
		}, nil
	})
}

// mctsAgent plays by Monte Carlo Tree Search with UCT selection and random
// rollouts. No learning is involved — it is pure search, budgeted by an
// iteration count or the context deadline (whichever comes first).
type mctsAgent struct {
	iterations int
	rng        *rand.Rand
}

type mctsNode struct {
	board    *chess.Board
	parent   *mctsNode
	move     chess.Move // move from parent that reached this node
	untried  []chess.Move
	children []*mctsNode
	wins     float64 // results from the perspective of the player who moved here
	visits   int
}

func newNode(b *chess.Board, parent *mctsNode, move chess.Move) *mctsNode {
	return &mctsNode{board: b, parent: parent, move: move, untried: b.LegalMoves()}
}

func (a *mctsAgent) Name() string { return "mcts" }

func (a *mctsAgent) SelectMove(ctx context.Context, b *chess.Board) (chess.Move, error) {
	if len(b.LegalMoves()) == 0 {
		return chess.NoMove, errNoMoves
	}
	root := newNode(b.Clone(), nil, chess.NoMove)

	for i := 0; i < a.iterations; i++ {
		if i%256 == 0 && ctx.Err() != nil {
			break
		}
		node := a.treePolicy(root)
		result := a.rollout(node.board) // White-perspective result in [0,1]
		backpropagate(node, result)
	}

	// Play the most-visited root move (the robust choice in MCTS).
	best := root.children[0]
	for _, c := range root.children[1:] {
		if c.visits > best.visits {
			best = c
		}
	}
	return best.move, nil
}

// treePolicy walks down the tree, expanding the first node with untried moves.
func (a *mctsAgent) treePolicy(node *mctsNode) *mctsNode {
	for {
		if len(node.untried) > 0 {
			return a.expand(node)
		}
		if len(node.children) == 0 {
			return node // terminal position
		}
		node = bestUCT(node)
	}
}

func (a *mctsAgent) expand(node *mctsNode) *mctsNode {
	i := a.rng.IntN(len(node.untried))
	m := node.untried[i]
	node.untried = append(node.untried[:i], node.untried[i+1:]...)
	child := newNode(node.board.ApplyMove(m), node, m)
	node.children = append(node.children, child)
	return child
}

// bestUCT picks the child maximizing the UCT score, evaluated from the
// perspective of the player to move at node (the player choosing the child).
func bestUCT(node *mctsNode) *mctsNode {
	const c = math.Sqrt2
	logN := math.Log(float64(node.visits))
	var best *mctsNode
	bestScore := math.Inf(-1)
	for _, child := range node.children {
		exploit := child.wins / float64(child.visits)
		explore := c * math.Sqrt(logN/float64(child.visits))
		if s := exploit + explore; s > bestScore {
			bestScore, best = s, child
		}
	}
	return best
}

// rollout plays random moves to a terminal position or the ply cap, returning
// the result from White's perspective in [0,1].
func (a *mctsAgent) rollout(b *chess.Board) float64 {
	sim := b.Clone()
	for ply := 0; ply < maxRolloutPlies; ply++ {
		res, _ := sim.Status()
		if res != chess.InProgress {
			return whiteResult(res)
		}
		moves := sim.LegalMoves()
		sim.MakeMove(moves[a.rng.IntN(len(moves))])
	}
	return sigmoid(whiteCentipawns(sim))
}

// backpropagate updates visit and win counts up to the root. A node's wins are
// stored from the perspective of the player who moved into it.
func backpropagate(node *mctsNode, whiteResult float64) {
	for n := node; n != nil; n = n.parent {
		n.visits++
		if n.parent != nil {
			// The mover into n is the side to move at n.parent.
			if n.parent.board.SideToMove() == chess.White {
				n.wins += whiteResult
			} else {
				n.wins += 1 - whiteResult
			}
		}
	}
}

func whiteResult(res chess.Result) float64 {
	switch res {
	case chess.WhiteWins:
		return 1
	case chess.BlackWins:
		return 0
	default:
		return 0.5
	}
}

func whiteCentipawns(b *chess.Board) int {
	s := eval.Material(b)
	if b.SideToMove() == chess.White {
		return s
	}
	return -s
}

func sigmoid(cp int) float64 {
	return 1 / (1 + math.Exp(-float64(cp)/400))
}
