package tview

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

const (
	// The minimum capacity of the text area's piece chain slice.
	pieceChainMinCap = 10

	// The minimum capacity of the text area's edit buffer.
	editBufferMinCap = 200

	// The maximum number of bytes making up a grapheme cluster. In theory, this
	// could be longer but it would be highly unusual.
	maxGraphemeClusterSize = 40

	// The minimum width of text (if available) to be shown left of the cursor.
	minCursorPrefix = 5

	// The minimum width of text (if available) to be shown right of the cursor.
	minCursorSuffix = 3
)

// Types of user actions on a text area.
type taAction int

const (
	taActionOther        taAction = iota
	taActionTypeSpace             // Typing a space character.
	taActionTypeNonSpace          // Typing a non-space character.
	taActionBackspace             // Deleting the previous character.
	taActionDelete                // Deleting the next character.
)

// NewLine is the string sequence to be inserted when hitting the Enter key in a
// TextArea. The default is "\n" but you may change it to "\r\n" if required.
var NewLine = "\n"

// textAreaSpan represents a range of text in a text area. The text area widget
// roughly follows the concept of Piece Chains outline in
// http://www.catch22.net/tuts/neatpad/piece-chains with some modifications.
// This type represents a "span" (or "piece") and thus refers to a subset of the
// text in the editor as part of a doubly-linked list.
//
// In most places where we reference a position in the text, we use a
// three-element int array. The first element is the index of the referenced
// span in the piece chain. The second element is the offset into the span's
// referenced text (relative to the span's start), its value is always >= 0 and
// < span.length. The third elements is the corresponding text parser's state.
//
// A range of text is represented by a span range which is a starting position
// (int array) and an ending position (int array). The starting position
// references the first character of the range, the ending position references
// the position after the last character of the range. The end of the text is
// therefore always [3]int{1, 0, 0}, position 0 of the ending sentinel.
type textAreaSpan struct {
	// Links to the previous and next textAreaSpan objects as indices into the
	// TextArea.spans slice. The sentinel spans (index 0 and 1) have -1 as their
	// previous or next links.
	previous, next int

	// The start index and the length of the text segment this span represents.
	// If "length" is negative, the span represents a substring of
	// TextArea.initialText and the actual length must be its absolute value. If
	// it is positive, the span represents a substring of TextArea.editText. For
	// the sentinel spans (index 0 and 1), both values will be 0. Others will
	// never have a zero length.
	offset, length int
}

// TextArea implements a simple text editor for multi-line text. Multi-color
// text is not supported. Word-wrapping is enabled by default but can be turned
// off or be changed to character-wrapping.
//
// # Navigation and Editing
//
// A text area is always in editing mode and no other mode exists. The following
// keys can be used to move the cursor:
//
//   - Left arrow: Move left.
//   - Right arrow: Move right.
//   - Down arrow: Move down.
//   - Up arrow: Move up.
//   - Ctrl-A, Home: Move to the beginning of the current line.
//   - Ctrl-E, End: Move to the end of the current line.
//   - Ctrl-F, page down: Move down by one page.
//   - Ctrl-B, page up: Move up by one page.
//   - Alt-Up arrow: Scroll the page up, leaving the cursor in its position.
//   - Alt-Down arrow: Scroll the page down, leaving the cursor in its position.
//   - Alt-Left arrow: Scroll the page to the left, leaving the cursor in its
//     position. Ignored if wrapping is enabled.
//   - Alt-Right arrow:  Scroll the page to the right, leaving the cursor in its
//     position. Ignored if wrapping is enabled.
//   - Alt-B: Jump to the beginning of the current or previous word.
//   - Alt-F: Jump to the end of the current or next word.
//
// Words are defined according to Unicode Standard Annex #29. We skip any words
// that contain only spaces or punctuation.
//
// Entering a character will insert it at the current cursor location.
// Subsequent characters are moved accordingly. If the cursor is outside the
// visible area, any changes to the text will move it into the visible area. The
// following keys can also be used to modify the text:
//
//   - Enter: Insert a newline character (see [NewLine]).
//   - Tab: Insert [TabSize] spaces.
//   - Ctrl-H, Backspace: Delete one character to the left of the cursor.
//   - Ctrl-D, Delete: Delete the character under the cursor (or the first
//     character on the next line if the cursor is at the end of a line).
//   - Ctrl-K: Delete everything under and to the right of the cursor until the
//     next newline character.
//   - Ctrl-W: Delete from the start of the current word to the left of the
//     cursor.
//   - Ctrl-U: Delete the current line, i.e. everything after the last newline
//     character before the cursor up until the next newline character. This may
//     span multiple lines if wrapping is enabled.
//
// Text can be selected by moving the cursor while holding the Shift key. Thus
// when text is selected:
//
//   - Entering a character will replace the selected text with the new
//     character.
//   - Backspace, delete, Ctrl-H, Ctrl-D: Delete the selected text.
//   - Ctrl-Q: Copy the selected text into the clipboard, unselect the text.
//   - Ctrl-X: Copy the selected text into the clipboard and delete it.
//   - Ctrl-V: Replace the selected text with the clipboard text. If no text is
//     selected, the clipboard text will be inserted at the cursor location.
//
// The Ctrl-Q key was chosen for the "copy" function because the Ctrl-C key is
// the default key to stop the application. If your application frees up the
// global Ctrl-C key and you want to bind it to the "copy to clipboard"
// function, you may use [Box.SetInputCapture] to override the Ctrl-Q key to
// implement copying to the clipboard. Note that using your terminal's /
// operating system's key bindings for copy+paste functionality may not have the
// expected effect as tview will not be able to handle these keys.
//
// The default clipboard is an internal text buffer, i.e. the operating system's
// clipboard is not used. If you want to implement your own clipboard (or make
// use of your operating system's clipboard), you can use
// [TextArea.SetClipboard] which  provides all the functionality needed to
// implement your own clipboard.
//
// The text area also supports Undo:
//
//   - Ctrl-Z: Undo the last change.
//   - Ctrl-Y: Redo the last Undo change.
//
// If the mouse is enabled, clicking on a screen cell will move the cursor to
// that location or to the end of the line if past the last character. Turning
// the scroll wheel will scroll the text. Text can also be selected by moving
// the mouse while pressing the left mouse button (see below for details). The
// word underneath the mouse cursor can be selected by double-clicking.
type TextArea struct {
	*Box

	// The text to be shown in the text area when it is empty.
	placeholder string

	// Styles:

	// The style of the text. Background colors different from the Box's
	// background color may lead to unwanted artefacts.
	textStyle tcell.Style

	// The style of the selected text.
	selectedStyle tcell.Style

	// The style of the placeholder text.
	placeholderStyle tcell.Style

	// Text manipulation related fields:

	// The text area's text prior to any editing. It is referenced by spans with
	// a negative length.
	initialText string

	// Any text that's been added by the user at some point. We only ever append
	// to this buffer. It is referenced by spans with a positive length.
	editText strings.Builder

	// The total length of all text in the text area.
	length int

	// The maximum number of bytes allowed in the text area. If 0, there is no
	// limit.
	maxLength int

	// The piece chain. The first two spans are sentinel spans which don't
	// reference anything and always remain in the same place. Spans are never
	// deleted.
	spans []textAreaSpan

	// Display, navigation, and cursor related fields:

	// If set to true, lines that are longer than the available width are
	// wrapped onto the next line. If set to false, any characters beyond the
	// available width are discarded.
	wrap bool

	// If set to true and if wrap is also true, lines are split at spaces or
	// after punctuation characters.
	wordWrap bool

	// The index of the first line shown in the text area.
	rowOffset int

	// The number of cells to be skipped on each line (not used in wrap mode).
	columnOffset int

	// The inner height and width of the text area the last time it was drawn.
	lastHeight, lastWidth int

	// The width of the currently known widest line, as determined by
	// [extendLines].
	widestLine int

	// Text positions and states of the start of lines. Each element is a span
	// position (see textAreaSpan) and a state as returned by uniseg.Step(). Not
	// all lines of the text may be contained at any time, extend as needed with
	// the TextArea.extendLines() function.
	lineStarts [][3]int

	// The cursor always points to the next position where a new character would
	// be placed. The selection start is the same as cursor as long as there is
	// no selection. When there is one, the selection is between selectionStart
	// and cursor.
	cursor, selectionStart struct {
		// The row and column in screen space but relative to the start of the
		// text which may be outside the text area's box. The column value may
		// be larger than where the cursor actually is if the line the cursor
		// is on is shorter. The actualColumn is the position as it is seen on
		// screen. These three values may not be determined yet, in which case
		// the row is negative.
		row, column, actualColumn int

		// The textAreaSpan position with state for the actual next character.
		pos [3]int
	}

	// Clipboard related fields:

	// The internal clipboard.
	clipboard string

	// The function to call when the user copies/cuts a text selection to the
	// clipboard.
	copyToClipboard func(string)

	// The function to call when the user pastes text from the clipboard.
	pasteFromClipboard func() string

	// Undo/redo related fields:

	// The last action performed by the user.
	lastAction taAction

	// The undo stack's items are the first of two consecutive indices into the
	// spans slice. The first referenced span is a copy of the one before the
	// modified span range, the second referenced span is a copy of the one
	// after the modified span range. To undo an action, these two spans are put
	// back into their original place. Undos and redos decrease or increase the
	// nextUndo value. Thus, the next undo action is not always the last item.
	undoStack []int

	// The current undo/redo position on the undo stack. If no undo or redo has
	// been performed yet, this is the same as len(undoStack).
	nextUndo int
}

