package chess

// Undo holds the information needed to reverse a MakeMove.
type Undo struct {
	captured  Piece
	castling  CastleRights
	enPassant Square
	halfMove  int
	fullMove  int
}

// castlingLost maps a square to the castling rights that are revoked when that
// square is the origin or destination of a move (king/rook move, or rook
// capture).
var castlingLost = func() [64]CastleRights {
	var t [64]CastleRights
	t[4] = WhiteKingside | WhiteQueenside  // e1
	t[0] = WhiteQueenside                  // a1
	t[7] = WhiteKingside                   // h1
	t[60] = BlackKingside | BlackQueenside // e8
	t[56] = BlackQueenside                 // a8
	t[63] = BlackKingside                  // h8
	return t
}()

// enPassantCaptureSquare returns the square of the pawn captured en passant when
// a pawn of color mover lands on `to`.
func enPassantCaptureSquare(to Square, mover Color) Square {
	if mover == White {
		return NewSquare(to.File(), to.Rank()-1)
	}
	return NewSquare(to.File(), to.Rank()+1)
}

// MakeMove applies a move to the board and returns an Undo that reverses it.
// The move must be legal (or at least pseudo-legal); behavior is undefined
// otherwise.
func (b *Board) MakeMove(m Move) Undo {
	from, to := m.From(), m.To()
	moving := b.squares[from]
	us := b.sideToMove
	flag := m.Flag()

	u := Undo{
		captured:  b.squares[to],
		castling:  b.castling,
		enPassant: b.enPassant,
		halfMove:  b.halfMove,
		fullMove:  b.fullMove,
	}

	b.enPassant = NoSquare
	b.halfMove++
	if moving.Type() == Pawn || !u.captured.IsEmpty() {
		b.halfMove = 0
	}

	if flag == FlagEnPassant {
		capSq := enPassantCaptureSquare(to, us)
		u.captured = b.squares[capSq]
		b.squares[capSq] = NoPiece
		b.halfMove = 0
	}

	// Move the piece (promoting if necessary).
	b.squares[from] = NoPiece
	placed := moving
	if m.IsPromotion() {
		placed = MakePiece(us, m.Promotion())
	}
	b.setPiece(to, placed)

	// Move the rook when castling.
	switch flag {
	case FlagCastleKingside:
		b.moveRook(rookSquare(to, true))
	case FlagCastleQueenside:
		b.moveRook(rookSquare(to, false))
	}

	b.castling &^= castlingLost[from]
	b.castling &^= castlingLost[to]

	if flag == FlagDoublePawnPush {
		b.enPassant = Square((int(from) + int(to)) / 2)
	}

	if us == Black {
		b.fullMove++
	}
	b.sideToMove = us.Opposite()

	return u
}

// rookSquare returns the (rookFrom, rookTo) squares for a castling move whose
// king lands on kingTo.
func rookSquare(kingTo Square, kingside bool) (rookFrom, rookTo Square) {
	rank := kingTo.Rank()
	if kingside {
		return NewSquare(7, rank), NewSquare(5, rank)
	}
	return NewSquare(0, rank), NewSquare(3, rank)
}

// moveRook relocates a rook from rf to rt during castling.
func (b *Board) moveRook(rf, rt Square) {
	b.squares[rt] = b.squares[rf]
	b.squares[rf] = NoPiece
}

// UnmakeMove reverses a MakeMove given the Undo it returned.
func (b *Board) UnmakeMove(m Move, u Undo) {
	from, to := m.From(), m.To()
	us := b.sideToMove.Opposite() // the side that made the move

	b.sideToMove = us
	b.castling = u.castling
	b.enPassant = u.enPassant
	b.halfMove = u.halfMove
	b.fullMove = u.fullMove

	flag := m.Flag()

	moved := b.squares[to]
	if m.IsPromotion() {
		moved = MakePiece(us, Pawn)
	}
	b.setPiece(from, moved)

	if flag == FlagEnPassant {
		b.squares[to] = NoPiece
		b.squares[enPassantCaptureSquare(to, us)] = u.captured
	} else {
		b.squares[to] = u.captured
	}

	switch flag {
	case FlagCastleKingside:
		rf, rt := rookSquare(to, true)
		b.moveRook(rt, rf)
	case FlagCastleQueenside:
		rf, rt := rookSquare(to, false)
		b.moveRook(rt, rf)
	}
}

// ApplyMove returns a new board with the move applied, leaving the receiver
// unchanged. This clone-based variant suits search strategies (e.g. beam
// search) that keep many independent positions alive at once.
func (b *Board) ApplyMove(m Move) *Board {
	nb := b.Clone()
	nb.MakeMove(m)
	return nb
}
