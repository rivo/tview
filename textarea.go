package tview

import (
	"strings"

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
)

var (
	// NewLine is the string sequence to be inserted when hitting the Enter key
	// in a TextArea. The default is "\n" but you may change it to "\r\n" if
	// required.
	NewLine = "\n"
)

// textAreaSpan represents a range of text in a text area. The text area widget
// roughly follows the concept of Piece Chains outline in
// http://www.catch22.net/tuts/neatpad/piece-chains with some modifications.
// This type represents a "span" (or "piece") and thus refers to a subset of the
// text in the editor as part of a doubly-linked list.
//
// In most places where we reference a position in the text, we use a
// two-element int array. The first element is the index of the referenced span
// in the piece chain. The second element is the offset into the span's
// referenced text (relative to the span's start), its value is always >= 0 and
// < span.length. Sometimes, we may use a three-element int array which also
// contains the corresponding text parser's state in the third position.
//
// A range of text is represented by a span range which is a starting position
// (int array) and an ending position (int array). The starting position
// references the first character of the range, the ending position references
// the position after the last character of the range. The end of the text is
// therefore always [2]int{1, 0}, position 0 of the ending sentinel.
type textAreaSpan struct {
	// Links to the previous and next textAreaSpan objects as indices into the
	// TextArea.spans slice. The sentinel spans (index 0 and 1) have -1 as their
	// previous or next links.
	previous, next int

	// The start index and the length of the text segment this span represents.
	// If "length" is negative, the span represents a substring of
	// TextArea.initialText and the actual length must be its absolute value. If
	// it is positive, the span represents a substring of TextArea.editText. For
	// the sentinel spans (index 0 and 1), both values will be 0.
	offset, length int
}

// TextArea implements a simple text editor for multi-line text. Multi-color
// text is not supported. Word-wrapping is enabled by default but can be turned
// off or be changed to character-wrapping.
//
// Navigation and Editing
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
//   - Alt-Left arrow: Scroll the page to the right, leaving the cursor in its
//     position. Ignored if wrapping is enabled.
//   - Alt-Right arrow: Scroll the page to the left, leaving the cursor in its
//     position. Ignored if wrapping is enabled.
//
// Entering a character (rune) will insert it at the current cursor location.
// Subsequent characters are moved accordingly. If the cursor is outside the
// visible area, any changes to the text will move it into the visible area. The
// following keys can also be used to modify the text:
//
//   - Enter: Insert a newline character (see NewLine).
//   - Tab: Insert TabSize spaces.
//   - Ctrl-H, Backspace: Delete one character to the left of the cursor.
//   - Ctrl-D, Delete: Delete the character under the cursor (or the first
//     character on the next line if the cursor is at the end of a line).
//   - Ctrl-K: Delete everything under and to the right of the cursor.
//   - Ctrl-W: Delete from the start of the current word to the left of the
//     cursor.
//   - Ctrl-U: Delete the current line, i.e. everything after the last newline
//     character before the cursor up until the next newline character. This may
//     span multiple lines if wrapping is enabled.
//
// Text can be selected by moving the cursor while holding the Shift key. Thus
// when text is selected:
//
//   - Entering a character (rune) will replace the selected text with the new
//     character.
//   - Backspace, delete: Delete the selected text.
//   - Ctrl-Q: Copy the selected text into the clipboard, unselect the text.
//   - Ctrl-X: Copy the selected text into the clipboard and delete it.
//   - Ctrl-V: Replace the selected text with the clipboard text. If no text is
//     selected, the clipboard text will be inserted at the cursor location.
//
// The default clipboard is an internal text buffer, i.e. the operating system's
// clipboard is not used. The Ctrl-Q key was chosen for the "copy" function
// because the Ctrl-C key is the default key to stop the application. If your
// application frees up the global Ctrl-C key and you want to bind it to the
// "copy to clipboard" function, you may use SetInputCapture() to override the
// Ctrl-Q key to implement copying to the clipboard.
//
// Similarly, if you want to implement your own clipboard (or make use of your
// operating system's clipboard), you can also use SetInputCapture() to override
// the key binds for copy, cut, and paste. The GetSelection(), ReplaceText(),
// and SetSelection() provide all the functionality needed for your own
// clipboard. TODO: This will need to be reviewed.
//
// The text area also supports Undo:
//
//    - Ctrl-Z: Undo the last change.
//    - Ctrl-Y: Redo the last Undo change.
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

	// The style of the placeholder text.
	placeholderStyle tcell.Style

	// Text manipulation related fields:

	// The text area's text prior to any editing.
	initialText string

	// Any text that's been added by the user at some point.
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

	// The undo stack's items are the first of two consecutive indices into the
	// spans slice. The first referenced span is a copy of the one before the
	// modified span range, thse second referenced span is a copy of the one
	// after the modified span range.
	undoStack []int

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

	// The inner width of the text area the last time it was drawn.
	lastWidth int

	// Text positions and states of the start of lines. Each element is a span
	// position (see textAreaSpan) and a state as returned by uniseg.Step(). Not
	// all lines of the text may be contained at any time, extend as needed with
	// the TextArea.extendLines() function.
	lineStarts [][3]int

	// The cursor always points to the next position where a new character would
	// be placed. This is the row (index 0) and column (index 1) in screen space
	// but relative to the start of the text which may be outside the text
	// area's box. The column value may be larger than where the cursor actually
	// is if the line the cursor is on is shorter. These two values may be
	// determined yet. The remaining values (indices 2-4) is a textAreaSpan
	// temporarily negative if the cursor's screen position hasn't been
	// position with state for the actual next character.
	cursor [5]int
}

