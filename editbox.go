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
	f.view.SetScrollable( true)
	f.view.SetRegions(false)
	f.view.lineOffset = 0
	f.view.trackOff = true
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
	log("DRAW 1 lineOffset", f.view.lineOffset)
	// draw textview
	f.view.Draw(screen)
	log("DRAW 2 lineOffset", f.view.lineOffset)

	// correct position of cursor
	f.cursorPositionCorrection()
	log("DRAW 3 lineOffset", f.view.lineOffset)

	// show cursor
	screen.ShowCursor(f.cursor.x, f.cursor.y)
	log("DRAW 4 lineOffset", f.view.lineOffset)
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

func (f *EditBox) insertRune(r rune) {
	log("	insertRune")
	log("		lineOffset ", f.view.lineOffset)
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
	log("		lineOffset ", f.view.lineOffset)
	f.updateBuffers()
	log("		lineOffset ", f.view.lineOffset)
	log("	end of insertRune")
}

func (f *EditBox) updateBuffers() {
	_, _, width, _ := f.GetInnerRect()
	// text := f.view.GetText(false)
	text := strings.Join(f.view.buffer, "\n")
	f.view.Clear()
	f.view.lastWidth = -1
	f.view.SetText(text)
	f.view.reindexBuffer(width)
}

func (f EditBox) cursorIndexLine() int {
	log("	cursorIndexLine : ")
	_, y, _, _ := f.GetInnerRect()
	indexLine := f.cursor.y - y + f.view.lineOffset
	log("	cursorIndexLine : ",indexLine)
	if indexLine < 0 {
		indexLine = 0
	}
	if size := len(f.view.index) - 1; size <= indexLine {
		indexLine = size
	}

	log("	cursorIndexLine : ",indexLine)
	log("			cursor : ", f.cursor, y, f.view.lineOffset)
	return indexLine
}

// cursorByScreen return position cursor in TextView.buffer coordinate.
// unit: rune
func (f EditBox) cursorByScreen() (bufferLine, bufferPosition int) {
	log("--------------------------------")
	log("cursorByScreen")
	log("	lineOffset: ",f.view.lineOffset)
	if len(f.view.index) == 0 {
		return
	}
	log("	lineOffset: ",f.view.lineOffset)
	x, _, _, _ := f.GetInnerRect()
	log("	lineOffset: ",f.view.lineOffset)
	indexLine := f.cursorIndexLine()
	log("	lineOffset: ",f.view.lineOffset)
	bufferLine = f.view.index[indexLine].Line
	log("	lineOffset: ",f.view.lineOffset)
	bytePos := f.view.index[indexLine].Pos

	// convert from screen grapheme to buffer position in rune

	buf := f.view.buffer[bufferLine]
	runePos := len([]rune(buf[:bytePos]))
	buf = buf[bytePos:]

	// bufferPosition = runePos + f.view.columnOffset + f.cursor.x - x
	// TODO: check for 2symbol grapheme
	widthOnScreen := f.view.columnOffset + f.cursor.x - x
	amountRunes := 0
	for ; ; amountRunes++ {
		log(" - ", amountRunes)
		if len([]rune(buf)) <= amountRunes {
			log(" break 1 ")
			break
			// bufferPosition = run
		}
		width := stringWidth(string(([]rune(buf))[:amountRunes]))
		log(" -- widths ", widthOnScreen, width)
		if widthOnScreen == width {
			log(" break 3 ", widthOnScreen, width)
			break
		}
		if widthOnScreen < width {
			amountRunes--
			log(" break 2 ", widthOnScreen, width)
			break
		}
	}
	bufferPosition = runePos + amountRunes
	// amountRunes := runeWidth(part, widthOnScreen)
	// if amountRunes-1 > widthOnScreen {
	// 	panic(fmt.Errorf("not valid calculation %v %v on line \"%s\"", amountRunes, widthOnScreen, part))
	// }
	// amountGraphemes := uniseg.GraphemeClusterCount(part)
	// if amountGraphemes == 0 && amountRunes != widthOnScreen {
	// 	panic(fmt.Errorf("cannot calculate amount runes %d on \"%s\"",
	// 		amountRunes, part))
	// }
	log("cursor", f.cursor)
	log("index line", indexLine)
	log("buffer line", bufferLine)
	log("buffer pos ", bufferPosition)
	//log("part ", part)
	log("amountRunes = ", amountRunes)
	log("--")
	log("widthOnScreen ", widthOnScreen)
	log("f.view.columnOffset ", f.view.columnOffset)
	log("x ", x)
	log("--------------------------------")
	log("\n")
	return
}

// TODO: using []rune convertion type is not valid for graphemes

//	* convert from buffer coordinate to line view index coordinate
//	* bufferLine must be inside size of buffers
//	* bufferPosition can be more then len(buffer[bufferLine])
func (f *EditBox) cursorByBuffer(bufferLine, bufferPosition int) {
	log("cursorByBuffer")
	log("cursor ", f.cursor)
	log("buffer line", bufferLine)
	log("buffer pos ", bufferPosition)

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
		// 	} else if size := stringWidth(buffers[bufferLine]); size < bufferPosition {
		// 		bufferPosition = size
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
		log("- pos ", pos, " - ", bufferPosition)
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
	log("indexLine ", indexLine)
	log("indexPos ", indexPos)
	//if indexLine < 0 {
	//	indexLine = len(indexes) - 1
	//}
	//if indexPos < 0 {
	//	indexPos = 0
	//}

	// convert position from indexes to grapheme for cursor
	var posInGrapheme int
	{
		buf := buffers[indexes[indexLine].Line]
		partBuf := buf[indexes[indexLine].Pos:]
		part2 := string(([]rune(partBuf))[:indexPos])
		posInGrapheme = stringWidth(part2)
	}
	log("> posInGrapheme", posInGrapheme)

	// if indexes[indexLine].Width+1 < indexPos {
	//     isOutsideBuffer = true
	// }
	// if indexes[indexLine].Width < indexPos {
	// if inde
	// 	indexPos = indexes[indexLine].Width + 1
	// } else if indexes[indexLine].Width+1 < indexPos && bufferLine < len(buffers)-1 {
	// 	f.cursorByBuffer(bufferLine+1, 0)
	// }
	// store last cursor position
	// lastCx := f.cursor.x
	lastCy := f.cursor.y
	log("lineOffset ", f.view.lineOffset)

	// cursor must be inside screen
	x, y, width, height := f.GetInnerRect()
	_ = width
	f.cursor.x = posInGrapheme + x - f.view.columnOffset
	f.cursor.y = indexLine + y - f.view.lineOffset
	if isOutsideBuffer {
		f.cursor.x++
	}
	// 	if x+width <= f.cursor.x {
	// 		x = 0
	// 	}
	if y+height <= f.cursor.y {
		diff := (indexLine - lastIndexLine) - (y + height - lastCy) + 1
		f.view.lineOffset += diff
		log("offset 1 ", diff)
	}
	if f.cursor.y < y {
		diff := -((lastIndexLine - indexLine) - lastCy) - 1
		f.view.lineOffset += diff
		log("offset 2 ", diff)
	}
	// TODO other cases
	log("cursor", f.cursor)
	log("isOutsideBuffer ", isOutsideBuffer)
	log("lineOffset ", f.view.lineOffset)
	log("\n")
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
		log("BEFORE PUTH THE BUTTON\n")
		line, pos := f.cursorByScreen()
		key := event.Key()
		log("PUTH THE BUTTON\n")
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
		log("PUTH THE BUTTON CORRECT BY BUFFER\n")
		f.cursorByBuffer(line, pos)
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
