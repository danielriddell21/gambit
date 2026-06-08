// Package chess is a self-contained chess engine: board state, pieces, move
// generation, legality checking, FEN parsing, make/unmake, and game-result
// detection.
//
// It is the project's only public, reusable surface — an external program can
// depend on it to represent a position and enumerate legal moves without
// pulling in any of gambit's agents, GUI or game-loop code.
//
// Squares are indexed 0..63 with A1=0 and H8=63. Colors, pieces and moves are
// compact value types so move lists are cheap to copy and sort during search.
package chess
