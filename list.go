package tview

import (
	"fmt"

	"github.com/gdamore/tcell"
)

// listItem represents one item in a List.
type listItem struct {
	MainText      string // The main text of the list item.
	SecondaryText string // A secondary text to be shown underneath the main text.
	Shortcut      rune   // The key to select the list item directly, 0 if there is no shortcut.
}

// List displays rows of items, each of which can be selected.
type List struct {
	*Box

	// The items of the list.
	items []*listItem

	// The index of the currently selected item.
	currentItem int

	// Whether or not to show the secondary item texts.
	showSecondaryText bool

	// The item main text color. Selected items have their background and text
	// color switched.
	mainTextColor tcell.Color

	// The item secondary text color.
	secondaryTextColor tcell.Color

	// The item shortcut text color.
	shortcutColor tcell.Color

	// An optional function which is called when a list item was selected.
	selected func(index int, mainText, secondaryText string, shortcut rune)
}

// NewList returns a new form.
func NewList() *List {
	return &List{
		Box:                NewBox(),
		showSecondaryText:  true,
		mainTextColor:      tcell.ColorWhite,
		secondaryTextColor: tcell.ColorGreen,
		shortcutColor:      tcell.ColorYellow,
	}
}

// SetMainTextColorColor sets the color of the items' main text.
func (l *List) SetMainTextColorColor(color tcell.Color) *List {
	l.mainTextColor = color
	return l
}

// SetSecondaryTextColorColor sets the color of the items' secondary text.
func (l *List) SetSecondaryTextColorColor(color tcell.Color) *List {
	l.secondaryTextColor = color
	return l
}

// SetShortcutColor sets the color of the items' shortcut.
func (l *List) SetShortcutColor(color tcell.Color) *List {
	l.shortcutColor = color
	return l
}

// ShowSecondaryText determines whether or not to show secondary item texts.
func (l *List) ShowSecondaryText(show bool) *List {
	l.showSecondaryText = show
	return l
}

// SetSelectedFunc sets the function which is called when the user selects a
// list item. The function receives the item's index in the list of items
// (starting with 0), its main text, secondary text, and its shortcut rune.
func (l *List) SetSelectedFunc(handler func(int, string, string, rune)) *List {
	l.selected = handler
	return l
}

// AddItem adds a new item to the list. An item has a main text which will be
// highlighted when selected. It also has a secondary text which is shown
// underneath the main text (if it is set to visible) but which may remain
// empty.
//
// The shortcut is a key binding. If the specified rune is entered, the item
// is selected immediately. Set to 0 for no binding.
func (l *List) AddItem(mainText, secondaryText string, shortcut rune) *List {
	l.items = append(l.items, &listItem{
		MainText:      mainText,
		SecondaryText: secondaryText,
		Shortcut:      shortcut,
	})
	return l
}

// Draw draws this primitive onto the screen.
func (l *List) Draw(screen tcell.Screen) {
	l.Box.Draw(screen)

	// Determine the dimensions.
	x := l.x
	y := l.y
	width := l.width
	bottomLimit := l.y + l.height
	if l.border {
		x++
		y++
		width -= 2
		bottomLimit -= 2
	}

	// Do we show any shortcuts?
	var showShortcuts bool
	for _, item := range l.items {
		if item.Shortcut != 0 {
			showShortcuts = true
			x += 4
			width -= 4
			break
		}
	}

	// Draw the list items.
	for index, item := range l.items {
		if y >= bottomLimit {
			break
		}

		// Shortcuts.
		if showShortcuts && item.Shortcut != 0 {
			Print(screen, fmt.Sprintf("(%s)", string(item.Shortcut)), x-2, y, width+4, AlignRight, l.shortcutColor)
		}

		// Main text.
		color := l.mainTextColor
		if l.focus.HasFocus() && index == l.currentItem {
			textLength := len([]rune(item.MainText))
			style := tcell.StyleDefault.Background(l.mainTextColor)
			for bx := 0; bx < textLength && bx < width; bx++ {
				screen.SetContent(x+bx, y, ' ', nil, style)
			}
			color = l.backgroundColor
		}
		Print(screen, item.MainText, x, y, width, AlignLeft, color)
		y++

		if y >= bottomLimit {
			break
		}

		// Secondary text.
		if l.showSecondaryText {
			Print(screen, item.SecondaryText, x, y, width, AlignLeft, l.secondaryTextColor)
			y++
		}
	}
}

// InputHandler returns the handler for this primitive.
func (l *List) InputHandler() func(event *tcell.EventKey) {
	return func(event *tcell.EventKey) {
		switch key := event.Key(); key {
		case tcell.KeyTab, tcell.KeyDown, tcell.KeyRight:
			l.currentItem++
		case tcell.KeyBacktab, tcell.KeyUp, tcell.KeyLeft:
			l.currentItem--
		case tcell.KeyHome:
			l.currentItem = 0
		case tcell.KeyEnd:
			l.currentItem = len(l.items) - 1
		case tcell.KeyPgDn:
			l.currentItem += 5
		case tcell.KeyPgUp:
			l.currentItem -= 5
		case tcell.KeyEnter:
			if l.selected != nil {
				item := l.items[l.currentItem]
				l.selected(l.currentItem, item.MainText, item.SecondaryText, item.Shortcut)
			}
		case tcell.KeyRune:
			ch := event.Rune()
			if ch != ' ' {
				// It's not a space bar. Is it a shortcut?
				var found bool
				for index, item := range l.items {
					if item.Shortcut == ch {
						// We have a shortcut.
						found = true
						l.currentItem = index
						break
					}
				}
				if !found {
					break
				}
			}
			if l.selected != nil {
				item := l.items[l.currentItem]
				l.selected(l.currentItem, item.MainText, item.SecondaryText, item.Shortcut)
			}
		}

		if l.currentItem < 0 {
			l.currentItem = len(l.items) - 1
		} else if l.currentItem >= len(l.items) {
			l.currentItem = 0
		}
	}
}
