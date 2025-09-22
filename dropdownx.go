package tview

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// MultiSelectDropDown implements a selection widget whose options become visible in a
// drop-down list when activated.
//
// See https://github.com/rivo/tview/wiki/MultiSelectDropDown for an example.
type MultiSelectDropDown struct {
	*Box

	// Whether or not this drop-down is disabled/read-only.
	disabled bool

	// Strings to be placed before and after each drop-down option.
	optionPrefix, optionSuffix string

	// Strings to be placed before and after the current option.
	currentOptionPrefix, currentOptionSuffix string

	// The text to be displayed when no option has yet been selected.
	noSelection string

	// Set to true if the options are visible and selectable.
	open bool

	// The input field containing the entered prefix for the current selection.
	// This is only visible when the drop-down is open. It never receives focus,
	// however. And it only receives events, we never call its Draw method.
	prefix *InputField

	// The list element for the options.
	list *MultiList

	// The text to be displayed before the input area.
	label string

	// The label style.
	labelStyle tcell.Style

	// The field style.
	fieldStyle tcell.Style

	// The style of the field when it is focused and the drop-down is closed.
	focusedStyle tcell.Style

	// The style of the field when it is disabled.
	disabledStyle tcell.Style

	// The style of the prefix.
	prefixStyle tcell.Style

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	// An optional function which is called when the user indicated that they
	// are done selecting options. The key which was pressed is provided (tab,
	// shift-tab, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)

	dragging bool // Set to true when mouse dragging is in progress.
}

// NewDropDown returns a new [MultiSelectDropDown].
func NewMultiSelectDropDown() *MultiSelectDropDown {
	list := NewMultiList()
	list.ShowSecondaryText(false).
		SetMainTextStyle(tcell.StyleDefault.Background(Styles.MoreContrastBackgroundColor).Foreground(Styles.PrimitiveBackgroundColor)).
		SetSelectedStyle(tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.PrimitiveBackgroundColor)).
		SetHighlightFullLine(true).
		SetBackgroundColor(Styles.MoreContrastBackgroundColor)

	prefix := NewInputField()
	prefix.SetDisabled(true)

	box := NewBox()
	d := &MultiSelectDropDown{
		Box:           box,
		list:          list,
		prefix:        prefix,
		labelStyle:    tcell.StyleDefault.Foreground(Styles.SecondaryTextColor),
		fieldStyle:    tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
		focusedStyle:  tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.ContrastBackgroundColor),
		disabledStyle: tcell.StyleDefault.Background(box.backgroundColor).Foreground(Styles.SecondaryTextColor),
		prefixStyle:   tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.ContrastBackgroundColor),
	}

	d.Box.Primitive = d
	return d
}

func (d *MultiSelectDropDown) SelectItems(indexes []int) *MultiSelectDropDown {
	d.list.SelectItems(indexes)
	return d
}

func (d *MultiSelectDropDown) SelectValues(vals []string) *MultiSelectDropDown {
	d.list.SelectValues(vals)
	return d
}

func (d *MultiSelectDropDown) GetSelected() ([]string, []int) {
	indexes := d.list.GetSelected()
	vals := d.list.GetValues(indexes)
	return vals, indexes
}

// SetTextOptions sets the text to be placed before and after each drop-down
// option (prefix/suffix), the text placed before and after the currently
// selected option (currentPrefix/currentSuffix) as well as the text to be
// displayed when no option is currently selected. Per default, all of these
// strings are empty.
func (d *MultiSelectDropDown) SetTextOptions(prefix, suffix, currentPrefix, currentSuffix, noSelection string) *MultiSelectDropDown {
	d.currentOptionPrefix = currentPrefix
	d.currentOptionSuffix = currentSuffix
	d.noSelection = noSelection
	d.optionPrefix = prefix
	d.optionSuffix = suffix
	for idx, item := range d.list.items {
		d.list.SetItemText(idx, prefix+item.MainText+suffix, "")
	}
	return d
}

// SetUseStyleTags sets a flag that determines whether tags found in the option
// texts are interpreted as tview tags. By default, this flag is enabled (for
// backwards compatibility reasons).
func (d *MultiSelectDropDown) SetUseStyleTags(useStyleTags bool) *MultiSelectDropDown {
	d.list.SetUseStyleTags(useStyleTags, useStyleTags)
	return d
}

