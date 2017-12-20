package tview

import (
	"math"
	"strings"

	"github.com/gdamore/tcell"
)

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

// PrintSimple prints white text to the screen at the given position.
func PrintSimple(screen tcell.Screen, text string, x, y int) {
	Print(screen, text, x, y, math.MaxInt64, AlignLeft, tcell.ColorWhite)
}

// WordWrap splits a text such that each resulting line does not exceed the
// given width. Possible split points are after commas, dots, dashes, and any
// whitespace. Whitespace at split points will be dropped.
//
// Text is always split at newline characters ('\n').
func WordWrap(text string, width int) (lines []string) {
	x := 0
	start := 0
	candidate := -1 // -1 = no candidate yet.
	startAfterCandidate := 0
	countAfterCandidate := 0
	var evaluatingCandidate bool
	text = strings.TrimSpace(text)

	for pos, ch := range text {
		if !evaluatingCandidate && x >= width {
			// We've exceeded the width, we must split.
			if candidate >= 0 {
				lines = append(lines, text[start:candidate])
				start = startAfterCandidate
				x = countAfterCandidate
			} else {
				lines = append(lines, text[start:pos])
				start = pos
				x = 0
			}
			candidate = -1
			evaluatingCandidate = false
		}

		switch ch {
		// We have a candidate.
		case ',', '.', '-':
			if x > 0 {
				candidate = pos + 1
				evaluatingCandidate = true
			}
		// If we've had a candidate, skip whitespace. If not, we have a candidate.
		case ' ', '\t':
			if x > 0 && !evaluatingCandidate {
				candidate = pos
				evaluatingCandidate = true
			}
		// Split in any case.
		case '\n':
			lines = append(lines, text[start:pos])
			start = pos + 1
			evaluatingCandidate = false
			countAfterCandidate = 0
			x = 0
			continue
		// If we've had a candidate, we have a new start.
		default:
			if evaluatingCandidate {
				startAfterCandidate = pos
				evaluatingCandidate = false
				countAfterCandidate = 0
			}
		}
		x++
		countAfterCandidate++
	}

	// Process remaining text.
	text = strings.TrimSpace(text[start:])
	if len(text) > 0 {
		lines = append(lines, text)
	}

	return
}
