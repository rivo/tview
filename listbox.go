//
// Copyright (c) 2018 Litmus Automation Inc.
// Author: Levko Burburas <levko.burburas.external@litmus.cloud>
//

package tview

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
)

// listBoxItem represents one item in a ListBox.
type listBoxItem struct {
	MainText      string // The main text of the list item.
	SecondaryText string // A secondary text to be shown underneath the main text.
	Shortcut      rune   // The key to select the list item directly, 0 if there is no shortcut.
	Selected      func() // The optional function which is called when the item is selected.
}

// ListBox displays rows of items, each of which can be selected.
//
// See https://github.com/rivo/tview/wiki/ListBox for an example.
type ListBox struct {
	*Box

	align int

	labelFiller string

	// The items of the list.
	items []*listBoxItem

	// The text to be displayed before the input area.
	label string

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// The index of the currently selected item.
	currentItem int

	offset int
	// Whether or not to show the secondary item texts.
	showSecondaryText bool

	// The item main text color.
	mainTextColor tcell.Color

	// The item secondary text color.
	secondaryTextColor tcell.Color

	// The item shortcut text color.
	shortcutColor tcell.Color

	// The text color for selected items.
	selectedTextColor tcell.Color

	// The background color for selected items.
	selectedBackgroundColor tcell.Color

	// An optional function which is called when the user has navigated to a list
	// item.
	changed func(index int, mainText, secondaryText string, shortcut rune)

	// An optional function which is called when a list item was selected. This
	// function will be called even if the list item defines its own callback.
	selected func(index int, mainText, secondaryText string, shortcut rune)

	// An optional function which is called when the user presses the Escape key.
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
}

// NewListBox returns a new form.
func NewListBox() *ListBox {
	l := &ListBox{
		Box:                     NewBox(),
		mainTextColor:           Styles.PrimaryTextColor,
		secondaryTextColor:      Styles.TertiaryTextColor,
		shortcutColor:           Styles.SecondaryTextColor,
		selectedTextColor:       Styles.PrimitiveBackgroundColor,
		selectedBackgroundColor: Styles.PrimaryTextColor,
		labelColor:              Styles.SecondaryTextColor,
		fieldBackgroundColor:    Styles.FieldBackgroundColor,
		fieldTextColor:          Styles.FieldTextColor,
		align:                   AlignLeft,
		labelFiller:             " ",
	}

	l.focus = l
	return l
}

// SetCurrentItem sets the currently selected item by its index. This triggers
// a "changed" event.
func (l *ListBox) SetCurrentItem(index int) *ListBox {
	_, _, _, height := l.GetInnerRect()
	l.currentItem = index
	l.offset = l.currentItem - height/2
	if l.currentItem < len(l.items) && l.changed != nil {
		item := l.items[l.currentItem]
		l.changed(l.currentItem, item.MainText, item.SecondaryText, item.Shortcut)
	}
	return l
}

// SetCurrentItemByText sets the currently selected item by its text. This triggers
// a "changed" event.
func (l *ListBox) SetCurrentItemByText(text string) *ListBox {
	for i := 0; i < len(l.items); i++ {
		if l.items[i].MainText == text {
			l.SetCurrentItem(i)
			break
		}
	}
	return l
}

// GetCurrentItem returns the index of the currently selected list item.
func (l *ListBox) GetCurrentItem() int {
	return l.currentItem
}

// GetCurrentItemText returns the index of the currently selected list item.
func (l *ListBox) GetCurrentItemText() string {
	if len(l.items) == 0 {
		return ""
	}
	return l.items[l.currentItem].MainText
}

// SetMainTextColor sets the color of the items' main text.
func (l *ListBox) SetMainTextColor(color tcell.Color) *ListBox {
	l.mainTextColor = color
	return l
}

// SetSecondaryTextColor sets the color of the items' secondary text.
func (l *ListBox) SetSecondaryTextColor(color tcell.Color) *ListBox {
	l.secondaryTextColor = color
	return l
}

// SetShortcutColor sets the color of the items' shortcut.
func (l *ListBox) SetShortcutColor(color tcell.Color) *ListBox {
	l.shortcutColor = color
	return l
}

// SetSelectedTextColor sets the text color of selected items.
func (l *ListBox) SetSelectedTextColor(color tcell.Color) *ListBox {
	l.selectedTextColor = color
	return l
}

// SetSelectedBackgroundColor sets the background color of selected items.
func (l *ListBox) SetSelectedBackgroundColor(color tcell.Color) *ListBox {
	l.selectedBackgroundColor = color
	return l
}

