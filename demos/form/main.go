// Demo code for the Form primitive.
package main

import (
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	maleTitles := []string{"Mr.", "Dr.", "Prof."}
	femaleTitles := []string{"Ms.", "Mrs.", "Dr.", "Prof."}

	form := tview.NewForm()
	form.AddDropDown("Title", maleTitles, 0, nil).
		AddInputField("First name", "", 20, nil, nil).
		AddInputField("Last name", "", 20, nil, nil).
		AddTextArea("Address", "", 40, 0, 0, nil).
		AddRadio("Sex", 0, true, func(newValue int) {
			dd := form.GetFormItem(0).(*tview.DropDown)
			if newValue == 0 {
				dd.SetOptions(maleTitles, nil)
			} else {
				dd.SetOptions(femaleTitles, nil)
			}
		}, "male", "female").
		AddTextView("Notes", "This is just a demo.\nYou can enter whatever you wish.\nMind how the radio changes title options", 40, 3, true, false).
		AddCheckbox("Age 18+", false, nil).
		AddPasswordField("Password", "", 10, '*', nil).
		AddButton("Save", nil).
		AddButton("Quit", func() {
			app.Stop()
		})
	form.SetBorder(true).SetTitle("Enter some data").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).EnableMouse(true).EnablePaste(true).Run(); err != nil {
		panic(err)
	}
}
