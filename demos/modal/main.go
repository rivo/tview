// Demo code for the Modal primitive.
package main

import (
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	modal := tview.NewModal().
		SetText("Do you want to quit the application?").
		AddButtons([]string{"Quit", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				app.Stop()
			}
		})
	if err := app.SetRoot(modal, false).Run(); err != nil {
		panic(err)
	}
}
