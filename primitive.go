package tview

import "github.com/gdamore/tcell"

// Primitive is the top-most interface for all graphical primitives.
type Primitive interface {
	// Draw draws this primitive onto the screen. Implementers can call the
	// screen's ShowCursor() function but should only do so when they have focus.
	// (They will need to keep track of this themselves.)
	Draw(screen tcell.Screen)

	// GetRect returns the current position of the primitive, x, y, width, and
	// height.
	GetRect() (int, int, int, int)

	// SetRect sets a new position of the primitive.
	SetRect(x, y, width, height int)

	// InputHandler returns a handler which receives key events when it has focus.
	// It is called by the Application class.
	//
	// A value of nil may also be returned, in which case this primitive cannot
	// receive focus and will not process any key events.
	//
	// The Application's Draw() function will be called automatically after the
	// handler returns.
	InputHandler() func(event *tcell.EventKey)

	// Focus is called by the application when the primitive receives focus.
	Focus(app *Application)

	// Blur is called by the application when the primitive loses focus.
	Blur()
}
