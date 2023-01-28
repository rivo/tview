package tview

import (
	"github.com/gdamore/tcell/v2"
)

// Center is a wrapper which adds space around another primitive to shot up in the middle.
type Center struct {
	*Box

	// the contained primitive
	primitive Primitive

	width, height int

	// keep a reference in case we need it when we change the primitive
	setFocus func(p Primitive)
}

// NewCenter returns a new frame around the given primitive.
func NewCenter(primitive Primitive, width, height int) *Center {
	return &Center{Box: NewBox(), primitive: primitive, width: width, height: height}
}

func (c *Center) Resize(width, height int) *Center {
	if width > 0 {
		c.width = width
	}
	if height > 0 {
		c.height = height
	}
	return c
}

// SetPrimitive replaces the contained primitive with the given one.
func (c *Center) SetPrimitive(p Primitive) *Center {
	hasFocus := c.primitive.HasFocus()
	c.primitive = p
	if hasFocus && c.setFocus != nil {
		c.setFocus(p) // Restore focus.
	}
	return c
}

// Primitive returns the primitive contained in this frame.
func (c *Center) Primitive() Primitive {
	return c.primitive
}

// Draw draws this primitive onto the screen.
func (c *Center) Draw(screen tcell.Screen) {
	x, y, inWidth, inHeight := c.GetInnerRect()
	width, height := c.width, c.height
	if width < inWidth {
		x += (inWidth - width) >> 1
	} else if width > inWidth {
		width = inWidth
	}
	if height < inHeight {
		y += (inHeight - height) >> 1
	} else if height > inHeight {
		height = inHeight
	}

	c.primitive.SetRect(x, y, width, height)
	c.primitive.Draw(screen)
}

// Focus is called when this primitive receives focus.
func (c *Center) Focus(delegate func(p Primitive)) {
	c.setFocus = delegate
	delegate(c.primitive)
	c.Box.Focus(delegate)
}

// HasFocus returns whether or not this primitive has focus.
func (c *Center) HasFocus() bool {
	return c.primitive.HasFocus()
}

// MouseHandler returns the mouse handler for this primitive.
func (c *Center) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return c.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if !c.InRect(event.Position()) {
			return false, nil
		}

		// Pass mouse events on to contained primitive.
		consumed, capture = c.primitive.MouseHandler()(action, event, setFocus)
		if consumed {
			return true, capture
		}

		// Clicking on the frame parts.
		if action == MouseLeftDown {
			setFocus(c)
			consumed = true
		}

		return
	})
}

// InputHandler returns the handler for this primitive.
func (c *Center) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if handler := c.primitive.InputHandler(); handler != nil {
			handler(event, setFocus)
			return
		}
	})
}
