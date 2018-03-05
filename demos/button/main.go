// Demo code for the Button primitive.
package main

import "github.com/rivo/tview"

func main() {
	app := tview.NewApplication()
	button := tview.NewButton("Hit Enter to close").SetSelectedFunc(func() {
		app.Stop()
	})
	button.SetBorder(true).SetRect(0, 0, 22, 3)
	if err := app.SetRoot(button, false).Run(); err != nil {
		panic(err)
	}
}
