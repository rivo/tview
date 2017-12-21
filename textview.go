package tview

import (
	"math"
	"regexp"
	"sync"
	"unicode/utf8"

	"github.com/gdamore/tcell"
)

// textColors maps color strings which may be embedded in text sent to a
// TextView to their tcell counterparts.
var textColors = map[string]tcell.Color{
	"red":    tcell.ColorRed,
	"white":  tcell.ColorWhite,
	"yellow": tcell.ColorYellow,
	"blue":   tcell.ColorBlue,
	"green":  tcell.ColorGreen,
}

// A regular expression commonly used throughout the TextView class.
var colorPattern = regexp.MustCompile(`\[(white|yellow|blue|green|red)\]`)

// textViewIndex contains information about each line displayed in the text
// view.
type textViewIndex struct {
	Line  int         // The index into the "buffer" variable.
	Pos   int         // The index into the "buffer" string.
	Color tcell.Color // The starting color.
}

// TextView is a box which displays text. It implements the Reader interface so
// you can stream text to it.
//
// If the text view is scrollable (the default), text is kept in a buffer and
// can be navigated using the arrow keys, Ctrl-F and Ctrl-B for page jumps, "g"
// for the beginning of the text, and "G" for the end of the text.
//
// If the text is not scrollable, any text above the top line is discarded.
//
// If dynamic colors are enabled, text color can be changed dynamically by
// embedding it into square brackets. For example,
//
//   "This is a [red]warning[white]!"
//
// will print the word "warning" in red. The following colors are currently
// supported: white, yellow, blue, green, red.
type TextView struct {
	sync.Mutex
	*Box

	// The text buffer.
	buffer []string

	// The processed line index. This is nil if the buffer has changed and needs
	// to be re-indexed.
	index []*textViewIndex

	// The display width for which the index is created.
	indexWidth int

	// The last bytes that have been received but are not part of the buffer yet.
	recentBytes []byte

	// The index of the first line shown in the text view.
	lineOffset int

	// If set to true, the text view will always remain at the end of the content.
	trackEnd bool

	// The number of characters to be skipped on each line (not in wrap mode).
	columnOffset int

	// The height of the content the last time the text view was drawn.
	pageSize int

	// If set to true, the text view will keep a buffer of text which can be
	// navigated when the text is longer than what fits into the box.
	scrollable bool

	// If set to true, lines that are longer than the available width are wrapped
	// onto the next line. If set to false, any characters beyond the available
	// width are discarded.
	wrap bool

	// The (starting) color of the text.
	textColor tcell.Color

	// If set to true, the text color can be changed dynamically by piping color
	// strings in square brackets to the text view.
	dynamicColors bool

	// An optional function which is called when the content of the text view has
	// changed.
	changed func()

	// An optional function which is called when the user presses one of the
	// following keys: Escape, Enter, Tab, Backtab.
	done func(tcell.Key)
}

// NewTextView returns a new text view.
func NewTextView() *TextView {
	return &TextView{
		Box:           NewBox(),
		lineOffset:    -1,
		scrollable:    true,
		wrap:          true,
		textColor:     tcell.ColorWhite,
		dynamicColors: true,
	}
}

// SetScrollable sets the flag that decides whether or not the text view is
// scollable. If true, text is kept in a buffer and can be navigated.
func (t *TextView) SetScrollable(scrollable bool) *TextView {
	t.scrollable = scrollable
	return t
}

// SetWrap sets the flag that, if true, leads to lines that are longer than the
// available width being wrapped onto the next line. If false, any characters
// beyond the available width are not displayed.
func (t *TextView) SetWrap(wrap bool) *TextView {
	if t.wrap != wrap {
		t.index = nil
	}
	t.wrap = wrap
	return t
}

// SetTextColor sets the initial color of the text (which can be changed
// dynamically by sending color strings in square brackets to the text view if
// dynamic colors are enabled).
func (t *TextView) SetTextColor(color tcell.Color) *TextView {
	t.textColor = color
	return t
}

// SetDynamicColors sets the flag that allows the text color to be changed
// dynamically. See type description for details.
func (t *TextView) SetDynamicColors(dynamic bool) *TextView {
	if t.dynamicColors != dynamic {
		t.index = nil
	}
	t.dynamicColors = dynamic
	return t
}

// SetChangedFunc sets a handler function which is called when the text of the
// text view has changed. This is typically used to cause the application to
// redraw the screen.
func (t *TextView) SetChangedFunc(handler func()) *TextView {
	t.changed = handler
	return t
}

// SetDoneFunc sets a handler which is called when the user presses on the
// following keys: Escape, Enter, Tab, Backtab. The key is passed to the
// handler.
func (t *TextView) SetDoneFunc(handler func(key tcell.Key)) *TextView {
	t.done = handler
	return t
}

// Clear removes all text from the buffer.
func (t *TextView) Clear() *TextView {
	t.buffer = nil
	t.recentBytes = nil
	t.index = nil
	return t
}

// Write lets us implement the io.Writer interface.
func (t *TextView) Write(p []byte) (n int, err error) {
	// Notify at the end.
	if t.changed != nil {
		defer t.changed()
	}

	t.Lock()
	defer t.Unlock()

	// Copy data over.
	newBytes := append(t.recentBytes, p...)
	t.recentBytes = nil

	// If we have a trailing invalid UTF-8 byte, we'll wait.
	if r, _ := utf8.DecodeLastRune(p); r == utf8.RuneError {
		t.recentBytes = newBytes
		return len(p), nil
	}

	// If we have a trailing open dynamic color, exclude it.
	if t.dynamicColors {
		openColor := regexp.MustCompile(`\[[a-z]+$`)
		location := openColor.FindIndex(newBytes)
		if location != nil {
			t.recentBytes = newBytes[location[0]:]
			newBytes = newBytes[:location[0]]
		}
	}

	// Transform the new bytes into strings.
	newLine := regexp.MustCompile(`\r?\n`)
	for index, line := range newLine.Split(string(newBytes), -1) {
		if index == 0 {
			if len(t.buffer) == 0 {
				t.buffer = []string{line}
			} else {
				t.buffer[len(t.buffer)-1] += line
			}
		} else {
			t.buffer = append(t.buffer, line)
		}
	}

	// Reset the index.
	t.index = nil

	return len(p), nil
}

