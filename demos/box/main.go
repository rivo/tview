// Demo code for the Box primitive.
package main

import "github.com/rivo/tview"

func main() {
	box := tview.NewBox().
		SetBorder(true).
		SetTitle("Box Demo")
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		panic(err)
	}
}