// ShowSecondaryText determines whether or not to show secondary item texts.
func (l *ListBox) ShowSecondaryText(show bool) *ListBox {
	l.showSecondaryText = show
	return l
}

// SetChangedFunc sets the function which is called when the user navigates to
// a list item. The function receives the item's index in the list of items
// (starting with 0), its main text, secondary text, and its shortcut rune.
//
// This function is also called when the first item is added or when
// SetCurrentItem() is called.
func (l *ListBox) SetChangedFunc(handler func(int, string, string, rune)) *ListBox {
	l.changed = handler
	return l
}

// SetSelectedFunc sets the function which is called when the user selects a
// list item by pressing Enter on the current selection. The function receives
// the item's index in the list of items (starting with 0), its main text,
// secondary text, and its shortcut rune.
func (l *ListBox) SetSelectedFunc(handler func(int, string, string, rune)) *ListBox {
	l.selected = handler
	return l
}

// SetDoneFunc sets a function which is called when the user presses the Escape
// key.
func (l *ListBox) SetDoneFunc(handler func(tcell.Key)) *ListBox {
	l.done = handler
	return l
}

// AddItem adds a new item to the list. An item has a main text which will be
// highlighted when selected. It also has a secondary text which is shown
// underneath the main text (if it is set to visible) but which may remain
// empty.
//
// The shortcut is a key binding. If the specified rune is entered, the item
// is selected immediately. Set to 0 for no binding.
//
// The "selected" callback will be invoked when the user selects the item. You
// may provide nil if no such item is needed or if all events are handled
// through the selected callback set with SetSelectedFunc().
func (l *ListBox) AddItem(mainText, secondaryText string, shortcut rune, selected func()) *ListBox {
	l.items = append(l.items, &listBoxItem{
		MainText:      mainText,
		SecondaryText: secondaryText,
		Shortcut:      shortcut,
		Selected:      selected,
	})
	if len(l.items) == 1 && l.changed != nil {
		item := l.items[0]
		l.changed(0, item.MainText, item.SecondaryText, item.Shortcut)
	}
	return l
}

// Clear removes all items from the list.
func (l *ListBox) Clear() *ListBox {
	l.items = nil
	l.currentItem = 0
	return l
}

// Draw draws this primitive onto the screen.
func (l *ListBox) Draw(screen tcell.Screen) {
	l.Box.Draw(screen)

	// Determine the dimensions.
	x, y, width, height := l.GetInnerRect()
	bottomLimit := y + height

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

	// We want to keep the current selection in view. What is our offset?
	if l.showSecondaryText {
		if l.currentItem >= height/2 {
			l.offset = l.currentItem + 1 - (height / 2)
		}
	} else {
		if l.offset > 0 && l.offset > l.currentItem {
			l.offset--
		} else if l.currentItem >= height && l.offset <= l.currentItem-height {
			l.offset = l.currentItem + 1 - height
		}
	}

	// Draw the list items.
	for index, item := range l.items {
		if index < l.offset {
			continue
		}

		if y >= bottomLimit {
			break
		}

		// Shortcuts.
		if showShortcuts && item.Shortcut != 0 {
			Print(screen, fmt.Sprintf("(%s)", string(item.Shortcut)), x-5, y, 4, AlignRight, l.shortcutColor)
		}

		// Main text.
		Print(screen, item.MainText, x, y, width, AlignLeft, l.mainTextColor)

		// Background color of selected text.
		if index == l.currentItem {
			textWidth := StringWidth(item.MainText)
			for bx := 0; bx < textWidth && bx < width; bx++ {
				m, c, style, _ := screen.GetContent(x+bx, y)
				fg, _, _ := style.Decompose()
				if fg == l.mainTextColor {
					fg = l.selectedTextColor
				}
				if l.focus.HasFocus() {
					style = style.Background(l.fieldBackgroundColor).Foreground(fg)
				} else {
					style = style.Background(l.fieldTextColor).Foreground(l.fieldBackgroundColor).Underline(true)
				}
				screen.SetContent(x+bx, y, m, c, style)
			}
		}
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
func (l *ListBox) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	_, _, _, height := l.GetInnerRect()
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		previousItem := l.currentItem

		switch key := event.Key(); key {
		case tcell.KeyTab, tcell.KeyBacktab: // We're done.
			if l.done != nil {
				l.done(key)
			}
			if l.finished != nil {
				l.finished(key)
			}
		case tcell.KeyDown, tcell.KeyRight:
			l.currentItem++
			if l.currentItem >= len(l.items) {
				l.currentItem = 0
				l.offset = 0
			}
		case tcell.KeyUp, tcell.KeyLeft:
			l.currentItem--
			if l.currentItem < 0 {
				l.currentItem = len(l.items) - 1
			}
		case tcell.KeyHome:
			l.currentItem = 0
		case tcell.KeyEnd:
			l.currentItem = len(l.items) - 1
		case tcell.KeyPgDn:
			if l.currentItem < l.offset+height-1 {
				l.currentItem = l.offset + height - 1
			} else {
				l.currentItem += height
			}
			if l.currentItem >= len(l.items) {
				l.currentItem = 0
				l.offset = 0
			}
		case tcell.KeyPgUp:
			if l.currentItem > l.offset {
				l.currentItem = l.offset
			} else {
				l.currentItem -= height
				l.offset = l.currentItem
			}
		case tcell.KeyEnter:
			if len(l.items) == 0 {
				break
			}
			item := l.items[l.currentItem]
			if item.Selected != nil {
				item.Selected()
			}
			if l.selected != nil {
				l.selected(l.currentItem, item.MainText, item.SecondaryText, item.Shortcut)
			}
		case tcell.KeyEscape:
			if l.done != nil {
				l.done(key)
			}
			if l.finished != nil {
				l.finished(key)
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
			item := l.items[l.currentItem]
			if item.Selected != nil {
				item.Selected()
			}
			if l.selected != nil {
				l.selected(l.currentItem, item.MainText, item.SecondaryText, item.Shortcut)
			}
		}

		if l.currentItem < 0 {
			l.currentItem = len(l.items) - 1
		} else if l.currentItem >= len(l.items) {
			l.currentItem = 0
		}

		if l.currentItem != previousItem && l.currentItem < len(l.items) && l.changed != nil {
			item := l.items[l.currentItem]
			l.changed(l.currentItem, item.MainText, item.SecondaryText, item.Shortcut)
		}
	})
}

