package tview

import "github.com/gdamore/tcell/v2"

var (
	// NewLine is the string sequence to be inserted when hitting the Enter key
	// in a TextArea. The default is "\n" but you may change it to "\r\n" if
	// required.
	NewLine = "\n"
)

// TextArea implements a simple text editor for multi-line text. Multi-color
// text is not supported. Text can be optionally word-wrapped to fit the
// available width.
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
// If the mouse is enabled, clicking on a screen cell will move the cursor to
// that location or to the end of the line if past the last character. Turning
// the scroll wheel will scroll the text. Text can also be selected by moving
// the mouse while pressing the left mouse button (see below for details).
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
// Text can be selected by moving the cursor while holding the Shift key or
// dragging the mouse. When text is selected:
//
//   - Entering a character (rune) will replace the selected text with the new
//     character. (The Enter key is an exception, see further below.)
//   - Backspace, delete: Delete the selected text.
//   - Enter: Copy the selected text into the clipboard, unselect the text.
//   - Ctrl-X: Copy the selected text into the clipboard and delete it.
//   - Ctrl-V: Replace the selected text with the clipboard text. If no text is
//     selected, the clipboard text will be inserted at the cursor location.
//
// The default clipboard is an internal text buffer, i.e. the operating system's
// clipboard is not used. The Enter key was chosen for the "copy" function
// because the Ctrl-C key is the default key to stop the application. If your
// application frees up the global Ctrl-C key and you want to bind it to the
// "copy to clipboard" function, you may use SetInputCapture() to override the
// Enter/Ctrl-C keys to implement copying to the clipboard.
//
// Similarly, if you want to implement your own clipboard (or make use of your
// operating system's clipboard), you can also use SetInputCapture() to override
// the key binds for copy, cut, and paste. The GetSelection(), ReplaceText(),
// and SetSelection() provide all the functionality needed for your own
// clipboard.
//
//    - Ctrl-Z: Undo the last change.
//    - Ctrl-Y: Redo the last change.
type TextArea struct {
	*Box

	// The text to be shown in the text area when it is empty.
	placeholder string

	// If set to true, lines that are longer than the available width are
	// wrapped onto the next line. If set to false, any characters beyond the
	// available width are discarded.
	wrap bool

	// If set to true and if wrap is also true, lines are split at spaces or
	// after punctuation characters.
	wordWrap bool

	// The maximum number of bytes allowed in the text area. If 0, there is no
	// limit.
	maxLength int

	// The index of the first line shown in the text area.
	lineOffset int

	// The number of cells to be skipped on each line (not used in wrap mode).
	columnOffset int

	// The height of the content the last time the text area was drawn.
	pageSize int

	// The style of the text. Background colors different from the Box's
	// background color may lead to unwanted artefacts.
	textStyle tcell.Style
}

// NewTextArea returns a new text area.
func NewTextArea() *TextArea {
	return &TextArea{
		Box:       NewBox(),
		wrap:      true,
		wordWrap:  true,
		textStyle: tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.PrimaryTextColor),
	}
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

// SetSelection selects the text starting at index "start" and ending just
// before the index "end". Any previous selection is discarded. If "start" and
// "end" are the same, currently selected text is unselected.
func (t *TextArea) SetSelection(start, end int) *TextArea {
	//TODO
	return t
}

// GetSelection returns the currently selected text or an empty string if no
// text is currently selected. The start and end indices (a half-open range)
// into the text area's text are also returned.
func (t *TextArea) GetSelection() (string, int, int) {
	return "", 0, 0 //TODO
}

// ReplaceText replaces the text in the given range with the given text. The
// range is half-open, that is, the character at the "end" index is not
// replaced. If the provided range overlaps with a selection, the selected text
// will be unselected.
func (t *TextArea) ReplaceText(start, end int, text string) *TextArea {
	//TODO
	return t
}
