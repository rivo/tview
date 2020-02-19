package tview

import "github.com/gdamore/tcell"

type PrimitiveValidationHandler interface {
	Validate(p Primitive, key *tcell.EventKey) bool
}
