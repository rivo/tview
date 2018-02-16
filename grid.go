package tview

import (
	"math"

	"github.com/gdamore/tcell"
)

// gridItem represents one primitive and its possible position on a grid.
type gridItem struct {
	Item                        Primitive // The item to be positioned. May be nil for an empty item.
	Row, Column                 int       // The top-left grid cell where the item is placed.
	Width, Height               int       // The number of rows and columns the item occupies.
	MinGridWidth, MinGridHeight int       // The minimum grid width/height for which this item is visible.
	Focus                       bool      // Whether or not this item attracts the layout's focus.

	visible    bool // Whether or not this item was visible the last time the grid was drawn.
	x, y, w, h int  // The last position of the item relative to the top-left corner of the grid. Undefined if visible is false.
}

// Grid is an implementation of a grid-based layout. It works by defining the
// size of the rows and columns, then placing primitives into the grid.
//
// Some settings can lead to the grid exceeding its available space. SetOffset()
// can then be used to scroll in steps of rows and columns. These offset values
// can also be controlled with the arrow keys while the grid has focus and none
// of its contained primitives do.
//
// See https://github.com/rivo/tview/wiki/Grid for an example.
type Grid struct {
	*Box

	// The items to be positioned.
	items []*gridItem

	// The definition of the rows and columns of the grid. See
	// SetRows()/SetColumns() for details.
	rows, columns []int

	// The minimum sizes for rows and columns.
	minWidth, minHeight int

	// The size of the gaps between neighboring primitives. This is automatically
	// set to 1 if borders is true.
	gapRows, gapColumns int

	// The number of rows and columns skipped before drawing the top-left corner
	// of the grid.
	rowOffset, columnOffset int

	// Whether or not borders are drawn around grid items. If this is set to true,
	// a gap size of 1 is automatically assumed (which is filled with the border
	// graphics).
	borders bool
}

// NewGrid returns a new grid-based layout container with no initial primitives.
func NewGrid() *Grid {
	g := &Grid{
		Box: NewBox(),
	}
	g.focus = g
	return g
}

// SetRows defines how the rows of the grid are distributed. Each value defines
// the size of one row, starting with the leftmost row. Values greater 0
// represent absolute row widths (gaps not included). Values less or equal 0
// represent proportional row widths or fractions of the remaining free space,
// where 0 is treated the same as -1. That is, a row with a value of -3 will
// have three times the width of a row with a value of -1 (or 0). The minimum
// width set with SetMinSize() is always observed.
//
// Primitives may extend beyond the rows defined explicitly with this function.
// A value of 0 is assumed for any undefined row. In fact, if you never call
// this function, all rows occupied by primitives will have the same width.
// On the other hand, unoccupied rows defined with this function will always
// take their place.
//
// Assuming a total width of the grid of 100 cells and a minimum width of 0, the
// following call will result in rows with widths of 30, 10, 15, 15, and 30
// cells:
//
//   grid.SetRows(30, 10, -1, -1, -2)
//
// If a primitive were then placed in the 6th and 7th row, the resulting widths
// would be: 30, 10, 10, 10, 20, 10, and 10 cells.
//
// If you then called SetMinSize() as follows:
//
//   grid.SetMinSize(15, 20)
//
// The resulting widths would be: 30, 15, 15, 15, 20, 15, and 15 cells, a total
// of 125 cells, 25 cells wider than the available grid width.
func (g *Grid) SetRows(rows ...int) *Grid {
	g.rows = rows
	return g
}

// SetColumns defines how the columns of the grid are distributed. These values
// behave the same as the row values provided with SetRows(), see there for
// a definition and examples.
//
// The provided values correspond to column heights, the first value definining
// the height of the topmost column.
func (g *Grid) SetColumns(columns ...int) *Grid {
	g.columns = columns
	return g
}

// SetMinSize sets an absolute minimum width for rows and an absolute minimum
// height for columns.
func (g *Grid) SetMinSize(row, column int) *Grid {
	g.minWidth, g.minHeight = row, column
	return g
}

