package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const logo = `
   __        _
  / /__   __(_)__ _      __
 / __/ | / / / _ \ | /| / /
/ /_ | |/ / /  __/ |/ |/ /
\__/ |___/_/\___/|__/|__/

`

const (
	subtitle   = `tview - Rich Widgets for Terminal UIs`
	navigation = `Ctrl-N: Next slide    Ctrl-P: Previous slide    Ctrl-C: Exit`
)

// Cover returns the cover page.
func Cover(nextSlide func()) (title string, content tview.Primitive) {
	// What's the size of the logo?
	lines := strings.Split(logo, "\n")
	logoWidth := 0
	logoHeight := len(lines)
	for _, line := range lines {
		if len(line) > logoWidth {
			logoWidth = len(line)
		}
	}
	logoBox := tview.NewTextView().
		SetTextColor(tcell.ColorGreen).
		SetDoneFunc(func(key tcell.Key) {
			nextSlide()
		})
	fmt.Fprint(logoBox, logo)

	// Create a frame for the subtitle and navigation infos.
	frame := tview.NewFrame(tview.NewBox()).
		SetBorders(0, 0, 0, 0, 0, 0).
		AddText(subtitle, true, tview.AlignCenter, tcell.ColorWhite).
		AddText("", true, tview.AlignCenter, tcell.ColorWhite).
		AddText(navigation, true, tview.AlignCenter, tcell.ColorDarkMagenta)

	// Create a Flex layout that centers the logo and subtitle.
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 7, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(logoBox, logoWidth, 1, true).
			AddItem(tview.NewBox(), 0, 1, false), logoHeight, 1, true).
		AddItem(frame, 0, 10, false)

	return "Start", flex
}
