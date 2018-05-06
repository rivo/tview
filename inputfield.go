package tview

import (
	"fmt"
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

	align int

	labelFiller string

	disable bool

	lockColors bool

	// The text that was entered.
	text string

	// The text to be displayed before the input area.
	label string

	// The text to be displayed before the input area.
	subLabel string

	// The item sub label color.
	subLabelColor tcell.Color

	// The text to be displayed in the input area when "text" is empty.
	placeholder string

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// The background color of the input area.
	fieldDisableBackgroundColor tcell.Color

	// The text color of the input area.
	fieldDisableTextColor tcell.Color

	// The text color of the placeholder.
	placeholderTextColor tcell.Color

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	cursorPosition int

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

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
}

// NewInputField returns a new input field.
func NewInputField() *InputField {
	input := &InputField{
		Box:                         NewBox(),
		labelColor:                  Styles.LabelTextColor,
		subLabelColor:               Styles.LabelTextColor,
		fieldBackgroundColor:        Styles.FieldBackgroundColor,
		fieldTextColor:              Styles.FieldTextColor,
		placeholderTextColor:        Styles.ContrastSecondaryTextColor,
		fieldDisableBackgroundColor: Styles.FieldDisableBackgroundColor,
		fieldDisableTextColor:       Styles.FieldDisableTextColor,
		align:       AlignLeft,
		labelFiller: " ",
	}
	input.height = 1

	return input
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

// SetLockColors locks the change of colors by form
func (i *InputField) SetLockColors(lock bool) *InputField {
	i.lockColors = lock
	return i
}

// SetLabel sets the text to be displayed before the input area.
func (i *InputField) SetLabel(label string) *InputField {
	if !strings.Contains(label, "%s") {
		label += "%s"
	}
	i.label = label
	return i
}

// GetLabel returns the text to be displayed before the input area.
func (i *InputField) GetLabel() string {
	return i.label
}

// GetLabelWidth returns label width.
func (i *InputField) GetLabelWidth() int {
	return StringWidth(strings.Replace(i.subLabel+i.label, "%s", "", -1))
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

// SetSubLabel sets the text to be displayed before the input area.
func (i *InputField) SetSubLabel(label string) *InputField {
	i.subLabel = label
	return i
}

// SetSubLabelColor sets the color of the subLabel.
func (i *InputField) SetSubLabelColor(color tcell.Color) *InputField {
	i.subLabelColor = color
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
func (i *InputField) SetFormAttributes(labelWidth, fieldWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	if i.fieldWidth == 0 {
		i.fieldWidth = fieldWidth
	}
	if i.labelWidth == 0 {
		i.labelWidth = labelWidth
	}
	if !i.lockColors {
		i.labelColor = labelColor
		i.backgroundColor = bgColor
		i.fieldTextColor = fieldTextColor
		i.fieldBackgroundColor = fieldBgColor
	}
	return i
}

// SetFieldAlign sets the input alignment within the radiobutton box. This must be
// either AlignLeft, AlignCenter, or AlignRight.
func (i *InputField) SetFieldAlign(align int) FormItem {
	i.align = align
	return i
}

// GetFieldAlign returns the input alignment within the radiobutton box.
func (i *InputField) GetFieldAlign() (align int) {
	return i.align
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

// SetDisable sets an input field like disabled
func (i *InputField) SetDisable(disable bool) *InputField {
	i.disable = disable
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

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (i *InputField) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	i.finished = handler
	return i
}

// GetFinishedFunc returns SetDoneFunc().
func (i *InputField) GetFinishedFunc() func(key tcell.Key) {
	return i.finished
}

// Draw draws this primitive onto the screen.
func (i *InputField) Draw(screen tcell.Screen) {
	i.Box.Draw(screen)

	fieldBackgroundColor := i.fieldBackgroundColor
	fieldTextColor := i.fieldTextColor

	if i.disable {
		fieldBackgroundColor = i.fieldDisableBackgroundColor
		fieldTextColor = i.fieldDisableTextColor
	}

	// Prepare
	x, y, width, height := i.GetInnerRect()
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}

	//fmt.Println("-", i.label, i.labelWidth, i.fieldWidth)

	// Draw label.
	var labels = []struct {
		text  string
		color tcell.Color
	}{{
		text:  i.subLabel,
		color: i.subLabelColor,
	}, {
		text:  i.label,
		color: i.labelColor,
	}}

	if len(labels) > 0 {
		labelWidth := i.labelWidth
		if labelWidth > rightLimit-x {
			labelWidth = rightLimit - x
		}

		addCount := labelWidth - i.GetLabelWidth()

		for _, label := range labels {
			if addCount > 0 && strings.Contains(label.text, "%s") {
				label.text = fmt.Sprintf(label.text, strings.Repeat(i.labelFiller, addCount))
				addCount = 0
			} else {
				label.text = strings.Replace(label.text, "%s", "", -1)
			}

			labelWidth = StringWidth(label.text)
			Print(screen, label.text, x, y, labelWidth, AlignLeft, label.color)
			x += labelWidth
		}
		x++
	}

	// Draw input area.
	fieldWidth := i.fieldWidth
	if fieldWidth == 0 {
		fieldWidth = math.MaxInt32
	}
	if rightLimit-x < fieldWidth {
		fieldWidth = rightLimit - x
	}
	fieldStyle := tcell.StyleDefault.Background(fieldBackgroundColor)
	for index := 0; index < fieldWidth; index++ {
		screen.SetContent(x+index, y, ' ', nil, fieldStyle)
	}

	// Draw placeholder text.
	text := i.text
	if text == "" && i.placeholder != "" {
		Print(screen, i.placeholder, x, y, fieldWidth, AlignLeft, i.placeholderTextColor)
	}

	textWidth := runewidth.StringWidth(text)
	// Draw entered text.
	if i.maskCharacter > 0 {
		text = strings.Repeat(string(i.maskCharacter), utf8.RuneCountInString(i.text))
	}
	if i.cursorPosition < 0 {
		i.cursorPosition = 0
	}
	if i.cursorPosition > textWidth {
		i.cursorPosition = textWidth
	}

	fieldWidth-- // We need one cell for the cursor.

	if fieldWidth < textWidth {
		runes := []rune(text)
		p := len(runes)
		if i.cursorPosition-1 < textWidth-fieldWidth {
			p = i.cursorPosition + fieldWidth
		}
		for pos := p - 1; pos >= 0; pos-- {
			ch := runes[pos]
			w := runewidth.RuneWidth(ch)
			if fieldWidth-w < 0 {
				break
			}
			_, _, style, _ := screen.GetContent(x+fieldWidth-w, y)
			style = style.Foreground(fieldTextColor)
			for w > 0 {
				fieldWidth--
				screen.SetContent(x+fieldWidth, y, ch, nil, style)
				w--
			}
		}
	} else {
		pos := 0
		for _, ch := range text {
			w := runewidth.RuneWidth(ch)
			_, _, style, _ := screen.GetContent(x+pos, y)
			style = style.Foreground(fieldTextColor)
			for w > 0 {
				screen.SetContent(x+pos, y, ch, nil, style)
				pos++
				w--
			}
		}
	}

	// Set cursor.
	if !i.disable && i.focus.HasFocus() {
		i.setCursor(screen)
	}
}

// setCursor sets the cursor position.
func (i *InputField) setCursor(screen tcell.Screen) {
	x := i.x + i.paddingLeft
	y := i.y + i.paddingTop
	rightLimit := x + i.width
	if i.border {
		x++
		y++
		rightLimit -= 2
	}

	cursorIndent := i.cursorPosition

	textWidth := runewidth.StringWidth(i.text)
	if textWidth > i.fieldWidth {
		overflow := textWidth - (i.fieldWidth - 2)
		if overflow > textWidth-cursorIndent {

		}
		cursorIndent -= overflow
	}
	if cursorIndent < 0 {
		cursorIndent = 0
	}

	if i.fieldWidth > 0 && cursorIndent > i.fieldWidth-1 {
		cursorIndent = i.fieldWidth - 1
	}
	if i.labelWidth > 0 {
		x += i.labelWidth + 1 + cursorIndent
	} else {
		x += StringWidth(i.subLabel+i.label) + 1 + cursorIndent
	}
	if x >= rightLimit {
		x = rightLimit - 1
	}
	screen.ShowCursor(x, y)
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
			newText := i.text[0:i.cursorPosition] + string(event.Rune()) + i.text[i.cursorPosition:]
			if i.accept != nil {
				if !i.accept(newText, event.Rune()) {
					break
				}
			}
			i.cursorPosition++
			i.text = newText
		case tcell.KeyCtrlU: // Delete all.
			i.text = ""
		case tcell.KeyCtrlW: // Delete last word.
			lastWord := regexp.MustCompile(`\s*\S+\s*$`)
			i.text = lastWord.ReplaceAllString(i.text, "")
		case tcell.KeyBackspace, tcell.KeyBackspace2: // Delete last character.
			if len(i.text) == 0 || i.cursorPosition < 1 {
				break
			}
			newText := i.text[0:i.cursorPosition-1] + i.text[i.cursorPosition:]
			runes := []rune(i.text)
			if i.accept != nil {
				if !i.accept(newText, runes[i.cursorPosition-1]) {
					break
				}
			}
			i.cursorPosition--
			i.text = newText
		case tcell.KeyDelete:
			if len(i.text) == 0 || i.cursorPosition < 1 || len(i.text) == i.cursorPosition {
				break
			}
			newText := i.text[0:i.cursorPosition] + i.text[i.cursorPosition+1:]
			runes := []rune(i.text)
			if i.accept != nil {
				if !i.accept(newText, runes[i.cursorPosition]) {
					break
				}
			}
			i.text = newText
		case tcell.KeyEnter, tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEscape: // We're done.
			if i.done != nil {
				i.done(key)
			}
			if i.finished != nil {
				i.finished(key)
			}
		case tcell.KeyLeft:
			i.cursorPosition--
		case tcell.KeyRight:
			i.cursorPosition++
		case tcell.KeyHome:
			i.cursorPosition = 0
		case tcell.KeyEnd:
			i.cursorPosition = runewidth.StringWidth(i.text)
		}
	})
}

// Focus is called when this primitive receives focus.
func (i *InputField) Focus(delegate func(p Primitive)) {
	if i.disable && i.finished != nil {
		i.finished(tcell.KeyTAB)
		return
	}
	i.hasFocus = true
}
