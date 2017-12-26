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

	// Key overrides.
	keyOverrides map[tcell.Key]func(p Primitive) bool

	// Rune overrides.
	runeOverrides map[rune]func(p Primitive) bool
}

// NewApplication creates and returns a new application.
func NewApplication() *Application {
	return &Application{
		keyOverrides:  make(map[tcell.Key]func(p Primitive) bool),
		runeOverrides: make(map[rune]func(p Primitive) bool),
	}
}

// SetKeyCapture installs a global capture function for the given key. It
// intercepts all events for the given key and routes them to the handler.
// The handler receives the Primitive to which the key is originally redirected,
// the one which has focus, or nil if it was not directed to a Primitive. The
// handler also returns whether or not the key event is then forwarded to that
// Primitive.
//
// Special keys (e.g. Escape, Enter, or Ctrl-A) are defined by the "key"
// argument. The "ch" rune is ignored. Other keys (e.g. "a", "h", or "5") are
// specified by their rune, with key set to tcell.KeyRune. See also
// https://godoc.org/github.com/gdamore/tcell#EventKey for more information.
//
// To remove a handler again, provide a nil handler for the same key.
//
// The application itself will exit when Ctrl-C is pressed. You can intercept
// this with this function as well.
func (a *Application) SetKeyCapture(key tcell.Key, ch rune, handler func(p Primitive) bool) *Application {
	if key == tcell.KeyRune {
		if handler != nil {
			a.runeOverrides[ch] = handler
		} else {
			if _, ok := a.runeOverrides[ch]; ok {
				delete(a.runeOverrides, ch)
			}
		}
	} else {
		if handler != nil {
			a.keyOverrides[key] = handler
		} else {
			if _, ok := a.keyOverrides[key]; ok {
				delete(a.keyOverrides, key)
			}
		}
	}
	return a
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
		if a.screen == nil {
			a.RUnlock()
			break
		}
		event := a.screen.PollEvent()
		a.RUnlock()
		if event == nil {
			break // The screen was finalized.
		}
		switch event := event.(type) {
		case *tcell.EventKey:
			a.RLock()
			p := a.focus
			a.RUnlock()

			// Intercept keys.
			if event.Key() == tcell.KeyRune {
				if handler, ok := a.runeOverrides[event.Rune()]; ok {
					if !handler(p) {
						break
					}
				}
			} else {
				if handler, ok := a.keyOverrides[event.Key()]; ok {
					pr := p
					if event.Key() == tcell.KeyCtrlC {
						pr = nil
					}
					if !handler(pr) {
						break
					}
				}
			}

			// Ctrl-C closes the application.
			if event.Key() == tcell.KeyCtrlC {
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
			a.Draw()
		}
	}

	return nil
}

// Stop stops the application, causing Run() to return.
func (a *Application) Stop() {
	a.RLock()
	defer a.RUnlock()
	if a.screen == nil {
		return
	}
	a.screen.Fini()
	a.screen = nil
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

	// Draw all primitives.
	a.root.Draw(a.screen)

	// Sync screen.
	a.screen.Show()

	return a
}

// SetRoot sets the root primitive for this application. This function must be
// called or nothing will be displayed when the application starts.
func (a *Application) SetRoot(root Primitive) *Application {
	a.Lock()
	defer a.Unlock()

	a.root = root

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
	a.RLock()
	defer a.RUnlock()
	return a.focus
}
