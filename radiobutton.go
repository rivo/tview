//
// Copyright (c) 2018 Litmus Automation Inc.
// Author: Levko Burburas <levko.burburas.external@litmus.cloud>
//

package tview

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell"
)

// RadioOption holds key and title for radio option
type RadioOption struct {
	Name  string
	Title string
}

// NewRadioOption returns a new option for RadioOption
func NewRadioOption(name, text string) *RadioOption {
	return &RadioOption{
		Name:  name,
		Title: text,
	}
}

// RadioButtons implements a simple primitive for radio button selections.
type RadioButtons struct {
	*Box
	options        []*RadioOption
	currentOption  int
	selectedOption int
	itemPadding    int
	align          int

	lockColors bool

	// The text to be displayed before the input area.
	subLabel string

	// The item sub label color.
	subLabelColor tcell.Color

	// The text to be displayed before the input area.
	label string

	labelFiller string

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// The item main text color.
	mainTextColor tcell.Color

	// The item secondary text color.
	secondaryTextColor tcell.Color

	// The text color for selected items.
	selectedTextColor tcell.Color

	// The background color for selected items.
	selectedBackgroundColor tcell.Color

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	labelWidth int

	joinElements   []*RadioButtons
	currentElement int
	// An optional function which is called when the user indicated that they
	// are done selecting options. The key which was pressed is provided (tab,
	// shift-tab, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)

	// If set to true, instead of position items and buttons from top to bottom,
	// they are positioned from left to right.
	horizontal bool

	horizontalSeparator string

	// An optional function which is called when the user has navigated to a list
	// item.
	changed func(*RadioOption)

	inputHandler func() func(event *tcell.EventKey, setFocus func(p Primitive))
}

// NewRadioButtons returns a new radio button primitive.
func NewRadioButtons() *RadioButtons {
	r := &RadioButtons{
		Box:                     NewBox(),
		labelColor:              tcell.ColorWhite,
		fieldBackgroundColor:    tcell.ColorBlack,
		fieldTextColor:          tcell.ColorWhite,
		mainTextColor:           Styles.PrimaryTextColor,
		secondaryTextColor:      Styles.TertiaryTextColor,
		selectedTextColor:       Styles.PrimitiveBackgroundColor,
		selectedBackgroundColor: Styles.PrimaryTextColor,
		align:       AlignLeft,
		labelFiller: " ",
	}

	r.focus = r
	r.joinElements = append(r.joinElements, r)
	r.inputHandler = r.defaultInputHandler

	return r
}

// Join combines two element the same type in one, for navigation
func (r *RadioButtons) Join(elements ...*RadioButtons) *RadioButtons {
	for i := 0; i < len(elements); i++ {
		elements[i].inputHandler = r.inputHandler
		elements[i].id = r.id
		elements[i].currentOption = -1
	}
	r.joinElements = append(r.joinElements, elements...)
	return r
}

// SetOptions replaces all current options with the ones provided
func (r *RadioButtons) SetOptions(options []*RadioOption) *RadioButtons {
	r.options = options
	return r
}

// SetItemPadding sets the number of empty rows between form items for vertical
// layouts and the number of empty cells between form items for horizontal
// layouts.
func (r *RadioButtons) SetItemPadding(padding int) *RadioButtons {
	r.itemPadding = padding
	return r
}

// SetAlign sets the radiobox alignment within the box. This must be
// either AlignLeft, AlignCenter, or AlignRight.
func (r *RadioButtons) SetAlign(align int) *RadioButtons {
	r.align = align
	return r
}

// Focus is called when this primitive receives focus.
func (r *RadioButtons) Focus(delegate func(p Primitive)) {
	if r != r.joinElements[r.currentElement] {
		delegate(r.joinElements[r.currentElement])
		return
	}
	r.hasFocus = true
}

// InputHandler returns the handler for this primitive.
func (r *RadioButtons) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	if r.inputHandler != nil {
		return r.inputHandler()
	}
	return r.Box.InputHandler()
}