// SetLabel sets the text to be displayed before the input area.
func (d *MultiSelectDropDown) SetLabel(label string) *MultiSelectDropDown {
	d.label = label
	return d
}

// GetLabel returns the text to be displayed before the input area.
func (d *MultiSelectDropDown) GetLabel() string {
	return d.label
}

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (d *MultiSelectDropDown) SetLabelWidth(width int) *MultiSelectDropDown {
	d.labelWidth = width
	return d
}

// SetLabelColor sets the color of the label.
func (d *MultiSelectDropDown) SetLabelColor(color tcell.Color) *MultiSelectDropDown {
	d.labelStyle = d.labelStyle.Foreground(color)
	return d
}

// SetLabelStyle sets the style of the label.
func (d *MultiSelectDropDown) SetLabelStyle(style tcell.Style) *MultiSelectDropDown {
	d.labelStyle = style
	return d
}

// SetFieldBackgroundColor sets the background color of the selected field.
// This also overrides the prefix background color.
func (d *MultiSelectDropDown) SetFieldBackgroundColor(color tcell.Color) *MultiSelectDropDown {
	d.fieldStyle = d.fieldStyle.Background(color)
	d.prefix.SetFieldBackgroundColor(color)
	return d
}

// SetFieldTextColor sets the text color of the options area.
func (d *MultiSelectDropDown) SetFieldTextColor(color tcell.Color) *MultiSelectDropDown {
	d.fieldStyle = d.fieldStyle.Foreground(color)
	return d
}

// SetFieldStyle sets the style of the options area.
func (d *MultiSelectDropDown) SetFieldStyle(style tcell.Style) *MultiSelectDropDown {
	d.fieldStyle = style
	return d
}

// SetFocusedStyle sets the style of the options area when the drop-down is
// focused and closed.
func (d *MultiSelectDropDown) SetFocusedStyle(style tcell.Style) *MultiSelectDropDown {
	d.focusedStyle = style
	return d
}

// SetDisabledStyle sets the style of the options area when the drop-down is
// disabled.
func (d *MultiSelectDropDown) SetDisabledStyle(style tcell.Style) *MultiSelectDropDown {
	d.disabledStyle = style
	return d
}

// SetPrefixTextColor sets the color of the prefix string. The prefix string is
// shown when the user starts typing text, which directly selects the first
// option that starts with the typed string.
func (d *MultiSelectDropDown) SetPrefixTextColor(color tcell.Color) *MultiSelectDropDown {
	d.prefixStyle = d.prefixStyle.Foreground(color)
	return d
}

// SetPrefixStyle sets the style of the prefix string. The prefix string is
// shown when the user starts typing text, which directly selects the first
// option that starts with the typed string.
func (d *MultiSelectDropDown) SetPrefixStyle(style tcell.Style) *MultiSelectDropDown {
	d.prefixStyle = style
	return d
}

// SetListStyles sets the styles of the items in the drop-down list (unselected
// as well as selected items). Style attributes are currently ignored but may be
// used in the future.
func (d *MultiSelectDropDown) SetListStyles(unselected, selected tcell.Style) *MultiSelectDropDown {
	d.list.SetMainTextStyle(unselected).SetSelectedStyle(selected)
	_, bg, _ := unselected.Decompose()
	d.list.SetBackgroundColor(bg)
	return d
}

// SetFormAttributes sets attributes shared by all form items.
func (d *MultiSelectDropDown) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	d.labelWidth = labelWidth
	d.SetLabelColor(labelColor)
	d.SetBackgroundColor(bgColor)
	d.SetFieldStyle(tcell.StyleDefault.Foreground(fieldTextColor).Background(fieldBgColor))
	return d
}

// SetFieldWidth sets the screen width of the options area. A value of 0 means
// extend to as long as the longest option text.
func (d *MultiSelectDropDown) SetFieldWidth(width int) *MultiSelectDropDown {
	d.fieldWidth = width
	return d
}

