package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	form := tview.NewFrame(tview.NewForm().
		AddItem("First name", "", 20, nil).
		AddItem("Last name", "", 20, nil).
		AddItem("Age", "", 4, nil).
		AddButton("Save", func() { app.Stop() }).
		AddButton("Cancel", nil)).
		AddText("Customer details", true, tview.AlignLeft, tcell.ColorRed)
	form.SetBorder(true)

	box := tview.NewFlex(tview.FlexColumn, []tview.Primitive{
		form,
		tview.NewFlex(tview.FlexRow, []tview.Primitive{
			tview.NewBox().SetBorder(true).SetTitle("Second"),
			tview.NewBox().SetBorder(true).SetTitle("Third"),
		}),
		tview.NewBox().SetBorder(true).SetTitle("Fourth"),
	})
	box.AddItem(tview.NewBox().SetBorder(true).SetTitle("Fifth"), 20)

	inputField := tview.NewInputField().
		SetLabel("Type something: ").
		SetFieldLength(10).
		SetAcceptanceFunc(tview.InputFieldFloat)
	inputField.SetBorder(true).SetTitle("Type!")

	final := tview.NewFlex(tview.FlexRow, []tview.Primitive{box})
	final.AddItem(inputField, 3)

	app.SetRoot(final, true).SetFocus(form)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
