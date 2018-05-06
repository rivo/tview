package tview

import (
	"math/rand"

	"github.com/gdamore/tcell"
)

// Box implements Primitive with a background and optional elements such as a
// border and a title. Most subclasses keep their content contained in the box
// but don't necessarily have to.
//
// Note that all classes which subclass from Box will also have access to its
// functions.
//
// See https://github.com/rivo/tview/wiki/Box for an example.
type Box struct {
	id int
	// The position of the rect.
	x, y, width, height int

	maxWidth, maxHeight, minWidth, minHeight int

	// The inner rect reserved for the box's content.
	innerX, innerY, innerWidth, innerHeight int

	// Border padding.
	paddingTop, paddingBottom, paddingLeft, paddingRight int

	// Border padding to be measured by percent.
	paddingPercentTop, paddingPercentBottom, paddingPercentLeft, paddingPercentRight int

	// The box's background color.
	backgroundColor tcell.Color

	// Whether or not a border is drawn, reducing the box's space for content by
	// two in width and height.
	border bool

	// The color of the border.
	borderColor tcell.Color

	// The title. Only visible if there is a border, too.
	title string

	// The color of the title.
	titleColor tcell.Color

	// The alignment of the title.
	titleAlign int

	// Provides a way to find out if this box has focus. We always go through
	// this interface because it may be overridden by implementing classes.
	focus Focusable

	// Whether or not this box has focus.
	hasFocus bool

	// If set to true, the inner rect of this box will be within the screen at the
	// last time the box was drawn.
	clampToScreen bool

	// An optional capture function which receives a key event and returns the
	// event to be forwarded to the primitive's default input handler (nil if
	// nothing should be forwarded).
	inputCapture func(event *tcell.EventKey) *tcell.EventKey

	// An optional function which is called before the box is drawn.
	draw func(screen tcell.Screen, x, y, width, height int) (int, int, int, int)
}

// NewBox returns a Box without a border.
func NewBox() *Box {
	b := &Box{
		id:              rand.Int(),
		width:           15,
		height:          10,
		innerX:          -1, // Mark as uninitialized.
		backgroundColor: Styles.PrimitiveBackgroundColor,
		borderColor:     Styles.BorderColor,
		titleColor:      Styles.TitleColor,
		titleAlign:      AlignCenter,
		clampToScreen:   true,
	}
	b.focus = b
	return b
}

// GetID returns unique id of box
func (b *Box) GetID() int {
	return b.id
}

// SetSize sets size of the box
func (b *Box) SetSize(width, height int) *Box {
	b.width = width
	b.height = height
	return b
}

// SetMinSize sets min size of the box
func (b *Box) SetMinSize(width, height int) *Box {
	b.minWidth = width
	b.minHeight = height
	return b
}

// SetMaxSize sets max size of the box
func (b *Box) SetMaxSize(width, height int) *Box {
	b.maxWidth = width
	b.maxHeight = height
	return b
}

// SetPercentPadding sets border in percents. This feature implements resizable paddingPercent of the box relative to the size of the screen
func (b *Box) SetPercentPadding(top, bottom, left, right int) *Box {
	b.paddingPercentTop, b.paddingPercentBottom, b.paddingPercentLeft, b.paddingPercentRight = top, bottom, left, right
	return b
}

// SetBorderPadding sets the size of the borders around the box content.
func (b *Box) SetBorderPadding(top, bottom, left, right int) *Box {
	b.paddingTop, b.paddingBottom, b.paddingLeft, b.paddingRight = top, bottom, left, right
	return b
}

// GetBorderPadding returns padding of the box
func (b *Box) GetBorderPadding() (top, bottom, left, right int) {
	outerX, outerY, outerWidth, outerHeight := b.GetRect()
	innerX, innerY, innerWidth, innerHeight := b.GetInnerRect()
	return innerY - outerY, (outerY + outerHeight) - (innerY + innerHeight), innerX - outerX, (outerX + outerWidth) - (innerX + innerWidth)
}

// GetRect returns the current position of the rectangle, x, y, width, and
// height.
func (b *Box) GetRect() (int, int, int, int) {
	width := b.width
	switch {
	case b.maxWidth > 0 && width > b.maxWidth:
		width = b.maxWidth
	case b.minWidth > 0 && width < b.minWidth:
		width = b.minWidth
	}

	height := b.height
	switch {
	case b.maxHeight > 0 && height > b.maxHeight:
		height = b.maxHeight
	case b.minHeight > 0 && height < b.minHeight:
		height = b.minHeight
	}

	return b.x, b.y, width, height
}

