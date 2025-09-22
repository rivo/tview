// Demo code for the Modal primitive.
package main

import (
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	list := tview.NewList()
	list.ShowSecondaryText(false)
	list.SetHighlightFullLine(true)

	for _, option := range []string{"Option 1", "Option 2", "Quit"} {
		list.AddItem(option, "", 0, nil)
	}

	modal := tview.NewModalList("Selection", list).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				app.Stop()
			}
		})
	if err := app.SetRoot(modal, false).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
