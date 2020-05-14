package tview

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell"
)

// Uncomment only for debugging
//
// var debugBuffer bytes.Buffer
//
// func log(args ...interface{}) {
// 	fmt.Fprintln(&debugBuffer, args...)
// }
//
// func init() {
// 	go func() {
// 		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 			fmt.Fprintf(w, "%s", debugBuffer.String())
// 		})
// 		http.ListenAndServe(":9090", nil)
// 	}()
// }

// TextArea is a wrapper which adds space around another primitive. In addition,
// the top area (header) and the bottom area (footer) may also contain text.
//
// See https://github.com/rivo/tview/wiki/TextArea for an example.
type TextArea struct {
	*Box
	view *TextView

	// absolute screen coordinate of cursor
	cursor struct{ x, y int }
	// TODO : add colors in edit area
	// TODO : implement Primitive interface
}

// NewTextArea returns a new textArea around the given primitive. The primitive's
// size will be changed to fit within this textArea.
func NewTextArea() *TextArea {
	f := &TextArea{}
	f.view = NewTextView()
	f.Box = f.view.Box
	// WordWrap is not acceptable, because cursor movement
	// is not valid for that case
	f.view.SetWordWrap(false)
	f.view.SetWrap(true)
	f.view.SetScrollable(true)
	f.view.SetRegions(false)
	f.view.lineOffset = 0
	f.view.trackOff = true
	return f
}

func (f *TextArea) GetBox() *Box {
	return f.Box
}

// AddText adds text to the textArea. Set "header" to true if the text is to appear
// in the header, above the contained primitive. Set it to false for it to
// appear in the footer, below the contained primitive. "align" must be one of
// the Align constants. Rows in the header are printed top to bottom, rows in
// the footer are printed bottom to top. Note that long text can overlap as
// different alignments will be placed on the same row.
func (f *TextArea) SetText(text string) *TextArea {
	f.view.SetText(text)
	return f
}

// GetText returns the current text of this text area.
func (f *TextArea) GetText() string {
	// Get the buffer.
	buffers := f.view.buffer

	// Add newlines again.
	text := strings.Join(buffers, "\n")

	return text
}

// Draw draws this primitive onto the screen.
func (f *TextArea) Draw(screen tcell.Screen) {
	// draw textview
	f.view.Draw(screen)

	// correct position of cursor
	f.cursorPositionCorrection()

	// show cursor
	screen.ShowCursor(f.cursor.x, f.cursor.y)
}

// cursorPositionCorrection modify position of cursor on acceptable
func (f *TextArea) cursorPositionCorrection() {
	x, y, width, height := f.GetInnerRect()
	// cursor is inside acceptable screen limits of TextArea
	borderLimit := func() {
		if f.cursor.x < x {
			f.cursor.x = x
		} else if x+width < f.cursor.x {
			// cursor on right border is acceptable
			f.cursor.x = x + width
		}
		if f.cursor.y < y {
			f.cursor.y = y
		} else if y+height-1 < f.cursor.y {
			f.cursor.y = y + height - 1
		}
		// limitation by offset
		if f.view.lineOffset < 0 {
			f.view.lineOffset = 0
		} else if len(f.view.index) <= f.view.lineOffset {
			f.view.lineOffset = len(f.view.index) - 1
		}
		if f.view.columnOffset < 0 {
			f.view.columnOffset = 0
		}
	}
	borderLimit()
	{
		// cursor is inside of text
		line, pos := f.cursorByScreen()
		f.cursorByBuffer(line, pos)
	}
	borderLimit()
}

// deleteRune remove rune at the left of cursor and return new position of
// cursor in buffer coordinates
func (f *TextArea) deleteRune() (newLine, newPos int) {
	// get position cursor in buffer
	line, pos := f.cursorByScreen()
	if pos == 0 && line == 0 {
		return
	}
	runes := []rune(f.view.buffer[line])
	if 0 < pos && pos < len(f.view.buffer[line])+1 {
		// delete rune
		// prepare split into new lines
		if len(runes)-1 < pos {
			// remove last rune
			runes = runes[:len(runes)-1]
		} else {
			runes = append(runes[:pos-1], runes[pos:]...)
		}
		// change buffer
		f.view.buffer[line] = string(runes)
		// move cursor
		newLine = line
		newPos = pos - 1
	} else if 0 < line {
		// delete newline
		size := len([]rune(f.view.buffer[line-1]))
		f.view.buffer[line-1] = f.view.buffer[line-1] + f.view.buffer[line]
		if line+1 < len(f.view.buffer) {
			f.view.buffer = append(f.view.buffer[:line], f.view.buffer[line+1:]...)
		} else {
			f.view.buffer = f.view.buffer[:line]
		}
		// move cursor
		newPos = size
		newLine = line - 1
	}
	// update a view
	f.updateBuffers()
	return
}