// GetInnerRect returns the position of the inner rectangle (x, y, width,
// height), without the border and without any padding.
func (b *Box) GetInnerRect() (int, int, int, int) {
	if b.innerX >= 0 {
		return b.innerX, b.innerY, b.innerWidth, b.innerHeight
	}
	x, y, width, height := b.GetRect()
	if b.border {
		x++
		y++
		width -= 2
		height -= 2
	}
	x, y, width, height = x+b.paddingLeft,
		y+b.paddingTop,
		width-b.paddingLeft-b.paddingRight,
		height-b.paddingTop-b.paddingBottom

	// Percent padding
	left := int(float64(width) / 100 * float64(b.paddingPercentLeft))
	top := int(float64(height) / 100 * float64(b.paddingPercentTop))
	right := int(float64(width) / 100 * float64(b.paddingPercentRight))
	bottom := int(float64(height) / 100 * float64(b.paddingPercentBottom))

	return x + left, y + top, width - left - right, height - top - bottom

}

// SetRect sets a new position of the primitive.
func (b *Box) SetRect(x, y, width, height int) {
	b.x = x
	b.y = y
	b.width = width
	b.height = height
}

// SetDrawFunc sets a callback function which is invoked after the box primitive
// has been drawn. This allows you to add a more individual style to the box
// (and all primitives which extend it).
//
// The function is provided with the box's dimensions (set via SetRect()). It
// must return the box's inner dimensions (x, y, width, height) which will be
// returned by GetInnerRect(), used by descendent primitives to draw their own
// content.
func (b *Box) SetDrawFunc(handler func(screen tcell.Screen, x, y, width, height int) (int, int, int, int)) *Box {
	b.draw = handler
	return b
}

// GetDrawFunc returns the callback function which was installed with
// SetDrawFunc() or nil if no such function has been installed.
func (b *Box) GetDrawFunc() func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	return b.draw
}

// WrapInputHandler wraps an input handler (see InputHandler()) with the
// functionality to capture input (see SetInputCapture()) before passing it
// on to the provided (default) input handler.
//
// This is only meant to be used by subclassing primitives.
func (b *Box) WrapInputHandler(inputHandler func(*tcell.EventKey, func(p Primitive))) func(*tcell.EventKey, func(p Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if b.inputCapture != nil {
			event = b.inputCapture(event)
		}
		if event != nil && inputHandler != nil {
			inputHandler(event, setFocus)
		}
	}
}

// InputHandler returns nil.
func (b *Box) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return b.WrapInputHandler(nil)
}

// SetInputCapture installs a function which captures key events before they are
// forwarded to the primitive's default key event handler. This function can
// then choose to forward that key event (or a different one) to the default
// handler by returning it. If nil is returned, the default handler will not
// be called.
//
// Providing a nil handler will remove a previously existing handler.
func (b *Box) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *Box {
	b.inputCapture = capture
	return b
}

// GetInputCapture returns the function installed with SetInputCapture() or nil
// if no such function has been installed.
func (b *Box) GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return b.inputCapture
}

// SetBackgroundColor sets the box's background color.
func (b *Box) SetBackgroundColor(color tcell.Color) *Box {
	b.backgroundColor = color
	return b
}

// SetBorder sets the flag indicating whether or not the box should have a
// border.
func (b *Box) SetBorder(show bool) *Box {
	b.border = show
	return b
}

// GetBorder returns the flag indicating whether or not the box was set
func (b *Box) GetBorder() bool {
	return b.border
}

// SetBorderColor sets the box's border color.
func (b *Box) SetBorderColor(color tcell.Color) *Box {
	b.borderColor = color
	return b
}

// SetTitle sets the box's title.
func (b *Box) SetTitle(title string) *Box {
	b.title = title
	return b
}

// SetTitleColor sets the box's title color.
func (b *Box) SetTitleColor(color tcell.Color) *Box {
	b.titleColor = color
	return b
}

