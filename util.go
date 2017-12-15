package tview

import "github.com/gdamore/tcell"

// Text alignment within a box.
const (
	AlignLeft = iota
	AlignCenter
	AlignRight
)

// Print prints text onto the screen at position (x,y). "align" is one of the
// Align constants and will affect the direction starting at (x,y) into which
// the text is printed. The screen's background color will be maintained. The
// number of runes printed will not exceed "maxWidth".
//
// Returns the number of runes printed.
func Print(screen tcell.Screen, text string, x, y, maxWidth, align int, color tcell.Color) int {
	// We deal with runes, not with bytes.
	runes := []rune(text)
	if maxWidth < 0 {
		return 0
	}

	// Shorten text if it's too long.
	if len(runes) > maxWidth {
		switch align {
		case AlignCenter:
			trim := (len(runes) - maxWidth) / 2
			runes = runes[trim : maxWidth+trim]
		case AlignRight:
			runes = runes[len(runes)-maxWidth:]
		default: // AlignLeft.
			runes = runes[:maxWidth]
		}
	}

	// Adjust x-position.
	if align == AlignCenter {
		x -= len(runes) / 2
	} else if align == AlignRight {
		x -= len(runes) - 1
	}

	// Draw text.
	for _, ch := range runes {
		_, _, style, _ := screen.GetContent(x, y)
		style = style.Foreground(color)
		screen.SetContent(x, y, ch, nil, style)
		x++
	}

	return len(runes)
}