// SetGap sets the size of the gaps between neighboring primitives on the grid.
// If borders are drawn (see SetBorders()), these values are ignored and a gap
// of 1 is assumed.
func (g *Grid) SetGap(row, column int) *Grid {
	g.gapRows, g.gapColumns = row, column
	return g
}

// SetBorders sets whether or not borders are drawn around grid items. Setting
// this value to true will cause the gap values (see SetGap()) to be ignored and
// automatically assumed to be 1 where the border graphics are drawn.
func (g *Grid) SetBorders(borders bool) *Grid {
	g.borders = borders
	return g
}

// AddItem adds a primitive and its position to the grid. The top-left corner
// of the primitive will be located in the top-left corner of the grid cell at
// the given row and column and will span "width" rows and "height" columns. For
// example, for a primitive to occupy rows 2, 3, and 4 and columns 5 and 6:
//
//   grid.AddItem(p, 2, 4, 3, 2, true)
//
// If width or height is 0, the primitive will not be drawn.
//
// You can add the same primitive multiple times with different grid positions.
// The minGridWidth and minGridHeight values will then determine which of those
// positions will be used. This is similar to CSS media queries. These minimum
// values refer to the overall size of the grid. If multiple items for the same
// primitive apply, the one with the highest minimum values (with a preference
// for the minimum width) will be used, or the primitive added last if those
// values are the same. Example:
//
//   grid.AddItem(p, 0, 0, 0, 0, 0, 0, true). // Hide in small grids.
//     AddItem(p, 0, 0, 1, 2, 100, 0, true).  // One-column layout for medium grids.
//     AddItem(p, 1, 1, 3, 2, 300, 0, true)   // Multi-column layout for large grids.
//
// To use the same grid layout for all sizes, simply set minGridWidth and
// minGridHeight to 0.
//
// If the item's focus is set to true, it will receive focus when the grid
// receives focus. If there are multiple items with a true focus flag, the last
// visible one that was added will receive focus.
func (g *Grid) AddItem(p Primitive, row, column, width, height, minGridWidth, minGridHeight int, focus bool) *Grid {
	return g
}

// SetOffset sets the number of rows and columns which are skipped before
// drawing the first grid cell in the top-left corner. As the grid will never
// completely move off the screen, these values may be adjusted the next time
// the grid is drawn. The actual position of the grid may also be adjusted such
// that contained primitives that have focus are visible.
func (g *Grid) SetOffset(rows, columns int) *Grid {
	g.rowOffset, g.columnOffset = rows, columns
	return g
}

// GetOffset returns the current row and column offset (see SetOffset() for
// details).
func (g *Grid) GetOffset() (rows, columns int) {
	return g.rowOffset, g.columnOffset
}

// Focus is called when this primitive receives focus.
func (g *Grid) Focus(delegate func(p Primitive)) {
	for _, item := range g.items {
		if item.visible && item.Focus {
			delegate(item.Item)
			return
		}
	}
	g.hasFocus = true
}

// Blur is called when this primitive loses focus.
func (g *Grid) Blur() {
	g.hasFocus = false
}

// HasFocus returns whether or not this primitive has focus.
func (g *Grid) HasFocus() bool {
	for _, item := range g.items {
		if item.visible && item.Item.GetFocusable().HasFocus() {
			return true
		}
	}
	return false
}

// InputHandler returns the handler for this primitive.
func (g *Grid) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return g.wrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'g':
				g.rowOffset, g.columnOffset = 0, 0
			case 'G':
				g.rowOffset = math.MaxInt32
			case 'j':
				g.rowOffset++
			case 'k':
				g.rowOffset--
			case 'h':
				g.columnOffset--
			case 'l':
				g.columnOffset++
			}
		case tcell.KeyHome:
			g.rowOffset, g.columnOffset = 0, 0
		case tcell.KeyEnd:
			g.rowOffset = math.MaxInt32
		case tcell.KeyUp:
			g.rowOffset--
		case tcell.KeyDown:
			g.rowOffset++
		case tcell.KeyLeft:
			g.columnOffset--
		case tcell.KeyRight:
			g.columnOffset++
		}
	})
}

// Draw draws this primitive onto the screen.
func (g *Grid) Draw(screen tcell.Screen) {
	g.Box.Draw(screen)
	//TODO
}
