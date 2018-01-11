package tview

import (
	"strings"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
)

// FormItem is the interface all form items must implement to be able to be
// included in a form.
type FormItem interface {
	Primitive

	// GetLabel returns the item's label text.
	GetLabel() string

	// SetFormAttributes sets a number of item attributes at once.
	SetFormAttributes(label string, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem

	// SetEnteredFunc sets the handler function for when the user finished
	// entering data into the item. The handler may receive events for the
	// Enter key (we're done), the Escape key (cancel input), the Tab key (move to
	// next field), and the Backtab key (move to previous field).
	SetFinishedFunc(handler func(key tcell.Key)) FormItem
}

// Form is a Box which contains multiple input fields, one per row.
//
// See https://github.com/rivo/tview/wiki/Form for an example.
type Form struct {
	*Box

	// The items of the form (one row per item).
	items []FormItem

	// The buttons of the form.
	buttons []*Button

	// The alignment of the buttons.
	buttonsAlign int

	// The number of empty rows between items.
	itemPadding int

	// The index of the item or button which has focus. (Items are counted first,
	// buttons are counted last.)
	focusedElement int

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// The background color of the buttons.
	buttonBackgroundColor tcell.Color

	// The color of the button text.
	buttonTextColor tcell.Color

	// An optional function which is called when the user hits Escape.
	cancel func()
}

// NewForm returns a new form.
func NewForm() *Form {
	box := NewBox().SetBorderPadding(1, 1, 1, 1)

	f := &Form{
		Box:                   box,
		itemPadding:           1,
		labelColor:            Styles.SecondaryTextColor,
		fieldBackgroundColor:  Styles.ContrastBackgroundColor,
		fieldTextColor:        Styles.PrimaryTextColor,
		buttonBackgroundColor: Styles.ContrastBackgroundColor,
		buttonTextColor:       Styles.PrimaryTextColor,
	}

	f.focus = f

	return f
}

// SetItemPadding sets the number of empty rows between form items.
func (f *Form) SetItemPadding(padding int) *Form {
	f.itemPadding = padding
	return f
}

// SetLabelColor sets the color of the labels.
func (f *Form) SetLabelColor(color tcell.Color) *Form {
	f.labelColor = color
	return f
}

// SetFieldBackgroundColor sets the background color of the input areas.
func (f *Form) SetFieldBackgroundColor(color tcell.Color) *Form {
	f.fieldBackgroundColor = color
	return f
}

// SetFieldTextColor sets the text color of the input areas.
func (f *Form) SetFieldTextColor(color tcell.Color) *Form {
	f.fieldTextColor = color
	return f
}

// SetButtonsAlign sets how the buttons align horizontally, one of AlignLeft
// (the default), AlignCenter, and AlignRight.
func (f *Form) SetButtonsAlign(align int) *Form {
	f.buttonsAlign = align
	return f
}

// SetButtonBackgroundColor sets the background color of the buttons.
func (f *Form) SetButtonBackgroundColor(color tcell.Color) *Form {
	f.buttonBackgroundColor = color
	return f
}

// SetButtonTextColor sets the color of the button texts.
func (f *Form) SetButtonTextColor(color tcell.Color) *Form {
	f.buttonTextColor = color
	return f
}

// AddInputField adds an input field to the form. It has a label, an optional
// initial value, a field length (a value of 0 extends it as far as possible),
// an optional accept function to validate the item's value (set to nil to
// accept any text), and an (optional) callback function which is invoked when
// the input field's text has changed.
func (f *Form) AddInputField(label, value string, fieldLength int, accept func(textToCheck string, lastChar rune) bool, changed func(text string)) *Form {
	f.items = append(f.items, NewInputField().
		SetLabel(label).
		SetText(value).
		SetFieldLength(fieldLength).
		SetAcceptanceFunc(accept).
		SetChangedFunc(changed))
	return f
}

// AddPasswordField adds a password field to the form. This is similar to an
// input field except that the user's input not shown. Instead, a "mask"
// character is displayed. The password field has a label, an optional initial
// value, a field length (a value of 0 extends it as far as possible), and an
// (optional) callback function which is invoked when the input field's text has
// changed.
func (f *Form) AddPasswordField(label, value string, fieldLength int, mask rune, changed func(text string)) *Form {
	if mask == 0 {
		mask = '*'
	}
	f.items = append(f.items, NewInputField().
		SetLabel(label).
		SetText(value).
		SetFieldLength(fieldLength).
		SetMaskCharacter(mask).
		SetChangedFunc(changed))
	return f
}

// AddDropDown adds a drop-down element to the form. It has a label, options,
// and an (optional) callback function which is invoked when an option was
// selected.
func (f *Form) AddDropDown(label string, options []string, initialOption int, selected func(option string, optionIndex int)) *Form {
	f.items = append(f.items, NewDropDown().
		SetLabel(label).
		SetCurrentOption(initialOption).
		SetOptions(options, selected))
	return f
}

// AddCheckbox adds a checkbox to the form. It has a label, an initial state,
// and an (optional) callback function which is invoked when the state of the
// checkbox was changed by the user.
func (f *Form) AddCheckbox(label string, checked bool, changed func(checked bool)) *Form {
	f.items = append(f.items, NewCheckbox().
		SetLabel(label).
		SetChecked(checked).
		SetChangedFunc(changed))
	return f
}

// AddButton adds a new button to the form. The "selected" function is called
// when the user selects this button. It may be nil.
func (f *Form) AddButton(label string, selected func()) *Form {
	f.buttons = append(f.buttons, NewButton(label).SetSelectedFunc(selected))
	return f
}

// GetElement returns the form element at the given position, starting with
// index 0. Elements are referenced in the order they were added. Buttons are
// not included.
func (f *Form) GetElement(index int) Primitive {
	return f.items[index]
}

// SetCancelFunc sets a handler which is called when the user hits the Escape
// key.
func (f *Form) SetCancelFunc(callback func()) *Form {
	f.cancel = callback
	return f
}

// Draw draws this primitive onto the screen.
func (f *Form) Draw(screen tcell.Screen) {
	f.Box.Draw(screen)

	// Determine the dimensions.
	x, y, width, height := f.GetInnerRect()
	bottomLimit := y + height
	rightLimit := x + width

	// Find the longest label.
	var labelLength int
	for _, item := range f.items {
		label := strings.TrimSpace(item.GetLabel())
		labelWidth := runewidth.StringWidth(label)
		if labelWidth > labelLength {
			labelLength = labelWidth
		}
	}
	labelLength++ // Add one space.

	// Set up and draw the input fields.
	for _, item := range f.items {
		if y >= bottomLimit {
			return // Stop here.
		}
		label := strings.TrimSpace(item.GetLabel())
		item.SetFormAttributes(
			label+strings.Repeat(" ", labelLength-runewidth.StringWidth(label)),
			f.labelColor,
			f.backgroundColor,
			f.fieldTextColor,
			f.fieldBackgroundColor,
		).SetRect(x, y, width, 1)
		if item.GetFocusable().HasFocus() {
			defer item.Draw(screen)
		} else {
			item.Draw(screen)
		}
		y += 1 + f.itemPadding
	}

	// How wide are the buttons?
	buttonWidths := make([]int, len(f.buttons))
	buttonsWidth := 0
	for index, button := range f.buttons {
		width := runewidth.StringWidth(button.GetLabel()) + 4
		buttonWidths[index] = width
		buttonsWidth += width + 2
	}
	buttonsWidth -= 2

	// Where do we place them?
	if x+buttonsWidth < rightLimit {
		if f.buttonsAlign == AlignRight {
			x = rightLimit - buttonsWidth
		} else if f.buttonsAlign == AlignCenter {
			x = (x + rightLimit - buttonsWidth) / 2
		}
	}

	// Draw them.
	if f.itemPadding == 0 {
		y++
	}
	if y >= bottomLimit {
		return // Stop here.
	}
	for index, button := range f.buttons {
		space := rightLimit - x
		if space < 1 {
			break // No space for this button anymore.
		}
		buttonWidth := buttonWidths[index]
		if buttonWidth > space {
			buttonWidth = space
		}
		button.SetLabelColor(f.buttonTextColor).
			SetLabelColorActivated(f.buttonBackgroundColor).
			SetBackgroundColorActivated(f.buttonTextColor).
			SetBackgroundColor(f.buttonBackgroundColor).
			SetRect(x, y, buttonWidth, 1)
		button.Draw(screen)

		x += buttonWidth + 2
	}
}

// Focus is called by the application when the primitive receives focus.
func (f *Form) Focus(delegate func(p Primitive)) {
	if len(f.items)+len(f.buttons) == 0 {
		return
	}

	// Hand on the focus to one of our child elements.
	if f.focusedElement < 0 || f.focusedElement >= len(f.items)+len(f.buttons) {
		f.focusedElement = 0
	}
	handler := func(key tcell.Key) {
		switch key {
		case tcell.KeyTab, tcell.KeyEnter:
			f.focusedElement++
			f.Focus(delegate)
		case tcell.KeyBacktab:
			f.focusedElement--
			if f.focusedElement < 0 {
				f.focusedElement = len(f.items) + len(f.buttons) - 1
			}
			f.Focus(delegate)
		case tcell.KeyEscape:
			if f.cancel != nil {
				f.cancel()
			} else {
				f.focusedElement = 0
				f.Focus(delegate)
			}
		}
	}

	if f.focusedElement < len(f.items) {
		// We're selecting an item.
		item := f.items[f.focusedElement]
		item.SetFinishedFunc(handler)
		delegate(item)
	} else {
		// We're selecting a button.
		button := f.buttons[f.focusedElement-len(f.items)]
		button.SetBlurFunc(handler)
		delegate(button)
	}
}

// HasFocus returns whether or not this primitive has focus.
func (f *Form) HasFocus() bool {
	for _, item := range f.items {
		if item.GetFocusable().HasFocus() {
			return true
		}
	}
	for _, button := range f.buttons {
		if button.focus.HasFocus() {
			return true
		}
	}
	return false
}
