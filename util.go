package tview

import (
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// Text alignment within a box. Also used to align images.
const (
	AlignLeft = iota
	AlignCenter
	AlignRight
	AlignTop    = 0
	AlignBottom = 2
)

// Common regular expressions.
var (
	colorPattern     = regexp.MustCompile(`\[([a-zA-Z]+|#[0-9a-zA-Z]{6}|\-)?(:([a-zA-Z]+|#[0-9a-zA-Z]{6}|\-)?(:([lbidrus]+|\-)?)?)?\]`)
	regionPattern    = regexp.MustCompile(`\["([a-zA-Z0-9_,;: \-\.]*)"\]`)
	escapePattern    = regexp.MustCompile(`\[([a-zA-Z0-9_,;: \-\."#]+)\[(\[*)\]`)
	nonEscapePattern = regexp.MustCompile(`(\[[a-zA-Z0-9_,;: \-\."#]+\[*)\]`)
	boundaryPattern  = regexp.MustCompile(`(([,\.\-:;!\?&#+]|\n)[ \t\f\r]*|([ \t\f\r]+))`)
	spacePattern     = regexp.MustCompile(`\s+`)
)

// Positions of substrings in regular expressions.
const (
	colorForegroundPos = 1
	colorBackgroundPos = 3
	colorFlagPos       = 5
)

// The number of colors available in the terminal.
var availableColors = 256

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

	// Determine the number of colors available in the terminal.
	info, err := tcell.LookupTerminfo(os.Getenv("TERM"))
	if err == nil {
		availableColors = info.Colors
	}
}

// styleFromTag takes the given style, defined by a foreground color (fgColor),
// a background color (bgColor), and style attributes, and modifies it based on
// the substrings (tagSubstrings) extracted by the regular expression for color
// tags. The new colors and attributes are returned where empty strings mean
// "don't modify" and a dash ("-") means "reset to default".
func styleFromTag(fgColor, bgColor, attributes string, tagSubstrings []string) (newFgColor, newBgColor, newAttributes string) {
	if tagSubstrings[colorForegroundPos] != "" {
		color := tagSubstrings[colorForegroundPos]
		if color == "-" {
			fgColor = "-"
		} else if color != "" {
			fgColor = color
		}
	}

	if tagSubstrings[colorBackgroundPos-1] != "" {
		color := tagSubstrings[colorBackgroundPos]
		if color == "-" {
			bgColor = "-"
		} else if color != "" {
			bgColor = color
		}
	}

	if tagSubstrings[colorFlagPos-1] != "" {
		flags := tagSubstrings[colorFlagPos]
		if flags == "-" {
			attributes = "-"
		} else if flags != "" {
			attributes = flags
		}
	}

	return fgColor, bgColor, attributes
}

// overlayStyle calculates a new style based on "style" and applying tag-based
// colors/attributes to it (see also styleFromTag()).
func overlayStyle(style tcell.Style, fgColor, bgColor, attributes string) tcell.Style {
	_, _, defAttr := style.Decompose()

	if fgColor != "" && fgColor != "-" {
		style = style.Foreground(tcell.GetColor(fgColor))
	}

	if bgColor != "" && bgColor != "-" {
		style = style.Background(tcell.GetColor(bgColor))
	}

	if attributes == "-" {
		style = style.Bold(defAttr&tcell.AttrBold > 0).
			Italic(defAttr&tcell.AttrItalic > 0).
			Blink(defAttr&tcell.AttrBlink > 0).
			Reverse(defAttr&tcell.AttrReverse > 0).
			Underline(defAttr&tcell.AttrUnderline > 0).
			Dim(defAttr&tcell.AttrDim > 0)
	} else if attributes != "" {
		style = style.Normal()
		for _, flag := range attributes {
			switch flag {
			case 'l':
				style = style.Blink(true)
			case 'b':
				style = style.Bold(true)
			case 'i':
				style = style.Italic(true)
			case 'd':
				style = style.Dim(true)
			case 'r':
				style = style.Reverse(true)
			case 'u':
				style = style.Underline(true)
			case 's':
				style = style.StrikeThrough(true)
			}
		}
	}

	return style
}

