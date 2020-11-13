package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Flex demonstrates flexbox layout.
func Flex(nextSlide func()) (title string, content tview.Primitive) {
	modalShown := false
	pages := tview.NewPages()
	flex := tview.NewFlex().
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Flexible width, twice of middle column"), 0, 2, true).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Flexible width"), 0, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Fixed height"), 15, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Flexible height"), 0, 1, false), 0, 1, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Fixed width"), 30, 1, false)
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if modalShown {
			nextSlide()
			modalShown = false
		} else {
			pages.ShowPage("modal")
			modalShown = true
		}
		return event
	})
	modal := tview.NewModal().
		SetText("Resize the window to see the effect of the flexbox parameters").
		AddButtons([]string{"Ok"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.HidePage("modal")
	})
	pages.AddPage("flex", flex, true, true).
		AddPage("modal", modal, false, false)
	return "Flex", pages
}
