package tview

import (
	"sync"

	"github.com/gdamore/tcell"
)

// Application represents the top node of an application.
type Application struct {
	sync.Mutex

	// The application's screen.
	screen tcell.Screen

	// The primitive which currently has the keyboard focus.
	focus Primitive

	// The root primitive to be seen on the screen.
	root Primitive

	// Whether or not the application resizes the root primitive.
	rootAutoSize bool
}

// NewApplication creates and returns a new application.
func NewApplication() *Application {
	return &Application{}
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
	if a.rootAutoSize && a.root != nil {
		width, height := a.screen.Size()
		a.root.SetRect(0, 0, width, height)
	}
	a.Unlock()
	a.Draw()

	// Start event loop.
	for {
		if a.screen == nil {
			break
		}
		event := a.screen.PollEvent()
		if event == nil {
			break // The screen was finalized.
		}
		switch event := event.(type) {
		case *tcell.EventKey:
			if event.Key() == tcell.KeyCtrlC {
				a.Stop() // Ctrl-C closes the application.
			}
			a.Lock()
			p := a.focus // Pass other key events to the currently focused primitive.
			a.Unlock()
			if p != nil {
				if handler := p.InputHandler(); handler != nil {
					handler(event, func(p Primitive) {
						a.SetFocus(p)
					})
					a.Draw()
				}
			}
		case *tcell.EventResize:
			if a.rootAutoSize && a.root != nil {
				a.Lock()
				width, height := a.screen.Size()
				a.root.SetRect(0, 0, width, height)
				a.Unlock()
				a.Draw()
			}
		}
	}

	return nil
}

// Stop stops the application, causing Run() to return.
func (a *Application) Stop() {
	if a.screen == nil {
		return
	}
	a.screen.Fini()
	a.screen = nil
}

// Draw refreshes the screen. It calls the Draw() function of the application's
// root primitive and then syncs the screen buffer.
func (a *Application) Draw() *Application {
	a.Lock()
	defer a.Unlock()

	// Maybe we're not ready yet or not anymore.
	if a.screen == nil || a.root == nil {
		return a
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
// If autoSize is set to true, the application will set the root primitive's
// position to (0,0) and its size to the screen's size. It will also resize and
// redraw it when the screen resizes.
func (a *Application) SetRoot(root Primitive, autoSize bool) *Application {
	a.Lock()
	defer a.Unlock()

	a.root = root
	a.rootAutoSize = autoSize

	return a
}

// SetFocus sets the focus on a new primitive. All key events will be redirected
// to that primitive. Callers must ensure that the primitive will handle key
// events.
//
// Blur() will be called on the previously focused primitive. Focus() will be
// called on the new primitive.
func (a *Application) SetFocus(p Primitive) *Application {
	if p.InputHandler() == nil {
		return a
	}

	a.Lock()
	if a.focus != nil {
		a.focus.Blur()
	}
	a.focus = p
	a.Unlock()
	p.Focus(func(p Primitive) {
		a.SetFocus(p)
	})

	return a
}

// GetFocus returns the primitive which has the current focus. If none has it,
// nil is returned.
func (a *Application) GetFocus() Primitive {
	return a.focus
}