// decomposeString returns information about a string which may contain color
// tags or region tags, depending on which ones are requested to be found. It
// returns the indices of the style tags (as returned by
// re.FindAllStringIndex()), the style tags themselves (as returned by
// re.FindAllStringSubmatch()), the indices of region tags and the region tags
// themselves, the indices of an escaped tags (only if at least style tags or
// region tags are requested), the string stripped by any tags and escaped, and
// the screen width of the stripped string.
func decomposeString(text string, findColors, findRegions bool) (colorIndices [][]int, colors [][]string, regionIndices [][]int, regions [][]string, escapeIndices [][]int, stripped string, width int) {
	// Shortcut for the trivial case.
	if !findColors && !findRegions {
		return nil, nil, nil, nil, nil, text, uniseg.StringWidth(text)
	}

	// Get positions of any tags.
	if findColors {
		colorIndices = colorPattern.FindAllStringIndex(text, -1)
		colors = colorPattern.FindAllStringSubmatch(text, -1)
	}
	if findRegions {
		regionIndices = regionPattern.FindAllStringIndex(text, -1)
		regions = regionPattern.FindAllStringSubmatch(text, -1)
	}
	escapeIndices = escapePattern.FindAllStringIndex(text, -1)

	// Because the color pattern detects empty tags, we need to filter them out.
	for i := len(colorIndices) - 1; i >= 0; i-- {
		if colorIndices[i][1]-colorIndices[i][0] == 2 {
			colorIndices = append(colorIndices[:i], colorIndices[i+1:]...)
			colors = append(colors[:i], colors[i+1:]...)
		}
	}

	// Make a (sorted) list of all tags.
	allIndices := make([][3]int, 0, len(colorIndices)+len(regionIndices)+len(escapeIndices))
	for indexType, index := range [][][]int{colorIndices, regionIndices, escapeIndices} {
		for _, tag := range index {
			allIndices = append(allIndices, [3]int{tag[0], tag[1], indexType})
		}
	}
	sort.Slice(allIndices, func(i int, j int) bool {
		return allIndices[i][0] < allIndices[j][0]
	})

	// Remove the tags from the original string.
	var from int
	buf := make([]byte, 0, len(text))
	for _, indices := range allIndices {
		if indices[2] == 2 { // Escape sequences are not simply removed.
			buf = append(buf, []byte(text[from:indices[1]-2])...)
			buf = append(buf, ']')
			from = indices[1]
		} else {
			buf = append(buf, []byte(text[from:indices[0]])...)
			from = indices[1]
		}
	}
	buf = append(buf, text[from:]...)
	stripped = string(buf)

	// Get the width of the stripped string.
	width = uniseg.StringWidth(stripped)

	return
}

// Print prints text onto the screen into the given box at (x,y,maxWidth,1),
// not exceeding that box. "align" is one of AlignLeft, AlignCenter, or
// AlignRight. The screen's background color will not be changed.
//
// You can change the colors and text styles mid-text by inserting a style tag.
// See the package description for details.
//
// Returns the number of actual bytes of the text printed (including style tags)
// and the actual width used for the printed runes.
func Print(screen tcell.Screen, text string, x, y, maxWidth, align int, color tcell.Color) (int, int) {
	start, end, width := printWithStyle(screen, text, x, y, 0, maxWidth, align, tcell.StyleDefault.Foreground(color), true)
	return end - start, width
}

// printWithStyle works like Print() but it takes a style instead of just a
// foreground color. The skipWidth parameter specifies the number of cells
// skipped at the beginning of the text. It returns the start index, end index
// (exclusively), and screen width of the text actually printed. If
// maintainBackground is "true", the existing screen background is not changed
// (i.e. the style's background color is ignored).
func printWithStyle(screen tcell.Screen, text string, x, y, skipWidth, maxWidth, align int, style tcell.Style, maintainBackground bool) (start, end, printedWidth int) {
	totalWidth, totalHeight := screen.Size()
	if maxWidth <= 0 || len(text) == 0 || y < 0 || y >= totalHeight {
		return 0, 0, 0
	}

	// If we don't overwrite the background, we use the default color.
	if maintainBackground {
		style = style.Background(tcell.ColorDefault)
	}

	// Skip beginning and measure width.
	var (
		state     *stepState
		textWidth int
	)
	str := text
	for len(str) > 0 {
		_, str, state = step(str, state, stepOptionsStyle)
		if skipWidth > 0 {
			skipWidth -= state.Width()
			if skipWidth <= 0 {
				text = str
				style = state.Style()
			}
			start += state.GrossLength()
		} else {
			textWidth += state.Width()
		}
	}

	// Reduce all alignments to AlignLeft.
	if align == AlignRight {
		// Chop off characters on the left until it fits.
		state = nil
		for len(text) > 0 && textWidth > maxWidth {
			_, text, state = step(text, state, stepOptionsStyle)
			textWidth -= state.Width()
			start += state.GrossLength()
			style = state.Style()
		}
		x, maxWidth = x+maxWidth-textWidth, textWidth
	} else if align == AlignCenter {
		// Chop off characters on the left until it fits.
		state = nil
		subtracted := (textWidth - maxWidth) / 2
		for len(text) > 0 && subtracted > 0 {
			_, text, state = step(text, state, stepOptionsStyle)
			subtracted -= state.Width()
			textWidth -= state.Width()
			start += state.GrossLength()
			style = state.Style()
		}
		if textWidth < maxWidth {
			x, maxWidth = x+maxWidth/2-textWidth/2, textWidth
		}
	}

	// Draw left-aligned text.
	end = start
	rightBorder := x + maxWidth
	state = &stepState{
		unisegState: -1,
		style:       style,
	}
	for len(text) > 0 && x < rightBorder && x < totalWidth {
		var c string
		c, text, state = step(text, state, stepOptionsStyle)
		if c == "" {
			break // We don't care about the style at the end.
		}
		runes := []rune(c)
		width := state.Width()

		finalStyle := state.Style()
		if maintainBackground {
			_, backgroundColor, _ := finalStyle.Decompose()
			if backgroundColor == tcell.ColorDefault {
				_, _, existingStyle, _ := screen.GetContent(x, y)
				_, background, _ := existingStyle.Decompose()
				finalStyle = finalStyle.Background(background)
			}
		}
		for offset := width - 1; offset >= 0; offset-- {
			// To avoid undesired effects, we populate all cells.
			if offset == 0 {
				screen.SetContent(x+offset, y, runes[0], runes[1:], finalStyle)
			} else {
				screen.SetContent(x+offset, y, ' ', nil, finalStyle)
			}
		}

		x += width
		end += state.GrossLength()
		printedWidth += width
	}

	return
}

