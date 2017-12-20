package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	pages := tview.NewPages()
	var list *tview.List

	form := tview.NewForm().
		AddInputField("First name", "", 20, nil).
		AddInputField("Last name", "", 20, nil).
		AddInputField("Age", "", 4, nil).
		AddDropDown("Select", []string{"One", "Two", "Three"}, 1, func(text string, index int) {
			if text == "Three" {
				app.Stop()
			}
		}).
		AddCheckbox("Check", false, nil).
		AddButton("Save", func() {
			previous := app.GetFocus()
			modal := tview.NewModal().
				SetText("Would you really like to save this customer to the database?").
				AddButtons([]string{"Save", "Cancel"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					pages.RemovePage("confirm")
					app.SetFocus(previous)
					app.Draw()
				})
			pages.AddPage("confirm", modal, true)
			app.SetFocus(modal)
			app.Draw()
		}).
		AddButton("Cancel", nil).
		AddButton("Go to list", func() { app.SetFocus(list) }).
		SetCancelFunc(func() {
			app.Stop()
		})
	form.SetTitle("Customer").SetBorder(true)

	list = tview.NewList().
		AddItem("Edit a form", "You can do whatever you want", 'e', func() { app.SetFocus(form) }).
		AddItem("Quit the program", "Do it!", 0, func() { app.Stop() })

	frame := tview.NewFrame(list).AddText("Choose!", true, tview.AlignCenter, tcell.ColorRed)
	frame.SetBorder(true)

	flex := tview.NewFlex().
		AddItem(form, 0).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(frame, 0).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Third"), 0), 0).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Fourth"), 0).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Fifth"), 20)

	inputField := tview.NewInputField().
		SetLabel("Type something: ").
		SetFieldLength(10).
		SetAcceptanceFunc(tview.InputFieldFloat)
	inputField.SetBorder(true).SetTitle("Type!")

	final := tview.NewFlex().
		SetFullScreen(true).
		SetDirection(tview.FlexRow).
		AddItem(flex, 0).
		AddItem(inputField, 3)

	pages.AddPage("flex", final, true)

	app.SetRoot(pages, false).SetFocus(list)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
