// Demo code for the finder primitive.
package main

import (
	"fmt"
	"github.com/rivo/tview"
)

type Item struct {
	Title string
}

func (i *Item) preview() string {
	return fmt.Sprintf("\033[1;33;4;44mTitle: %s\033[0m", i.Title)
}

func main() {

	app := tview.NewApplication()

	items := []*Item{
		{Title: "foo"},
		{Title: "bar"},
		{Title: "foo foo"},
		{Title: "baz one - this is a very long item which should overflow horizontally at some point"},
		{Title: "baz two"},
		{Title: "lorem ipsum"},
	}

	selectedView := tview.NewTextView()
	selectedView.SetBorder(true)
	selectedView.SetTitle("Selected item")
	selectedView.SetDynamicColors(true)

	finder := tview.NewFinder().
		SetItems(len(items), func(index int) string {
			return items[index].Title
		}).
		SetDoneFunc(func(index int) {
			app.Stop()
			fmt.Printf("Selected index: %d", index)
		}).
		SetChangedFunc(func(index int) {
			if index >= 0 {
				selectedView.SetText("")
				writer := tview.ANSIWriter(selectedView)
				_, _ = writer.Write([]byte(items[index].preview()))
			} else {
				selectedView.SetText("")
			}
		})

	finder.
		SetBorder(true).
		SetTitle("Pick an item")

	flex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(finder, 0, 1, true).
		AddItem(selectedView, 0, 1, false)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