// GetFieldWidth returns this primitive's field screen width.
func (d *MultiSelectDropDown) GetFieldWidth() int {
	if d.fieldWidth > 0 {
		return d.fieldWidth
	}
	fieldWidth := 0
	for _, option := range d.list.items {
		width := TaggedStringWidth(option.MainText)
		if width > fieldWidth {
			fieldWidth = width
		}
	}
	return fieldWidth
}

// GetFieldHeight returns this primitive's field height.
func (d *MultiSelectDropDown) GetFieldHeight() int {
	return 1
}

// SetDisabled sets whether or not the item is disabled / read-only.
func (d *MultiSelectDropDown) SetDisabled(disabled bool) FormItem {
	d.disabled = disabled
	if d.finished != nil {
		d.finished(-1)
	}
	return d
}

// GetDisabled returns whether or not the item is disabled / read-only.
func (d *MultiSelectDropDown) GetDisabled() bool {
	return d.disabled
}

// AddOption adds a new selectable option to this drop-down. The "selected"
// callback is called when this option was selected. It may be nil.
func (d *MultiSelectDropDown) AddOption(text string, selected func()) *MultiSelectDropDown {
	d.list.AddItem(d.optionPrefix+text+d.optionSuffix, "", selected)
	return d
}

// SetOptions replaces all current options with the ones provided and installs
// one callback function which is called when one of the options is selected.
// It will be called with the option's text and its index into the options
// slice. The "selected" parameter may be nil.
func (d *MultiSelectDropDown) SetOptions(texts []string, selected func(texts []string, indexes []int)) *MultiSelectDropDown {
	d.list.Clear()
	for _, text := range texts {
		d.AddOption(text, nil)
	}
	d.list.selected = selected
	return d
}

// GetOptionCount returns the number of options in the drop-down.
func (d *MultiSelectDropDown) GetOptionCount() int {
	return d.list.GetItemCount()
}

// RemoveOption removes the specified option from the drop-down. Panics if the
// index is out of range. If the currently selected option is removed, no option
// will be selected.
func (d *MultiSelectDropDown) RemoveOption(index int) *MultiSelectDropDown {
	d.list.RemoveItem(index)
	return d
}

// SetSelectedFunc sets a handler which is called when the user changes the
// drop-down's option. This handler will be called in addition and prior to
// an option's optional individual handler. The handler is provided with the
// selected option's text and index. If "no option" was selected, these values
// are an empty string and -1.
func (d *MultiSelectDropDown) SetSelectedFunc(handler func(texts []string, indexes []int)) *MultiSelectDropDown {
	d.list.selected = handler
	return d
}

// SetDoneFunc sets a handler which is called when the user is done selecting
// options. The callback function is provided with the key that was pressed,
// which is one of the following:
//
//   - KeyEscape: Abort selection.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (d *MultiSelectDropDown) SetDoneFunc(handler func(key tcell.Key)) *MultiSelectDropDown {
	d.done = handler
	return d
}

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (d *MultiSelectDropDown) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	d.finished = handler
	return d
}

