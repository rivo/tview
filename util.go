package tview

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
)

// Text alignment within a box.
const (
	AlignLeft = iota
	AlignCenter
	AlignRight
)

// Semigraphical runes.
const (
	GraphicsHoriBar             = '\u2500'
	GraphicsVertBar             = '\u2502'
	GraphicsTopLeftCorner       = '\u250c'
	GraphicsTopRightCorner      = '\u2510'
	GraphicsBottomRightCorner   = '\u2518'
	GraphicsBottomLeftCorner    = '\u2514'
	GraphicsDbVertBar           = '\u2550'
	GraphicsDbHorBar            = '\u2551'
	GraphicsDbTopLeftCorner     = '\u2554'
	GraphicsDbTopRightCorner    = '\u2557'
	GraphicsDbBottomRightCorner = '\u255d'
	GraphicsDbBottomLeftCorner  = '\u255a'
	GraphicsRightT              = '\u2524'
	GraphicsLeftT               = '\u251c'
	GraphicsTopT                = '\u252c'
	GraphicsBottomT             = '\u2534'
	GraphicsCross               = '\u253c'
	GraphicsEllipsis            = '\u2026'
)

// Common regular expressions.
var (
	colorPattern    = regexp.MustCompile(`\[([a-zA-Z]+|#[0-9a-zA-Z]{6})\]`)
	regionPattern   = regexp.MustCompile(`\["([a-zA-Z0-9_,;: \-\.]*)"\]`)
	escapePattern   = regexp.MustCompile(`\[("[a-zA-Z0-9_,;: \-\.]*"|[a-zA-Z]+|#[0-9a-zA-Z]{6})\[(\[*)\]`)
	boundaryPattern = regexp.MustCompile("([[:punct:]]\\s*|\\s+)")
	spacePattern    = regexp.MustCompile(`\s+`)
)

// Predefined InputField acceptance functions.
var (
	// InputFieldInteger accepts integers.
	InputFieldInteger func(text string, ch rune) bool

	// InputFieldFloat accepts floating-point numbers.
	InputFieldFloat func(text string, ch rune) bool

	// InputFieldMaxLength returns an input field accept handler which accepts
	// input strings up to a given length. Use it like this:
	//
	//   inputField.SetAcceptanceFunc(InputFieldMaxLength(10)) // Accept up to 10 characters.
	InputFieldMaxLength func(maxLength int) func(text string, ch rune) bool
)

// Package initialization.
func init() {
	// Initialize the predefined input field handlers.
	InputFieldInteger = func(text string, ch rune) bool {
		if text == "-" {
			return true
		}
		_, err := strconv.Atoi(text)
		return err == nil
	}
	InputFieldFloat = func(text string, ch rune) bool {
		if text == "-" || text == "." || text == "-." {
			return true
		}
		_, err := strconv.ParseFloat(text, 64)
		return err == nil
	}
	InputFieldMaxLength = func(maxLength int) func(text string, ch rune) bool {
		return func(text string, ch rune) bool {
			return len([]rune(text)) <= maxLength
		}
	}
}

