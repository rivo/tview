package tview

import "github.com/gdamore/tcell"

// Characters to draw the box border.
const (
	BoxVertBar           = '\u2500'
	BoxHorBar            = '\u2502'
	BoxTopLeftCorner     = '\u250c'
	BoxTopRightCorner    = '\u2510'
	BoxBottomRightCorner = '\u2518'
	BoxBottomLeftCorner  = '\u2514'
	BoxEllipsis          = '\u2026'
)

// Box implements Rect with a background and optional elements such as a border
// and a title.
type Box struct {
	// The position of the rect.
	x, y, width, height int

	// Whether or not the box has focus.
	hasFocus bool

	// The box's background color.
	backgroundColor tcell.Color

	// Whether or not a border is drawn, reducing the box's space for content by
	// two in width and height.
	border bool

	// The color of the border.
	borderColor tcell.Color

	// The color of the border when the box has focus.
	focusedBorderColor tcell.Color

	// The title. Only visible if there is a border, too.
	title string

	// The color of the title.
	titleColor tcell.Color
}

// NewBox returns a Box without a border.
func NewBox() *Box {
	return &Box{
		width:              15,
		height:             10,
		borderColor:        tcell.ColorWhite,
		focusedBorderColor: tcell.ColorYellow,
		titleColor:         tcell.ColorWhite,
	}
}

// Draw draws this primitive onto the screen.
func (b *Box) Draw(screen tcell.Screen) {
	// Don't draw anything if there is no space.
	if b.width <= 0 || b.height <= 0 {
		return
	}

	def := tcell.StyleDefault

	// Fill background.
	background := def.Background(b.backgroundColor)
	for y := b.y; y < b.y+b.height; y++ {
		for x := b.x; x < b.x+b.width; x++ {
			screen.SetContent(x, y, ' ', nil, background)
		}
	}

	// Draw border.
	if b.border && b.width >= 2 && b.height >= 2 {
		border := background.Foreground(b.borderColor)
		if b.hasFocus {
			border = background.Foreground(b.focusedBorderColor)
		}
		for x := b.x + 1; x < b.x+b.width-1; x++ {
			screen.SetContent(x, b.y, BoxVertBar, nil, border)
			screen.SetContent(x, b.y+b.height-1, BoxVertBar, nil, border)
		}
		for y := b.y + 1; y < b.y+b.height-1; y++ {
			screen.SetContent(b.x, y, BoxHorBar, nil, border)
			screen.SetContent(b.x+b.width-1, y, BoxHorBar, nil, border)
		}
		screen.SetContent(b.x, b.y, BoxTopLeftCorner, nil, border)
		screen.SetContent(b.x+b.width-1, b.y, BoxTopRightCorner, nil, border)
		screen.SetContent(b.x, b.y+b.height-1, BoxBottomLeftCorner, nil, border)
		screen.SetContent(b.x+b.width-1, b.y+b.height-1, BoxBottomRightCorner, nil, border)

		// Draw title.
		if b.title != "" && b.width >= 4 {
			title := background.Foreground(b.titleColor)
			x := b.x
			for index, ch := range b.title {
				x++
				if x >= b.x+b.width-1 {
					break
				}
				if x == b.x+b.width-2 && index < len(b.title)-1 {
					ch = BoxEllipsis
				}
				screen.SetContent(x, b.y, ch, nil, title)
			}
		}
	}
}

// GetRect returns the current position of the rectangle, x, y, width, and
// height.
func (b *Box) GetRect() (int, int, int, int) {
	return b.x, b.y, b.width, b.height
}

// SetRect sets a new position of the rectangle.
func (b *Box) SetRect(x, y, width, height int) {
	b.x = x
	b.y = y
	b.width = width
	b.height = height
}

// InputHandler returns nil.
func (b *Box) InputHandler() func(event *tcell.EventKey) {
	return nil
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

// SetBorderColor sets the box's border color.
func (b *Box) SetBorderColor(color tcell.Color) *Box {
	b.borderColor = color
	return b
}

// SetFocusedBorderColor sets the box's border color for when the box has focus.
func (b *Box) SetFocusedBorderColor(color tcell.Color) *Box {
	b.focusedBorderColor = color
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

// Focus is called when this primitive receives focus.
func (b *Box) Focus(app *Application) {
	b.hasFocus = true
}

// Blur is called when this primitive loses focus.
func (b *Box) Blur() {
	b.hasFocus = false
}
