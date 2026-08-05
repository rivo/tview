// Demo code for the Form primitive.
package main

import (
	"fmt"

	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	form := tview.NewForm().
		AddDropDown("Title", []string{"Mr.", "Ms.", "Mrs.", "Dr.", "Prof."}, 0, nil).
		AddInputField("First name", "", 20, nil, nil).
		AddInputField("Last name", "", 20, nil, nil).
		AddTextArea("Address", "", 40, 0, 0, nil).
		AddTextView("Notes", "This is just a demo.\nYou can enter whatever you wish.", 40, 2, true, false).
		AddCheckbox("Age 18+", false, nil).
		AddPasswordField("Password", "", 10, '*', nil).
		AddButton("Save", func() {
			app.Stop()
			fmt.Println("User clicked Save")
		}).
		AddButton("Quit", func() {
			app.Stop()
			fmt.Println("User clicked Quit")
		}).
		SetCancelFunc(func() {
			app.Stop()
			fmt.Println("Form entry was canceled")
		}).
		SetSubmitFunc(func() {
			app.Stop()
			fmt.Println("Form was submitted")
		})
	form.SetBorder(true).SetTitle("Enter some data").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
		panic(err)
	}
}
