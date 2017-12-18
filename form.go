package tview

import (
	"strings"

	"github.com/gdamore/tcell"
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

	// GetFocusable returns the item's Focusable.
	GetFocusable() Focusable
}

// Form is a Box which contains multiple input fields, one per row.
type Form struct {
	*Box

	// The items of the form (one row per item).
	items []FormItem

	// The buttons of the form.
	buttons []*Button

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
}

// NewForm returns a new form.
func NewForm() *Form {
	box := NewBox()

	f := &Form{
		Box:                  box,
		itemPadding:          1,
		labelColor:           tcell.ColorYellow,
		fieldBackgroundColor: tcell.ColorBlue,
		fieldTextColor:       tcell.ColorWhite,
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

// AddInputField adds an input field to the form. It has a label, an optional
// initial value, a field length (a value of 0 extends it as far as possible),
// and an optional accept function to validate the item's value (set to nil to
// accept any text).
func (f *Form) AddInputField(label, value string, fieldLength int, accept func(textToCheck string, lastChar rune) bool) *Form {
	f.items = append(f.items, NewInputField().
		SetLabel(label).
		SetText(value).
		SetFieldLength(fieldLength).
		SetAcceptanceFunc(accept))
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

// AddButton adds a new button to the form. The "selected" function is called
// when the user selects this button. It may be nil.
func (f *Form) AddButton(label string, selected func()) *Form {
	f.buttons = append(f.buttons, NewButton(label).SetSelectedFunc(selected))
	return f
}

// Draw draws this primitive onto the screen.
func (f *Form) Draw(screen tcell.Screen) {
	f.Box.Draw(screen)

	// Determine the dimensions.
	x := f.x
	y := f.y
	width := f.width
	bottomLimit := f.y + f.height
	if f.border {
		x++
		y++
		width -= 2
		bottomLimit -= 2
	}
	rightLimit := x + width

	// Find the longest label.
	var labelLength int
	for _, item := range f.items {
		label := strings.TrimSpace(item.GetLabel())
		if len([]rune(label)) > labelLength {
			labelLength = len([]rune(label))
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
			label+strings.Repeat(" ", labelLength-len([]rune(label))),
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

	// Draw the buttons.
	if f.itemPadding == 0 {
		y++
	}
	if y >= bottomLimit {
		return // Stop here.
	}
	for _, button := range f.buttons {
		space := rightLimit - x
		if space < 1 {
			return // No space for this button anymore.
		}
		buttonWidth := len([]rune(button.GetLabel())) + 4
		if buttonWidth > space {
			buttonWidth = space
		}
		button.SetRect(x, y, buttonWidth, 1)
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
		case tcell.KeyBacktab:
			f.focusedElement--
			if f.focusedElement < 0 {
				f.focusedElement = len(f.items) + len(f.buttons) - 1
			}
		case tcell.KeyEscape:
			f.focusedElement = 0
		}
		f.Focus(delegate)
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

// InputHandler returns the handler for this primitive.
func (f *Form) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p Primitive)) {}
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
