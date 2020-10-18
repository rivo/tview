// Demo code for the List primitive.
package main

import (
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	list := tview.NewList()

	reset := func() {
		list.
			Clear().
			AddItem("List item 1", "Some explanatory text", 'a', nil).
			AddItem("List item 2", "Some explanatory text", 'b', nil).
			AddItem("List item 3", "Some explanatory text", 'c', nil).
			AddItem("List item 4", "Some explanatory text", 'd', nil).
			AddItem("Quit", "Press to exit", 'q', func() {
				app.Stop()
			})

		list.ContextMenuList().SetItemEnabled(3, false)
	}

	list.AddContextItem("Delete item", 'i', func(index int) {
		list.RemoveItem(index)

		if list.GetItemCount() == 0 {
			list.ContextMenuList().SetItemEnabled(0, false)
			list.ContextMenuList().SetItemEnabled(1, false)
		}
		list.ContextMenuList().SetItemEnabled(3, true)
	})

	list.AddContextItem("Delete all", 'a', func(index int) {
		list.Clear()

		list.ContextMenuList().SetItemEnabled(0, false)
		list.ContextMenuList().SetItemEnabled(1, false)
		list.ContextMenuList().SetItemEnabled(3, true)
	})

	list.AddContextItem("", 0, nil)

	list.AddContextItem("Reset", 'r', func(index int) {
		reset()

		list.ContextMenuList().SetItemEnabled(0, true)
		list.ContextMenuList().SetItemEnabled(1, true)
		list.ContextMenuList().SetItemEnabled(3, false)
	})

	reset()
	if err := app.SetRoot(list, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