// PrintSimple prints white text to the screen at the given position.
func PrintSimple(screen tcell.Screen, text string, x, y int) {
	Print(screen, text, x, y, math.MaxInt32, AlignLeft, Styles.PrimaryTextColor)
}

// TaggedStringWidth returns the width of the given string needed to print it on
// screen. The text may contain style tags which are not counted.
func TaggedStringWidth(text string) (width int) {
	var state *stepState
	for len(text) > 0 {
		_, text, state = step(text, state, stepOptionsStyle)
		width += state.Width()
	}
	return
}

// WordWrap splits a text such that each resulting line does not exceed the
// given screen width. Split points are determined using the algorithm described
// in [Unicode Standard Annex #14] .
//
// This function considers style tags to have no width.
//
// [Unicode Standard Annex #14]: https://www.unicode.org/reports/tr14/
func WordWrap(text string, width int) (lines []string) {
	if width <= 0 {
		return
	}

	var (
		state                                              *stepState
		lineWidth, lineLength, lastOption, lastOptionWidth int
	)
	str := text
	for len(str) > 0 {
		// Parse the next character.
		var c string
		c, str, state = step(str, state, stepOptionsStyle)
		cWidth := state.Width()

		// Would it exceed the line width?
		if lineWidth+cWidth > width {
			if lastOptionWidth == 0 {
				// No split point so far. Just split at the current position.
				lines = append(lines, text[:lineLength])
				text = text[lineLength:]
				lineWidth, lineLength, lastOption, lastOptionWidth = 0, 0, 0, 0
			} else {
				// Split at the last split point.
				lines = append(lines, text[:lastOption])
				text = text[lastOption:]
				lineWidth -= lastOptionWidth
				lineLength -= lastOption
				lastOption, lastOptionWidth = 0, 0
			}
		}

		// Move ahead.
		lineWidth += cWidth
		lineLength += state.GrossLength()

		// Check for split points.
		if lineBreak, optional := state.LineBreak(); lineBreak {
			if optional {
				// Remember this split point.
				lastOption = lineLength
				lastOptionWidth = lineWidth
			} else if str != "" || c != "" && uniseg.HasTrailingLineBreakInString(c) {
				// We must split here.
				lines = append(lines, strings.TrimRight(text[:lineLength], "\n\r"))
				text = text[lineLength:]
				lineWidth, lineLength, lastOption, lastOptionWidth = 0, 0, 0, 0
			}
		}
	}
	lines = append(lines, text)

	return
}

// Escape escapes the given text such that color and/or region tags are not
// recognized and substituted by the print functions of this package. For
// example, to include a tag-like string in a box title or in a TextView:
//
//	box.SetTitle(tview.Escape("[squarebrackets]"))
//	fmt.Fprint(textView, tview.Escape(`["quoted"]`))
func Escape(text string) string {
	return nonEscapePattern.ReplaceAllString(text, "$1[]")
}

// iterateString iterates through the given string one printed character at a
// time. For each such character, the callback function is called with the
// Unicode code points of the character (the first rune and any combining runes
// which may be nil if there aren't any), the starting position (in bytes)
// within the original string, its length in bytes, the screen position of the
// character, the screen width of it, and a boundaries value which includes
// word/sentence boundary or line break information (see the
// github.com/rivo/uniseg package, Step() function, for more information). The
// iteration stops if the callback returns true. This function returns true if
// the iteration was stopped before the last character.
func iterateString(text string, callback func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth, boundaries int) bool) bool {
	var screenPos, textPos, boundaries int

	state := -1
	for len(text) > 0 {
		var cluster string
		cluster, text, boundaries, state = uniseg.StepString(text, state)

		width := boundaries >> uniseg.ShiftWidth
		runes := []rune(cluster)
		var comb []rune
		if len(runes) > 1 {
			comb = runes[1:]
		}

		if callback(runes[0], comb, textPos, len(cluster), screenPos, width, boundaries) {
			return true
		}

		screenPos += width
		textPos += len(cluster)
	}

	return false
}

// stripTags strips style tags from the given string. (Region tags are not
// stripped.)
func stripTags(text string) string {
	var (
		str   strings.Builder
		state *stepState
	)
	for len(text) > 0 {
		var c string
		c, text, state = step(text, state, stepOptionsStyle)
		str.WriteString(c)
	}
	return str.String()
}