// insertNewLine split buffer by left of cursor position.
func (f *TextArea) insertNewLine() {
	// get position cursor in buffer
	line, pos := f.cursorByScreen()
	if len(f.view.buffer) == 0 {
		f.view.buffer = []string{"\n"}
		return
	} else {
		// prepare split into new lines
		runes := []rune(f.view.buffer[line])
		var runeLineBefore []rune
		if pos < len(runes) {
			runeLineBefore = make([]rune, pos)
			copy(runeLineBefore, runes[:pos])
		} else {
			runeLineBefore = make([]rune, len(runes))
			copy(runeLineBefore, runes)
		}
		var runeLineAfter []rune
		if l := len(runes) - pos; 0 < l {
			runeLineAfter = make([]rune, l)
			copy(runeLineAfter, runes[pos:])
		}
		// change buffer
		f.view.buffer[line] = string(runeLineBefore)
		if line == len(f.view.buffer)-1 {
			f.view.buffer = append(f.view.buffer, string(runeLineAfter))
		} else {
			f.view.buffer = append(
				f.view.buffer[:line+1],
				append([]string{string(runeLineAfter)},
					f.view.buffer[line+1:]...)...)
		}
	}
	// update a view
	f.updateBuffers()
}

// insertRune add rune at teh left position of cursor
func (f *TextArea) insertRune(r rune) {
	// get position cursor in buffer
	line, pos := f.cursorByScreen()
	if len(f.view.buffer) == 0 {
		f.view.buffer = []string{""}
	}
	// prepare new line
	runes := []rune(f.view.buffer[line])
	str := string(r)
	if str == "\t" {
		str = strings.Repeat(" ", TabSize)
	}
	if pos < len(runes) {
		runes = append(runes[:pos], append([]rune(str), runes[pos:]...)...)
	} else {
		runes = append(runes[:pos], []rune(str)...)
	}
	// change buffer
	f.view.buffer[line] = string(runes)
	// update a view
	f.updateBuffers()
}

// updateBuffers is update all buffers of TextView if any changes is happen.
//	TODO: for optimization - need update not all textViewIndex
func (f *TextArea) updateBuffers() {
	_, _, width, _ := f.GetInnerRect()
	text := strings.Join(f.view.buffer, "\n")
	f.view.Clear()
	f.view.lastWidth = -1
	f.view.SetText(text)
	f.view.reindexBuffer(width)
}

// cursorIndexLine return line in TextView.view with cursor
func (f TextArea) cursorIndexLine() int {
	_, y, _, _ := f.GetInnerRect()
	indexLine := f.cursor.y - y + f.view.lineOffset
	if indexLine < 0 {
		indexLine = 0
	}
	if size := len(f.view.index) - 1; size <= indexLine {
		indexLine = size
	}
	return indexLine
}

// cursorByScreen return position cursor in TextView.buffer coordinate.
// unit: rune
func (f TextArea) cursorByScreen() (bufferLine, bufferPosition int) {
	if len(f.view.index) == 0 {
		return
	}
	x, _, _, _ := f.GetInnerRect()
	indexLine := f.cursorIndexLine()
	bufferLine = f.view.index[indexLine].Line
	bytePos := f.view.index[indexLine].Pos

	// convert from screen grapheme to buffer position in rune
	buf := f.view.buffer[bufferLine]
	runePos := len([]rune(buf[:bytePos]))
	buf = buf[bytePos:]

	// find amount of runes in view graphemes
	widthOnScreen := f.view.columnOffset + f.cursor.x - x
	amountRunes := 0
	for ; ; amountRunes++ {
		if len([]rune(buf)) <= amountRunes {
			break
		}
		width := stringWidth(string(([]rune(buf))[:amountRunes]))
		if widthOnScreen == width {
			break
		}
		if widthOnScreen < width {
			amountRunes--
			break
		}
	}
	// position in buffer
	bufferPosition = runePos + amountRunes
	return
}

