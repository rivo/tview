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

	// The text to be displayed in the input area when "text" is empty.
	placeholder string

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// The text color of the placeholder.
	placeholderTextColor tcell.Color

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	// A character to mask entered text (useful for password fields). A value of 0
	// disables masking.
	maskCharacter rune

	// The current cursor index within the string.
	cursor int

	// An optional function which may reject the last character that was entered.
	accept func(text string, ch rune) bool

	// An optional function which is called when the input has changed.
	changed func(text string)

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, enter, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
}

// NewInputField returns a new input field.
func NewInputField() *InputField {
	return &InputField{
		Box:                  NewBox(),
		labelColor:           Styles.SecondaryTextColor,
		fieldBackgroundColor: Styles.ContrastBackgroundColor,
		fieldTextColor:       Styles.PrimaryTextColor,
		placeholderTextColor: Styles.ContrastSecondaryTextColor,
	}
}

// SetText sets the current text of the input field.
func (i *InputField) SetText(text string) *InputField {
	i.text = text
	if i.changed != nil {
		i.changed(text)
	}
	i.cursor = len(i.text)
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

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (i *InputField) SetLabelWidth(width int) *InputField {
	i.labelWidth = width
	return i
}

// SetPlaceholder sets the text to be displayed when the input text is empty.
func (i *InputField) SetPlaceholder(text string) *InputField {
	i.placeholder = text
	return i
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

// SetPlaceholderTextColor sets the text color of placeholder text.
func (i *InputField) SetPlaceholderTextColor(color tcell.Color) *InputField {
	i.placeholderTextColor = color
	return i
}

// SetFormAttributes sets attributes shared by all form items.
func (i *InputField) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	i.labelWidth = labelWidth
	i.labelColor = labelColor
	i.backgroundColor = bgColor
	i.fieldTextColor = fieldTextColor
	i.fieldBackgroundColor = fieldBgColor
	return i
}

// SetFieldWidth sets the screen width of the input area. A value of 0 means
// extend as much as possible.
func (i *InputField) SetFieldWidth(width int) *InputField {
	i.fieldWidth = width
	return i
}

// GetFieldWidth returns this primitive's field width.
func (i *InputField) GetFieldWidth() int {
	return i.fieldWidth
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

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (i *InputField) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	i.finished = handler
	return i
}

// GetCursorPosition returns the current cursor's index of the input's text.
func (i *InputField) GetCursorPosition() int {
	return i.cursor
}

// SetCursorPosition sets the current cursor position within the input's text,
// clamping the values to the size of the input's text.
func (i *InputField) SetCursorPosition(position int) {
	i.cursor = clamp(position, 0, len(i.text))
}

func (i *InputField) getFieldText(text string, fieldWidth int) (string, int) {
	bufferSize := fieldWidth / 2
	left := clamp(i.cursor-bufferSize, 0, len(text))
	right := clamp(i.cursor+bufferSize, 0, len(text))
	return i.text[left:right], left
}

func (i *InputField) getFieldCursorPosition(fieldText string, leftOffset int) int {
	return clamp(i.cursor-leftOffset, 0, len(fieldText))
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
	if i.labelWidth > 0 {
		labelWidth := i.labelWidth
		if labelWidth > rightLimit-x {
			labelWidth = rightLimit - x
		}
		Print(screen, i.label, x, y, labelWidth, AlignLeft, i.labelColor)
		x += labelWidth
	} else {
		_, drawnWidth := Print(screen, i.label, x, y, rightLimit-x, AlignLeft, i.labelColor)
		x += drawnWidth
	}

	// Draw input area.
	fieldWidth := i.fieldWidth
	if fieldWidth == 0 {
		fieldWidth = math.MaxInt32
	}
	if rightLimit-x < fieldWidth {
		fieldWidth = rightLimit - x
	}
	fieldStyle := tcell.StyleDefault.Background(i.fieldBackgroundColor)
	for index := 0; index < fieldWidth; index++ {
		screen.SetContent(x+index, y, ' ', nil, fieldStyle)
	}

	// Draw placeholder text.
	text := i.text
	if text == "" && i.placeholder != "" {
		Print(screen, i.placeholder, x, y, fieldWidth, AlignLeft, i.placeholderTextColor)
	} else if len(text) > 0 {
		// Draw entered text.
		if i.maskCharacter > 0 {
			text = strings.Repeat(string(i.maskCharacter), utf8.RuneCountInString(i.text))
		} else {
			text = Escape(text)
		}

		// Clip the input's visible text to that which is around the cursor's position.
		fieldText, offset := i.getFieldText(text, i.Box.width)

		// Invert the color of the cursor's cell, use a space if on the "new character" cell
		cursorPos := i.getFieldCursorPosition(fieldText, offset)
		cursorChar := " "
		if len(fieldText) > 0 && cursorPos < len(fieldText) {
			cursorChar = string(fieldText[cursorPos])
		}

		alignment := AlignLeft
		if fieldWidth < runewidth.StringWidth(text) {
			alignment = AlignRight
		}

		Print(screen, fieldText, x, y, fieldWidth, alignment, i.fieldTextColor)

		if i.focus.HasFocus() {
			printWithStyle(screen, cursorChar, x+cursorPos, y, 1, alignment, tcell.StyleDefault.
				Foreground(i.fieldBackgroundColor).
				Background(i.fieldTextColor))
		}
	}
}

// InputHandler returns the handler for this primitive.
func (i *InputField) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return i.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
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
			left := ""
			right := ""
			if i.cursor > 0 {
				left = i.text[:i.cursor]
			}
			if i.cursor < len(i.text) {
				right = i.text[i.cursor:len(i.text)]
			}
			newText := left + string(event.Rune()) + right
			if i.accept != nil {
				if !i.accept(newText, event.Rune()) {
					break
				}
			}
			i.text = newText
			i.cursor++
		case tcell.KeyLeft:
			i.cursor = clamp(i.cursor-1, 0, len(i.text))
		case tcell.KeyRight:
			i.cursor = clamp(i.cursor+1, 0, len(i.text))
		case tcell.KeyCtrlU: // Delete all.
			i.text = ""
			i.cursor = 0
		case tcell.KeyCtrlW: // Delete last word.
			lastWord := regexp.MustCompile(`\s*\S+\s*$`)
			i.text = lastWord.ReplaceAllString(i.text, "")
			i.cursor = len(i.text)
		case tcell.KeyBackspace, tcell.KeyBackspace2: // Delete last character.
			if len(i.text) == 0 {
				break
			}
			leftLimit := clamp(i.cursor-1, 0, len(i.text))
			i.text = i.text[0:leftLimit] + i.text[i.cursor:len(i.text)]
			i.cursor = clamp(i.cursor-1, 0, len(i.text))
		case tcell.KeyEnter, tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEscape: // We're done.
			if i.done != nil {
				i.done(key)
			}
			if i.finished != nil {
				i.finished(key)
			}
		}
	})
}
