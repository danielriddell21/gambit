// Package assets embeds binary assets (currently the font used to render chess
// pieces) so they can be shared by the GUI and the demo generator without a
// separate copy.
package assets

import _ "embed"

// Font is the DejaVu Sans TTF, which includes the Unicode chess glyphs
// (U+2654..U+265F). DejaVu Sans is freely redistributable.
//
//go:embed DejaVuSans.ttf
var Font []byte
