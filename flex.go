package tview

import "github.com/gdamore/tcell"

// Configuration values.
const (
	FlexRow = iota
	FlexColumn
)

// FlexItem holds layout options for one item.
type FlexItem struct {
	Item      Primitive // The item to be positioned.
	FixedSize int       // The item's fixed size which may not be changed, 0 if it has no fixed size.
}

// Flex is a basic implementation of a flexbox layout.
type Flex struct {
	x, y, width, height int        // The size and position of this primitive.
	Items               []FlexItem // The items to be positioned.
	Direction           int        // FlexRow or FlexColumn.
}

// NewFlex returns a new flexbox layout container with the given primitives.
// The items all have no fixed size. If more control is needed, call AddItem().
// The direction argument must be FlexRow or FlexColumn.
func NewFlex(direction int, items []Primitive) *Flex {
	box := &Flex{
		width:     15,
		height:    10,
		Direction: direction,
	}
	for _, item := range items {
		box.Items = append(box.Items, FlexItem{Item: item})
	}
	return box
}

// AddItem adds a new item to the container. fixedSize is a size that may not be
// changed. A value of 0 means that its size may be changed.
func (f *Flex) AddItem(item Primitive, fixedSize int) *Flex {
	f.Items = append(f.Items, FlexItem{Item: item, FixedSize: fixedSize})
	return f
}

// Draw draws this primitive onto the screen.
func (f *Flex) Draw(screen tcell.Screen) {
	// Calculate size and position of the items.

	// How much space can we distribute?
	var variables int
	distSize := f.width
	if f.Direction == FlexRow {
		distSize = f.height
	}
	for _, item := range f.Items {
		if item.FixedSize > 0 {
			distSize -= item.FixedSize
		} else {
			variables++
		}
	}

	// Calculate positions and draw items.
	pos := f.x
	if f.Direction == FlexRow {
		pos = f.y
	}
	for _, item := range f.Items {
		size := item.FixedSize
		if size <= 0 {
			size = distSize / variables
			distSize -= size
			variables--
		}
		if f.Direction == FlexColumn {
			item.Item.SetRect(pos, f.y, size, f.height)
		} else {
			item.Item.SetRect(f.x, pos, f.width, size)
		}
		pos += size

		item.Item.Draw(screen)
	}
}

// GetRect returns the current position of the primitive, x, y, width, and
// height.
func (f *Flex) GetRect() (int, int, int, int) {
	return f.x, f.y, f.width, f.height
}

// SetRect sets a new position of the primitive.
func (f *Flex) SetRect(x, y, width, height int) {
	f.x = x
	f.y = y
	f.width = width
	f.height = height
}

// InputHandler returns nil.
func (f *Flex) InputHandler() func(event *tcell.EventKey) {
	return nil
}

// Focus is called when this primitive receives focus.
func (f *Flex) Focus(app *Application) {
}

// Blur is called when this primitive loses focus.
func (f *Flex) Blur() {
}