// InputHandler returns the handler for this primitive.
func (r *RadioButtons) defaultInputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		parent := r
		r = parent.joinElements[parent.currentElement]

		switch key := event.Key(); key {
		case tcell.KeyUp, tcell.KeyLeft:
			r.currentOption--
			if r.currentOption < 0 {
				if len(parent.joinElements) > 1 {
					parent.currentElement--
					if parent.currentElement < 0 {
						parent.currentElement = len(parent.joinElements) - 1
					}
					nextElement := parent.joinElements[parent.currentElement]
					setFocus(nextElement)
					nextElement.currentOption = len(nextElement.options) - 1
					break
				}
				r.currentOption = len(r.options) - 1 - int(math.Mod(float64(len(r.options)), float64(r.currentOption)))
			}
			if r.changed != nil {
				r.changed(r.options[r.currentOption])
			}
		case tcell.KeyDown, tcell.KeyRight:
			r.currentOption++
			if r.currentOption >= len(r.options) {
				if len(parent.joinElements) > 1 {
					parent.currentElement++
					if parent.currentElement >= len(parent.joinElements) {
						parent.currentElement = 0
					}
					nextElement := parent.joinElements[parent.currentElement]
					setFocus(nextElement)
					nextElement.currentOption = 0
					break
				}
				r.currentOption = int(math.Mod(float64(len(r.options)), float64(r.currentOption)))
			}
			if r.changed != nil {
				r.changed(r.options[r.currentOption])
			}
		case tcell.KeyEnter, tcell.KeyRune: // We're done.
			for index, element := range parent.joinElements {
				if parent.currentElement != index {
					element.currentOption = -1
				}
			}
			if r.changed != nil {
				r.changed(r.options[r.currentOption])
			}
		case tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEscape: // We're done.
			if parent.done != nil {
				parent.done(key)
			}
			if parent.finished != nil {
				parent.finished(key)
			}
		}
	})
}

// GetRect returns the current position of the rectangle, x, y, width, and
// height.
func (r *RadioButtons) GetRect() (int, int, int, int) {
	x, y, width, _ := r.Box.GetRect()
	optionsCount := len(r.options)
	if optionsCount == 0 {
		optionsCount = 1
	}
	height := 1
	if !r.horizontal {
		height = (optionsCount * (r.itemPadding + 1)) - r.itemPadding
	}
	if height > 0 && r.Box.GetBorder() {
		height++
	}
	return x, y, width, height
}

// GetLabelWidth returns label width.
func (r *RadioButtons) GetLabelWidth() int {
	return StringWidth(strings.Replace(r.subLabel+r.label, "%s", "", -1))
}

// GetFieldWidth returns field width.
func (r *RadioButtons) GetFieldWidth() int {
	if r.fieldWidth > 0 {
		return r.fieldWidth
	}

	var maxWidth int
	for i := 0; i < len(r.options); i++ {
		line := fmt.Sprintf(`%s[white] %s`, Styles.GraphicsRadioUnchecked, r.options[i].Title)
		if r.horizontal {
			maxWidth += StringWidth(line)
			if i < len(r.options)-1 {
				maxWidth += r.itemPadding + 1
			}
		} else if maxWidth < StringWidth(line) {
			maxWidth = StringWidth(line)
		}
	}

	return maxWidth
}

// SetFieldWidth sets the screen width of the options area. A value of 0 means
// extend to as long as the longest option text.
func (r *RadioButtons) SetFieldWidth(width int) FormItem {
	r.fieldWidth = width
	return r
}

// SetFormAttributes sets attributes shared by all form items.
func (r *RadioButtons) SetFormAttributes(labelWidth, fieldWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	if r.fieldWidth == 0 {
		r.fieldWidth = fieldWidth
	}
	if r.labelWidth == 0 {
		r.labelWidth = labelWidth
	}
	if !r.lockColors {
		r.labelColor = labelColor
		r.backgroundColor = bgColor
		r.fieldTextColor = fieldTextColor
		r.fieldBackgroundColor = fieldBgColor
	}
	return r
}

