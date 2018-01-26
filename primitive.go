package tview

import "github.com/gdamore/tcell"

// Primitive is the top-most interface for all graphical primitives.
type Primitive interface {
	// everything should have a unique Id, defaults to a uuid
	Id() string

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
	// The handler will receive any events and a function that allows it to
	// set the focus to a different primitive, so that future key events are sent
	// to that primitive.
	//
	// The Application's Draw() function will be called automatically after the
	// handler returns.
	//
	// The Box class provides functionality to intercept keyboard input. If you
	// subclass from Box, it is recommended that you wrap your handler using
	// Box.wrapInputHandler() so you inherit that functionality.
	InputHandler() func(event tcell.Event, setFocus func(p Primitive))

	// Focus is called by the application when the primitive receives focus.
	// Implementers may call delegate() to pass the focus on to another primitive.
	Focus(delegate func(p Primitive))

	// Blur is called by the application when the primitive loses focus.
	Blur()

	// HasFocus returns true if the primitive has focus
	HasFocus() bool

	// GetFocusable returns the item's Focusable.
	GetFocusable() Focusable

	// Mount is a longer term context for bringing a widget into scope
	Mount(context map[string]interface{}) error

	// Refresh is a longer term context for bringing a widget into scope
	Refresh(context map[string]interface{}) error

	// Unmount is the opposite of mount
	Unmount() error

	// IsMounted returns true if the primitive is mounted
	IsMounted() bool

	// Render is called when something in the future does so.
	Render() error

	// GetProps returns the primitive's prop
	GetProp(prop string) (interface{}, bool)

	// GetProps returns the primitive's props
	GetProps() map[string]interface{}

	// SetProp sets a primitive's props
	// It is up to the implementor to ensure correctness.
	SetProp(props string, value interface{}) error

	// SetProps replaces the primitive's props
	// It is up to the implementor to ensure correctness.
	SetProps(newProps map[string]interface{}) error
}
