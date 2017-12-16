package tview

import (
	"strings"

	"github.com/gdamore/tcell"
)

// Form is a Box which contains multiple input fields, one per row.
type Form struct {
	*Box

	// The items of the form (one row per item).
	items []*InputField

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

// AddItem adds a new item to the form. It has a label, an optional initial
// value, a field length (a value of 0 extends it as far as possible), and
// an optional accept function to validate the item's value (set to nil to
// accept any text).
func (f *Form) AddItem(label, value string, fieldLength int, accept func(textToCheck string, lastChar rune) bool) *Form {
	f.items = append(f.items, NewInputField().
		SetLabel(label).
		SetText(value).
		SetFieldLength(fieldLength).
		SetAcceptanceFunc(accept))
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
	for _, inputField := range f.items {
		label := strings.TrimSpace(inputField.GetLabel())
		if len([]rune(label)) > labelLength {
			labelLength = len([]rune(label))
		}
	}
	labelLength++ // Add one space.

	// Set up and draw the input fields.
	for _, inputField := range f.items {
		if y >= bottomLimit {
			return // Stop here.
		}
		label := strings.TrimSpace(inputField.GetLabel())
		inputField.SetLabelColor(f.labelColor).
			SetFieldBackgroundColor(f.fieldBackgroundColor).
			SetFieldTextColor(f.fieldTextColor).
			SetLabel(label+strings.Repeat(" ", labelLength-len([]rune(label)))).
			SetBackgroundColor(f.backgroundColor).
			SetRect(x, y, width, 1)
		inputField.Draw(screen)
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
func (f *Form) Focus(app *Application) {
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
		f.Focus(app)
	}
	if f.focusedElement < len(f.items) {
		// We're selecting an item.
		inputField := f.items[f.focusedElement]
		inputField.SetDoneFunc(handler)
		app.SetFocus(inputField)
	} else {
		// We're selecting a button.
		button := f.buttons[f.focusedElement-len(f.items)]
		button.SetBlurFunc(handler)
		app.SetFocus(button)
	}
}

// InputHandler returns the handler for this primitive.
func (f *Form) InputHandler() func(event *tcell.EventKey) {
	return func(event *tcell.EventKey) {}
}

// HasFocus returns whether or not this primitive has focus.
func (f *Form) HasFocus() bool {
	for _, item := range f.items {
		if item.focus.HasFocus() {
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
