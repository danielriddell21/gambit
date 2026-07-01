package chess

// Perft counts the leaf nodes of the move tree to the given depth. It is the
// standard correctness check for move generation.
func Perft(b *Board, depth int) uint64 {
	if depth == 0 {
		return 1
	}
	moves := b.GenerateMoves(nil)
	if depth == 1 {
		return uint64(len(moves))
	}
	var nodes uint64
	for _, m := range moves {
		u := b.MakeMove(m)
		nodes += Perft(b, depth-1)
		b.UnmakeMove(m, u)
	}
	return nodes
}
