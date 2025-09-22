// Demo code for the List primitive.
package main

import (
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	list := tview.NewMultiList().
		AddItem("List item 1", "Some explanatory text", nil).
		AddItem("List item 2", "Some explanatory text", nil).
		AddItem("List item 3", "Some explanatory text", nil).
		AddItem("List item 4", "Some explanatory text", nil).
		AddItem("Quit", "Press to exit", func() {
			app.Stop()
		})
	if err := app.SetRoot(list, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
