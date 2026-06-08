package chess

// knightDeltas are the (file, rank) offsets of a knight's moves.
var knightDeltas = [8][2]int{
	{1, 2}, {2, 1}, {2, -1}, {1, -2},
	{-1, -2}, {-2, -1}, {-2, 1}, {-1, 2},
}

// kingDeltas are the (file, rank) offsets of a king's moves.
var kingDeltas = [8][2]int{
	{1, 0}, {1, 1}, {0, 1}, {-1, 1},
	{-1, 0}, {-1, -1}, {0, -1}, {1, -1},
}

// bishopDeltas are the diagonal sliding directions.
var bishopDeltas = [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

// rookDeltas are the orthogonal sliding directions.
var rookDeltas = [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

// IsSquareAttacked reports whether square s is attacked by any piece of the
// given color.
func (b *Board) IsSquareAttacked(s Square, by Color) bool {
	f, r := s.File(), s.Rank()

	// Pawns: a pawn of color `by` attacks s from one rank "behind" s relative
	// to its direction of travel.
	pawnRank := r - 1 // white pawns sit below s
	if by == Black {
		pawnRank = r + 1 // black pawns sit above s
	}
	if pawnRank >= 0 && pawnRank <= 7 {
		for _, df := range [2]int{-1, 1} {
			pf := f + df
			if pf < 0 || pf > 7 {
				continue
			}
			p := b.squares[NewSquare(pf, pawnRank)]
			if !p.IsEmpty() && p.Color() == by && p.Type() == Pawn {
				return true
			}
		}
	}

	if b.attackedByStep(f, r, knightDeltas[:], by, Knight) {
		return true
	}
	if b.attackedByStep(f, r, kingDeltas[:], by, King) {
		return true
	}
	if b.attackedBySlide(f, r, bishopDeltas[:], by, Bishop, Queen) {
		return true
	}
	if b.attackedBySlide(f, r, rookDeltas[:], by, Rook, Queen) {
		return true
	}
	return false
}

// attackedByStep checks the fixed (non-sliding) offsets for an attacker of the
// given type and color.
func (b *Board) attackedByStep(f, r int, deltas [][2]int, by Color, t PieceType) bool {
	for _, d := range deltas {
		nf, nr := f+d[0], r+d[1]
		if nf < 0 || nf > 7 || nr < 0 || nr > 7 {
			continue
		}
		p := b.squares[NewSquare(nf, nr)]
		if !p.IsEmpty() && p.Color() == by && p.Type() == t {
			return true
		}
	}
	return false
}

// attackedBySlide walks each sliding direction looking for an attacker that is
// either of the primary type or the secondary type (queen).
func (b *Board) attackedBySlide(f, r int, deltas [][2]int, by Color, t1, t2 PieceType) bool {
	for _, d := range deltas {
		nf, nr := f+d[0], r+d[1]
		for nf >= 0 && nf <= 7 && nr >= 0 && nr <= 7 {
			p := b.squares[NewSquare(nf, nr)]
			if !p.IsEmpty() {
				if p.Color() == by && (p.Type() == t1 || p.Type() == t2) {
					return true
				}
				break
			}
			nf += d[0]
			nr += d[1]
		}
	}
	return false
}

// InCheck reports whether the side to move is in check.
func (b *Board) InCheck() bool {
	return b.IsSquareAttacked(b.kings[b.sideToMove], b.sideToMove.Opposite())
}

// isColorInCheck reports whether the given color's king is attacked.
func (b *Board) isColorInCheck(c Color) bool {
	return b.IsSquareAttacked(b.kings[c], c.Opposite())
}