// NewTextArea returns a new text area. Use [TextArea.SetText] to set the
// initial text.
func NewTextArea() *TextArea {
	t := &TextArea{
		Box:              NewBox(),
		wrap:             true,
		wordWrap:         true,
		placeholderStyle: tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.TertiaryTextColor),
		textStyle:        tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.PrimaryTextColor),
		selectedStyle:    tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.PrimitiveBackgroundColor),
		spans:            make([]textAreaSpan, 2, pieceChainMinCap), // We reserve some space to avoid reallocations right when editing starts.
		lastAction:       taActionOther,
	}
	t.editText.Grow(editBufferMinCap)
	t.spans[0] = textAreaSpan{previous: -1, next: 1}
	t.spans[1] = textAreaSpan{previous: 0, next: -1}
	t.cursor.pos = [3]int{1, 0, -1}
	t.selectionStart = t.cursor
	t.SetClipboard(nil, nil)

	return t
}

// SetText sets the text of the text area. All text is deleted and replaced with
// the new text. Any edits are discarded, no undos are available. This function
// is typically only used to initialize the text area with a text after it has
// been created. To clear the text area's text (again, no undos), provide an
// empty string.
//
// If cursorAtTheEnd is false, the cursor is placed at the start of the text. If
// it is true, it is placed at the end of the text. For very long texts, placing
// the cursor at the end can be an expensive operation because the entire text
// needs to be parsed and laid out.
func (t *TextArea) SetText(text string, cursorAtTheEnd bool) *TextArea {
	t.spans = t.spans[:2]
	t.initialText = text
	t.editText.Reset()
	t.lineStarts = nil
	t.length = len(text)
	t.rowOffset = 0
	t.columnOffset = 0
	t.resetLines()
	t.cursor.column = 0
	t.cursor.pos = [3]int{1, 0, -1}
	//TODO: Reset undo.

	if len(text) > 0 {
		t.spans = append(t.spans, textAreaSpan{
			previous: 0,
			next:     1,
			offset:   0,
			length:   -len(text),
		})
		t.spans[0].next = 2
		t.spans[1].previous = 2
		if !cursorAtTheEnd {
			t.cursor.row = 0
		}
	} else {
		t.spans[0].next = 1
		t.spans[1].previous = 0
		t.cursor.row, t.cursor.actualColumn = 0, 0
	}
	t.selectionStart = t.cursor

	return t
}

// GetText returns the entire text of the text area. Note that this will newly
// allocate the entire text.
func (t *TextArea) GetText() string {
	if t.length == 0 {
		return ""
	}

	var text strings.Builder
	text.Grow(t.length)
	spanIndex := t.spans[0].next
	for spanIndex != 1 {
		span := &t.spans[spanIndex]
		if span.length < 0 {
			text.WriteString(t.initialText[span.offset : span.offset-span.length])
		} else {
			text.WriteString(t.editText.String()[span.offset : span.offset+span.length])
		}
		spanIndex = t.spans[spanIndex].next
	}

	return text.String()

}

// SetWrap sets the flag that, if true, leads to lines that are longer than the
// available width being wrapped onto the next line. If false, any characters
// beyond the available width are not displayed.
func (t *TextArea) SetWrap(wrap bool) *TextArea {
	if t.wrap != wrap {
		t.wrap = wrap
		t.resetLines()
	}
	return t
}

// SetWordWrap sets the flag that, if true and if the "wrap" flag is also true
// (see SetWrap()), wraps the line at spaces or after punctuation marks.
//
// This flag is ignored if the "wrap" flag is false.
func (t *TextArea) SetWordWrap(wrapOnWords bool) *TextArea {
	if t.wordWrap != wrapOnWords {
		t.wordWrap = wrapOnWords
		t.resetLines()
	}
	return t
}

// SetPlaceholder sets the text to be displayed when the text area is empty.
func (t *TextArea) SetPlaceholder(placeholder string) *TextArea {
	t.placeholder = placeholder
	return t
}

// SetMaxLength sets the maximum number of bytes allowed in the text area. If 0,
// there is no limit. If the text area currently contains more bytes than this,
// it may violate this constraint.
func (t *TextArea) SetMaxLength(maxLength int) *TextArea {
	t.maxLength = maxLength
	return t
}

// SetTextStyle sets the style of the text. Background colors different from the
// Box's background color may lead to unwanted artefacts.
func (t *TextArea) SetTextStyle(style tcell.Style) *TextArea {
	t.textStyle = style
	return t
}

// SetSelectedStyle sets the style of the selected text.
func (t *TextArea) SetSelectedStyle(style tcell.Style) *TextArea {
	t.selectedStyle = style
	return t
}

// SetPlaceholderStyle sets the style of the placeholder text.
func (t *TextArea) SetPlaceholderStyle(style tcell.Style) *TextArea {
	t.placeholderStyle = style
	return t
}

