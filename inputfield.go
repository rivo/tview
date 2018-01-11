package tview

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
)

// InputField is a one-line box (three lines if there is a title) where the
// user can enter text.
//
// Use SetMaskCharacter() to hide input from onlookers (e.g. for password
// input).
//
// See https://github.com/rivo/tview/wiki/InputField for an example.
type InputField struct {
	*Box

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

	// A character to mask entered text (useful for password fields). A value of 0
	// disables masking.
	maskCharacter rune

	// An optional function which may reject the last character that was entered.
	accept func(text string, ch rune) bool

	// An optional function which is called when the input has changed.
	changed func(text string)

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, enter, or escape).
	done func(tcell.Key)
}

// NewInputField returns a new input field.
func NewInputField() *InputField {
	return &InputField{
		Box:                  NewBox(),
		labelColor:           Styles.SecondaryTextColor,
		fieldBackgroundColor: Styles.ContrastBackgroundColor,
		fieldTextColor:       Styles.PrimaryTextColor,
	}
}

// SetText sets the current text of the input field.
func (i *InputField) SetText(text string) *InputField {
	i.text = text
	if i.changed != nil {
		i.changed(text)
	}
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

// SetFormAttributes sets attributes shared by all form items.
func (i *InputField) SetFormAttributes(label string, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	i.label = label
	i.labelColor = labelColor
	i.backgroundColor = bgColor
	i.fieldTextColor = fieldTextColor
	i.fieldBackgroundColor = fieldBgColor
	return i
}

// SetFieldLength sets the length of the input area. A value of 0 means extend
// as much as possible.
func (i *InputField) SetFieldLength(length int) *InputField {
	i.fieldLength = length
	return i
}

// SetMaskCharacter sets a character that masks user input on a screen. A value
// of 0 disables masking.
func (i *InputField) SetMaskCharacter(mask rune) *InputField {
	i.maskCharacter = mask
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

// SetChangedFunc sets a handler which is called whenever the text of the input
// field has changed. It receives the current text (after the change).
func (i *InputField) SetChangedFunc(handler func(text string)) *InputField {
	i.changed = handler
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

// SetFinishedFunc calls SetDoneFunc().
func (i *InputField) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	return i.SetDoneFunc(handler)
}

// Draw draws this primitive onto the screen.
func (i *InputField) Draw(screen tcell.Screen) {
	i.Box.Draw(screen)

	// Prepare
	x, y, width, height := i.GetInnerRect()
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}

	// Draw label.
	_, drawnWidth := Print(screen, i.label, x, y, rightLimit-x, AlignLeft, i.labelColor)
	x += drawnWidth

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
	text := i.text
	if i.maskCharacter > 0 {
		text = strings.Repeat(string(i.maskCharacter), utf8.RuneCountInString(i.text))
	}
	fieldLength-- // We need one cell for the cursor.
	if fieldLength < runewidth.StringWidth(i.text) {
		Print(screen, text, x, y, fieldLength, AlignRight, i.fieldTextColor)
	} else {
		Print(screen, text, x, y, fieldLength, AlignLeft, i.fieldTextColor)
	}

	// Set cursor.
	if i.focus.HasFocus() {
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
	fieldLength := runewidth.StringWidth(i.text)
	if i.fieldLength > 0 && fieldLength > i.fieldLength-1 {
		fieldLength = i.fieldLength - 1
	}
	x += runewidth.StringWidth(i.label) + fieldLength
	if x >= rightLimit {
		x = rightLimit - 1
	}
	screen.ShowCursor(x, y)
}

// InputHandler returns the handler for this primitive.
func (i *InputField) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p Primitive)) {
		// Trigger changed events.
		currentText := i.text
		defer func() {
			if i.text != currentText && i.changed != nil {
				i.changed(i.text)
			}
		}()

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
			if len(i.text) == 0 {
				break
			}
			runes := []rune(i.text)
			i.text = string(runes[:len(runes)-1])
		case tcell.KeyEnter, tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEscape: // We're done.
			if i.done != nil {
				i.done(key)
			}
		}
	}
}