// SetFieldAlign sets the input alignment within the radiobutton box. This must be
// either AlignLeft, AlignCenter, or AlignRight.
func (r *RadioButtons) SetFieldAlign(align int) FormItem {
	r.align = align
	return r
}

// SetLabelFiller sets a sign which will be fill the label when this one need to stretch
func (r *RadioButtons) SetLabelFiller(filler string) FormItem {
	r.labelFiller = filler
	return r
}

// GetFieldAlign returns the input alignment within the radiobutton box.
func (r *RadioButtons) GetFieldAlign() (align int) {
	return r.align
}

// SetDoneFunc sets a handler which is called when the user is done selecting
// options. The callback function is provided with the key that was pressed,
// which is one of the following:
//
//   - KeyEscape: Abort selection.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (r *RadioButtons) SetDoneFunc(handler func(key tcell.Key)) *RadioButtons {
	r.done = handler
	return r
}

// SetFinishedFunc calls SetDoneFunc().
func (r *RadioButtons) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	r.finished = handler
	return r
}

// GetFinishedFunc returns SetDoneFunc().
func (r *RadioButtons) GetFinishedFunc() func(key tcell.Key) {
	return r.finished
}

// SetLabel sets the text to be displayed before the input area.
func (r *RadioButtons) SetLabel(label string) *RadioButtons {
	if !strings.Contains(label, "%s") {
		label += "%s"
	}
	r.label = label
	return r
}

// GetLabel returns the text to be displayed before the input area.
func (r *RadioButtons) GetLabel() string {
	return r.label
}

// SetLockColors locks the change of colors by form
func (r *RadioButtons) SetLockColors(lock bool) *RadioButtons {
	r.lockColors = lock
	return r
}

// SetLabelColor sets the color of the label.
func (r *RadioButtons) SetLabelColor(color tcell.Color) *RadioButtons {
	r.labelColor = color
	return r
}

// SetSubLabel sets the text to be displayed before the input area.
func (r *RadioButtons) SetSubLabel(label string) *RadioButtons {
	r.subLabel = label
	return r
}

// SetSubLabelColor sets the color of the subLabel.
func (r *RadioButtons) SetSubLabelColor(color tcell.Color) *RadioButtons {
	r.subLabelColor = color
	return r
}

// SetFieldBackgroundColor sets the background color of the options area.
func (r *RadioButtons) SetFieldBackgroundColor(color tcell.Color) *RadioButtons {
	r.fieldBackgroundColor = color
	return r
}

// SetFieldTextColor sets the text color of the options area.
func (r *RadioButtons) SetFieldTextColor(color tcell.Color) *RadioButtons {
	r.fieldTextColor = color
	return r
}

// SetHorizontal sets the direction the form elements are laid out. If set to
// true, instead of positioning them from top to bottom (the default), they are
// positioned from left to right, moving into the next row if there is not
// enough space.
func (r *RadioButtons) SetHorizontal(horizontal bool) *RadioButtons {
	r.horizontal = horizontal
	return r
}

// SetCurrentOptionByName sets the index of the currently selected option. This may
// be a negative value to indicate that no option is currently selected.
func (r *RadioButtons) SetCurrentOptionByName(name string) *RadioButtons {
	for i := 0; i < len(r.options); i++ {
		if r.options[i].Name == name {
			r.currentOption = i
			if r.changed != nil {
				r.changed(r.options[r.currentOption])
			}
			break
		}
	}
	return r
}

// SetCurrentOption sets the index of the currently selected option. This may
// be a negative value to indicate that no option is currently selected.
func (r *RadioButtons) SetCurrentOption(index int) *RadioButtons {
	r.currentOption = index
	if r.changed != nil {
		r.changed(r.options[r.currentOption])
	}
	return r
}