// GetOffset returns the text's offset, that is, the number of rows and columns
// skipped during drawing at the top or on the left, respectively. Note that the
// column offset is ignored if wrapping is enabled.
func (t *TextArea) GetOffset() (row, column int) {
	return t.rowOffset, t.columnOffset
}

// SetOffset sets the text's offset, that is, the number of rows and columns
// skipped during drawing at the top or on the left, respectively. If wrapping
// is enabled, the column offset is ignored.
func (t *TextArea) SetOffset(row, column int) *TextArea {
	t.rowOffset, t.columnOffset = row, column
	return t
}

// SetClipboard allows you to implement your own clipboard by providing a
// function that is called when the user wishes to store text in the clipboard
// (copyToClipboard) and a function that is called when the user wishes to
// retrieve text from the clipboard (pasteFromClipboard).
//
// Providing nil values will cause the default clipboard implementation to be
// used.
func (t *TextArea) SetClipboard(copyToClipboard func(string), pasteFromClipboard func() string) *TextArea {
	t.copyToClipboard = copyToClipboard
	if t.copyToClipboard == nil {
		t.copyToClipboard = func(text string) {
			t.clipboard = text
		}
	}

	t.pasteFromClipboard = pasteFromClipboard
	if t.pasteFromClipboard == nil {
		t.pasteFromClipboard = func() string {
			return t.clipboard
		}
	}

	return t
}

// replace deletes a range of text and inserts the given text at that position.
// If the resulting text would exceed the maximum length, the function does not
// do anything. The function returns the end position of the deleted/inserted
// range.
//
// The function can hang if "deleteStart" is located after "deleteEnd".
//
// Undo events are always generated unless continuation is true and text is
// either appended to the end of a span or a span is shortened at the beginning
// or the end (and nothing else).
//
// This function does not modify [TextArea.lineStarts].
func (t *TextArea) replace(deleteStart, deleteEnd [3]int, insert string, continuation bool) [3]int {
	// Maybe nothing needs to be done?
	if deleteStart == deleteEnd && insert == "" || t.maxLength > 0 && t.length+len(insert) >= t.maxLength {
		return deleteEnd
	}

	// Handle the few cases where we don't put anything onto the undo stack.
	if continuation {
		// Same action as the one before. An undo item was already generated for
		// this block of (same) actions.
		if insert == "" {
			if deleteStart[1] == 0 && deleteEnd[1] == 0 {
				// We're deleting an entire span (only one because we delete at
				// most one grapheme). Where does it connect to the last undo
				// item?
				if t.spans[t.undoStack[t.nextUndo-1]].next == deleteStart[0] {
					// Backspace. Extend the "before" span of the undo stack's
					// last item.
					previous := t.spans[deleteStart[0]].previous
					if previous != 0 {
						undoPrevious := t.undoStack[t.nextUndo-1]
						spansLength := len(t.spans) // We make two new entries for the modified undo item.
						t.spans = append(t.spans, t.spans[previous])
						t.spans = append(t.spans, t.spans[undoPrevious+1])
						t.spans[undoPrevious].previous = spansLength
						t.spans[spansLength].next = undoPrevious
						t.undoStack[t.nextUndo-1] = spansLength
					}
					t.spans[previous].next = deleteEnd[0]
					t.spans[deleteEnd[0]].previous = previous
					length := t.spans[deleteStart[0]].length
					if length < 0 {
						t.length += length
					} else {
						t.length -= length
					}
					return deleteEnd
				}
				// Delete. Extend the "after" span of the undo stack's last
				// item.
				if deleteEnd[0] != 1 {
					undoNext := t.undoStack[t.nextUndo-1] + 1
					spansLength := len(t.spans) // We make two new entries for the modified undo item.
					t.spans = append(t.spans, t.spans[undoNext-1])
					t.spans = append(t.spans, t.spans[undoNext])
					t.spans[undoNext].next = spansLength + 1
					t.spans[spansLength+1].previous = undoNext
					t.undoStack[t.nextUndo-1] = spansLength
				}
				previous := t.spans[deleteStart[0]].previous
				t.spans[deleteEnd[0]].previous = previous
				t.spans[previous].next = deleteEnd[0]
				length := t.spans[deleteStart[0]].length
				if length < 0 {
					t.length += length
				} else {
					t.length -= length
				}
				return deleteEnd
			} else if deleteEnd[1] == 0 {
				// Simple backspace. Just shorten this span.
				length := t.spans[deleteStart[0]].length
				if length < 0 {
					length = -deleteStart[1]
				}
				t.spans[deleteStart[0]].length = length
				t.length -= deleteStart[1]
				return deleteEnd
			} else if deleteStart[1] == 0 {
				// Simple delete. Just clip the beginning of this span.
				t.spans[deleteEnd[0]].offset += deleteEnd[1]
				if t.spans[deleteEnd[0]].length < 0 {
					t.spans[deleteEnd[0]].length += deleteEnd[1]
				} else {
					t.spans[deleteEnd[0]].length -= deleteEnd[1]
				}
				t.length -= deleteEnd[1]
				return deleteEnd
			}
		} else if insert != "" && deleteStart == deleteEnd && deleteEnd[1] == 0 {
			previous := t.spans[deleteStart[0]].previous
			bufferSpan := t.spans[previous]
			if bufferSpan.length > 0 && bufferSpan.offset+bufferSpan.length == t.editText.Len() {
				// Typing individual characters. Simply extend the edit buffer.
				length, _ := t.editText.WriteString(insert)
				t.spans[previous].length += length
				t.length += length
				return deleteEnd
			}
		}
	}

	// All other cases generate an undo item.
	before := t.spans[deleteStart[0]].previous
	after := deleteEnd[0]
	if deleteEnd[1] > 0 {
		after = t.spans[deleteEnd[0]].next
	}
	t.undoStack = t.undoStack[:t.nextUndo]
	t.undoStack = append(t.undoStack, len(t.spans))
	t.spans = append(t.spans, t.spans[before])
	t.spans = append(t.spans, t.spans[after])
	t.nextUndo++

	// Adjust total text length by subtracting everything between "before" and
	// "after". Inserted spans will be added back.
	for index := deleteStart[0]; index != after; index = t.spans[index].next {
		if t.spans[index].length < 0 {
			t.length += t.spans[index].length
		} else {
			t.length -= t.spans[index].length
		}
	}
	t.spans[before].next = after
	t.spans[after].previous = before

	// We go from left to right, connecting new spans as needed. We update
	// "before" as the span to connect new spans to.

	// If we start deleting in the middle of a span, connect a partial span.
	if deleteStart[1] != 0 {
		span := textAreaSpan{
			previous: before,
			next:     after,
			offset:   t.spans[deleteStart[0]].offset,
			length:   deleteStart[1],
		}
		if t.spans[deleteStart[0]].length < 0 {
			span.length = -span.length
		}
		t.length += deleteStart[1] // This was previously subtracted.
		t.spans[before].next = len(t.spans)
		t.spans[after].previous = len(t.spans)
		before = len(t.spans)
		for row, lineStart := range t.lineStarts { // Also redirect line starts until the end of this new span.
			if lineStart[0] == deleteStart[0] {
				if lineStart[1] >= deleteStart[1] {
					t.lineStarts = t.lineStarts[:row] // Everything else is unknown at this point.
					break
				}
				t.lineStarts[row][0] = len(t.spans)
			}
		}
		t.spans = append(t.spans, span)
	}

	// If we insert text, connect a new span.
	if insert != "" {
		span := textAreaSpan{
			previous: before,
			next:     after,
			offset:   t.editText.Len(),
		}
		span.length, _ = t.editText.WriteString(insert)
		t.length += span.length
		t.spans[before].next = len(t.spans)
		t.spans[after].previous = len(t.spans)
		before = len(t.spans)
		t.spans = append(t.spans, span)
	}

	// If we stop deleting in the middle of a span, connect a partial span.
	if deleteEnd[1] != 0 {
		span := textAreaSpan{
			previous: before,
			next:     after,
			offset:   t.spans[deleteEnd[0]].offset + deleteEnd[1],
		}
		length := t.spans[deleteEnd[0]].length
		if length < 0 {
			span.length = length + deleteEnd[1]
			t.length -= span.length // This was previously subtracted.
		} else {
			span.length = length - deleteEnd[1]
			t.length += span.length // This was previously subtracted.
		}
		t.spans[before].next = len(t.spans)
		t.spans[after].previous = len(t.spans)
		deleteEnd[0], deleteEnd[1] = len(t.spans), 0
		t.spans = append(t.spans, span)
	}

	return deleteEnd
}

