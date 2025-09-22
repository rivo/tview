// Demo code for the DropDown primitive.
package main

import "github.com/rivo/tview"

func main() {
	app := tview.NewApplication()
	dropdown := tview.NewMultiSelectDropDown().
		SetLabel("Select an option (hit Enter): ").
		SetOptions([]string{"First Item", "Second Item", "Third Item", "Fourth Item", "Fifth Item"}, nil)
	if err := app.SetRoot(dropdown, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