// reindexBuffer re-indexes the buffer such that we can use it to easily draw
// the buffer onto the screen. Each line in the index will contain a pointer
// into the buffer from which on we will print text. It will also contain the
// color with which the line starts.
func (t *TextView) reindexBuffer(width int) {
	if t.index != nil && width == t.indexWidth {
		return // Nothing has changed. We can still use the current index.
	}
	t.index = nil

	color := t.textColor
	if !t.wrap {
		width = math.MaxInt64
	}
	for index, str := range t.buffer {
		// Find all color tags in this line.
		var (
			colorTagIndices [][]int
			colorTags       [][]string
		)
		if t.dynamicColors {
			colorTagIndices = colorPattern.FindAllStringIndex(str, -1)
			colorTags = colorPattern.FindAllStringSubmatch(str, -1)
		}

		// Break down the line.
		var currentTag, currentWidth int
		for pos := range str {
			// Skip any color tags.
			if currentTag < len(colorTags) {
				if pos >= colorTagIndices[currentTag][0] && pos < colorTagIndices[currentTag][1] {
					color = textColors[colorTags[currentTag][1]]
					continue
				} else if pos >= colorTagIndices[currentTag][1] {
					currentTag++
				}
			}

			// Add this line.
			if currentWidth == 0 {
				t.index = append(t.index, &textViewIndex{
					Line:  index,
					Pos:   pos,
					Color: color,
				})
			}

			currentWidth++

			// Have we crossed the width?
			if t.wrap && currentWidth >= width {
				currentWidth = 0
			}
		}
	}

	t.indexWidth = width
}

// Draw draws this primitive onto the screen.
func (t *TextView) Draw(screen tcell.Screen) {
	t.Lock()
	defer t.Unlock()
	t.Box.Draw(screen)

	// Get the available size.
	x, y, width, height := t.GetInnerRect()
	t.pageSize = height

	// Re-index.
	t.reindexBuffer(width)

	// Adjust line offset.
	if t.lineOffset+height > len(t.index) {
		t.trackEnd = true
	}
	if t.trackEnd {
		t.lineOffset = len(t.index) - height
	}
	if t.lineOffset < 0 {
		t.lineOffset = 0
	}

	// Draw the buffer.
	style := tcell.StyleDefault.Background(t.backgroundColor)
	for line := t.lineOffset; line < len(t.index); line++ {
		// Are we done?
		if line-t.lineOffset >= height {
			break
		}

		// Get the text for this line.
		index := t.index[line]
		text := t.buffer[index.Line][index.Pos:]
		style = style.Foreground(index.Color)

		// Get color tags.
		var (
			colorTagIndices [][]int
			colorTags       [][]string
		)
		if t.dynamicColors {
			colorTagIndices = colorPattern.FindAllStringIndex(text, -1)
			colorTags = colorPattern.FindAllStringSubmatch(text, -1)
		}

		// Print one line.
		var currentTag, skip, posX int
		for pos, ch := range text {
			if currentTag < len(colorTags) {
				if pos >= colorTagIndices[currentTag][0] && pos < colorTagIndices[currentTag][1] {
					style = style.Foreground(textColors[colorTags[currentTag][1]])
					continue
				} else if pos >= colorTagIndices[currentTag][1] {
					currentTag++
				}
			}

			// Skip to the right.
			if !t.wrap && skip < t.columnOffset {
				skip++
				continue
			}

			// Stop at the right border.
			if posX >= width {
				break
			}

			screen.SetContent(x+posX, y+line-t.lineOffset, ch, nil, style)

			posX++
		}
	}
}

// InputHandler returns the handler for this primitive.
func (t *TextView) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p Primitive)) {
		switch key := event.Key(); key {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g': // Home.
				t.trackEnd = false
				t.lineOffset = 0
				t.columnOffset = 0
			case 'G': // End.
				t.trackEnd = true
				t.columnOffset = 0
			case 'j': // Down.
				t.lineOffset++
			case 'k': // Up.
				t.trackEnd = false
				t.lineOffset--
			case 'h': // Left.
				t.columnOffset--
				if t.columnOffset < 0 {
					t.columnOffset = 0
				}
			case 'l': // Right.
				t.columnOffset++
			}
		case tcell.KeyHome:
			t.trackEnd = false
			t.lineOffset = 0
			t.columnOffset = 0
		case tcell.KeyEnd:
			t.trackEnd = true
			t.columnOffset = 0
		case tcell.KeyUp:
			t.trackEnd = false
			t.lineOffset--
		case tcell.KeyDown:
			t.lineOffset++
		case tcell.KeyLeft:
			t.columnOffset--
			if t.columnOffset < 0 {
				t.columnOffset = 0
			}
		case tcell.KeyRight:
			t.columnOffset++
		case tcell.KeyPgDn, tcell.KeyCtrlF:
			t.lineOffset += t.pageSize
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			t.trackEnd = false
			t.lineOffset -= t.pageSize
		case tcell.KeyEscape, tcell.KeyEnter, tcell.KeyTab, tcell.KeyBacktab:
			if t.done != nil {
				t.done(key)
			}
		}
	}
}