// NewTextArea returns a new text area. For an empty text area, provide an empty
// string.
func NewTextArea(text string) *TextArea {
	t := &TextArea{
		Box:              NewBox(),
		wrap:             true,
		wordWrap:         true,
		placeholderStyle: tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.TertiaryTextColor),
		textStyle:        tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.PrimaryTextColor),
		initialText:      text,
		spans:            make([]textAreaSpan, 2, pieceChainMinCap), // We reserve some space to avoid reallocations right when editing starts.
	}
	t.editText.Grow(editBufferMinCap)
	t.spans[0] = textAreaSpan{previous: -1, next: 1}
	t.spans[1] = textAreaSpan{previous: 0, next: -1}
	if len(text) > 0 {
		t.spans = append(t.spans, textAreaSpan{
			previous: 0,
			next:     1,
			offset:   0,
			length:   -len(text),
		})
		t.spans[0].next = 2
		t.spans[1].previous = 2
		t.length = len(text)
		t.cursor = [5]int{-1, -1, 1, 0, -1}
	} else {
		t.cursor = [5]int{0, 0, 1, 0, -1}
	}
	return t
}

// SetWrap sets the flag that, if true, leads to lines that are longer than the
// available width being wrapped onto the next line. If false, any characters
// beyond the available width are not displayed.
func (t *TextArea) SetWrap(wrap bool) *TextArea {
	//TODO: Existing text needs reformatting.
	t.wrap = wrap
	return t
}

// SetWordWrap sets the flag that, if true and if the "wrap" flag is also true
// (see SetWrap()), wraps the line at spaces or after punctuation marks.
//
// This flag is ignored if the "wrap" flag is false.
func (t *TextArea) SetWordWrap(wrapOnWords bool) *TextArea {
	//TODO: Existing text needs reformatting.
	t.wordWrap = wrapOnWords
	return t
}

// SetPlaceholder sets the text to be displayed when the text area is empty.
func (t *TextArea) SetPlaceholder(placeholder string) *TextArea {
	t.placeholder = placeholder
	return t
}

// SetMaxLength sets the maximum number of bytes allowed in the text area. If 0,
// there is no limit.
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

