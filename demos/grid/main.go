package main

import "github.com/rivo/tview"

func main() {
	grid := tview.NewGrid().
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Top"), 0, 0, 1, 2, 0, 0, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Left"), 1, 0, 1, 1, 0, 0, true).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Right"), 1, 1, 1, 1, 0, 0, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom"), 2, 0, 1, 2, 0, 0, false)
	if err := tview.NewApplication().SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}
