package tview

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell"
)

// EditBox is a wrapper which adds space around another primitive. In addition,
// the top area (header) and the bottom area (footer) may also contain text.
//
// See https://github.com/rivo/tview/wiki/EditBox for an example.
type EditBox struct {
	*Box
	view *TextView

	cursor struct{ x, y int }
}

// NewEditBox returns a new editBox around the given primitive. The primitive's
// size will be changed to fit within this editBox.
func NewEditBox() *EditBox {
	f := &EditBox{}
	f.view = NewTextView()
	f.Box = f.view.Box
	f.view.SetWordWrap(true)
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

	f.view.Draw(screen)
	f.cursorLimiting()

	log("\nDRAW")
	log("Drww cursor", f.cursor)
	line, pos, _, _ := f.GetCursor()
	log("->", line, pos)
	f.MoveCursor(line, pos)
	line2, pos2, _, _ := f.GetCursor()
	log("--->", line2, pos2)
	if line != line2 || pos != pos2 {
		log(fmt.Sprintf("cursor: [%d,%d] != [%d,%d] ", line, pos, line2, pos2))
	}

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

func (f *EditBox) cursorLimiting() {
	log("cursorLimiting")
	log("before", f.cursor)
	defer func() {
		log("after", f.cursor)
		log("\n")
	}()
	x, y, width, height := f.GetInnerRect()
	// cursor must be inside editable part of box
	borderLimit := func() {
		if f.cursor.x < x {
			f.cursor.x = x
		} else if f.cursor.x > x+width-1 {
			f.cursor.x = x + width - 1
		}
		if f.cursor.y < y {
			f.cursor.y = y
		} else if f.cursor.y > y+height-1 {
			f.cursor.y = y + height - 1
		}
		// limitation by offset
		if f.view.lineOffset < 0 {
			f.view.lineOffset = 0
		} else if f.view.lineOffset >= len(f.view.index) {
			f.view.lineOffset = len(f.view.index) - 1
		}
		if f.view.columnOffset < 0 {
			f.view.columnOffset = 0
		}
	}
	// cursor cannot be outside text
	borderLimit()
	{
		if diff := (f.cursor.y + f.view.lineOffset - y) - (len(f.view.index) - 1); diff > 0 {
			f.cursor.y -= diff
		}
		presentLine := f.cursor.y + f.view.lineOffset - y
		if presentLine > len(f.view.index) {
			presentLine = len(f.view.index) - 1
		}
		xLimit := f.view.index[presentLine].Width + icel
		if f.cursor.x > xLimit {
			f.cursor.x = xLimit
		}
	}
	borderLimit()
}

const (
	// for insert char at the end of line
	icel int = 1
)

func (f *EditBox) deleteRune(shift bool) {
	// get position cursor in buffer
	line, pos, haveRune, _ := f.GetCursor()
	if pos == 0 && line == 0 {
		return
	}
	log("deleteRune")
	log(line, pos, haveRune)
	runes := []rune(f.view.buffer[line])
	if shift && pos == 0 && len(runes) > 0 {
		// prepare split into new lines
		runes = runes[1:]
		// change buffer
		f.view.buffer[line] = string(runes)
	} else if pos > 0 {
		// delete rune
		// prepare split into new lines
		runes = append(runes[:pos-1], runes[pos:]...)
		// change buffer
		f.view.buffer[line] = string(runes)
	} else if line > 0 {
		// delete newline
		f.view.buffer[line-1] = f.view.buffer[line-1] + f.view.buffer[line]
		if line+1 < len(f.view.buffer) {
			f.view.buffer = append(f.view.buffer[:line], f.view.buffer[line+1:]...)
		} else {
			f.view.buffer = f.view.buffer[:line]
		}
	}
	if !shift {
		// move cursor
		event := tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone)
		f.InputHandler()(event, nil)
	}

	// update a view
	f.updateBuffers()
	f.cursorLimiting()
}

func (f *EditBox) insertNewLine() {
	// get position cursor in buffer
	line, pos, _, _ := f.GetCursor()
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
	if l := len(runes) - pos; l > 0 {
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
	// move cursor
	event := tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone)
	f.InputHandler()(event, nil)
	f.cursorLimiting()
}

func (f *EditBox) insertRune(r rune) {
	// get position cursor in buffer
	line, pos, haveRune, _ := f.GetCursor()
	_ = haveRune
	log("insertRune")
	log(line, pos)
	// prepare new line
	runes := []rune(f.view.buffer[line])
	str := string(r)
	if str == "\t" {
		str = strings.Repeat(" ", TabSize)
	}
	runes = append(runes[:pos], append([]rune(str), runes[pos:]...)...)
	// change buffer
	f.view.buffer[line] = string(runes)
	// update a view
	f.updateBuffers()

	//
	// 	// move cursor
	// 	// x, _, width, _ := f.GetInnerRect()
	// 	// 	f.cursor.x++
	// 	// if f.cursor.x == x+width {
	// 	// 	f.cursor.y++
	// 	// 	f.cursor.x = 0
	// 	// } else {
	// 	// }
	//
	// 	//if haveRune {
	// 	event := tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone)
	// 	f.InputHandler()(event, nil)
	// 	//}
	//
	// 	log("cursor pos :", f.cursor)
	f.MoveCursor(line, pos+len(str))
}