// Draw draws this primitive onto the screen.
func (t *TextArea) Draw(screen tcell.Screen) {
	t.Box.DrawForSubclass(screen, t)

	// Prepare
	x, y, width, height := t.GetInnerRect()
	if width == 0 || height == 0 {
		return // We have no space for anything.
	}
	columnOffset := t.columnOffset
	if t.wrap {
		columnOffset = 0
	}

	// Show/hide the cursor at the end.
	defer func() {
		if t.HasFocus() {
			if t.cursor.row >= 0 &&
				t.cursor.row-t.rowOffset >= 0 && t.cursor.row-t.rowOffset < height &&
				t.cursor.actualColumn-columnOffset >= 0 && t.cursor.actualColumn-columnOffset < width {
				screen.ShowCursor(x+t.cursor.actualColumn-columnOffset, y+t.cursor.row-t.rowOffset)
			} else {
				screen.HideCursor()
			}
		}
	}()

	// Placeholder.
	if t.length == 0 && len(t.placeholder) > 0 {
		t.drawPlaceholder(screen, x, y, width, height)
		return // We're done already.
	}

	// Make sure the visible lines are broken over.
	if t.lastWidth != width && t.lineStarts != nil {
		t.resetLines()
	}
	t.lastHeight, t.lastWidth = height, width
	t.extendLines(width, t.rowOffset+height)
	if len(t.lineStarts) <= t.rowOffset {
		return // It's scrolled out of view.
	}

	// If the cursor position is unknown, find it. This usually only happens
	// before the screen is drawn for the first time.
	if t.cursor.row < 0 {
		t.clampToCursor(0)
		if t.selectionStart.row < 0 {
			t.selectionStart = t.cursor
		}
	}

	// Print the text.
	var cluster, text string
	line := t.rowOffset
	pos := t.lineStarts[line]
	endPos := pos
	posX, posY := 0, 0
	for pos[0] != 1 {
		cluster, text, _, pos, endPos = t.step(text, pos, endPos)

		// Prepare drawing.
		clusterWidth := stringWidth(cluster)
		runes := []rune(cluster)
		style := t.selectedStyle
		fromRow, fromColumn := t.cursor.row, t.cursor.actualColumn
		toRow, toColumn := t.selectionStart.row, t.selectionStart.actualColumn
		if fromRow > toRow || fromRow == toRow && fromColumn > toColumn {
			fromRow, fromColumn, toRow, toColumn = toRow, toColumn, fromRow, fromColumn
		}
		if toRow < line ||
			toRow == line && toColumn <= posX ||
			fromRow > line ||
			fromRow == line && fromColumn > posX {
			style = t.textStyle
		}

		// Draw character.
		if posX+clusterWidth-columnOffset <= width && posX-columnOffset >= 0 && clusterWidth > 0 {
			screen.SetContent(x+posX-columnOffset, y+posY, runes[0], runes[1:], style)
		}

		// Advance.
		posX += clusterWidth
		if line+1 < len(t.lineStarts) && t.lineStarts[line+1] == pos {
			// We must break over.
			posY++
			if posY >= height {
				break // Done.
			}
			posX = 0
			line++
		}
	}
}

// drawPlaceholder draws the placeholder text into the given rectangle. It does
// not do anything if the text area already contains text or if there is no
// placeholder text.
func (t *TextArea) drawPlaceholder(screen tcell.Screen, x, y, width, height int) {
	posX, posY := x, y
	lastLineBreak, lastGraphemeBreak := x, x // Screen positions of the last possible line/grapheme break.
	iterateString(t.placeholder, func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth, boundaries int) bool {
		if posX+screenWidth > x+width {
			// This character doesn't fit. Break over to the next line.
			// Perform word wrapping first by copying the last word over to
			// the next line.
			clearX := lastLineBreak
			if lastLineBreak == x {
				clearX = lastGraphemeBreak
			}
			posY++
			if posY >= y+height {
				return true
			}
			newPosX := x
			for clearX < posX {
				main, comb, _, _ := screen.GetContent(clearX, posY-1)
				screen.SetContent(clearX, posY-1, ' ', nil, tcell.StyleDefault.Background(t.backgroundColor))
				screen.SetContent(newPosX, posY, main, comb, t.placeholderStyle)
				clearX++
				newPosX++
			}
			lastLineBreak, lastGraphemeBreak, posX = x, x, newPosX
		}

		// Draw this character.
		screen.SetContent(posX, posY, main, comb, t.placeholderStyle)
		posX += screenWidth
		switch boundaries & uniseg.MaskLine {
		case uniseg.LineMustBreak:
			posY++
			if posY >= y+height {
				return true
			}
			posX = x
		case uniseg.LineCanBreak:
			lastLineBreak = posX
		}
		lastGraphemeBreak = posX

		return false
	})
}

// resetLines resets the [lineStarts] array so [extendLines] has to be called
// again to access text information.
func (t *TextArea) resetLines() {
	t.truncateLines(0)
	t.cursor.row = -1
	t.widestLine = 0
}

