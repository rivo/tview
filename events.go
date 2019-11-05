package tview

import "github.com/gdamore/tcell"

// EventKey is the key input event info.
// This exists for some consistency with EventMouse,
// even though it's just an alias to tcell.EventKey for backwards compatibility.
type EventKey = tcell.EventKey

// MouseAction are bit flags indicating what the mouse is logically doing.
type MouseAction int

const (
	MouseDown MouseAction = 1 << iota
	MouseUp
	MouseClick // Button1 only.
	MouseMove  // The mouse position changed.
)

// EventMouse is the mouse event info.
type EventMouse struct {
	*tcell.EventMouse
	target Primitive
	app    *Application
	action MouseAction
}

// Target gets the target Primitive of the mouse event.
func (e *EventMouse) Target() Primitive {
	return e.target
}

// Application gets the event originating *Application.
func (e *EventMouse) Application() *Application {
	return e.app
}

// MouseAction gets the mouse action of this event.
func (e *EventMouse) Action() MouseAction {
	return e.action
}

// SetFocus will set focus to the primitive.
func (e *EventMouse) SetFocus(p Primitive) {
	e.app.SetFocus(p)
}

// NewEventMouse creates a new mouse event.
func NewEventMouse(base *tcell.EventMouse, target Primitive, app *Application, action MouseAction) *EventMouse {
	return &EventMouse{base, target, app, action}
}