// cursorByBuffer modify position of cursor in according to position in buffers.
func (f *TextArea) cursorByBuffer(bufferLine, bufferPosition int) {
	lastIndexLine := f.cursorIndexLine()

	buffers := f.view.buffer
	if len(buffers) == 0 || len(f.view.index) == 0 {
		f.cursor.x = 0
		f.cursor.y = 0
		return
	}
	// correction bufferLine
	if bufferLine < 0 {
		bufferLine = 0
	} else if len(buffers)-1 < bufferLine {
		bufferLine = len(buffers) - 1
	}
	// correction bufferPosition
	if bufferPosition < 0 {
		bufferPosition = 0
	}
	// find index
	indexes := f.view.index
	indexLine := -1 // position in slice indexes
	indexPos := -1  // amount rune from indexes[i].Pos
	isOutsideBuffer := false
	for i := len(indexes) - 1; i >= 0; i-- {
		if indexes[i].Line != bufferLine {
			continue
		}
		pos := len([]rune(buffers[bufferLine][:indexes[i].Pos]))
		if pos <= bufferPosition {
			indexLine = i
			if size := len([]rune(buffers[bufferLine])); size < bufferPosition {
				bufferPosition = size
				isOutsideBuffer = true
			}
			indexPos = bufferPosition - pos
			break
		}
	}
	if indexLine < 0 {
		// TODO: find that situation
		indexLine = 0
	}
	// convert position from indexes to grapheme for cursor
	var posInGrapheme int
	{
		buf := buffers[indexes[indexLine].Line]
		partBuf := buf[indexes[indexLine].Pos:]
		part2 := string(([]rune(partBuf))[:indexPos])
		posInGrapheme = stringWidth(part2)
	}

	// store last cursor position
	lastCy := f.cursor.y

	// cursor must be inside screen
	x, y, width, height := f.GetInnerRect()
	_ = width
	f.cursor.x = posInGrapheme + x - f.view.columnOffset
	f.cursor.y = indexLine + y - f.view.lineOffset
	if isOutsideBuffer {
		f.cursor.x++
	}
	if y+height <= f.cursor.y {
		diff := (indexLine - lastIndexLine) - (y + height - lastCy) + 1
		f.view.lineOffset += diff
	}
	if f.cursor.y < y {
		diff := -((lastIndexLine - indexLine) - lastCy) - 1
		f.view.lineOffset += diff
	}
	// TODO: other cases with columnOffset
}

// InputHandler returns the handler for this primitive.
func (f *TextArea) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return f.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		// Moving strategy
		//
		//	* Moving only in buffer coordinate, not on screen coordinate
		//	* Move up/down - moving by buffer lines
		//	* Move left/right - moving by buffer runes
		//
		line, pos := f.cursorByScreen()
		key := event.Key()
		switch key {
		case tcell.KeyUp:
			if line <= 0 {
				// do nothing
			} else {
				line--
			}
		case tcell.KeyDown:
			if len(f.view.buffer)-1 <= line {
				// do nothing
			} else {
				line++
			}
		case tcell.KeyLeft:
			if pos == 0 {
				// do nothing
			} else {
				pos--
			}
		case tcell.KeyRight:
			if len(f.view.buffer) == 0 {
				break
			}
			if stringWidth(f.view.buffer[line]) <= pos {
				// do nothing
			} else {
				pos++
			}
		case tcell.KeyHome:
			pos = 0
		case tcell.KeyEnd:
			if 0 <= line && line < len(f.view.buffer) {
				pos = len([]rune(f.view.buffer[line]))
			}
		case tcell.KeyEnter:
			f.insertNewLine()
			line++
			pos = 0
		case tcell.KeyDelete:
			pos++
			defer func() {
				f.deleteRune()
				pos--
				f.cursorByBuffer(line, pos)
			}()
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			line, pos = f.deleteRune()
		default:
			r := event.Rune()
			if unicode.IsLetter(r) ||
				unicode.IsNumber(r) ||
				unicode.IsMark(r) ||
				unicode.IsSpace(r) {
				f.insertRune(r)
				pos++
			}
		}
		f.cursorByBuffer(line, pos)
	})
}

// MouseHandler returns the mouse handler for this primitive.
func (f *TextArea) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return f.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {

		if action == MouseLeftClick && f.InRect(event.Position()) {
			setFocus(f)
			consumed = true
		}

		x, y := event.Position()
		if !f.InRect(x, y) {
			return false, nil
		}

		switch action {
		case MouseLeftClick:
			f.cursor.x = x
			f.cursor.y = y
			consumed = true
			setFocus(f)
		case MouseScrollUp:
			f.view.lineOffset--
			consumed = true
		case MouseScrollDown:
			f.view.lineOffset++
			consumed = true
		default:
			return
		}

		// correct position of cursor
		f.cursorPositionCorrection()

		return
	})
}
