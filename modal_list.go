package tview

import (
	"github.com/gdamore/tcell/v2"
)

type ModalList struct {
	*Box

	// The list embedded in the modal's frame.
	list *List

	// The frame embedded in the modal.
	frame *Frame

	// The optional callback for when the user clicked one of the items. It
	// receives the index of the clicked item and the item's text.
	done func(int, string)
}

func NewModalList(title string, list *List) *ModalList {
	m := &ModalList{Box: NewBox()}

	list.SetSelectedFunc(func(i int, main string, _ string, _ rune) {
		if m.done != nil {
			m.done(i, main)
		}
	})
	list.SetDoneFunc(func() {
		if m.done != nil {
			m.done(-1, "")
		}
	})
	m.list = list

	frame := NewFrame(m.list).SetBorders(0, 0, 1, 0, 0, 0)
	frame.SetBorder(true).
		SetBackgroundColor(Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)
	frame.SetTitle(title)
	frame.SetTitleColor(Styles.TitleColor)
	m.SetFrame(frame)
	m.Box.Primitive = m
	m.dontClear = true
	return m
}

func (m *ModalList) SetFrame(frame *Frame) *ModalList {
	m.frame = frame
	return m
}

// SetBackgroundColor sets the color of the modal frame background.
func (m *ModalList) SetBackgroundColor(color tcell.Color) *ModalList {
	m.list.SetBackgroundColor(color)
	m.frame.SetBackgroundColor(color)
	return m
}

// Draw draws this primitive onto the screen.
func (m *ModalList) Draw(screen tcell.Screen) {
	// Calculate the width of this modal.
	width := 0
	for i := 0; i < m.list.GetItemCount(); i++ {
		main, secondary := m.list.GetItemText(i)
		width = max(width, len(main)+len(secondary)+2)
	}

	screenWidth, screenHeight := screen.Size()

	// Set the modal's position and size.
	height := m.list.GetItemCount() + 4
	width += 2
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.frame.Clear()
	m.Box.DrawForSubclass(screen, m)
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

func (m *ModalList) SetDoneFunc(handler func(int, string)) *ModalList {
	m.done = handler
	return m
}

// Focus is called when this primitive receives focus.
func (m *ModalList) Focus(delegate func(p Primitive)) {
	delegate(m.list)
}

// HasFocus returns whether this primitive has focus.
func (m *ModalList) HasFocus() bool {
	return m.list.HasFocus()
}

// FocusChain implements the [Primitive]'s FocusChain method.
func (m *ModalList) FocusChain(chain *[]Primitive) bool {
	if hasFocus := m.list.FocusChain(chain); hasFocus {
		if chain != nil {
			*chain = append(*chain, m)
		}
		return true
	}
	return m.Box.FocusChain(chain)
}

// MouseHandler returns the mouse handler for this primitive.
func (m *ModalList) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return m.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		// Pass mouse events on to the form.
		consumed, capture = m.list.MouseHandler()(action, event, setFocus)
		if !consumed && action == MouseLeftClick && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}

// InputHandler returns the handler for this primitive.
func (m *ModalList) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if m.frame.HasFocus() {
			if handler := m.frame.InputHandler(); handler != nil {
				handler(event, setFocus)
				return
			}
		}
	})
}
