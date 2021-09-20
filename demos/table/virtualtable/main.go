package main

import (
	"fmt"
	"math"

	"github.com/rivo/tview"
)

type TableData struct {
	tview.TableContentReadOnly
}

func (d *TableData) GetCell(row, column int) *tview.TableCell {
	letters := [...]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 'A' + byte(row%26)} // log(math.MaxInt64) / log(26) ~= 14
	start := len(letters) - 1
	row /= 26
	for row > 0 {
		start--
		row--
		letters[start] = 'A' + byte(row%26)
		row /= 26
	}
	return tview.NewTableCell(fmt.Sprintf("[red]%s[green]%d", letters[start:], column))
}

func (d *TableData) GetRowCount() int {
	return math.MaxInt64
}

func (d *TableData) GetColumnCount() int {
	return math.MaxInt64
}

func main() {
	data := &TableData{}
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, true).
		SetContent(data)
	if err := tview.NewApplication().SetRoot(table, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
