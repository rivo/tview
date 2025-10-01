// Demo code for the Modal primitive.
package main

import (
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter)
	f.SetBorderPadding(0, 0, 0, 0)
	f.SetFocus(0)
	f.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	f.AddInputField("Name", "", 30, nil, nil)
	f.AddCheckbox("Is Active", false, nil)

	modal := tview.NewModalForm("< Modal Form >", f)
	modal.SetText("Do you want to quit the application?")
	modal.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	modal.AddButtons([]string{"Quit", "Cancel"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Quit" {
			app.Stop()
		}
	})

	if err := app.SetRoot(modal, false).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