// extendLines traverses the current text and extends t.lineStarts such that it
// describes at least maxLines+1 lines (or less if the text is shorter). Text is
// laid out for the given width while respecting the wrapping settings. It is
// assumed that if t.lineStarts already has entries, they obey the same rules.
//
// If width is 0, nothing happens.
func (t *TextArea) extendLines(width, maxLines int) {
	if width <= 0 {
		return
	}

	// Start with the first span.
	if len(t.lineStarts) == 0 {
		if len(t.spans) > 2 {
			t.lineStarts = append(t.lineStarts, [3]int{t.spans[0].next, 0, -1})
		} else {
			return // No text.
		}
	}

	// Determine starting positions and starting spans.
	pos := t.lineStarts[len(t.lineStarts)-1] // The starting position is the last known line.
	endPos := pos
	var (
		cluster, text                    string
		lineWidth, boundaries            int
		lastGraphemeBreak, lastLineBreak [3]int
		widthSinceLineBreak              int
	)
	for pos[0] != 1 {
		// Get the next grapheme cluster.
		cluster, text, boundaries, pos, endPos = t.step(text, pos, endPos)
		clusterWidth := stringWidth(cluster)
		lineWidth += clusterWidth
		widthSinceLineBreak += clusterWidth

		// Any line breaks?
		if !t.wrap || lineWidth <= width {
			if boundaries&uniseg.MaskLine == uniseg.LineMustBreak && (len(text) > 0 || uniseg.HasTrailingLineBreakInString(cluster)) {
				// We must break over.
				t.lineStarts = append(t.lineStarts, pos)
				if lineWidth > t.widestLine {
					t.widestLine = lineWidth
				}
				lineWidth = 0
				lastGraphemeBreak = [3]int{}
				lastLineBreak = [3]int{}
				widthSinceLineBreak = 0
				if len(t.lineStarts) > maxLines {
					break // We have enough lines, we can stop.
				}
				continue
			}
		} else { // t.wrap && lineWidth > width
			if !t.wordWrap || lastLineBreak == [3]int{} {
				if lastGraphemeBreak != [3]int{} { // We have at least one character on each line.
					// Break after last grapheme.
					t.lineStarts = append(t.lineStarts, lastGraphemeBreak)
					if lineWidth > t.widestLine {
						t.widestLine = lineWidth
					}
					lineWidth = clusterWidth
					lastLineBreak = [3]int{}
				}
			} else { // t.wordWrap && lastLineBreak != [3]int{}
				// Break after last line break opportunity.
				t.lineStarts = append(t.lineStarts, lastLineBreak)
				if lineWidth > t.widestLine {
					t.widestLine = lineWidth
				}
				lineWidth = widthSinceLineBreak
				lastLineBreak = [3]int{}
			}
		}

		// Analyze break opportunities.
		if boundaries&uniseg.MaskLine == uniseg.LineCanBreak {
			lastLineBreak = pos
			widthSinceLineBreak = 0
		}
		lastGraphemeBreak = pos

		// Can we stop?
		if len(t.lineStarts) > maxLines {
			break
		}
	}
}

// truncateLines truncates the trailing lines of the [TextArea.lineStarts]
// slice such that len(lineStarts) <= fromLine. If fromLine is negative, a value
// of 0 is assumed. If it is greater than the length of lineStarts, nothing
// happens.
func (t *TextArea) truncateLines(fromLine int) {
	if fromLine < 0 {
		fromLine = 0
	}
	if fromLine < len(t.lineStarts) {
		t.lineStarts = t.lineStarts[:fromLine]
	}
}

// clampToCursor ensures that the cursor is visible in the text area. If the
// cursor position is unknown, "startRow" helps reduce processing time by
// indicating the lowest row in which searching should start. Set this to 0 if
// you don't have any information where the cursor might be (but know that this
// is expensive for long texts). This function also sets the cursor clamp flag
// to false.
func (t *TextArea) clampToCursor(startRow int) {
	if t.cursor.row >= 0 {
		// This is the simple case because the current cursor position is known.
		if t.cursor.row < t.rowOffset {
			// We're above the viewport.
			t.rowOffset = t.cursor.row
		} else if t.cursor.row >= t.rowOffset+t.lastHeight {
			// We're below the viewport.
			t.rowOffset = t.cursor.row - t.lastHeight + 1
			if t.rowOffset >= len(t.lineStarts) {
				t.extendLines(t.lastWidth, t.rowOffset)
				if t.rowOffset >= len(t.lineStarts) {
					t.rowOffset = len(t.lineStarts) - 1
					if t.rowOffset < 0 {
						t.rowOffset = 0
					}
				}
			}
		}
		if !t.wrap {
			if t.cursor.actualColumn < t.columnOffset+minCursorPrefix {
				// We're left of the viewport.
				t.columnOffset = t.cursor.actualColumn - minCursorPrefix
				if t.columnOffset < 0 {
					t.columnOffset = 0
				}
			} else if t.cursor.actualColumn >= t.columnOffset+t.lastWidth-minCursorSuffix {
				// We're right of the viewport.
				t.columnOffset = t.cursor.actualColumn - t.lastWidth + minCursorSuffix
				if t.columnOffset >= t.widestLine {
					t.columnOffset = t.widestLine - 1
					if t.columnOffset < 0 {
						t.columnOffset = 0
					}
				}
			}
		}
		return
	}

	// The screen position of the cursor is unknown. Find it. This is expensive.
	// First, find the row.
	row := startRow
	if row < 0 {
		row = 0
	}
RowLoop:
	for {
		// Examine the current row.
		if row+1 >= len(t.lineStarts) {
			t.extendLines(t.lastWidth, row+1)
		}
		if row >= len(t.lineStarts) {
			t.cursor.row, t.cursor.actualColumn, t.cursor.pos = row, 0, [3]int{1, 0, -1}
			break // It's the end of the text.
		}

		// Check this row's spans to see if the cursor is in this row.
		pos := t.lineStarts[row]
		for pos[0] != 1 {
			if row+1 >= len(t.lineStarts) {
				break // It's the last row so the cursor must be in this row.
			}
			if t.cursor.pos[0] == pos[0] {
				// The cursor is in this span.
				if t.lineStarts[row+1][0] == pos[0] {
					// The next row starts with the same span.
					if t.cursor.pos[1] >= t.lineStarts[row+1][1] {
						// The cursor is not in this row.
						row++
						continue RowLoop
					} else {
						// The cursor is in this row.
						break
					}
				} else {
					// The next row starts with a different span. The cursor
					// must be in this row.
					break
				}
			} else {
				// The cursor is in a different span.
				if t.lineStarts[row+1][0] == pos[0] {
					// The next row starts with the same span. This row is
					// irrelevant.
					row++
					continue RowLoop
				} else {
					// The next row starts with a different span. Move towards it.
					pos = [3]int{t.spans[pos[0]].next, 0, -1}
				}
			}
		}

		// Try to find the screen position in this row.
		pos = t.lineStarts[row]
		endPos := pos
		column := 0
		var cluster, text string
		for {
			if pos[0] == 1 || t.cursor.pos[0] == pos[0] && t.cursor.pos[1] == pos[1] {
				// We found the position. We're done.
				t.cursor.row, t.cursor.actualColumn, t.cursor.pos = row, column, pos
				break RowLoop
			}
			cluster, text, _, pos, endPos = t.step(text, pos, endPos)
			if row+1 < len(t.lineStarts) && t.lineStarts[row+1] == pos {
				// We reached the end of the line. Go to the next one.
				row++
				continue RowLoop
			}
			clusterWidth := stringWidth(cluster)
			column += clusterWidth
		}
	}

	if t.cursor.row >= 0 {
		// We know the position now. Adapt offsets.
		t.clampToCursor(startRow)
	}
}

