package tview

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func newTestScreen(t *testing.T, width, height int) tcell.SimulationScreen {
	t.Helper()
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to initialize simulation screen: %v", err)
	}
	screen.SetSize(width, height)
	return screen
}

func sendKey(table *Table, key tcell.Key) {
	handler := table.InputHandler()
	handler(tcell.NewEventKey(key, 0, tcell.ModNone), nil)
}

func TestTableHorizontalScrollRemainingColumnsFitRowsSelectable(t *testing.T) {
	screen := newTestScreen(t, 50, 20)
	table := NewTable().
		SetFixed(1, 0).
		SetSelectable(true, false)

	widths := []int{30, 19, 10, 5}
	for row := 0; row < 20; row++ {
		for col, width := range widths {
			text := strings.Repeat(string(rune('a'+col)), width)
			table.SetCell(row, col, NewTableCell(text))
		}
	}
	table.SetRect(0, 0, 50, 20)

	table.Draw(screen)
	_, colOffset := table.GetOffset()
	if colOffset != 0 {
		t.Fatalf("initial column offset expected 0, got %d", colOffset)
	}

	// Scroll right. The remaining columns (1, 2, 3) total 19+10+5+2 = 36 <= 50,
	// so column offset should advance to 1 and stay there.
	sendKey(table, tcell.KeyRight)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("after right arrow, column offset expected 1, got %d", colOffset)
	}

	// Selecting another row triggers a redraw; column offset must not jump back to 0.
	sendKey(table, tcell.KeyDown)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("after down arrow, column offset expected 1, got %d", colOffset)
	}

	// An explicit redraw must also preserve column offset.
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("after redraw, column offset expected 1, got %d", colOffset)
	}

	// Pressing right arrow again when remaining columns already fit should stay at 1.
	sendKey(table, tcell.KeyRight)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("after second right arrow, column offset expected 1, got %d", colOffset)
	}

	// Pressing left arrow should scroll back to column offset 0.
	sendKey(table, tcell.KeyLeft)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 0 {
		t.Fatalf("after left arrow, column offset expected 0, got %d", colOffset)
	}
}

func TestTableHorizontalScrollSelectableCellsHiddenColumn(t *testing.T) {
	screen := newTestScreen(t, 50, 20)
	table := NewTable().
		SetFixed(1, 0).
		SetSelectable(true, true)

	widths := []int{30, 19, 10, 5}
	for row := 0; row < 20; row++ {
		for col, width := range widths {
			text := strings.Repeat(string(rune('a'+col)), width)
			table.SetCell(row, col, NewTableCell(text))
		}
	}
	table.SetRect(0, 0, 50, 20)

	table.Draw(screen)
	_, colOffset := table.GetOffset()
	if colOffset != 0 {
		t.Fatalf("initial column offset expected 0, got %d", colOffset)
	}

	// Move selection to column 1 (still visible).
	sendKey(table, tcell.KeyRight)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 0 {
		t.Fatalf("at column 1, column offset expected 0, got %d", colOffset)
	}

	// Move selection to column 2 (hidden on the right initially).
	sendKey(table, tcell.KeyRight)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("at column 2, column offset expected 1, got %d", colOffset)
	}

	// Redraw must keep column 2 visible and not jump back to 0.
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("after redraw with column 2 selected, column offset expected 1, got %d", colOffset)
	}

	// Move selection to column 3.
	sendKey(table, tcell.KeyRight)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("at column 3, column offset expected 1, got %d", colOffset)
	}

	// Move row selection down, column offset should remain stable.
	sendKey(table, tcell.KeyDown)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("after down arrow at column 3, column offset expected 1, got %d", colOffset)
	}

	// Move selection back to column 0.
	sendKey(table, tcell.KeyLeft) // col 2
	table.Draw(screen)
	sendKey(table, tcell.KeyLeft) // col 1
	table.Draw(screen)
	sendKey(table, tcell.KeyLeft) // col 0
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 0 {
		t.Fatalf("after returning to column 0, column offset expected 0, got %d", colOffset)
	}
}

func TestTableHorizontalScrollWithFixedColumns(t *testing.T) {
	screen := newTestScreen(t, 60, 20)
	table := NewTable().
		SetFixed(1, 1).
		SetSelectable(true, false)

	// Fixed column 0 has width 10.
	// Columns 1..4 have widths 30, 19, 10, 5. Total width = 10 + 1 + 30 + 1 + 19 + 1 + 10 + 1 + 5 = 78.
	widths := []int{10, 30, 19, 10, 5}
	for row := 0; row < 20; row++ {
		for col, width := range widths {
			text := strings.Repeat(string(rune('a'+col)), width)
			table.SetCell(row, col, NewTableCell(text))
		}
	}
	table.SetRect(0, 0, 60, 20)

	table.Draw(screen)
	_, colOffset := table.GetOffset()
	if colOffset != 0 {
		t.Fatalf("initial column offset expected 0, got %d", colOffset)
	}

	// Scroll right: fixed column 0 is retained, non-fixed columns start from 2.
	// Non-fixed columns 2, 3, 4 width = 19 + 1 + 10 + 1 + 5 = 36. Plus fixed col 10 + 1 = 47 <= 60.
	sendKey(table, tcell.KeyRight)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("after right arrow, column offset expected 1, got %d", colOffset)
	}

	// Redraw must remain stable at offset 1.
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("after redraw, column offset expected 1, got %d", colOffset)
	}

	sendKey(table, tcell.KeyDown)
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 1 {
		t.Fatalf("after row move, column offset expected 1, got %d", colOffset)
	}
}

func TestTableHorizontalScrollMuchWiderThanScreen(t *testing.T) {
	screen := newTestScreen(t, 50, 20)
	table := NewTable().
		SetFixed(0, 0).
		SetSelectable(true, false)

	// 10 columns of width 15 each. Total width = 15*10 + 9 = 159 > 50.
	for row := 0; row < 10; row++ {
		for col := 0; col < 10; col++ {
			text := strings.Repeat(string(rune('0'+col)), 15)
			table.SetCell(row, col, NewTableCell(text))
		}
	}
	table.SetRect(0, 0, 50, 20)

	table.Draw(screen)
	_, colOffset := table.GetOffset()
	if colOffset != 0 {
		t.Fatalf("initial column offset expected 0, got %d", colOffset)
	}

	for step := 1; step <= 5; step++ {
		sendKey(table, tcell.KeyRight)
		table.Draw(screen)
		_, colOffset = table.GetOffset()
		if colOffset != step {
			t.Fatalf("step %d: column offset expected %d, got %d", step, step, colOffset)
		}
	}
}

func TestTableSingleColumnWiderThanScreen(t *testing.T) {
	screen := newTestScreen(t, 50, 20)
	table := NewTable().
		SetFixed(0, 0).
		SetSelectable(true, false)

	table.SetCell(0, 0, NewTableCell(strings.Repeat("x", 100)))
	table.SetRect(0, 0, 50, 20)

	table.Draw(screen)
	_, colOffset := table.GetOffset()
	if colOffset != 0 {
		t.Fatalf("column offset expected 0, got %d", colOffset)
	}

	// Redraw does not change offset.
	table.Draw(screen)
	_, colOffset = table.GetOffset()
	if colOffset != 0 {
		t.Fatalf("after redraw, column offset expected 0, got %d", colOffset)
	}
}
