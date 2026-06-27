package chess

// LegalMoves returns all fully-legal moves for the side to move. It allocates a
// fresh slice; hot search loops should prefer GenerateMoves with a reused
// buffer.
func (b *Board) LegalMoves() []Move {
	return b.GenerateMoves(nil)
}

// GenerateMoves appends all fully-legal moves for the side to move to dst and
// returns the extended slice.
func (b *Board) GenerateMoves(dst []Move) []Move {
	pseudo := b.GeneratePseudoLegal(nil)
	us := b.sideToMove
	for _, m := range pseudo {
		u := b.MakeMove(m)
		if !b.isColorInCheck(us) {
			dst = append(dst, m)
		}
		b.UnmakeMove(m, u)
	}
	return dst
}

// GeneratePseudoLegal appends all pseudo-legal moves (legal except that they may
// leave the mover's king in check) to dst and returns the extended slice.
func (b *Board) GeneratePseudoLegal(dst []Move) []Move {
	us := b.sideToMove
	for s := Square(0); s < 64; s++ {
		p := b.squares[s]
		if p.IsEmpty() || p.Color() != us {
			continue
		}
		switch p.Type() {
		case Pawn:
			dst = b.genPawn(dst, s)
		case Knight:
			dst = b.genStep(dst, s, knightDeltas[:])
		case Bishop:
			dst = b.genSlide(dst, s, bishopDeltas[:])
		case Rook:
			dst = b.genSlide(dst, s, rookDeltas[:])
		case Queen:
			dst = b.genSlide(dst, s, bishopDeltas[:])
			dst = b.genSlide(dst, s, rookDeltas[:])
		case King:
			dst = b.genStep(dst, s, kingDeltas[:])
			dst = b.genCastling(dst)
		}
	}
	return dst
}

func (b *Board) genStep(dst []Move, from Square, deltas [][2]int) []Move {
	us := b.sideToMove
	f, r := from.File(), from.Rank()
	for _, d := range deltas {
		nf, nr := f+d[0], r+d[1]
		if nf < 0 || nf > 7 || nr < 0 || nr > 7 {
			continue
		}
		to := NewSquare(nf, nr)
		target := b.squares[to]
		if target.IsEmpty() || target.Color() != us {
			dst = append(dst, NewMove(from, to, FlagNormal))
		}
	}
	return dst
}

func (b *Board) genSlide(dst []Move, from Square, deltas [][2]int) []Move {
	us := b.sideToMove
	f, r := from.File(), from.Rank()
	for _, d := range deltas {
		nf, nr := f+d[0], r+d[1]
		for nf >= 0 && nf <= 7 && nr >= 0 && nr <= 7 {
			to := NewSquare(nf, nr)
			target := b.squares[to]
			if target.IsEmpty() {
				dst = append(dst, NewMove(from, to, FlagNormal))
			} else {
				if target.Color() != us {
					dst = append(dst, NewMove(from, to, FlagNormal))
				}
				break
			}
			nf += d[0]
			nr += d[1]
		}
	}
	return dst
}

func (b *Board) genPawn(dst []Move, from Square) []Move {
	us := b.sideToMove
	f, r := from.File(), from.Rank()

	dir, startRank, promoRank := 1, 1, 7
	if us == Black {
		dir, startRank, promoRank = -1, 6, 0
	}

	oneRank := r + dir
	if oneRank < 0 || oneRank > 7 {
		return dst
	}
	dst = b.genPawnPushes(dst, from, f, r, dir, startRank, oneRank, promoRank)
	dst = b.genPawnCaptures(dst, from, f, oneRank, promoRank, us)
	return dst
}

// genPawnPushes adds the single (and, from the start rank, double) forward pushes
// for the pawn on from when the squares ahead are empty.
func (b *Board) genPawnPushes(dst []Move, from Square, f, r, dir, startRank, oneRank, promoRank int) []Move {
	one := NewSquare(f, oneRank)
	if !b.squares[one].IsEmpty() {
		return dst
	}
	dst = b.addPawnMove(dst, from, one, oneRank == promoRank, FlagNormal)
	if r == startRank { // double push
		two := NewSquare(f, r+2*dir)
		if b.squares[two].IsEmpty() {
			dst = append(dst, NewMove(from, two, FlagDoublePawnPush))
		}
	}
	return dst
}

// genPawnCaptures adds the diagonal captures for the pawn on from, including en
// passant.
func (b *Board) genPawnCaptures(dst []Move, from Square, f, oneRank, promoRank int, us Color) []Move {
	for _, df := range [2]int{-1, 1} {
		nf := f + df
		if nf < 0 || nf > 7 {
			continue
		}
		to := NewSquare(nf, oneRank)
		target := b.squares[to]
		switch {
		case !target.IsEmpty() && target.Color() != us:
			dst = b.addPawnMove(dst, from, to, oneRank == promoRank, FlagNormal)
		case to == b.enPassant:
			dst = append(dst, NewMove(from, to, FlagEnPassant))
		}
	}
	return dst
}

// addPawnMove appends either four promotion moves or a single non-promotion
// move. flag is used only for the non-promotion case.
func (b *Board) addPawnMove(dst []Move, from, to Square, promo bool, flag MoveFlag) []Move {
	if promo {
		return append(dst,
			NewMove(from, to, FlagPromoQueen),
			NewMove(from, to, FlagPromoRook),
			NewMove(from, to, FlagPromoBishop),
			NewMove(from, to, FlagPromoKnight),
		)
	}
	return append(dst, NewMove(from, to, flag))
}

func (b *Board) genCastling(dst []Move) []Move {
	us := b.sideToMove
	them := us.Opposite()

	if b.isColorInCheck(us) {
		return dst
	}

	for _, c := range castlingOptions(us) {
		if b.castling.Has(c.right) && b.squaresEmpty(c.empty) && b.squaresSafe(c.safe, them) {
			dst = append(dst, NewMove(c.from, c.to, c.flag))
		}
	}
	return dst
}

// castleOption describes one castling move: the right it needs, the squares that
// must be empty, the squares that must be unattacked, and the king's move.
type castleOption struct {
	right    CastleRights
	empty    []Square
	safe     []Square
	from, to Square
	flag     MoveFlag
}

// castlingOptions returns the two castling moves for the side to move.
func castlingOptions(us Color) []castleOption {
	if us == White {
		return []castleOption{
			{WhiteKingside, []Square{5, 6}, []Square{5, 6}, 4, 6, FlagCastleKingside},
			{WhiteQueenside, []Square{3, 2, 1}, []Square{3, 2}, 4, 2, FlagCastleQueenside},
		}
	}
	return []castleOption{
		{BlackKingside, []Square{61, 62}, []Square{61, 62}, 60, 62, FlagCastleKingside},
		{BlackQueenside, []Square{59, 58, 57}, []Square{59, 58}, 60, 58, FlagCastleQueenside},
	}
}

// squaresEmpty reports whether all the given squares are unoccupied.
func (b *Board) squaresEmpty(sqs []Square) bool {
	for _, s := range sqs {
		if !b.squares[s].IsEmpty() {
			return false
		}
	}
	return true
}

// squaresSafe reports whether none of the given squares is attacked by the side.
func (b *Board) squaresSafe(sqs []Square, by Color) bool {
	for _, s := range sqs {
		if b.IsSquareAttacked(s, by) {
			return false
		}
	}
	return true
}