// replace deletes a range of text and inserts the given text at that position.
// If the resulting text would exceed the maximum length, the function does not
// do anything. See textAreaSpan for information about text positions and span
// ranges.
//
// This function does not generate Undo events. Undo events are generated
// elsewhere, when the user changes their type of edit.
func (t *TextArea) replace(deleteStart, deleteEnd [2]int, insert string) {
	// Check max length.
	if t.maxLength > 0 && t.length+len(insert) > t.maxLength {
		return
	}

	// Delete.
	for deleteStart[0] != deleteEnd[0] {
		if deleteStart[1] == 0 {
			// Delete this entire span.
			deleteStart[0] = t.deleteSpan(deleteStart[0])
			deleteStart[1] = 0
		} else {
			// Delete a partial span at the end.
			if t.spans[deleteStart[0]].length < 0 {
				// Initial text span. Has negative length.
				t.length -= -t.spans[deleteStart[0]].length - deleteStart[1]
				t.spans[deleteStart[0]].length = -deleteStart[1]
			} else {
				// Edit buffer span. Has positive length.
				t.length -= t.spans[deleteStart[0]].length - deleteStart[1]
				t.spans[deleteStart[0]].length = deleteStart[1]
			}
			deleteStart[0] = t.spans[deleteStart[0]].next
			deleteStart[1] = 0
		}
	} // At this point, deleteStart[0] == deleteEnd[0].
	if deleteEnd[1] > deleteStart[1] {
		if deleteStart[1] == 0 {
			// Delete a partial span at the beginning.
			t.length -= deleteEnd[1]
			if t.spans[deleteEnd[0]].length < 0 {
				// Initial text span. Has negative length.
				t.spans[deleteEnd[0]].length += deleteEnd[1]
			} else {
				// Edit buffer span. Has positive length.
				t.spans[deleteEnd[0]].length -= deleteEnd[1]
			}
			t.spans[deleteEnd[0]].offset += deleteEnd[1]
		} else {
			// Delete in the middle by splitting the span.
			deleteStart[0] = t.splitSpan(deleteStart[0], deleteStart[1])
			deleteStart[1] = 0
		}
	}

	// Insert.
	if len(insert) > 0 {
		spanIndex, offset := deleteStart[0], deleteStart[1]
		span := t.spans[spanIndex]

		if offset == 0 {
			previousSpan := t.spans[span.previous]
			if previousSpan.length > 0 && previousSpan.offset+previousSpan.length == t.editText.Len() {
				// We can simply append to the edit buffer.
				length, _ := t.editText.WriteString(insert)
				previousSpan.length += length
				t.length += length
			} else {
				// Insert a new span.
				t.insertSpan(insert, spanIndex)
			}
		} else {
			// Split and insert.
			spanIndex = t.splitSpan(spanIndex, offset)
			t.insertSpan(insert, spanIndex)
		}
	}
}

// deleteSpan removes the span with the given index from the piece chain. It
// returns the index of the span after the deleted span (or the provided index
// if no span was deleted due to an invalid span index).
//
// This function also adjusts TextArea.length.
func (t *TextArea) deleteSpan(index int) int {
	if index < 2 || index >= len(t.spans) {
		return index
	}

	// Remove from piece chain.
	previous := t.spans[index].previous
	next := t.spans[index].next
	t.spans[previous].next = next
	t.spans[next].previous = previous

	// Adjust total length.
	length := t.spans[index].length
	if length < 0 {
		length = -length
	}
	t.length -= length

	return next
}

// splitSpan splits the span with the given index at the given offset into two
// spans. It returns the index of the span after the split or the provided
// index if no span was split due to an invalid span index or an invalid
// offset.
func (t *TextArea) splitSpan(index, offset int) int {
	if index < 2 || index >= len(t.spans) || offset <= 0 || offset >= t.spans[index].length {
		return index
	}

	// Make a new trailing span.
	span := t.spans[index]
	newSpan := textAreaSpan{
		previous: index,
		next:     span.next,
		offset:   span.offset + offset,
	}

	// Adjust lengths.
	if span.length < 0 {
		// Initial text span. Has negative length.
		newSpan.length = span.length + offset
		span.length = -offset
	} else {
		// Edit buffer span. Has positive length.
		newSpan.length = span.length - offset
		span.length = offset
	}

	// Insert it.
	newIndex := len(t.spans)
	t.spans = append(t.spans, newSpan)
	t.spans[span.next].previous = newIndex
	span.next = newIndex

	return newIndex
}

