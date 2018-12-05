// Demo code for the Checkbox primitive.
package main

import "github.com/ravenops/tview"

func main() {
	app := tview.NewApplication()
	checklist := tview.NewChecklist("A", "B", "C")
	if err := app.SetRoot(checklist, true).Run(); err != nil {
		panic(err)
	}
}