// step is similar to uniseg.StepString() but it iterates over the piece chain,
// starting with "pos", a span position plus state (which may be -1 for the
// start of the text). The returned "boundaries" value is same value returned by
// uniseg.StepString(). The "pos" and "endPos" positions refer to the start and
// the end of the "text" string, respectively. For the first call, text may be
// empty and pos/endPos may be the same. For consecutive calls, provide "rest"
// as the text and "newPos" and "newEndPos" as the new positions/states. An
// empty "rest" string indicates the end of the text. The "endPos" state is not
// used.
func (t *TextArea) step(text string, pos, endPos [3]int) (cluster, rest string, boundaries int, newPos, newEndPos [3]int) {
	if pos[0] == 1 {
		return // We're already past the end.
	}

	// We want to make sure we have a text at least the size of a grapheme
	// cluster.
	span := t.spans[pos[0]]
	if len(text) < maxGraphemeClusterSize &&
		(span.length < 0 && -span.length-pos[1] >= maxGraphemeClusterSize ||
			span.length > 0 && t.spans[pos[0]].length-pos[1] >= maxGraphemeClusterSize) {
		// We can use a substring of one span.
		if span.length < 0 {
			text = t.initialText[span.offset+pos[1] : span.offset-span.length]
		} else {
			text = t.editText.String()[span.offset+pos[1] : span.offset+span.length]
		}
		endPos = [3]int{span.next, 0, -1}
	} else {
		// We have to compose the text from multiple spans.
		for len(text) < maxGraphemeClusterSize && endPos[0] != 1 {
			endSpan := t.spans[endPos[0]]
			var moreText string
			if endSpan.length < 0 {
				moreText = t.initialText[endSpan.offset+endPos[1] : endSpan.offset-endSpan.length]
			} else {
				moreText = t.editText.String()[endSpan.offset+endPos[1] : endSpan.offset+endSpan.length]
			}
			if len(moreText) > maxGraphemeClusterSize {
				moreText = moreText[:maxGraphemeClusterSize]
			}
			text += moreText
			endPos[1] += len(moreText)
			if endPos[1] >= endSpan.length {
				endPos[0], endPos[1] = endSpan.next, 0
			}
		}
	}

	// Run the grapheme cluster iterator.
	cluster, text, boundaries, pos[2] = uniseg.StepString(text, pos[2])
	pos[1] += len(cluster)
	for pos[0] != 1 && (span.length < 0 && pos[1] >= -span.length || span.length >= 0 && pos[1] >= span.length) {
		pos[0] = span.next
		if span.length < 0 {
			pos[1] += span.length
		} else {
			pos[1] -= span.length
		}
		span = t.spans[pos[0]]
	}

	return cluster, text, boundaries, pos, endPos
}

// moveCursor sets the cursor's screen position and span position for the given
// row and column which are screen space coordinates relative to the top-left
// corner of the text area's full text (visible or not). The column value may be
// negative, in which case, the cursor will be placed at the end of the line.
// The cursor's actual position will be aligned with a grapheme cluster
// boundary. The next call to [Draw] will attempt to keep the cursor in the
// viewport.
func (t *TextArea) moveCursor(row, column int) {
	// Are we within the range of rows?
	if len(t.lineStarts) <= row {
		// No. Extent the line buffer.
		t.extendLines(t.lastWidth, row)
	}
	if len(t.lineStarts) == 0 {
		return // No lines. Nothing to do.
	}
	if row < 0 {
		// We're at the start of the text.
		row = 0
		column = 0
	} else if row >= len(t.lineStarts) {
		// We're already past the end.
		row = len(t.lineStarts) - 1
		column = -1
	}

	// Iterate through this row until we find the position.
	t.cursor.row, t.cursor.actualColumn = row, 0
	if t.wrap {
		t.cursor.actualColumn = 0
	}
	pos := t.lineStarts[row]
	endPos := pos
	var cluster, text string
	for pos[0] != 1 {
		oldPos := pos // We may have to revert to this position.
		cluster, text, _, pos, endPos = t.step(text, pos, endPos)
		clusterWidth := stringWidth(cluster)
		if len(t.lineStarts) > row+1 && pos == t.lineStarts[row+1] || // We've reached the end of the line.
			column >= 0 && t.cursor.actualColumn+clusterWidth > column { // We're past the requested column.
			pos = oldPos
			break
		}
		t.cursor.actualColumn += clusterWidth
	}

	if column < 0 {
		t.cursor.column = t.cursor.actualColumn
	} else {
		t.cursor.column = column
	}
	t.cursor.pos = pos
	t.clampToCursor(row)
}

// moveWordRight moves the cursor to the end of the current or next word. The
// next call to [Draw] will attempt to keep the cursor in the viewport.
func (t *TextArea) moveWordRight() {
	// Because we rely on clampToCursor to calculate the new screen position,
	// this is an expensive operation for large texts.
	pos := t.cursor.pos
	endPos := pos
	var (
		cluster, text string
		inWord        bool
	)
	for pos[0] != 0 {
		var boundaries int
		oldPos := pos
		cluster, text, boundaries, pos, endPos = t.step(text, pos, endPos)
		if oldPos == t.cursor.pos {
			continue // Skip the first character.
		}
		firstRune, _ := utf8.DecodeRuneInString(cluster)
		if !unicode.IsSpace(firstRune) && !unicode.IsPunct(firstRune) {
			inWord = true
		}
		if inWord && boundaries&uniseg.MaskWord != 0 {
			pos = oldPos
			break
		}
	}
	startRow := t.cursor.row
	t.cursor.row, t.cursor.column, t.cursor.actualColumn = -1, 0, 0
	t.cursor.pos = pos
	t.clampToCursor(startRow)
}

// moveWordLeft moves the cursor to the beginning of the current or previous
// word.
func (t *TextArea) moveWordLeft() {
	// We go back row by row, trying to find the last word boundary before the
	// cursor.
	row := t.cursor.row
	if row+1 < len(t.lineStarts) {
		t.extendLines(t.lastWidth, row+1)
	}
	if row >= len(t.lineStarts) {
		row = len(t.lineStarts) - 1
	}
	for row >= 0 {
		pos := t.lineStarts[row]
		endPos := pos
		var lastWordBoundary [3]int
		var (
			cluster, text string
			inWord        bool
			boundaries    int
		)
		for pos[0] != 1 && pos != t.cursor.pos {
			oldBoundaries := boundaries
			oldPos := pos
			cluster, text, boundaries, pos, endPos = t.step(text, pos, endPos)
			firstRune, _ := utf8.DecodeRuneInString(cluster)
			wordRune := !unicode.IsSpace(firstRune) && !unicode.IsPunct(firstRune)
			if oldBoundaries&uniseg.MaskWord != 0 {
				if pos != t.cursor.pos && !inWord && wordRune {
					// A boundary transitioning from a space/punctuation word to
					// a letter word.
					lastWordBoundary = oldPos
				}
				inWord = false
			}
			if wordRune {
				inWord = true
			}
		}
		if lastWordBoundary[0] != 0 {
			// We found something.
			t.cursor.pos = lastWordBoundary
			break
		}
		row--
	}
	if row < 0 {
		// We didn't find anything. We're at the start of the text.
		t.cursor.pos = [3]int{t.spans[0].next, 0, -1}
		row = 0
	}
	t.cursor.row, t.cursor.column, t.cursor.actualColumn = -1, 0, 0
	t.clampToCursor(row)
}

