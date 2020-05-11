package tview

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/gdamore/tcell"
)

var debugBuffer bytes.Buffer

func log(args ...interface{}) {
	fmt.Fprintln(&debugBuffer, args...)
}

func init() {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%s", debugBuffer.String())
		})
		http.ListenAndServe(":9090", nil)
	}()
}

// EditBox is a wrapper which adds space around another primitive. In addition,
// the top area (header) and the bottom area (footer) may also contain text.
//
// See https://github.com/rivo/tview/wiki/EditBox for an example.
type EditBox struct {
	*Box
	view *TextView

	cursor struct {
		// absolute screen coordinate of cursor
		x, y int
	}
}

// NewEditBox returns a new editBox around the given primitive. The primitive's
// size will be changed to fit within this editBox.
func NewEditBox() *EditBox {
	f := &EditBox{}
	f.view = NewTextView()
	f.Box = f.view.Box
	// WordWrap is not acceptable, because cursor movement is valid for that case
	f.view.SetWordWrap(false)
	f.view.SetWrap(true)
	f.view.SetScrollable(true)
	return f
}

func (f *EditBox) GetBox() *Box {
	return f.Box
}

// AddText adds text to the editBox. Set "header" to true if the text is to appear
// in the header, above the contained primitive. Set it to false for it to
// appear in the footer, below the contained primitive. "align" must be one of
// the Align constants. Rows in the header are printed top to bottom, rows in
// the footer are printed bottom to top. Note that long text can overlap as
// different alignments will be placed on the same row.
func (f *EditBox) SetText(text string) *EditBox {
	f.view.SetText(text)
	return f
}

// TODO: GetText

// Draw draws this primitive onto the screen.
func (f *EditBox) Draw(screen tcell.Screen) {
	// draw textview
	f.view.Draw(screen)

	// correct position of cursor
	f.cursorPositionCorrection()

	// show cursor
	screen.ShowCursor(f.cursor.x, f.cursor.y)
}

// // Focus is called when this primitive receives focus.
// func (f *EditBox) Focus(delegate func(p Primitive)) {
// 	delegate(f.primitive)
// }
//
// // HasFocus returns whether or not this primitive has focus.
// func (f *EditBox) HasFocus() bool {
// 	focusable, ok := f.primitive.(Focusable)
// 	if ok {
// 		return focusable.HasFocus()
// 	}
// 	return false
// }

