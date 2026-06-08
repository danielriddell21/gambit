package chess

import "math/rand/v2"

// Zobrist keys. They are generated once with a fixed seed so hashes are stable
// across runs (useful for tests and transposition tables).
var (
	zobristPieces   [16][64]uint64 // indexed by Piece value then square
	zobristCastling [16]uint64     // indexed by CastleRights value
	zobristEnPass   [8]uint64      // indexed by file
	zobristSide     uint64
)

func init() {
	rng := rand.New(rand.NewPCG(0x9E3779B97F4A7C15, 0xBF58476D1CE4E5B9))
	for p := range zobristPieces {
		for s := range zobristPieces[p] {
			zobristPieces[p][s] = rng.Uint64()
		}
	}
	for i := range zobristCastling {
		zobristCastling[i] = rng.Uint64()
	}
	for i := range zobristEnPass {
		zobristEnPass[i] = rng.Uint64()
	}
	zobristSide = rng.Uint64()
}

// Hash returns a Zobrist hash of the position, suitable for repetition
// detection. Positions that are equal for repetition purposes (same pieces,
// side to move, castling rights and en-passant possibility) hash equally.
func (b *Board) Hash() uint64 {
	var h uint64
	for s := Square(0); s < 64; s++ {
		p := b.squares[s]
		if !p.IsEmpty() {
			h ^= zobristPieces[p][s]
		}
	}
	h ^= zobristCastling[b.castling]
	if b.enPassant != NoSquare {
		h ^= zobristEnPass[b.enPassant.File()]
	}
	if b.sideToMove == Black {
		h ^= zobristSide
	}
	return h
}
