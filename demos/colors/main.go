// Demo code for the Flex primitive.
package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Color struct {
	Name  string
	Color tcell.Color
}

func main() {
	app := tview.NewApplication()

	colorsCount := len(tcell.ColorNames)
	colorsCountPerRow := 6
	rowsCount := colorsCount / colorsCountPerRow
	lastRowCount := colorsCount % colorsCountPerRow

	colors := make([]Color, 0, colorsCount)
	for name, color := range tcell.ColorNames {
		colors = append(colors, Color{Name: name, Color: color})
	}

	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	for i := 0; i < rowsCount-1; i++ {
		row := tview.NewFlex().SetDirection(tview.FlexColumn)
		for j := 0; j < colorsCountPerRow; j++ {
			color := colors[i*colorsCountPerRow+j]
			box := tview.NewBox().SetBorder(true).SetTitle(color.Name).SetBackgroundColor(color.Color)
			row.AddItem(box, 0, 1, false)
		}
		flex.AddItem(row, 0, 1, false)
	}

	row := tview.NewFlex().SetDirection(tview.FlexColumn)
	for j := 0; j < lastRowCount; j++ {
		color := colors[rowsCount*colorsCountPerRow+j]
		box := tview.NewBox().SetBorder(true).SetTitle(color.Name).SetBackgroundColor(color.Color)
		row.AddItem(box, 0, 1, false)
	}
	flex.AddItem(row, 0, 1, false)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
