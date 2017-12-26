package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func main() {
	l, _ := os.Create("/tmp/tview.log")
	defer l.Close()
	log.SetOutput(l)

	app := tview.NewApplication()
	pages := tview.NewPages()
	list := tview.NewList()

	app.SetKeyCapture(tcell.KeyCtrlQ, 0, func(p tview.Primitive) bool {
		app.Stop()
		return false
	})

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

	textView := tview.NewTextView()
	textView.SetWrap(true).
		SetDynamicColors(true).
		SetScrollable(true).
		SetRegions(true).
		SetChangedFunc(func() { app.Draw() }).
		SetDoneFunc(func(key tcell.Key) { textView.ScrollToHighlight(); app.SetFocus(list) })
	textView.SetBorder(true).SetTitle("Text view")
	go func() {
		for i := 0; i < 200; i++ {
			fmt.Fprintf(textView, "[\"%d\"]%d\n", i, i)
		}
		textView.Highlight("199")
	}()

	frame := tview.NewFrame(list).AddText("Choose!", true, tview.AlignCenter, tcell.ColorRed)
	frame.SetBorder(true)

	table := tview.NewTable().SetBorders(true).SetSeparator(tview.GraphicsVertBar).SetSelectable(false, false)
	lorem := strings.Split("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.", " ")
	cols, rows := 20, 120
	word := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			color := tcell.ColorWhite
			if c < 2 || r < 2 {
				color = tcell.ColorYellow
			}
			table.SetCell(r, c, &tview.TableCell{
				Text:  lorem[word],
				Color: color,
				Align: tview.AlignCenter,
			})
			word++
			if word >= len(lorem) {
				word = 0
			}
		}
	}
	table.SetSelected(0, 0).SetFixed(2, 2).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.SetFocus(list)
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {
		cell := table.GetCell(row, column)
		cell.Color = tcell.ColorRed
		table.SetSelectable(false, false)
	})
	table.SetBorder(true).SetBorderPadding(1, 1, 1, 1)

	list.AddItem("Edit a form", "You can do whatever you want", 'e', func() { app.SetFocus(form) }).
		AddItem("Navigate text", "Try all the navigations", 't', func() { app.SetFocus(textView) }).
		AddItem("Navigate table", "Rows and columns", 'a', func() { app.SetFocus(table) }).
		AddItem("Quit the program", "Do it!", 0, func() { app.Stop() })

	flex := tview.NewFlex().
		AddItem(form, 0).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(frame, 0).
			AddItem(textView, 0), 0).
		AddItem(table, 0).
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

	app.SetRoot(pages).SetFocus(list)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
