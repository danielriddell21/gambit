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

	// Single push.
	oneRank := r + dir
	if oneRank >= 0 && oneRank <= 7 {
		one := NewSquare(f, oneRank)
		if b.squares[one].IsEmpty() {
			dst = b.addPawnMove(dst, from, one, oneRank == promoRank, FlagNormal)
			// Double push.
			if r == startRank {
				two := NewSquare(f, r+2*dir)
				if b.squares[two].IsEmpty() {
					dst = append(dst, NewMove(from, two, FlagDoublePawnPush))
				}
			}
		}
	}

	// Captures (including en passant).
	if oneRank >= 0 && oneRank <= 7 {
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

	if us == White {
		if b.castling.Has(WhiteKingside) &&
			b.squares[5].IsEmpty() && b.squares[6].IsEmpty() &&
			!b.IsSquareAttacked(5, them) && !b.IsSquareAttacked(6, them) {
			dst = append(dst, NewMove(4, 6, FlagCastleKingside))
		}
		if b.castling.Has(WhiteQueenside) &&
			b.squares[3].IsEmpty() && b.squares[2].IsEmpty() && b.squares[1].IsEmpty() &&
			!b.IsSquareAttacked(3, them) && !b.IsSquareAttacked(2, them) {
			dst = append(dst, NewMove(4, 2, FlagCastleQueenside))
		}
	} else {
		if b.castling.Has(BlackKingside) &&
			b.squares[61].IsEmpty() && b.squares[62].IsEmpty() &&
			!b.IsSquareAttacked(61, them) && !b.IsSquareAttacked(62, them) {
			dst = append(dst, NewMove(60, 62, FlagCastleKingside))
		}
		if b.castling.Has(BlackQueenside) &&
			b.squares[59].IsEmpty() && b.squares[58].IsEmpty() && b.squares[57].IsEmpty() &&
			!b.IsSquareAttacked(59, them) && !b.IsSquareAttacked(58, them) {
			dst = append(dst, NewMove(60, 58, FlagCastleQueenside))
		}
	}

	return dst
}