// SetTitleAlign sets the alignment of the title, one of AlignLeft, AlignCenter,
// or AlignRight.
func (b *Box) SetTitleAlign(align int) *Box {
	b.titleAlign = align
	return b
}

// Draw draws this primitive onto the screen.
func (b *Box) Draw(screen tcell.Screen) {
	_, _, width, height := b.GetRect()
	// Don't draw anything if there is no space.
	if width <= 0 || height <= 0 {
		return
	}

	def := tcell.StyleDefault

	// Fill background.
	background := def.Background(b.backgroundColor)
	for y := b.y; y < b.y+height; y++ {
		for x := b.x; x < b.x+width; x++ {
			screen.SetContent(x, y, ' ', nil, background)
		}
	}

	// Draw border.
	if b.border && width >= 2 && height >= 2 {
		border := background.Foreground(b.borderColor)
		var vertical, horizontal, topLeft, topRight, bottomLeft, bottomRight rune
		if b.focus.HasFocus() {
			horizontal = Styles.GraphicsDbVertBar
			vertical = Styles.GraphicsDbHorBar
			topLeft = Styles.GraphicsDbTopLeftCorner
			topRight = Styles.GraphicsDbTopRightCorner
			bottomLeft = Styles.GraphicsDbBottomLeftCorner
			bottomRight = Styles.GraphicsDbBottomRightCorner
		} else {
			horizontal = Styles.GraphicsVertBar
			vertical = Styles.GraphicsHoriBar
			topLeft = Styles.GraphicsTopLeftCorner
			topRight = Styles.GraphicsTopRightCorner
			bottomLeft = Styles.GraphicsBottomLeftCorner
			bottomRight = Styles.GraphicsBottomRightCorner
		}
		for x := b.x + 1; x < b.x+width-1; x++ {
			screen.SetContent(x, b.y, vertical, nil, border)
			screen.SetContent(x, b.y+height-1, vertical, nil, border)
		}
		for y := b.y + 1; y < b.y+height-1; y++ {
			screen.SetContent(b.x, y, horizontal, nil, border)
			screen.SetContent(b.x+width-1, y, horizontal, nil, border)
		}
		screen.SetContent(b.x, b.y, topLeft, nil, border)
		screen.SetContent(b.x+width-1, b.y, topRight, nil, border)
		screen.SetContent(b.x, b.y+height-1, bottomLeft, nil, border)
		screen.SetContent(b.x+width-1, b.y+height-1, bottomRight, nil, border)

		// Draw title.
		if b.title != "" && width >= 4 {
			title := "  " + b.title + "  "
			_, printed := Print(screen, title, b.x+1, b.y, width-2, b.titleAlign, b.titleColor)
			if StringWidth(title)-printed > 0 && printed > 0 {
				_, _, style, _ := screen.GetContent(b.x+width-2, b.y)
				fg, _, _ := style.Decompose()
				Print(screen, string(Styles.GraphicsEllipsis), b.x+width-2, b.y, 1, AlignLeft, fg)
			}
		}
	}

	// Call custom draw function.
	if b.draw != nil {
		b.innerX, b.innerY, b.innerWidth, b.innerHeight = b.draw(screen, b.x, b.y, width, height)
	} else {
		// Remember the inner rect.
		b.innerX = -1
		b.innerX, b.innerY, b.innerWidth, b.innerHeight = b.GetInnerRect()
	}

	// Clamp inner rect to screen.
	if b.clampToScreen {
		width, height := screen.Size()
		if b.innerX < 0 {
			b.innerWidth += b.innerX
			b.innerX = 0
		}
		if b.innerX+b.innerWidth >= width {
			b.innerWidth = width - b.innerX
		}
		if b.innerY+b.innerHeight >= height {
			b.innerHeight = height - b.innerY
		}
		if b.innerY < 0 {
			b.innerHeight += b.innerY
			b.innerY = 0
		}
	}
}

// Focus is called when this primitive receives focus.
func (b *Box) Focus(delegate func(p Primitive)) {
	b.hasFocus = true
}

// Blur is called when this primitive loses focus.
func (b *Box) Blur() {
	b.hasFocus = false
}

// HasFocus returns whether or not this primitive has focus.
func (b *Box) HasFocus() bool {
	return b.hasFocus
}

// GetFocusable returns the item's Focusable.
func (b *Box) GetFocusable() Focusable {
	return b.focus
}
