package tview

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// TestTableInputHandlerNoSelectableCells covers the case where every cell in
// the table is non-selectable. Before the forward/backwards clamping, the
// down-arrow handler entered an unbounded loop in this scenario because a
// previous Draw left t.selectedRow == rowCount, which the input handler then
// passed as an unreachable finalRow target to backwards. See rivo/tview#1146.
func TestTableInputHandlerNoSelectableCells(t *testing.T) {
	table := NewTable().SetSelectable(true, false)
	for i, col := range []string{"cpu", "memory", "disk"} {
		table.SetCell(0, i, NewTableCell(col).SetSelectable(false))
	}

	// Simulate the state Draw leaves behind when nothing is selectable: it
	// walks selectedRow past the last row.
	table.selectedRow = table.GetRowCount()
	table.selectedColumn = 0

	done := make(chan struct{})
	go func() {
		table.InputHandler()(tcell.NewEventKey(tcell.KeyDown, ' ', 0), func(Primitive) {})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("InputHandler hung on KeyDown when every cell is non-selectable")
	}
}