func (f *EditBox) cursorPositionCorrection() {
	x, y, width, height := f.GetInnerRect()
	// cursor is inside acceptable screen limits of EditBox
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

const (
	// for insert char at the end of line
	icel int = 1
)

func (f *EditBox) deleteRune() {
	// get position cursor in buffer
	line, pos := f.cursorByScreen()
	if pos == 0 && line == 0 {
		return
	}
	runes := []rune(f.view.buffer[line])
	if 0 < pos && pos < len(f.view.buffer[line])+1 {
		// delete rune
		// prepare split into new lines
		if len(f.view.buffer)-1 < pos {
			runes = runes[:pos-1]
		} else {
			runes = append(runes[:pos-1], runes[pos:]...)
		}
		// change buffer
		f.view.buffer[line] = string(runes)
	} else if 0 < line {
		// delete newline
		f.view.buffer[line-1] = f.view.buffer[line-1] + f.view.buffer[line]
		if line+1 < len(f.view.buffer) {
			f.view.buffer = append(f.view.buffer[:line], f.view.buffer[line+1:]...)
		} else {
			f.view.buffer = f.view.buffer[:line]
		}
	}
	// update a view
	f.updateBuffers()
	//f.cursorLimiting()
}

func (f *EditBox) insertNewLine() {
	// get position cursor in buffer
	line, pos := f.cursorByScreen()
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
	// update a view
	f.updateBuffers()
}

func (f *EditBox) insertRune(r rune) {
	// get position cursor in buffer
	line, pos := f.cursorByScreen()
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

func (f *EditBox) updateBuffers() {
	_, _, width, _ := f.GetInnerRect()
	text := f.view.GetText(false)
	f.view.Clear()
	f.view.lastWidth = -1
	f.view.SetText(text)
	f.view.reindexBuffer(width)
}

func (f EditBox) cursorIndexLine() int {
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
// TODO : unit: grapheme ??? rune ??
func (f EditBox) cursorByScreen() (bufferLine, bufferPosition int) {
	x, _, _, _ := f.GetInnerRect()
	indexLine := f.cursorIndexLine()
	bufferLine = f.view.index[indexLine].Line
	bytePos := f.view.index[indexLine].Pos
	runePos := stringWidth(f.view.buffer[bufferLine][:bytePos])

	bufferPosition = runePos + f.view.columnOffset + f.cursor.x - x
	// TODO: check
	// widthOnScreen := f.view.columnOffset + f.cursor.x - x
	// amountRunes := runeWidth(f.view.buffer[bufferLine][bytePos:], widthOnScreen)
	// bufferPosition = runePos + amountRunes
	return
}

// TODO: using []rune convertion type is not valid for graphemes

//	* convert from buffer coordinate to line view index coordinate
//	* bufferLine must be inside size of buffers
//	* bufferPosition can be more then len(buffer[bufferLine])
func (f *EditBox) cursorByBuffer(bufferLine, bufferPosition int) {

	lastIndexLine := f.cursorIndexLine()

	defer log("\n\n")
	log("cursorByBuffer : ", bufferLine, bufferPosition)
	buffers := f.view.buffer
	if len(buffers) == 0 {
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
	} else if size := stringWidth(buffers[bufferLine]); size < bufferPosition {
		bufferPosition = size
	}
	// find index
	indexes := f.view.index
	indexLine := -1
	indexPos := -1
	for i := len(indexes) - 1; i >= 0; i-- {
		if indexes[i].Line != bufferLine {
			continue
		}
		pos := stringWidth(buffers[bufferLine][:indexes[i].Pos])
		// log("---- ", i, indexes[i].Line == bufferLine, pos, fmt.Sprintf("%#v", indexes[i]))
		log("= ", i, pos, bufferPosition)
		if pos <= bufferPosition {
			log("FOUND")
			indexLine = i
			indexPos = bufferPosition - pos // TODO: need to use stringWidth
			break
		}
	}
	log("cursorByBuffer ", indexLine, indexPos)
	if indexLine < 0 {
		indexLine = len(indexes) - 1
	}
	if indexPos < 0 {
		indexPos = 0
	}
	if indexPos == indexes[indexLine].Width+1 {
		indexPos = indexes[indexLine].Width + 1
	} else if indexes[indexLine].Width+1 < indexPos && bufferLine < len(buffers)-1 {
		f.cursorByBuffer(bufferLine+1, 0)
	}
	// store last cursor position
	// lastCx := f.cursor.x
	lastCy := f.cursor.y

	// cursor must be inside screen
	x, y, width, height := f.GetInnerRect()
	log("cursorByBuffer 1 : --- : ", f.cursor.x, indexPos, x, f.view.columnOffset)
	f.cursor.x = indexPos + x - f.view.columnOffset
	log("cursorByBuffer 2 : --- : ", f.cursor.y, indexLine, y, f.view.lineOffset)
	f.cursor.y = indexLine + y - f.view.lineOffset
	log("cursorByBuffer 3 : --- : ", f.cursor)
	log("cursor position ", f.cursor)
	log("limits ", x, y, width, height)
	// 	if x+width <= f.cursor.x {
	// 		x = 0
	// 		log("x = 0")
	// 	}
	if y+height <= f.cursor.y {
		log("*1*", f.cursor.y, y, f.view.lineOffset, indexLine)
		diff := (indexLine - lastIndexLine) - (y + height - lastCy) + 1
		f.view.lineOffset += diff
		log("y diff = ", diff)
	}
	if f.cursor.y < y {
		log("*2*", f.cursor.y, y, f.view.lineOffset, indexLine)
		diff := -((lastIndexLine - indexLine) - lastCy) - 1
		f.view.lineOffset += diff
		log("y diff = ", diff)
	}
	log("cursor position after correction", f.cursor)
	// TODO other cases
}

// InputHandler returns the handler for this primitive.
func (f *EditBox) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return f.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		//	f.cursorLimiting()

		// Moving strategy
		//
		//	* Moving only in buffer coordinate, not on screen coordinate
		//	* Move up/down - moving by buffer lines
		//	* Move left/right - moving by buffer runes
		//
		line, pos := f.cursorByScreen()
		log("IN inpit", line, pos)
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
			if stringWidth(f.view.buffer[line]) <= pos {
				// do nothing
			} else {
				pos++
			}
		case tcell.KeyHome:
			pos = 0
		case tcell.KeyEnd:
			pos = stringWidth(f.view.buffer[line])
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
			f.deleteRune()
			pos--
		default:
			r := event.Rune()
			if unicode.IsLetter(r) || r == ' ' {
				f.insertRune(r)
				pos++
			}
		}
		log("PRE inpt", line, pos)
		f.cursorByBuffer(line, pos)
		log("OUT inpt", line, pos)
	})
}

// MouseHandler returns the mouse handler for this primitive.
func (f *EditBox) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
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

		// this is a workarount of TextView
		f.view.trackEnd = false

		// correct position of cursor
		f.cursorPositionCorrection()

		return
	})
}
