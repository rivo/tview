package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// End shows the final slide.
func End(nextSlide func()) (title string, content tview.Primitive) {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetDoneFunc(func(key tcell.Key) {
			nextSlide()
		})
	url := "[:::https://github.com/rivo/tview]https://github.com/rivo/tview"
	fmt.Fprint(textView, url)
	return "End", Center(tview.TaggedStringWidth(url), 1, textView)
}
