package main

import (
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const colorsText = `You can use color tags almost everywhere to partially change the color of a string. Simply put a color name or hex string in square brackets to change the following characters' color. H[green]er[white]e i[yellow]s a[darkcyan]n ex[red]amp[white]le. The [black:red]tags [black:green]look [black:yellow]like [::u]this: [blue:yellow:u[] [#00ff00[]`

// Colors demonstrates how to use colors.
func Colors(nextSlide func()) (title string, content tview.Primitive) {
	table := tview.NewTable().
		SetBorders(true).
		SetBordersColor(tcell.ColorBlue).
		SetDoneFunc(func(key tcell.Key) {
			nextSlide()
		})
	var row, column int
	for _, word := range strings.Split(colorsText, " ") {
		table.SetCellSimple(row, column, word)
		column++
		if column > 6 {
			column = 0
			row++
		}
	}
	table.SetBorderPadding(1, 1, 2, 2).
		SetBorder(true).
		SetTitle("A [red]c[yellow]o[green]l[darkcyan]o[blue]r[darkmagenta]f[red]u[yellow]l[white] [black:red]c[:yellow]o[:green]l[:darkcyan]o[:blue]r[:darkmagenta]f[:red]u[:yellow]l[white:] [::bu]title")
	return "Colors", Center(78, 19, table)
}