func (f *EditBox) updateBuffers() {
	_, _, width, _ := f.GetInnerRect()
	text := f.view.GetText(false)
	f.view.Clear()
	f.view.lastWidth = -1
	f.view.SetText(text)
	f.view.reindexBuffer(width)
}

func (f *EditBox) GetCursor() (
	bufferLine, bufferPosition int,
	haveRune bool, r rune,
) {
	f.cursorLimiting()

	x, y, _, _ := f.GetInnerRect()
	presentLine := f.cursor.y - y + f.view.lineOffset
	bufferLine = f.view.index[presentLine].Line
	bytePos := f.view.index[presentLine].Pos

	runePos := len([]rune(f.view.buffer[bufferLine][:bytePos]))
	bufferPosition = runePos + f.view.columnOffset + f.cursor.x - x

	runeLine := []rune(f.view.buffer[bufferLine])
	haveRune = bufferPosition < len(runeLine)
	if haveRune {
		r = runeLine[bufferPosition]
	}
	log("GetCursor", bufferLine, bufferPosition, haveRune, "|||", bytePos, runePos, f.view.columnOffset, f.cursor.x, x)
	log("GetCursor 2 ^^ ", len([]rune(f.view.buffer[bufferLine])))
	return
}

func (f *EditBox) MoveCursor(bufferLine, bufferPosition int) {
	log("MoveCursor === ", bufferLine, bufferPosition)
	x, y, _, _ := f.GetInnerRect()
	for i := range f.view.index {
		if i+1 == len(f.view.index)-1 {
			break
		}
		if f.view.index[i].Line != bufferLine {
			continue
		}
		pos := len([]rune(f.view.buffer[bufferLine][:f.view.index[i].Pos]))
		log(i, bufferLine, bufferPosition, "=======", f.view.index[i].Line, pos, pos+f.view.index[i].Width)
		if pos > bufferPosition {
			continue
		}
		if bufferPosition > pos+f.view.index[i].Width+icel {
			continue
		}
		f.cursor.y = i - f.view.lineOffset + y
		f.cursor.x = bufferPosition - pos + x - f.view.columnOffset
		log("--------------------------------- X ", bufferPosition, f.view.index[i].Pos, f.view.index[i].Width, icel, f.view.columnOffset)
		log("--------------------------------- Y ", i, f.view.lineOffset, y)
		log("cursor pos :", f.cursor)
		//  break
	}
	log("\n")
}

// InputHandler returns the handler for this primitive.
func (f *EditBox) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return f.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		f.cursorLimiting()
		x, y, width, height := f.GetInnerRect()
		_ = width
		_ = height
		presentLine := f.cursor.y + f.view.lineOffset - y
		if presentLine < 0 || presentLine >= len(f.view.index) {
			panic(fmt.Errorf("presentLine is %d. cursor.y=%d. lineOffset=%d. y=%d", presentLine, f.cursor.y, f.view.lineOffset, y))
		}
		key := event.Key()
		log("InpitHa", key)
		switch key {
		case tcell.KeyUp:
			if f.cursor.y == y {
				f.view.lineOffset--
			} else {
				f.cursor.y--
			}
		case tcell.KeyDown:
			if f.cursor.y == y+height-1 {
				f.view.lineOffset++
			} else {
				f.cursor.y++
			}
		case tcell.KeyLeft:
			line, pos, _, _ := f.GetCursor()
			if line == 0 && f.cursor.x == x {
				// do nothing
			} else if line > 0 && f.cursor.x == x { // on left border
				if f.cursor.y == y { // on up-left corner
					f.view.lineOffset--
				} else {
					f.cursor.y--
					f.cursor.x = x + width + 1
				}
			} else {
				f.MoveCursor(line, pos-1)
			}
		case tcell.KeyRight:
			line, pos, _, _ := f.GetCursor()
			if len(f.view.index) > 0 {
				if line == len(f.view.buffer) && f.cursor.x-x >= f.view.index[len(f.view.index)-1].Width+icel {
					// do nothing
				} else if line < len(f.view.buffer) && f.cursor.x-x >= f.view.index[presentLine].Width-1 { // on right border
					log("\n\n00000 :", f.cursor, y, height, "\n\n")
					if f.cursor.y == y+height-1 {
						log("111111111111")
						f.view.lineOffset++
					} else {
						log("2222222222222")
						f.cursor.y++
					}
					f.cursor.x = 0
				} else {
					log("333333333333 ::: ", line, len(f.view.buffer), ":::", f.cursor.x, x, f.view.index[presentLine].Width)
					f.MoveCursor(line, pos+1)
				}
			}

		case tcell.KeyHome:
			f.view.lineOffset = 0
			f.view.columnOffset = 0
		case tcell.KeyPgDn, tcell.KeyCtrlF:
			f.view.lineOffset += height
		case tcell.KeyPgUp, tcell.KeyCtrlB:
			f.view.lineOffset -= height
		case tcell.KeyEnd:
			f.cursor.x = x + width
		case tcell.KeyEnter:
			f.insertNewLine()
		case tcell.KeyDelete:
			f.deleteRune(true)
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			f.deleteRune(false)
		default:
			r := event.Rune()
			if unicode.IsLetter(r) || r == ' ' {
				f.insertRune(r)
			}
		}
		f.cursorLimiting()
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

		f.view.trackEnd = false
		f.cursorLimiting()
		return
	})
}