// insertSpan inserts the a span with the given text into the piece chain before
// the span with the given index and returns the index of the newly inserted
// span. If index <= 0, nothing happens and 1 is returned. The text is appended
// to the edit buffer. The length of the text is added to TextArea.length.
func (t *TextArea) insertSpan(text string, index int) int {
	if index < 1 || index >= len(t.spans) {
		return 1
	}

	// Make a new span.
	nextSpan := t.spans[index]
	span := textAreaSpan{
		previous: nextSpan.previous,
		next:     index,
		offset:   t.editText.Len(),
	}
	span.length, _ = t.editText.WriteString(text)

	// Insert into piece chain.
	newIndex := len(t.spans)
	t.spans[nextSpan.previous].next = newIndex
	nextSpan.previous = newIndex
	t.spans = append(t.spans, span)

	// Adjust text area length.
	t.length += span.length

	return newIndex
}

// Draw draws this primitive onto the screen.
func (t *TextArea) Draw(screen tcell.Screen) {
	t.Box.DrawForSubclass(screen, t)

	// Prepare
	x, y, width, height := t.GetInnerRect()
	if width == 0 || height == 0 {
		return // We have no space for anything.
	}

	// Placeholder.
	if t.length == 0 && len(t.placeholder) > 0 {
		t.drawPlaceholder(screen, x, y, width, height)
		return // We're done already.
	}

	// Make sure the visible lines are broken over.
	if t.lastWidth != width && t.lineStarts != nil {
		t.lineStarts = t.lineStarts[:0]
		t.lastWidth = width
		t.cursor[0], t.cursor[1] = -1, -1
	}
	t.extendLines(width, t.rowOffset+height)
	if len(t.lineStarts) <= t.rowOffset {
		return // It's scrolled out of view.
	}

	// Print the text.
	var (
		cluster, text string
		cursorVisible bool
	)
	line := t.rowOffset
	pos := t.lineStarts[line]
	endPos := pos
	posX, posY := 0, 0
	columnOffset := t.columnOffset
	if t.wrap {
		columnOffset = 0
	}
	if t.cursor[0] < 0 && t.cursor[2] == pos[0] && t.cursor[3] == pos[1] {
		// The cursor is in the first cell.
		t.cursor[0], t.cursor[1], t.cursor[4] = line, columnOffset, pos[2]
		cursorVisible = true
	}
	for pos[0] != 1 {
		cluster, text, _, pos, endPos = t.step(text, pos, endPos)
		clusterWidth := stringWidth(cluster)
		runes := []rune(cluster)
		if posX+clusterWidth-columnOffset <= width && posX-columnOffset >= 0 && clusterWidth > 0 {
			screen.SetContent(x+posX-columnOffset, y+posY, runes[0], runes[1:], t.textStyle)
		}
		posX += clusterWidth
		if line+1 < len(t.lineStarts) && t.lineStarts[line+1] == pos {
			// We must break over.
			posY++
			if posY >= height {
				break // Don&& x+posX->=olumnOffset<0
			}
			posX = 0
			line++
		}
		if t.cursor[0] < 0 && t.cursor[2] == pos[0] && t.cursor[3] == pos[1] {
			// We hit the cursor.
			t.cursor[0], t.cursor[1], t.cursor[4] = posY+t.rowOffset, posX, pos[2]
			cursorVisible = t.cursor[0]-t.rowOffset < height && t.cursor[1]-columnOffset < width
		}
	}

	if t.HasFocus() && cursorVisible {
		screen.ShowCursor(x+t.cursor[1]-columnOffset, y+t.cursor[0]-t.rowOffset)
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
			if pos[0] != 1 && boundaries&uniseg.MaskLine == uniseg.LineMustBreak {
				// We must break over.
				t.lineStarts = append(t.lineStarts, pos)
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
					lineWidth = clusterWidth
					lastLineBreak = [3]int{}
				}
			} else { // t.wordWrap && lastLineBreak != [3]int{}
				// Break after last line break opportunity.
				t.lineStarts = append(t.lineStarts, lastLineBreak)
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
