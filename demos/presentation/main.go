/*
Navigation

  - Ctrl-N: Jump to next slide
  - Ctrl-P: Jump to previous slide
*/
package main

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// Slide is a function which returns the slide's main primitive and its title.
// It receives a "nextSlide" function which can be called to advance the
// presentation to the next slide.
type Slide func(nextSlide func()) (title string, content tview.Primitive)

// The application.
var app = tview.NewApplication()

// Starting point for the presentation.
func main() {
	// The presentation slides.
	slides := []Slide{
		Cover,
		Introduction,
		HelloWorld,
		InputField,
		Form,
		TextView1,
		TextView2,
		Table,
		Flex,
		End,
	}

	// The bottom row has some info on where we are.
	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	// Create the pages for all slides.
	currentSlide := 0
	info.Highlight(strconv.Itoa(currentSlide))
	pages := tview.NewPages()
	previousSlide := func() {
		currentSlide = (currentSlide - 1 + len(slides)) % len(slides)
		info.Highlight(strconv.Itoa(currentSlide))
		pages.SwitchToPage(strconv.Itoa(currentSlide))
	}
	nextSlide := func() {
		currentSlide = (currentSlide + 1) % len(slides)
		info.Highlight(strconv.Itoa(currentSlide))
		pages.SwitchToPage(strconv.Itoa(currentSlide))
	}
	for index, slide := range slides {
		title, primitive := slide(nextSlide)
		pages.AddPage(strconv.Itoa(index), primitive, true, index == currentSlide)
		fmt.Fprintf(info, `%d ["%d"][cyan]%s[white][""]  `, index+1, index, title)
	}

	// Create the main layout.
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(info, 1, 1, false)

	// Shortcuts to navigate the slides.
	app.SetKeyCapture(tcell.KeyCtrlN, 0, func(p tview.Primitive) bool {
		nextSlide()
		return false
	}).SetKeyCapture(tcell.KeyCtrlP, 0, func(p tview.Primitive) bool {
		previousSlide()
		return false
	})

	// Start the application.
	if err := app.SetRoot(layout, true).SetFocus(layout).Run(); err != nil {
		panic(err)
	}
}
