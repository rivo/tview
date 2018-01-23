package tview

import (
	"sync"

	"github.com/gdamore/tcell"
)

// Application represents the top node of an application.
//
// It is not strictly required to use this class as none of the other classes
// depend on it. However, it provides useful tools to set up an application and
// plays nicely with all widgets.
type Application struct {
	sync.RWMutex

	// The application's screen.
	screen tcell.Screen

	// The primitive which currently has the keyboard focus.
	focus Primitive

	// The root primitive to be seen on the screen.
	root Primitive

	// Whether or not the application resizes the root primitive.
	rootAutoSize bool

	// An optional capture function which receives a key event and returns the
	// event to be forwarded to the default input handler (nil if nothing should
	// be forwarded).
	inputCapture func(event tcell.Event) tcell.Event
}

// NewApplication creates and returns a new application.
func NewApplication() *Application {
	return &Application{}
}

// SetInputCapture sets a function which captures all key events before they are
// forwarded to the key event handler of the primitive which currently has
// focus. This function can then choose to forward that key event (or a
// different one) by returning it or stop the key event processing by returning
// nil.
//
// Note that this also affects the default event handling of the application
// itself: Such a handler can intercept the Ctrl-C event which closes the
// applicatoon.
func (a *Application) SetInputCapture(capture func(event tcell.Event) tcell.Event) *Application {
	a.inputCapture = capture
	return a
}

func (a *Application) Screen() tcell.Screen {
	return a.screen
}

// Run starts the application and thus the event loop. This function returns
// when Stop() was called.
func (a *Application) Run() error {
	var err error
	a.Lock()

	// Make a screen.
	a.screen, err = tcell.NewScreen()
	if err != nil {
		a.Unlock()
		return err
	}
	if err = a.screen.Init(); err != nil {
		a.Unlock()
		return err
	}

	// We catch panics to clean up because they mess up the terminal.
	defer func() {
		if p := recover(); p != nil {
			if a.screen != nil {
				a.screen.Fini()
			}
			panic(p)
		}
	}()

	// Draw the screen for the first time.
	a.Unlock()
	a.Draw()

	// Start event loop.
	for {
		a.RLock()
		screen := a.screen
		a.RUnlock()
		if screen == nil {
			break
		}

		// Wait for next event.
		event := a.screen.PollEvent()
		if event == nil {
			break // The screen was finalized.
		}

		// Intercept all events.
		if a.inputCapture != nil {
			event = a.inputCapture(event)
			if event == nil {
				break // Don't forward event.
			}
		}

		switch evt := event.(type) {
		case *tcell.EventKey:
			a.RLock()
			p := a.focus
			a.RUnlock()

			// Ctrl-C closes the application.
			if evt.Key() == tcell.KeyCtrlC {
				a.Stop()
			}

			// Pass other key events to the currently focused primitive.
			if p != nil {
				if handler := p.InputHandler(); handler != nil {
					handler(event, func(p Primitive) {
						a.SetFocus(p)
					})
					a.Draw()
				}
			}
		case *tcell.EventResize:
			a.Lock()
			screen := a.screen
			if a.rootAutoSize && a.root != nil {
				width, height := screen.Size()
				a.root.SetRect(0, 0, width, height)
			}
			a.Unlock()
			screen.Clear()
			a.Draw()
		}
	}

	return nil
}

// Stop stops the application, causing Run() to return.
func (a *Application) Stop() error {
	a.Lock()
	defer a.Unlock()
	if a.screen == nil {
		return nil
	}
	err := a.screen.Fini()
	a.screen = nil
	return err
}

// Draw refreshes the screen. It calls the Draw() function of the application's
// root primitive and then syncs the screen buffer.
func (a *Application) Draw() *Application {
	a.RLock()
	defer a.RUnlock()

	// Maybe we're not ready yet or not anymore.
	if a.screen == nil || a.root == nil {
		return a
	}

	// Resize if requested.
	if a.rootAutoSize && a.root != nil {
		width, height := a.screen.Size()
		a.root.SetRect(0, 0, width, height)
	}

	// Draw all primitives.
	a.root.Draw(a.screen)

	// Sync screen.
	a.screen.Show()

	return a
}

// SetRoot sets the root primitive for this application. This function must be
// called or nothing will be displayed when the application starts.
//
// It also calls SetFocus() on the primitive.
func (a *Application) SetRoot(root Primitive, autoSize bool) *Application {

	a.Lock()
	a.root = root
	a.rootAutoSize = autoSize
	if a.screen != nil {
		a.screen.Clear()
	}
	a.Unlock()

	a.SetFocus(root)

	return a
}

// ResizeToFullScreen resizes the given primitive such that it fills the entire
// screen.
func (a *Application) ResizeToFullScreen(p Primitive) *Application {
	a.RLock()
	width, height := a.screen.Size()
	a.RUnlock()
	p.SetRect(0, 0, width, height)
	return a
}

// SetFocus sets the focus on a new primitive. All key events will be redirected
// to that primitive. Callers must ensure that the primitive will handle key
// events.
//
// Blur() will be called on the previously focused primitive. Focus() will be
// called on the new primitive.
func (a *Application) SetFocus(p Primitive) *Application {
	a.Lock()
	if a.focus != nil {
		a.focus.Blur()
	}
	a.focus = p
	if a.screen != nil {
		a.screen.HideCursor()
	}
	a.Unlock()
	p.Focus(func(p Primitive) {
		a.SetFocus(p)
	})

	return a
}

// GetFocus returns the primitive which has the current focus. If none has it,
// nil is returned.
func (a *Application) GetFocus() Primitive {
	a.RLock()
	defer a.RUnlock()
	return a.focus
}
