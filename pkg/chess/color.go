package chess

// Color identifies which side a piece belongs to or which side is to move.
type Color uint8

// The two colors.
const (
	White Color = iota
	Black
)

// Opposite returns the other color.
func (c Color) Opposite() Color {
	return c ^ 1
}

// String returns "white" or "black".
func (c Color) String() string {
	if c == White {
		return "white"
	}
	return "black"
}
