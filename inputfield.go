package tview

import (
	"math"
	"regexp"
	"strconv"

	"github.com/gdamore/tcell"
)

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
	// Initialize the predefined handlers.

	InputFieldInteger = func(text string, ch rune) bool {
		if text == "-" {
			return true
		}
		_, err := strconv.Atoi(text)
		return err == nil
	}

	InputFieldFloat = func(text string, ch rune) bool {
		if text == "-" || text == "." {
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

// InputField is a one-line box (three lines if there is a title) where the
// user can enter text.
type InputField struct {
	Box

	// The text that was entered.
	text string

	// The text to be displayed before the input area.
	label string

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// The length of the input area. A value of 0 means extend as much as
	// possible.
	fieldLength int

	// An optional function which may reject the last character that was entered.
	accept func(text string, ch rune) bool

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, enter, or escape).
	done func(tcell.Key)
}

// NewInputField returns a new input field.
func NewInputField() *InputField {
	return &InputField{
		Box:                  *NewBox(),
		labelColor:           tcell.ColorYellow,
		fieldBackgroundColor: tcell.ColorBlue,
		fieldTextColor:       tcell.ColorWhite,
	}
}

// SetText sets the current text of the input field.
func (i *InputField) SetText(text string) *InputField {
	i.text = text
	return i
}

// GetText returns the current text of the input field.
func (i *InputField) GetText() string {
	return i.text
}

// SetLabel sets the text to be displayed before the input area.
func (i *InputField) SetLabel(label string) *InputField {
	i.label = label
	return i
}

// GetLabel returns the text to be displayed before the input area.
func (i *InputField) GetLabel() string {
	return i.label
}

// SetLabelColor sets the color of the label.
func (i *InputField) SetLabelColor(color tcell.Color) *InputField {
	i.labelColor = color
	return i
}

// SetFieldBackgroundColor sets the background color of the input area.
func (i *InputField) SetFieldBackgroundColor(color tcell.Color) *InputField {
	i.fieldBackgroundColor = color
	return i
}

// SetFieldTextColor sets the text color of the input area.
func (i *InputField) SetFieldTextColor(color tcell.Color) *InputField {
	i.fieldTextColor = color
	return i
}

// SetFieldLength sets the length of the input area. A value of 0 means extend
// as much as possible.
func (i *InputField) SetFieldLength(length int) *InputField {
	i.fieldLength = length
	return i
}

// SetAcceptanceFunc sets a handler which may reject the last character that was
// entered (by returning false).
//
// This package defines a number of variables Prefixed with InputField which may
// be used for common input (e.g. numbers, maximum text length).
func (i *InputField) SetAcceptanceFunc(handler func(textToCheck string, lastChar rune) bool) *InputField {
	i.accept = handler
	return i
}

// SetDoneFunc sets a handler which is called when the user is done entering
// text. The callback function is provided with the key that was pressed, which
// is one of the following:
//
//   - KeyEnter: Done entering text.
//   - KeyEscape: Abort text input.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (i *InputField) SetDoneFunc(handler func(key tcell.Key)) *InputField {
	i.done = handler
	return i
}

// Draw draws this primitive onto the screen.
func (i *InputField) Draw(screen tcell.Screen) {
	i.Box.Draw(screen)

	// Prepare
	x := i.x
	y := i.y
	rightLimit := x + i.width
	height := i.height
	if i.border {
		x++
		y++
		rightLimit -= 2
		height -= 2
	}
	if height < 1 || rightLimit <= x {
		return
	}

	// Draw label.
	x += Print(screen, i.label, x, y, rightLimit-x, AlignLeft, i.labelColor)

	// Draw input area.
	fieldLength := i.fieldLength
	if fieldLength == 0 {
		fieldLength = math.MaxInt64
	}
	if rightLimit-x < fieldLength {
		fieldLength = rightLimit - x
	}
	fieldStyle := tcell.StyleDefault.Background(i.fieldBackgroundColor)
	for index := 0; index < fieldLength; index++ {
		screen.SetContent(x+index, y, ' ', nil, fieldStyle)
	}

	// Draw entered text.
	fieldLength-- // We need one cell for the cursor.
	if fieldLength < len([]rune(i.text)) {
		Print(screen, i.text, x+fieldLength-1, y, fieldLength, AlignRight, i.fieldTextColor)
	} else {
		Print(screen, i.text, x, y, fieldLength, AlignLeft, i.fieldTextColor)
	}

	// Set cursor.
	if i.hasFocus {
		i.setCursor(screen)
	}
}

// setCursor sets the cursor position.
func (i *InputField) setCursor(screen tcell.Screen) {
	x := i.x
	y := i.y
	rightLimit := x + i.width
	if i.border {
		x++
		y++
		rightLimit -= 2
	}
	fieldLength := len([]rune(i.text))
	if fieldLength > i.fieldLength-1 {
		fieldLength = i.fieldLength - 1
	}
	x += len([]rune(i.label)) + fieldLength
	if x >= rightLimit {
		x = rightLimit - 1
	}
	screen.ShowCursor(x, y)
}

// InputHandler returns the handler for this primitive.
func (i *InputField) InputHandler() func(event *tcell.EventKey) {
	return func(event *tcell.EventKey) {
		// Process key event.
		switch key := event.Key(); key {
		case tcell.KeyRune: // Regular character.
			newText := i.text + string(event.Rune())
			if i.accept != nil {
				if !i.accept(newText, event.Rune()) {
					break
				}
			}
			i.text = newText
		case tcell.KeyCtrlU: // Delete all.
			i.text = ""
		case tcell.KeyCtrlW: // Delete last word.
			lastWord := regexp.MustCompile(`\s*\S+\s*$`)
			i.text = lastWord.ReplaceAllString(i.text, "")
		case tcell.KeyBackspace, tcell.KeyBackspace2: // Delete last character.
			if len([]rune(i.text)) == 0 {
				break
			}
			i.text = i.text[:len([]rune(i.text))-1]
		case tcell.KeyEnter, tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEscape: // We're done.
			if i.done != nil {
				i.done(key)
			}
		}
	}
}
