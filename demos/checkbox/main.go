// Demo code for the Checkbox primitive.
package main

import "github.com/rivo/tview"

func main() {
	app := tview.NewApplication()
	checkbox := tview.NewCheckbox().SetLabel("Hit Enter to check box: ").SetMessage("start race")
	if err := app.SetRoot(checkbox, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