// Draw draws this primitive onto the screen.
func (d *MultiSelectDropDown) Draw(screen tcell.Screen) {
	d.Box.DrawForSubclass(screen, d)

	// Prepare.
	x, y, width, height := d.GetInnerRect()
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}
	useStyleTags, _ := d.list.GetUseStyleTags()

	// Draw label.
	if d.labelWidth > 0 {
		labelWidth := d.labelWidth
		if labelWidth > rightLimit-x {
			labelWidth = rightLimit - x
		}
		printWithStyle(screen, d.label, x, y, 0, labelWidth, AlignLeft, d.labelStyle, true)
		x += labelWidth
	} else {
		_, _, drawnWidth := printWithStyle(screen, d.label, x, y, 0, rightLimit-x, AlignLeft, d.labelStyle, true)
		x += drawnWidth
	}

	// What's the longest option text?
	maxWidth := 0
	for _, option := range d.list.items {
		str := d.optionPrefix + option.MainText + d.optionSuffix
		if !useStyleTags {
			str = Escape(str)
		}
		strWidth := 5 + TaggedStringWidth(str)
		if strWidth > maxWidth {
			maxWidth = strWidth
		}
		str = d.currentOptionPrefix + option.MainText + d.currentOptionSuffix
		if !useStyleTags {
			str = Escape(str)
		}
		strWidth = TaggedStringWidth(str)
		if strWidth > maxWidth {
			maxWidth = strWidth
		}
	}

	// Draw selection area.
	currentOption := d.list.currentItem
	fieldWidth := d.fieldWidth
	if fieldWidth == 0 {
		fieldWidth = maxWidth
		if currentOption < 0 {
			noSelectionWidth := TaggedStringWidth(d.noSelection)
			if noSelectionWidth > fieldWidth {
				fieldWidth = noSelectionWidth
			}
		} else if currentOption < len(d.list.items) {
			currentOptionWidth := TaggedStringWidth(d.currentOptionPrefix + d.list.items[currentOption].MainText + d.currentOptionSuffix)
			if currentOptionWidth > fieldWidth {
				fieldWidth = currentOptionWidth
			}
		}
	}
	if rightLimit-x < fieldWidth {
		fieldWidth = rightLimit - x
	}
	fieldStyle := d.fieldStyle
	if d.disabled {
		fieldStyle = d.disabledStyle
	} else if d.HasFocus() && !d.open {
		fieldStyle = d.focusedStyle
	}
	for index := 0; index < fieldWidth; index++ {
		screen.SetContent(x+index, y, ' ', nil, fieldStyle)
	}

	// Draw selected text.
	text := strings.Join(d.list.GetSelectedValues(), ",")
	if text == "" {
		text = d.noSelection
	}
	if !useStyleTags {
		text = Escape(text)
	}
	printWithStyle(screen, text, x, y, 0, fieldWidth, AlignLeft, fieldStyle, false)

	// Draw options list.
	if d.HasFocus() && d.open {
		lx := x
		ly := y + 1
		lwidth := maxWidth
		lheight := len(d.list.items)
		swidth, sheight := screen.Size()
		// We prefer to align the left sides of the list and the main widget, but
		// if there is no space to the right, then shift the list to the left.
		if lx+lwidth >= swidth {
			lx = swidth - lwidth
			if lx < 0 {
				lx = 0
			}
		}
		// We prefer to drop down but if there is no space, maybe drop up?
		if ly+lheight >= sheight && ly-2 > lheight-ly {
			ly = y - lheight
			if ly < 0 {
				ly = 0
			}
		}
		if ly+lheight >= sheight {
			lheight = sheight - ly
		}
		d.list.SetRect(lx, ly, lwidth, lheight)
		d.list.Draw(screen)
	}
}

// InputHandler returns the handler for this primitive.
func (d *MultiSelectDropDown) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if d.disabled {
			return
		}

		// Process key event.
		switch key := event.Key(); key {
		case tcell.KeyDown, tcell.KeyUp, tcell.KeyHome, tcell.KeyEnd, tcell.KeyPgDn, tcell.KeyPgUp:
			// Open the list and forward the event to it.
			d.openList(setFocus)
			if handler := d.list.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
			//d.prefix.SetText(strings.Join(d.list.GetSelectedValues(), ","))
		case tcell.KeyEnter:
			// If the list is closed, open it. Otherwise, forward the event to
			// it.
			if !d.open {
				d.openList(setFocus)
			} else {
				d.list.ToggleCurrentItem()
				d.prefix.SetText(strings.Join(d.list.GetSelectedValues(), ","))
			}
		case tcell.KeyEscape, tcell.KeyTab, tcell.KeyBacktab:
			// Done selecting.
			if d.done != nil {
				d.done(key)
			}
			if d.finished != nil {
				d.finished(key)
			}
			d.closeList(setFocus)
		case tcell.KeyRune:
			if d.open && event.Rune() == ' ' {
				d.list.ToggleCurrentItem()
				d.prefix.SetText(strings.Join(d.list.GetSelectedValues(), ","))
			}
		}
	})
}

