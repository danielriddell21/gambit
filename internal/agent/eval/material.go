// Package eval provides static evaluation functions for chess positions. It is
// kept separate from the search agents so evaluation and search can evolve
// independently.
package eval

import "github.com/danielriddell21/gambit/pkg/chess"

// Func scores a position in centipawns from the perspective of the side to
// move: a positive value favors the side to move.
type Func func(b *chess.Board) int

// pieceValue is the centipawn value of each piece type.
var pieceValue = [...]int{
	chess.Pawn:   100,
	chess.Knight: 320,
	chess.Bishop: 330,
	chess.Rook:   500,
	chess.Queen:  900,
	chess.King:   0,
}

// Material scores a position by material balance plus a small piece-square
// preference for central development. The result is from the perspective of the
// side to move.
func Material(b *chess.Board) int {
	score := 0
	b.Each(func(s chess.Square, p chess.Piece) {
		if p.IsEmpty() {
			return
		}
		v := pieceValue[p.Type()] + pst(p.Type(), s, p.Color())
		if p.Color() == chess.White {
			score += v
		} else {
			score -= v
		}
	})
	if b.SideToMove() == chess.Black {
		return -score
	}
	return score
}

// pst returns a small positional bonus encouraging knights, bishops and pawns
// toward the center. Values are symmetric, mirrored by color.
func pst(t chess.PieceType, s chess.Square, c chess.Color) int {
	file, rank := s.File(), s.Rank()
	if c == chess.Black {
		rank = 7 - rank
	}
	centerFile := 3 - absInt(file*2-7)/2 // 0 at the edges, up to ~3 centrally
	centerRank := 3 - absInt(rank*2-7)/2
	switch t {
	case chess.Knight, chess.Bishop:
		return (centerFile + centerRank) * 4
	case chess.Pawn:
		return centerFile*2 + rank*2 // reward central, advanced pawns
	default:
		return 0
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
