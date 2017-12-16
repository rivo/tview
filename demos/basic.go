package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	var list *tview.List

	frame := tview.NewFrame(tview.NewForm().
		AddItem("First name", "", 20, nil).
		AddItem("Last name", "", 20, nil).
		AddItem("Age", "", 4, nil).
		AddButton("Save", func() { app.Stop() }).
		AddButton("Cancel", nil).
		AddButton("Go to list", func() { app.SetFocus(list) })).
		AddText("Customer details", true, tview.AlignLeft, tcell.ColorRed).
		AddText("Customer details", false, tview.AlignCenter, tcell.ColorRed)
	frame.SetBorder(true)

	list = tview.NewList().
		AddItem("Edit a form", "You can do whatever you want", 'e').
		AddItem("Quit the program", "Do it!", 0).
		SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			if shortcut == 'e' {
				app.SetFocus(frame)
			}
		})
	list.SetBorder(true)

	flex := tview.NewFlex(tview.FlexColumn, []tview.Primitive{
		frame,
		tview.NewFlex(tview.FlexRow, []tview.Primitive{
			list,
			tview.NewBox().SetBorder(true).SetTitle("Third"),
		}),
		tview.NewBox().SetBorder(true).SetTitle("Fourth"),
	})
	flex.AddItem(tview.NewBox().SetBorder(true).SetTitle("Fifth"), 20)

	inputField := tview.NewInputField().
		SetLabel("Type something: ").
		SetFieldLength(10).
		SetAcceptanceFunc(tview.InputFieldFloat)
	inputField.SetBorder(true).SetTitle("Type!")

	final := tview.NewFlex(tview.FlexRow, []tview.Primitive{flex})
	final.AddItem(inputField, 3)

	app.SetRoot(final, true).SetFocus(list)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