// deleteLine deletes all characters between the last newline before the cursor
// and the next newline after the cursor (inclusive).
func (t *TextArea) deleteLine() {
	// We go back startRow by startRow, trying to find the last mandatory line break
	// before the cursor.
	startRow := t.cursor.row
	if startRow+1 < len(t.lineStarts) {
		t.extendLines(t.lastWidth, startRow+1)
	}
	if len(t.lineStarts) == 0 {
		return // Nothing to delete.
	}
	if startRow >= len(t.lineStarts) {
		startRow = len(t.lineStarts) - 1
	}
	for startRow >= 0 {
		// What's the last rune before the start of the line?
		pos := t.lineStarts[startRow]
		span := t.spans[pos[0]]
		var text string
		if pos[1] > 0 {
			// Extract text from this span.
			if span.length < 0 {
				text = t.initialText
			} else {
				text = t.editText.String()
			}
			text = text[:span.offset+pos[1]]
		} else {
			// Extract text from the previous span.
			if span.previous != 0 {
				span = t.spans[span.previous]
				if span.length < 0 {
					text = t.initialText[:span.offset-span.length]
				} else {
					text = t.editText.String()[:span.offset+span.length]
				}
			}
		}
		if uniseg.HasTrailingLineBreakInString(text) {
			// The row before this one ends with a mandatory line break. This is
			// the first line we will delete.
			break
		}
		startRow--
	}
	if startRow < 0 {
		// We didn't find anything. It'll be the first line.
		startRow = 0
	}

	// Find the next line break after the cursor.
	pos := t.cursor.pos
	endPos := pos
	var cluster, text string
	for pos[0] != 1 {
		cluster, text, _, pos, endPos = t.step(text, pos, endPos)
		if uniseg.HasTrailingLineBreakInString(cluster) {
			break
		}
	}

	// Delete the text.
	t.cursor.pos = t.replace(t.lineStarts[startRow], pos, "", false)
	t.cursor.row = -1
	t.truncateLines(startRow)
	t.clampToCursor(startRow)
}

// getSelection returns the current selection as span locations where the first
// returned location is always before or the same as the second returned
// location. This assumes that the cursor and selection positions are known. The
// third return value is the starting row of the selection.
func (t *TextArea) getSelection() ([3]int, [3]int, int) {
	from := t.selectionStart.pos
	to := t.cursor.pos
	row := t.selectionStart.row
	if t.cursor.row < t.selectionStart.row ||
		(t.cursor.row == t.selectionStart.row && t.cursor.column < t.selectionStart.column) {
		from, to = to, from
		row = t.cursor.row
	}
	return from, to, row
}

// getSelectedText returns the text of the current selection.
func (t *TextArea) getSelectedText() string {
	var text strings.Builder

	from, to, _ := t.getSelection()
	for from[0] != to[0] {
		span := t.spans[from[0]]
		if span.length < 0 {
			text.WriteString(t.initialText[span.offset+from[1] : span.offset-span.length])
		} else {
			text.WriteString(t.editText.String()[span.offset+from[1] : span.offset+span.length])
		}
		from[0], from[1] = span.next, 0
	}
	if from[0] != 1 && from[1] < to[1] {
		span := t.spans[from[0]]
		if span.length < 0 {
			text.WriteString(t.initialText[span.offset+from[1] : span.offset+to[1]])
		} else {
			text.WriteString(t.editText.String()[span.offset+from[1] : span.offset+to[1]])
		}
	}

	return text.String()
}

