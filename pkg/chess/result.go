package chess

// Result is the outcome of a game.
type Result uint8

// Game outcomes.
const (
	InProgress Result = iota
	WhiteWins
	BlackWins
	Draw
)

// String returns a PGN-style result token.
func (r Result) String() string {
	switch r {
	case WhiteWins:
		return "1-0"
	case BlackWins:
		return "0-1"
	case Draw:
		return "1/2-1/2"
	default:
		return "*"
	}
}

// DrawReason explains why a game was drawn.
type DrawReason uint8

// Draw reasons.
const (
	NotDraw DrawReason = iota
	Stalemate
	FiftyMoveRule
	ThreefoldRepetition
	InsufficientMaterial
)

// String returns a human-readable draw reason.
func (d DrawReason) String() string {
	switch d {
	case Stalemate:
		return "stalemate"
	case FiftyMoveRule:
		return "fifty-move rule"
	case ThreefoldRepetition:
		return "threefold repetition"
	case InsufficientMaterial:
		return "insufficient material"
	default:
		return ""
	}
}

// hasLegalMoves reports whether the side to move has at least one legal move.
func (b *Board) hasLegalMoves() bool {
	return len(b.GenerateMoves(nil)) > 0
}

// IsCheckmate reports whether the side to move is checkmated.
func (b *Board) IsCheckmate() bool {
	return b.InCheck() && !b.hasLegalMoves()
}

// IsStalemate reports whether the side to move is stalemated.
func (b *Board) IsStalemate() bool {
	return !b.InCheck() && !b.hasLegalMoves()
}

// IsInsufficientMaterial reports whether neither side has enough material to
// force checkmate (K vs K, K+minor vs K, and K+B vs K+B with same-colored
// bishops are treated as insufficient).
func (b *Board) IsInsufficientMaterial() bool {
	var knights, bishops int
	bishopColors := 0
	for s := Square(0); s < 64; s++ {
		p := b.squares[s]
		if p.IsEmpty() {
			continue
		}
		switch p.Type() {
		case King:
			// kings are always present
		case Knight:
			knights++
		case Bishop:
			bishops++
			if (s.File()+s.Rank())%2 == 0 {
				bishopColors |= 1
			} else {
				bishopColors |= 2
			}
		default:
			// any pawn, rook or queen is sufficient material
			return false
		}
	}

	switch {
	case knights == 0 && bishops == 0:
		return true // K vs K
	case knights == 1 && bishops == 0:
		return true // K+N vs K
	case knights == 0 && bishops >= 1 && bishopColors != 3:
		return true // only bishops, all on one color complex
	default:
		return false
	}
}

// Status returns the outcome derivable from the current position alone
// (checkmate, stalemate, insufficient material). Threefold repetition and the
// fifty-move rule depend on game history and are handled by callers; the
// fifty-move clock is exposed via HalfMoveClock.
func (b *Board) Status() (Result, DrawReason) {
	if !b.hasLegalMoves() {
		if b.InCheck() {
			if b.sideToMove == White {
				return BlackWins, NotDraw
			}
			return WhiteWins, NotDraw
		}
		return Draw, Stalemate
	}
	if b.IsInsufficientMaterial() {
		return Draw, InsufficientMaterial
	}
	return InProgress, NotDraw
}