// Print prints text onto the screen into the given box at (x,y,maxWidth,1),
// not exceeding that box. "align" is one of AlignLeft, AlignCenter, or
// AlignRight. The screen's background color will not be changed.
//
// You can change the text color mid-text by inserting a color tag. See the
// package description for details.
//
// Returns the number of actual runes printed (not including color tags) and the
// actual width used for the printed runes.
func Print(screen tcell.Screen, text string, x, y, maxWidth, align int, color tcell.Color) (int, int) {
	if maxWidth < 0 {
		return 0, 0
	}

	// Get positions of color and escape tags. Remove them from original string.
	colorIndices := colorPattern.FindAllStringIndex(text, -1)
	colors := colorPattern.FindAllStringSubmatch(text, -1)
	escapeIndices := escapePattern.FindAllStringIndex(text, -1)
	strippedText := escapePattern.ReplaceAllString(colorPattern.ReplaceAllString(text, ""), "[$1$2]")

	// We deal with runes, not with bytes.
	runes := []rune(strippedText)

	// This helper function takes positions for a substring of "runes" and a start
	// color and returns the substring with the original tags and the new start
	// color.
	substring := func(from, to int, color tcell.Color) (string, tcell.Color) {
		var colorPos, escapePos, runePos, startPos int
		for pos := range text {
			// Handle color tags.
			if colorPos < len(colorIndices) && pos >= colorIndices[colorPos][0] && pos < colorIndices[colorPos][1] {
				if pos == colorIndices[colorPos][1]-1 {
					if runePos <= from {
						color = tcell.GetColor(colors[colorPos][1])
					}
					colorPos++
				}
				continue
			}

			// Handle escape tags.
			if escapePos < len(escapeIndices) && pos >= escapeIndices[escapePos][0] && pos < escapeIndices[escapePos][1] {
				if pos == escapeIndices[escapePos][1]-1 {
					escapePos++
				} else if pos == escapeIndices[escapePos][1]-2 {
					continue
				}
			}

			// Check boundaries.
			if runePos == from {
				startPos = pos
			} else if runePos >= to {
				return text[startPos:pos], color
			}

			runePos++
		}

		return text[startPos:len(text)], color
	}

	// We want to reduce everything to AlignLeft.
	if align == AlignRight {
		width := 0
		start := len(runes)
		for index := start - 1; index >= 0; index-- {
			w := runewidth.RuneWidth(runes[index])
			if width+w > maxWidth {
				break
			}
			width += w
			start = index
		}
		text, color = substring(start, len(runes), color)
		return Print(screen, text, x+maxWidth-width, y, width, AlignLeft, color)
	} else if align == AlignCenter {
		width := runewidth.StringWidth(strippedText)
		if width == maxWidth {
			// Use the exact space.
			return Print(screen, text, x, y, maxWidth, AlignLeft, color)
		} else if width < maxWidth {
			// We have more space than we need.
			half := (maxWidth - width) / 2
			return Print(screen, text, x+half, y, maxWidth-half, AlignLeft, color)
		} else {
			// Chop off runes until we have a perfect fit.
			var choppedLeft, choppedRight, leftIndex, rightIndex int
			rightIndex = len(runes) - 1
			for rightIndex > leftIndex && width-choppedLeft-choppedRight > maxWidth {
				leftWidth := runewidth.RuneWidth(runes[leftIndex])
				rightWidth := runewidth.RuneWidth(runes[rightIndex])
				if choppedLeft < choppedRight {
					choppedLeft += leftWidth
					leftIndex++
				} else {
					choppedRight += rightWidth
					rightIndex--
				}
			}
			text, color = substring(leftIndex, rightIndex, color)
			return Print(screen, text, x, y, maxWidth, AlignLeft, color)
		}
	}

	// Draw text.
	drawn := 0
	drawnWidth := 0
	var colorPos, escapePos int
	for pos, ch := range text {
		// Handle color tags.
		if colorPos < len(colorIndices) && pos >= colorIndices[colorPos][0] && pos < colorIndices[colorPos][1] {
			if pos == colorIndices[colorPos][1]-1 {
				color = tcell.GetColor(colors[colorPos][1])
				colorPos++
			}
			continue
		}

		// Handle escape tags.
		if escapePos < len(escapeIndices) && pos >= escapeIndices[escapePos][0] && pos < escapeIndices[escapePos][1] {
			if pos == escapeIndices[escapePos][1]-1 {
				escapePos++
			} else if pos == escapeIndices[escapePos][1]-2 {
				continue
			}
		}

		// Check if we have enough space for this rune.
		chWidth := runewidth.RuneWidth(ch)
		if drawnWidth+chWidth > maxWidth {
			break
		}
		finalX := x + drawnWidth

		// Print the rune.
		_, _, style, _ := screen.GetContent(finalX, y)
		style = style.Foreground(color)
		for offset := 0; offset < chWidth; offset++ {
			// To avoid undesired effects, we place the same character in all cells.
			screen.SetContent(finalX+offset, y, ch, nil, style)
		}

		drawn++
		drawnWidth += chWidth
	}

	return drawn, drawnWidth
}

// PrintSimple prints white text to the screen at the given position.
func PrintSimple(screen tcell.Screen, text string, x, y int) {
	Print(screen, text, x, y, math.MaxInt32, AlignLeft, Styles.PrimaryTextColor)
}

// StringWidth returns the width of the given string needed to print it on
// screen. The text may contain color tags which are not counted.
func StringWidth(text string) int {
	return runewidth.StringWidth(escapePattern.ReplaceAllString(colorPattern.ReplaceAllString(text, ""), "[$1$2]"))
}

// WordWrap splits a text such that each resulting line does not exceed the
// given screen width. Possible split points are after any punctuation or
// whitespace. Whitespace after split points will be dropped.
//
// This function considers color tags to have no width.
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
	colorIndices := colorPattern.FindAllStringIndex(text, -1)
	escapeIndices := escapePattern.FindAllStringIndex(text, -1)

	var colorPos, escapePos int
	for pos, ch := range text {
		// Skip color tags.
		if colorPos < len(colorIndices) && pos >= colorIndices[colorPos][0] && pos < colorIndices[colorPos][1] {
			if pos == colorIndices[colorPos][1]-1 {
				colorPos++
			}
			continue
		}

		// Handle escape tags.
		if escapePos < len(escapeIndices) && pos >= escapeIndices[escapePos][0] && pos < escapeIndices[escapePos][1] {
			if pos == escapeIndices[escapePos][1]-1 {
				escapePos++
			} else if pos == escapeIndices[escapePos][1]-2 {
				continue
			}
		}

		// What's the width of this rune?
		chWidth := runewidth.RuneWidth(ch)

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

		switch {
		// We have a candidate.
		case ch >= '!' && ch <= '/', ch >= ':' && ch <= '@', ch >= '[' && ch <= '`', ch >= '{' && ch <= '~':
			if x > 0 {
				candidate = pos + 1
				evaluatingCandidate = true
			}
			// If we've had a candidate, skip whitespace. If not, we have a candidate.
		case ch == ' ', ch == '\t':
			if x > 0 && !evaluatingCandidate {
				candidate = pos
				evaluatingCandidate = true
			}
			// Split in any case.
		case ch == '\n':
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
		x += chWidth
		countAfterCandidate += chWidth
	}

	// Process remaining text.
	text = strings.TrimSpace(text[start:])
	if len(text) > 0 {
		lines = append(lines, text)
	}

	return
}