// InputHandler returns the handler for this primitive.
func (t *TextArea) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		// All actions except a few specific ones are "other" actions.
		newLastAction := taActionOther
		defer func() {
			t.lastAction = newLastAction
		}()

		// Process the different key events.
		switch key := event.Key(); key {
		case tcell.KeyLeft: // Move one grapheme cluster to the left.
			if event.Modifiers()&tcell.ModAlt == 0 {
				// Regular movement.
				if t.cursor.actualColumn == 0 {
					// Move to the end of the previous row.
					if t.cursor.row > 0 {
						t.moveCursor(t.cursor.row-1, -1)
					}
				} else {
					// Move one grapheme cluster to the left.
					t.moveCursor(t.cursor.row, t.cursor.actualColumn-1)
				}
				if event.Modifiers()&tcell.ModShift == 0 {
					t.selectionStart = t.cursor
				}
			} else if !t.wrap {
				// Just scroll.
				t.columnOffset--
				if t.columnOffset < 0 {
					t.columnOffset = 0
				}
			}
		case tcell.KeyRight: // Move one grapheme cluster to the right.
			if event.Modifiers()&tcell.ModAlt == 0 {
				// Regular movement.
				if t.cursor.pos[0] != 1 {
					var cluster string
					cluster, _, _, t.cursor.pos, _ = t.step("", t.cursor.pos, t.cursor.pos)
					if len(t.lineStarts) <= t.cursor.row+1 {
						t.extendLines(t.lastWidth, t.cursor.row+1)
					}
					if t.cursor.row+1 < len(t.lineStarts) && t.lineStarts[t.cursor.row+1] == t.cursor.pos {
						// We've reached the end of the line.
						t.cursor.row++
						t.cursor.actualColumn = 0
						t.cursor.column = 0
						t.clampToCursor(t.cursor.row)
					} else {
						// Move one character to the right.
						t.moveCursor(t.cursor.row, t.cursor.actualColumn+stringWidth(cluster))
					}
				}
				if event.Modifiers()&tcell.ModShift == 0 {
					t.selectionStart = t.cursor
				}
			} else if !t.wrap {
				// Just scroll.
				t.columnOffset++
				if t.columnOffset >= t.widestLine {
					t.columnOffset = t.widestLine - 1
					if t.columnOffset < 0 {
						t.columnOffset = 0
					}
				}
			}
		case tcell.KeyDown: // Move one row down.
			if event.Modifiers()&tcell.ModAlt == 0 {
				// Regular movement.
				t.moveCursor(t.cursor.row+1, t.cursor.column)
				if event.Modifiers()&tcell.ModShift == 0 {
					t.selectionStart = t.cursor
				}
			} else {
				// Just scroll.
				t.rowOffset++
				if t.rowOffset >= len(t.lineStarts) {
					t.extendLines(t.lastWidth, t.rowOffset)
					if t.rowOffset >= len(t.lineStarts) {
						t.rowOffset = len(t.lineStarts) - 1
						if t.rowOffset < 0 {
							t.rowOffset = 0
						}
					}
				}
			}
		case tcell.KeyUp: // Move one row up.
			if event.Modifiers()&tcell.ModAlt == 0 {
				// Regular movement.
				t.moveCursor(t.cursor.row-1, t.cursor.column)
				if event.Modifiers()&tcell.ModShift == 0 {
					t.selectionStart = t.cursor
				}
			} else {
				// Just scroll.
				t.rowOffset--
				if t.rowOffset < 0 {
					t.rowOffset = 0
				}
			}
		case tcell.KeyHome, tcell.KeyCtrlA: // Move to the start of the line.
			t.moveCursor(t.cursor.row, 0)
			if event.Modifiers()&tcell.ModShift == 0 {
				t.selectionStart = t.cursor
			}
		case tcell.KeyEnd, tcell.KeyCtrlE: // Move to the end of the line.
			t.moveCursor(t.cursor.row, -1)
			if event.Modifiers()&tcell.ModShift == 0 {
				t.selectionStart = t.cursor
			}
		case tcell.KeyPgDn, tcell.KeyCtrlF: // Move one page down.
			t.moveCursor(t.cursor.row+t.lastHeight, t.cursor.column)
			if event.Modifiers()&tcell.ModShift == 0 {
				t.selectionStart = t.cursor
			}
		case tcell.KeyPgUp, tcell.KeyCtrlB: // Move one page up.
			t.moveCursor(t.cursor.row-t.lastHeight, t.cursor.column)
			if event.Modifiers()&tcell.ModShift == 0 {
				t.selectionStart = t.cursor
			}
		case tcell.KeyEnter: // Insert a newline.
			from, to, row := t.getSelection()
			t.cursor.pos = t.replace(from, to, NewLine, t.lastAction != taActionTypeSpace)
			t.cursor.row = -1
			t.truncateLines(row - 1)
			t.clampToCursor(row)
			t.selectionStart = t.cursor
			newLastAction = taActionTypeSpace
		case tcell.KeyTab: // Insert TabSize spaces.
			from, to, row := t.getSelection()
			t.cursor.pos = t.replace(from, to, strings.Repeat(" ", TabSize), t.lastAction != taActionTypeSpace)
			t.cursor.row = -1
			t.truncateLines(row - 1)
			t.clampToCursor(row)
			t.selectionStart = t.cursor
			newLastAction = taActionTypeSpace
		case tcell.KeyRune:
			if event.Modifiers()&tcell.ModAlt > 0 {
				// We accept some Alt- key combinations.
				switch event.Rune() {
				case 'f':
					t.moveWordRight()
					if event.Modifiers()&tcell.ModShift == 0 {
						t.selectionStart = t.cursor
					}
				case 'b':
					t.moveWordLeft()
					if event.Modifiers()&tcell.ModShift == 0 {
						t.selectionStart = t.cursor
					}
				}
			} else {
				// Other keys are simply accepted as regular characters.
				r := event.Rune()
				from, to, row := t.getSelection()
				newLastAction = taActionTypeNonSpace
				if unicode.IsSpace(r) {
					newLastAction = taActionTypeSpace
				}
				t.cursor.pos = t.replace(from, to, string(r), newLastAction == t.lastAction || t.lastAction == taActionTypeNonSpace && newLastAction == taActionTypeSpace)
				t.cursor.row = -1
				t.truncateLines(row - 1)
				t.clampToCursor(row)
				t.selectionStart = t.cursor
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2: // Delete backwards. tcell.KeyBackspace is the same as tcell.CtrlH.
			from, to, row := t.getSelection()
			if from != to {
				// Simply delete the current selection.
				t.cursor.pos = t.replace(from, to, "", false)
				t.cursor.row = -1
				t.truncateLines(row - 1)
				t.clampToCursor(row)
				t.selectionStart = t.cursor
				break
			}

			// Move the cursor back by one grapheme cluster.
			endPos := t.cursor.pos
			if t.cursor.actualColumn == 0 {
				// Move to the end of the previous row.
				if t.cursor.row > 0 {
					t.moveCursor(t.cursor.row-1, -1)
				}
			} else {
				// Move one grapheme cluster to the left.
				t.moveCursor(t.cursor.row, t.cursor.actualColumn-1)
			}

			// Remove that last grapheme cluster.
			if t.cursor.pos != endPos {
				t.cursor.pos = t.replace(t.cursor.pos, endPos, "", t.lastAction != taActionBackspace) // Delete the character.
				t.cursor.pos[2] = endPos[2]
				t.truncateLines(t.cursor.row - 1)
				t.clampToCursor(t.cursor.row)
				newLastAction = taActionBackspace
			}
			t.selectionStart = t.cursor
		case tcell.KeyDelete, tcell.KeyCtrlD: // Delete forward.
			from, to, row := t.getSelection()
			if from != to {
				// Simply delete the current selection.
				t.cursor.pos = t.replace(from, to, "", false)
				t.cursor.row = -1
				t.truncateLines(row - 1)
				t.clampToCursor(row)
				t.selectionStart = t.cursor
				break
			}

			if t.cursor.pos[0] != 1 {
				_, _, _, endPos, _ := t.step("", t.cursor.pos, t.cursor.pos)
				t.cursor.pos = t.replace(t.cursor.pos, endPos, "", t.lastAction != taActionDelete) // Delete the character.
				t.cursor.pos[2] = endPos[2]
				t.truncateLines(t.cursor.row - 1)
				t.clampToCursor(t.cursor.row)
				newLastAction = taActionDelete
			}
			t.selectionStart = t.cursor
		case tcell.KeyCtrlK: // Delete everything under and to the right of the cursor until before the next newline character.
			pos := t.cursor.pos
			endPos := pos
			var cluster, text string
			for pos[0] != 1 {
				var boundaries int
				oldPos := pos
				cluster, text, boundaries, pos, endPos = t.step(text, pos, endPos)
				if boundaries&uniseg.MaskLine == uniseg.LineMustBreak {
					if uniseg.HasTrailingLineBreakInString(cluster) {
						pos = oldPos
					}
					break
				}
			}
			t.cursor.pos = t.replace(t.cursor.pos, pos, "", false)
			row := t.cursor.row
			t.cursor.row = -1
			t.truncateLines(row - 1)
			t.clampToCursor(row)
			t.selectionStart = t.cursor
		case tcell.KeyCtrlW: // Delete from the start of the current word to the left of the cursor.
			pos := t.cursor.pos
			t.moveWordLeft()
			t.cursor.pos = t.replace(t.cursor.pos, pos, "", false)
			row := t.cursor.row - 1
			t.cursor.row = -1
			t.truncateLines(row)
			t.clampToCursor(row)
			t.selectionStart = t.cursor
		case tcell.KeyCtrlU: // Delete the current line.
			t.deleteLine()
			t.selectionStart = t.cursor
		case tcell.KeyCtrlQ: // Copy to clipboard.
			if t.cursor != t.selectionStart {
				t.copyToClipboard(t.getSelectedText())
				t.selectionStart = t.cursor
			}
		case tcell.KeyCtrlX: // Cut to clipboard.
			if t.cursor != t.selectionStart {
				t.copyToClipboard(t.getSelectedText())
				from, to, row := t.getSelection()
				t.cursor.pos = t.replace(from, to, "", false)
				t.cursor.row = -1
				t.truncateLines(row - 1)
				t.clampToCursor(row)
				t.selectionStart = t.cursor
			}
		case tcell.KeyCtrlV: // Paste from clipboard.
			from, to, row := t.getSelection()
			t.cursor.pos = t.replace(from, to, t.pasteFromClipboard(), false)
			t.cursor.row = -1
			t.truncateLines(row - 1)
			t.clampToCursor(row)
			t.selectionStart = t.cursor
		}
	})
}