// openList hands control over to the embedded List primitive.
func (d *MultiSelectDropDown) openList(setFocus func(Primitive)) {
	if d.open {
		return
	}

	d.open = true

	// d.list.SetSelectedFunc(func(index int, mainText, secondaryText string) {
	// 	if d.dragging {
	// 		return // If we're dragging the mouse, we don't want to trigger any events.
	// 	}

	// 	// An option was selected. Close the list again.
	// 	d.list.currentItem = index
	// 	//d.closeList(setFocus)

	// 	d.prefix.SetText(strings.Join(d.list.GetSelectedValues(), ","))

	// 	// Trigger "selected" event.
	// 	currentOption := d.list.items[d.list.currentItem]
	// 	// if d.selected != nil {
	// 	// 	d.selected(currentOption.MainText, d.list.currentItem)
	// 	// }
	// 	if currentOption.Selected != nil {
	// 		currentOption.Selected()
	// 	}
	// })
	d.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch key := event.Key(); key {
		case tcell.KeyDown, tcell.KeyUp, tcell.KeyPgDn, tcell.KeyPgUp, tcell.KeyHome, tcell.KeyEnd, tcell.KeyEnter: // Basic list navigation.
			break
		case tcell.KeyEscape: // Abort selection.
			d.closeList(setFocus)
			return nil
		}

		return event
	})

	setFocus(d.list)
}

// closeList closes the embedded List element by hiding it and removing focus
// from it.
func (d *MultiSelectDropDown) closeList(setFocus func(Primitive)) {
	if !d.open {
		return
	}
	d.open = false
	if d.list.HasFocus() {
		setFocus(d)
	}
}

// IsOpen returns true if the drop-down list is currently open.
func (d *MultiSelectDropDown) IsOpen() bool {
	return d.open
}

// Focus is called by the application when the primitive receives focus.
func (d *MultiSelectDropDown) Focus(delegate func(p Primitive)) {
	// If we're part of a form and this item is disabled, there's nothing the
	// user can do here so we're finished.
	if d.finished != nil && d.disabled {
		d.finished(-1)
		return
	}

	if d.open {
		delegate(d.list)
	} else {
		d.Box.Focus(delegate)
	}
}

// FocusChain implements the [Primitive]'s FocusChain method.
func (d *MultiSelectDropDown) FocusChain(chain *[]Primitive) bool {
	if d.open {
		if hasFocus := d.list.FocusChain(chain); hasFocus {
			if chain != nil {
				*chain = append(*chain, d)
			}
			return true
		}
	}
	return d.Box.FocusChain(chain)
}

// MouseHandler returns the mouse handler for this primitive.
func (d *MultiSelectDropDown) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return d.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if d.disabled {
			return false, nil
		}

		// Was the mouse event in the drop-down box itself (or on its label)?
		x, y := event.Position()
		inRect := d.InInnerRect(x, y)
		if !d.open && !inRect {
			return d.InRect(x, y), nil // No, and it's not expanded either. Ignore.
		}

		// As long as the drop-down is open, we capture all mouse events.
		if d.open {
			capture = d
		}

		switch action {
		case MouseLeftDown:
			consumed = d.open || inRect
			capture = d
			if !d.open {
				d.openList(setFocus)
				d.dragging = true
			} else if consumed, _ := d.list.MouseHandler()(MouseLeftClick, event, setFocus); !consumed {
				d.closeList(setFocus) // Close drop-down if clicked outside of it.
			}
		case MouseMove:
			if d.dragging {
				// We pretend it's a left click so we can see the selection during
				// dragging. Because we don't act upon it, it's not a problem.
				d.list.MouseHandler()(MouseLeftClick, event, setFocus)
				consumed = true
			}
		case MouseLeftUp:
			if d.dragging {
				d.dragging = false
				d.list.MouseHandler()(MouseLeftClick, event, setFocus)
				consumed = true
			}
		}

		return
	})
}

// PasteHandler returns the handler for this primitive.
func (d *MultiSelectDropDown) PasteHandler() func(pastedText string, setFocus func(p Primitive)) {
	return d.WrapPasteHandler(func(pastedText string, setFocus func(p Primitive)) {
		// if !d.open || d.disabled {
		// 	return
		// }

		// // Strip any newline characters (simple version).
		// pastedText = regexp.MustCompile(`\r?\n`).ReplaceAllString(pastedText, "")

		// // Forward the pasted text to the input field.
		// d.prefix.PasteHandler()(pastedText, setFocus)
	})
}
