package tview

import "github.com/gdamore/tcell"

// Configuration values.
const (
	FlexRow = iota
	FlexColumn
)

// flexItem holds layout options for one item.
type flexItem struct {
	Item      Primitive // The item to be positioned.
	FixedSize int       // The item's fixed size which may not be changed, 0 if it has no fixed size.
}

// Flex is a basic implementation of a flexbox layout.
type Flex struct {
	*Box

	// The items to be positioned.
	items []flexItem

	// FlexRow or FlexColumn.
	direction int

	// If set to true, will use the entire screen as its available space instead
	// its box dimensions.
	fullScreen bool
}

// NewFlex returns a new flexbox layout container with the given primitives.
// The items all have no fixed size. If more control is needed, call AddItem().
// The direction argument must be FlexRow or FlexColumn.
func NewFlex() *Flex {
	f := &Flex{
		Box:       NewBox(),
		direction: FlexColumn,
	}
	f.focus = f
	return f
}

// SetDirection sets the direction in which the contained primitives are
// distributed. This can be either FlexColumn (default) or FlexRow.
func (f *Flex) SetDirection(direction int) *Flex {
	f.direction = direction
	return f
}

// SetFullScreen sets the flag which, when true, causes the flex layout to use
// the entire screen space instead of whatever size it is currently assigned to.
func (f *Flex) SetFullScreen(fullScreen bool) *Flex {
	f.fullScreen = fullScreen
	return f
}

// AddItem adds a new item to the container. fixedSize is a size that may not be
// changed. A value of 0 means that its size may be changed.
func (f *Flex) AddItem(item Primitive, fixedSize int) *Flex {
	f.items = append(f.items, flexItem{Item: item, FixedSize: fixedSize})
	return f
}

// Draw draws this primitive onto the screen.
func (f *Flex) Draw(screen tcell.Screen) {
	// Calculate size and position of the items.

	// Do we use the entire screen?
	if f.fullScreen {
		f.x = 0
		f.y = 0
		width, height := screen.Size()
		f.width = width
		f.height = height
	}

	// How much space can we distribute?
	var variables int
	distSize := f.width
	if f.direction == FlexRow {
		distSize = f.height
	}
	for _, item := range f.items {
		if item.FixedSize > 0 {
			distSize -= item.FixedSize
		} else {
			variables++
		}
	}

	// Calculate positions and draw items.
	pos := f.x
	if f.direction == FlexRow {
		pos = f.y
	}
	for _, item := range f.items {
		size := item.FixedSize
		if size <= 0 {
			size = distSize / variables
			distSize -= size
			variables--
		}
		if f.direction == FlexColumn {
			item.Item.SetRect(pos, f.y, size, f.height)
		} else {
			item.Item.SetRect(f.x, pos, f.width, size)
		}
		pos += size

		if item.Item.GetFocusable().HasFocus() {
			defer item.Item.Draw(screen)
		} else {
			item.Item.Draw(screen)
		}
	}
}

// Focus is called when this primitive receives focus.
func (f *Flex) Focus(delegate func(p Primitive)) {
	if len(f.items) > 0 {
		delegate(f.items[0].Item)
	}
}

// HasFocus returns whether or not this primitive has focus.
func (f *Flex) HasFocus() bool {
	for _, item := range f.items {
		if item.Item.GetFocusable().HasFocus() {
			return true
		}
	}
	return false
}