// SetFieldAlign sets the input alignment within the inputfield. This must be
// either AlignLeft, AlignCenter, or AlignRight.
func (l *ListBox) SetFieldAlign(align int) FormItem {
	l.align = align
	return l
}

// GetFieldAlign returns the input alignment within the inputfield.
func (l *ListBox) GetFieldAlign() (align int) {
	return l.align
}

// SetLabelFiller sets a sign which will be fill the label when this one need to stretch
func (l *ListBox) SetLabelFiller(labelFiller string) FormItem {
	l.labelFiller = labelFiller
	return l
}

// GetLabelWidth returns label width.
func (l *ListBox) GetLabelWidth() int {
	return StringWidth(strings.Replace(l.label, "%s", "", -1))
}

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (l *ListBox) SetLabelWidth(width int) *ListBox {
	l.labelWidth = width
	return l
}

// GetLabelFiller gets a sign which uses for stretching
func (l *ListBox) GetLabelFiller() (labelFiller string) {
	return l.labelFiller
}

// SetFieldWidth sets the screen width of the input area. A value of 0 means
// extend as much as possible.
func (l *ListBox) SetFieldWidth(width int) FormItem {
	l.fieldWidth = width
	return l
}

// GetFieldWidth returns this primitive's field width.
func (l *ListBox) GetFieldWidth() int {
	if l.fieldWidth == 0 {
		_, _, l.fieldWidth, _ = l.GetInnerRect()
	}
	return l.fieldWidth
}

// SetLabel sets the text to be displayed before the input area.
func (l *ListBox) SetLabel(label string) *ListBox {
	l.label = label
	return l
}

// GetLabel returns the text to be displayed before the input area.
func (l *ListBox) GetLabel() string {
	return l.label
}

// SetFinishedFunc calls SetDoneFunc().
func (l *ListBox) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	l.finished = handler
	return l
}

// GetFinishedFunc returns SetDoneFunc().
func (l *ListBox) GetFinishedFunc() func(key tcell.Key) {
	return l.finished
}

// SetFormAttributes sets attributes shared by all form items.
func (l *ListBox) SetFormAttributes(labelWidth, fieldWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	if l.fieldWidth == 0 {
		l.fieldWidth = fieldWidth
	}
	if l.labelWidth == 0 {
		l.labelWidth = labelWidth
	}

	l.labelColor = labelColor
	l.SetBackgroundColor(bgColor)
	l.fieldTextColor = fieldTextColor
	l.fieldBackgroundColor = fieldBgColor
	return l
}