// GetCurrentOption returns the index of the currently selected option as well
// as its text. If no option was selected, -1 and an empty string is returned.
func (r *RadioButtons) GetCurrentOption() *RadioOption {
	r = r.joinElements[r.currentElement]
	if len(r.options) > r.currentOption {
		return r.options[r.currentOption]
	}
	return nil
}

// GetCurrentOptionName returns the name of the currently selected option.
func (r *RadioButtons) GetCurrentOptionName() string {
	r = r.joinElements[r.currentElement]
	if len(r.options) > r.currentOption {
		return r.options[r.currentOption].Name
	}
	return ""
}

// SetChangedFunc sets the function which is called when the user navigates to
// a list item. The function receives the item's index in the list of items
// (starting with 0), its main text, secondary text, and its shortcut rune.
//
// This function is also called when the first item is added or when
// SetCurrentItem() is called.
func (r *RadioButtons) SetChangedFunc(handler func(*RadioOption)) *RadioButtons {
	r.changed = handler
	return r
}

// Draw draws this primitive onto the screen.
func (r *RadioButtons) Draw(screen tcell.Screen) {
	r.Box.Draw(screen)
	x, y, width, height := r.GetInnerRect()

	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}

	// Draw label.
	var labels = []struct {
		text  string
		color tcell.Color
	}{{
		text:  r.subLabel,
		color: r.subLabelColor,
	}, {
		text:  r.label,
		color: r.labelColor,
	}}

	if len(labels) > 0 {
		labelWidth := r.labelWidth
		if labelWidth > rightLimit-x {
			labelWidth = rightLimit - x
		}

		addCount := labelWidth - r.GetLabelWidth()

		for _, label := range labels {
			if addCount > 0 && strings.Contains(label.text, "%s") {
				label.text = fmt.Sprintf(label.text, strings.Repeat(r.labelFiller, addCount))
				addCount = 0
			} else {
				label.text = strings.Replace(label.text, "%s", "", -1)
			}

			labelWidth = StringWidth(label.text)
			Print(screen, label.text, x, y, labelWidth, AlignLeft, label.color)
			x += labelWidth
		}
	}

	var lineWidth int
	for index, option := range r.options {
		if index >= height && !r.horizontal {
			break
		}
		radioButton := Styles.GraphicsRadioUnchecked // Unchecked.
		if index == r.currentOption {
			radioButton = Styles.GraphicsRadioChecked // Checked.
		}

		line := fmt.Sprintf(`%s[white] %s`, radioButton, option.Title)
		if r.horizontal {
			Print(screen, line, x+lineWidth, y, width, AlignLeft, tcell.ColorWhite)
		} else {
			Print(screen, line, x, y+(index*(r.itemPadding+1)), width, AlignLeft, tcell.ColorWhite)
		}

		// Background color of selected text.
		if r.HasFocus() && index == r.currentOption {
			textWidth := StringWidth(line)
			for bx := 0; bx < textWidth && bx < width; bx++ {
				if r.horizontal {
					m, c, style, _ := screen.GetContent(x+lineWidth+bx, y)
					fg, _, _ := style.Decompose()
					if fg == r.mainTextColor {
						fg = r.selectedTextColor
					}
					style = style.Background(r.selectedBackgroundColor).Foreground(fg)
					screen.SetContent(x+lineWidth+bx, y, m, c, style)
				} else {
					m, c, style, _ := screen.GetContent(x+bx, y+(index*(r.itemPadding+1)))
					fg, _, _ := style.Decompose()
					if fg == r.mainTextColor {
						fg = r.selectedTextColor
					}
					style = style.Background(r.selectedBackgroundColor).Foreground(fg)
					screen.SetContent(x+bx, y+(index*(r.itemPadding+1)), m, c, style)
				}
			}
		}

		if r.horizontal {
			lineWidth += StringWidth(line) + (r.itemPadding + 1)
		}
	}
}
