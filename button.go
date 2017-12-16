package tview

import (
	"github.com/gdamore/tcell"
)

// Button is labeled box that triggers an action when selected.
type Button struct {
	*Box

	// The text to be displayed before the input area.
	label string

	// The label color.
	labelColor tcell.Color

	// The label color when the button is in focus.
	labelColorActivated tcell.Color

	// The background color when the button is in focus.
	backgroundColorActivated tcell.Color

	// An optional function which is called when the button was selected.
	selected func()

	// An optional function which is called when the user leaves the button. A
	// key is provided indicating which key was pressed to leave (tab or backtab).
	blur func(tcell.Key)
}

// NewButton returns a new input field.
func NewButton(label string) *Button {
	box := NewBox().SetBackgroundColor(tcell.ColorBlue)
	return &Button{
		Box:                      box,
		label:                    label,
		labelColor:               tcell.ColorWhite,
		labelColorActivated:      tcell.ColorBlue,
		backgroundColorActivated: tcell.ColorWhite,
	}
}

// SetLabel sets the button text.
func (b *Button) SetLabel(label string) *Button {
	b.label = label
	return b
}

// GetLabel returns the button text.
func (b *Button) GetLabel() string {
	return b.label
}

// SetLabelColor sets the color of the button text.
func (b *Button) SetLabelColor(color tcell.Color) *Button {
	b.labelColor = color
	return b
}

// SetLabelColorActivated sets the color of the button text when the button is
// in focus.
func (b *Button) SetLabelColorActivated(color tcell.Color) *Button {
	b.labelColorActivated = color
	return b
}

// SetBackgroundColorActivated sets the background color of the button text when
// the button is in focus.
func (b *Button) SetBackgroundColorActivated(color tcell.Color) *Button {
	b.backgroundColorActivated = color
	return b
}

// SetSelectedFunc sets a handler which is called when the button was selected.
func (b *Button) SetSelectedFunc(handler func()) *Button {
	b.selected = handler
	return b
}

// SetBlurFunc sets a handler which is called when the user leaves the button.
// The callback function is provided with the key that was pressed, which is one
// of the following:
//
//   - KeyEscape: Leaving the button with no specific direction.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (b *Button) SetBlurFunc(handler func(key tcell.Key)) *Button {
	b.blur = handler
	return b
}

// Draw draws this primitive onto the screen.
func (b *Button) Draw(screen tcell.Screen) {
	// Draw the box.
	backgroundColor := b.backgroundColor
	if b.focus.HasFocus() {
		b.backgroundColor = b.backgroundColorActivated
	}
	b.Box.Draw(screen)
	b.backgroundColor = backgroundColor

	// Draw label.
	x := b.x + b.width/2
	y := b.y + b.height/2
	width := b.width
	if b.border {
		width -= 2
	}
	labelColor := b.labelColor
	if b.focus.HasFocus() {
		labelColor = b.labelColorActivated
	}
	Print(screen, b.label, x, y, width, AlignCenter, labelColor)

	if b.focus.HasFocus() {
		screen.HideCursor()
	}
}

// InputHandler returns the handler for this primitive.
func (b *Button) InputHandler() func(event *tcell.EventKey) {
	return func(event *tcell.EventKey) {
		// Process key event.
		switch key := event.Key(); key {
		case tcell.KeyEnter: // Selected.
			if b.selected != nil {
				b.selected()
			}
		case tcell.KeyBacktab, tcell.KeyTab, tcell.KeyEscape: // Leave. No action.
			if b.blur != nil {
				b.blur(key)
			}
		}
	}
}
