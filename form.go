package tview

import (
	"strings"

	"github.com/gdamore/tcell"
)

// Form is a Box which contains multiple input fields, one per row.
type Form struct {
	Box

	// The items of the form (one row per item).
	items []*InputField

	// The number of empty rows between items.
	itemPadding int

	// The index of the item which has focus.
	focusedItem int

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color
}

// NewForm returns a new form.
func NewForm() *Form {
	return &Form{
		Box:                  *NewBox(),
		itemPadding:          1,
		labelColor:           tcell.ColorYellow,
		fieldBackgroundColor: tcell.ColorBlue,
		fieldTextColor:       tcell.ColorWhite,
	}
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
			break
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
}

// Focus is called by the application when the primitive receives focus.
func (f *Form) Focus(app *Application) {
	f.Box.Focus(app)

	if len(f.items) == 0 {
		return
	}

	// Hand on the focus to one of our items.
	if f.focusedItem < 0 || f.focusedItem >= len(f.items) {
		f.focusedItem = 0
	}
	f.hasFocus = false
	inputField := f.items[f.focusedItem]
	inputField.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyTab:
			f.focusedItem++
			f.Focus(app)
		}
	})
	app.SetFocus(inputField)
}

// InputHandler returns the handler for this primitive.
func (f *Form) InputHandler() func(event *tcell.EventKey) {
	return func(event *tcell.EventKey) {
	}
}
