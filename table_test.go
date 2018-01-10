package tview

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell"
)

func TestGetLastPosition(t *testing.T) {
	tests := []struct {
		c     *TableCell
		x     int
		y     int
		width int
	}{
		{&TableCell{x: 0, y: 0, width: 0}, 0, 0, 0},
		{&TableCell{x: 1, y: 1, width: 1}, 1, 1, 1},
		{&TableCell{x: -1, y: -1, width: -1}, -1, -1, -1},
	}
	for i, tt := range tests {
		x, y, w := tt.c.GetLastPosition()
		if x != tt.x {
			t.Errorf("case: %d, expected: %d, actual %d\n", i, x, tt.x)
		}
		if y != tt.y {
			t.Errorf("case: %d, expected: %d, actual %d\n", i, y, tt.y)
		}
		if w != tt.width {
			t.Errorf("case: %d, expected: %d, actual %d\n", i, w, tt.width)
		}
	}
}

func TestNewTable(t *testing.T) {
	//testNewTable
	table := NewTable()
	if !reflect.DeepEqual(table.Box, NewBox()) {
		t.Error("table.Box is wrong")
	}
	if table.borders {
		t.Error("table.borders is true")
	}
	if !reflect.DeepEqual(table.bordersColor, tcell.ColorWhite) {
		t.Error("table.bordersColer is wrong")
	}
	if table.separator != rune(' ') {
		t.Error("table.separator is not ' '")
	}
	if table.cells != nil {
		t.Error("table.cells not nil")
	}
	if table.lastColumn != -1 {
		t.Error("table.lastColumn is wrong")
	}
	if table.fixedRows != 0 {
		t.Error("table.fixedRows is wrong")
	}
	if table.fixedColumns != 0 {
		t.Error("table.fixedColumns is wrong")
	}
	if table.rowsSelectable {
		t.Error("table.RowsSelectedRow is true")
	}
	if table.columnsSelectable {
		t.Error("table.fixedRows is true")
	}
	if table.rowOffset != 0 {
		t.Error("table.rowOffset is wrong")
	}
	if table.columnOffset != 0 {
		t.Error("table.ColumnOffset is wrong")
	}
	if table.trackEnd {
		t.Error("table.trackEnd is true")
	}
	if table.visibleRows != 0 {
		t.Error("table.visibleRows is wrong")
	}
	if table.selected != nil {
		t.Error("table.selected is not nil")
	}
	if table.done != nil {
		t.Error("table.Done is not nil")
	}
}

func TestTableClear(t *testing.T) {
	tables := []*Table{
		NewTable(),
		NewTable(),
		NewTable(),
	}
	tables[1].cells = make([][]*TableCell, 0)
	tables[2].lastColumn = 0
	for i, tt := range tables {
		table := tt.Clear()
		if table.cells != nil {
			t.Errorf("case: %d, table.cells not nil", i)
		}
		if table.lastColumn != -1 {
			t.Errorf("case %d, table.lastColumn is wrong", i)
		}
	}
}
