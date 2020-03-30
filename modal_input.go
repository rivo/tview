package tview

import (
	"github.com/gdamore/tcell"
)

// ModalInput is a centered message window used to inform the user or prompt them
// for response to a question. It needs to have at least one button (added via
// AddButtons()) or it will never disappear.
//
// See https://github.com/rivo/tview/wiki/ModalInput for an example.
type ModalInput struct {
	*Box

	// The framed embedded in the modal.
	frame *Frame

	// The form embedded in the modal's frame.
	form *Form

	// The message text (original, not word-wrapped).
	text string

	// The text color.
	textColor tcell.Color

	// The optional callback for when the user clicked one of the buttons. It
	// receives the index of the clicked button and the button's label.
	done func(buttonIndex int, buttonLabel, answer string)

	// The index of our input element
	inputIndex int
}

// NewModal returns a new modal message window for input.
// It takes a label for the input and an optional (pre-set) value
func NewModalInput(label, value string) *ModalInput {
	m := &ModalInput{
		Box:       NewBox(),
		textColor: Styles.PrimaryTextColor,
	}
	m.form = NewForm().
		SetButtonsAlign(AlignCenter).
		SetButtonBackgroundColor(Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(Styles.PrimaryTextColor)
	m.form.SetBackgroundColor(Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	m.form.SetCancelFunc(func() {
		if m.done != nil {
			m.done(-1, "", "")
		}
	})
	m.inputIndex = len(m.form.items)
	m.form.AddInputField(label, value, 0, nil, nil)

	m.frame = NewFrame(m.form).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBorder(true).
		SetBackgroundColor(Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)

	m.focus = m
	return m
}

// SetBackgroundColor sets the color of the modal frame background.
func (m *ModalInput) SetBackgroundColor(color tcell.Color) *ModalInput {
	m.form.SetBackgroundColor(color)
	m.frame.SetBackgroundColor(color)
	return m
}

// SetTextColor sets the color of the message text.
func (m *ModalInput) SetTextColor(color tcell.Color) *ModalInput {
	m.textColor = color
	return m
}

// SetButtonBackgroundColor sets the background color of the buttons.
func (m *ModalInput) SetButtonBackgroundColor(color tcell.Color) *ModalInput {
	m.form.SetButtonBackgroundColor(color)
	return m
}

// SetButtonTextColor sets the color of the button texts.
func (m *ModalInput) SetButtonTextColor(color tcell.Color) *ModalInput {
	m.form.SetButtonTextColor(color)
	return m
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the user presses the Escape key. The index will
// then be negative and the label text an emptry string.
func (m *ModalInput) SetDoneFunc(handler func(buttonIndex int, buttonLabel, input string)) *ModalInput {
	m.done = handler
	return m
}

// SetText sets the message text of the window. The text may contain line
// breaks. Note that words are wrapped, too, based on the final size of the
// window.
func (m *ModalInput) SetText(text string) *ModalInput {
	m.text = text
	return m
}

// AddButtons adds buttons to the window. There must be at least one button and
// a "done" handler so the window can be closed again.
func (m *ModalInput) AddButtons(labels []string) *ModalInput {
	for index, label := range labels {
		func(i int, l string) {
			m.form.AddButton(label, func() {
				txt := ""
				item := m.form.GetFormItem(m.inputIndex)
				if input, ok := item.(*InputField); ok {
					txt = input.GetText()
				}
				if m.done != nil {
					m.done(i, l, txt)
				}
			})
			button := m.form.GetButton(m.form.GetButtonCount() - 1)
			button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyDown, tcell.KeyRight:
					return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
				case tcell.KeyUp, tcell.KeyLeft:
					return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
				}
				return event
			})
		}(index, label)
	}
	return m
}

// ClearButtons removes all buttons from the window.
func (m *ModalInput) ClearButtons() *ModalInput {
	m.form.ClearButtons()
	return m
}

// SetFocus shifts the focus to the button with the given index.
func (m *ModalInput) SetFocus(index int) *ModalInput {
	m.form.SetFocus(index)
	return m
}

// Focus is called when this primitive receives focus.
func (m *ModalInput) Focus(delegate func(p Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *ModalInput) HasFocus() bool {
	return m.form.HasFocus()
}

// Draw draws this primitive onto the screen.
func (m *ModalInput) Draw(screen tcell.Screen) {
	// Calculate the width of this modal.
	buttonsWidth := 0
	for i := 0; i < m.form.GetButtonCount(); i++ {
		button := m.form.GetButton(i)
		buttonsWidth += TaggedStringWidth(button.GetLabel()) + 4 + 2
	}
	buttonsWidth -= 2
	screenWidth, screenHeight := screen.Size()
	width := screenWidth / 3
	if width < buttonsWidth {
		width = buttonsWidth
	}
	// width is now without the box border.

	// Reset the text and find out how wide it is.
	m.frame.Clear()
	lines := WordWrap(m.text, width)
	for _, line := range lines {
		m.frame.AddText(line, true, AlignCenter, m.textColor)
	}
	// Add lines for input

	// Set the modal's position and size.
	height := len(lines) + 6 + 2
	width += 4
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}
