package main

import "github.com/rivo/tview"

// Introduction returns a tview.List with the highlights of the tview package.
func Introduction(nextSlide func()) (title string, content tview.Primitive) {
	list := tview.NewList().
		AddItem("A Go package for terminal based UIs", "with a special focus on rich interactive widgets", '1', nextSlide).
		AddItem("Based on github.com/gdamore/tcell", "Like termbox but better (see tcell docs)", '2', nextSlide).
		AddItem("Designed to be simple", `"Hello world" is 5 lines of code`, '3', nextSlide).
		AddItem("Good for data entry", `For charts, use "termui" - for low-level views, use "gocui" - ...`, '4', nextSlide).
		AddItem("Extensive documentation", "Everything is documented, examples in GitHub wiki, demo code for each widget", '5', nextSlide)
	return "Introduction", Center(80, 10, list)
}
