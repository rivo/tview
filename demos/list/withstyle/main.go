// Demo code for the List primitive with style.
package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	var mStyle, sStyle tcell.Style

	app := tview.NewApplication()
	list := tview.NewList()

	mStyle = tcell.StyleDefault.Foreground(tcell.ColorBlue)
	sStyle = tcell.StyleDefault.Foreground(tcell.ColorBeige)
	list.AddItemWithStyle("List item 1", "Some explanatory text", 'a', nil, mStyle, sStyle)

	mStyle = tcell.StyleDefault.Foreground(tcell.ColorPink)
	sStyle = tcell.StyleDefault.Foreground(tcell.ColorDarkGreen)
	list.AddItemWithStyle("List item 2", "Some explanatory text", 'b', nil, mStyle, sStyle)

	mStyle = tcell.StyleDefault.Foreground(tcell.ColorViolet)
	sStyle = tcell.StyleDefault.Foreground(tcell.ColorSlateGrey)
	list.AddItemWithStyle("List item 3", "Some explanatory text", 'c', nil, mStyle, sStyle)

	mStyle = tcell.StyleDefault.Foreground(tcell.ColorGold)
	sStyle = tcell.StyleDefault.Foreground(tcell.ColorWhite)
	list.AddItemWithStyle("List item 4", "Some explanatory text", 'd', nil, mStyle, sStyle)

	mStyle = tcell.StyleDefault.Foreground(tcell.ColorRed)
	sStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack)
	list.AddItemWithStyle("Quit", "Press to exit", 'q', func() {
		app.Stop()
	}, mStyle, sStyle)
	if err := app.SetRoot(list, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
