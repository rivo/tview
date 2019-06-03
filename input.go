package tview

import "github.com/diamondburned/tcell"

// MouseSupport is the interface which determines a primitive's
// mouse capabilities
type MouseSupport interface {
	MouseHandler() func(*tcell.EventMouse) bool
}
