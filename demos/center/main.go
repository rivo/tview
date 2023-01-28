// Demo code for the Centered Modal primitive.
package main

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	background := tview.NewTextView().
		SetTextColor(tcell.ColorBlue).
		SetText(strings.Repeat("background ", 1000))

	box := tview.NewBox().
		SetBorder(true).
		SetTitle("Centered Box")

	pages := tview.NewPages().
		AddPage("background", background, true, true).
		AddPage("modal", tview.NewCenter(box, 40, 10), true, true)

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}
