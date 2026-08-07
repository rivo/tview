package tview

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

var (
	// RadioCheckedString and RadioUncheckedString are visible characters of checked and unchecked
	// radio buttons. They can be set globally for the whole app through these variables.
	RadioCheckedString   = "\u25c9"
	RadioUncheckedString = "\u25ef"
)

type Radio struct {
	*Box

	// The currently selected value.
	value int

	// The text to be displayed before the input area.
	label string

	// The list of choosable texts.
	options []string

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// A callback function when the user changes the radion button value.
	onSetValue func(int)

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// If set to true, options are positioned from left to right, instead of top to bottom.
	horizontal bool

	// A callback function set by the Form class and called when the user leaves this form item.
	finished func(tcell.Key)
}

// NewRadio creates a radio button group with the given options.
func NewRadio(options ...string) *Radio {
	if len(options) == 0 {
		options = []string{"noOptions"}
	}
	return &Radio{
		Box:                  NewBox(),
		options:              options,
		labelColor:           Styles.SecondaryTextColor,
		fieldBackgroundColor: Styles.ContrastBackgroundColor,
		fieldTextColor:       Styles.PrimaryTextColor,
	}
}

// SetValue sets the current value of the radio group.
func (r *Radio) SetValue(value int) *Radio {
	if r.value == value {
		return r
	}
	if value < 0 {
		value = 0
	} else if value >= len(r.options) {
		value = len(r.options) - 1
	}
	r.changeValue(value)
	return r
}

// Value returns current radio value.
func (r *Radio) Value() int {
	return r.value
}

// changeValue changes the current value, and calls change callback if exists.
func (r *Radio) changeValue(value int) {
	r.value = value
	if r.onSetValue != nil {
		r.onSetValue(value)
	}
}

// SetOnSetValue sets callback handler of a value change.
func (r *Radio) SetOnSetValue(handler func(int)) *Radio {
	r.onSetValue = handler
	return r
}

// SetHorizontal sets the direction the options are laid out. If set to true, instead
// of positioning them from top to bottom (the default), they are positioned from left
// to right, moving into the next row if there is not enough space.
func (r *Radio) SetHorizontal(horizontal bool) *Radio {
	r.horizontal = horizontal
	return r
}

// InputHandler returns the handler for this primitive.
func (r *Radio) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return r.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		key := event.Key()
		if r.value > 0 &&
			((key == tcell.KeyLeft && r.horizontal) ||
				(key == tcell.KeyUp && !r.horizontal)) {
			r.changeValue(r.value - 1)
			return
		}
		if r.value < len(r.options)-1 &&
			((key == tcell.KeyRight && r.horizontal) ||
				(key == tcell.KeyDown && !r.horizontal)) {
			r.changeValue(r.value + 1)
			return
		}
		switch key {
		case tcell.KeyEnter, tcell.KeyTab, tcell.KeyBacktab:
			if r.finished != nil {
				r.finished(key)
			}
		}
	})
}

func (r *Radio) GetLabel() string {
	return r.label
}

func (r *Radio) SetLabel(l string) *Radio {
	r.label = l
	return r
}

// SetFormAttributes sets attributes shared by all form items.
func (r *Radio) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	r.labelWidth = labelWidth
	r.labelColor = labelColor
	r.backgroundColor = bgColor
	r.fieldTextColor = fieldTextColor
	r.fieldBackgroundColor = fieldBgColor
	return r
}

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (r *Radio) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	r.finished = handler
	return r
}

// GetFieldHeight returns this primitive's field height.
func (r *Radio) GetFieldHeight() int {
	if r.horizontal {
		return 1
	}
	return len(r.options)
}

// GetFieldWidth returns this primitive's field width.
func (r *Radio) GetFieldWidth() int {
	w := 0
	for _, option := range r.options {
		if r.horizontal {
			w += len(option) + 3 // checkbox + space + option + space
			continue
		}
		if len(option) > w {
			w = len(option)
		}
	}
	if r.horizontal {
		return w - 1
	}
	return w + 2
}

func (r *Radio) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()
	if width < 1 || height < 1 {
		return
	}

	// Draw label.
	var labelBg tcell.Color
	labelStyle := tcell.StyleDefault.Background(r.fieldBackgroundColor).Foreground(r.labelColor)
	if r.hasFocus {
		labelBg = Styles.MoreContrastBackgroundColor
		labelStyle = labelStyle.Background(Styles.InverseTextColor)
	} else {
		_, labelBg, _ = tcell.StyleDefault.Decompose()
	}
	if r.labelWidth > 0 {
		labelWidth := r.labelWidth
		if labelWidth > width {
			labelWidth = width
		}
		printWithStyle(screen, r.label, x, y, 0, labelWidth, AlignLeft, labelStyle, labelBg == tcell.ColorDefault)
		x += labelWidth
	} else {
		_, drawnWidth, _, _ := printWithStyle(screen, r.label, x, y, 0, width, AlignLeft, labelStyle, labelBg == tcell.ColorDefault)
		x += drawnWidth
	}

	// Draw radio buttons.
	fieldStyle := tcell.StyleDefault.Background(r.fieldBackgroundColor).Foreground(r.fieldTextColor)
	for i, option := range r.options {
		rb := RadioUncheckedString
		if i == r.value {
			rb = RadioCheckedString
		}
		line := fmt.Sprintf("%s %s", rb, option)
		printWithStyle(screen, line, x, y, 0, width, AlignLeft, fieldStyle, !r.hasFocus || i != r.value)
		if r.horizontal {
			x += uniseg.GraphemeClusterCount(line) + 1
		} else {
			y += 1
		}
	}
}

// MouseHandler returns the mouse handler for this primitive.
func (r *Radio) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return r.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if action != MouseLeftDown && action != MouseLeftClick {
			return // only interested in these two
		}
		x, y := event.Position()
		if !r.InRect(x, y) {
			return // out of widget
		}
		if action == MouseLeftDown {
			setFocus(r) // mouse down then moved: focus radio only
			return true, nil
		}
		rectX, rectY, _, _ := r.GetRect()
		x -= rectX + r.labelWidth
		y -= rectY
		if x < 0 {
			return // clicked on the label
		}
		// countOptLen counts this option's width
		countOptLen := func(i int, option string) int {
			res := 0
			if i != r.value {
				res += uniseg.GraphemeClusterCount(RadioUncheckedString)
			} else {
				res += uniseg.GraphemeClusterCount(RadioCheckedString)
			}
			res++
			res += uniseg.GraphemeClusterCount(option)
			return res
		}
		if !r.horizontal {
			if y < 0 || len(r.options) <= y { // shouldn't be necessary, make sure not to index out
				return
			}
			if x >= countOptLen(y, r.options[y]) {
				return // clicked to the right of this option
			}
			r.SetValue(y) // clicked on this option
			return true, nil
		}
		if y != 0 {
			return // horizontal radio means single line
		}
		for i, option := range r.options { // sum option widths until match
			x -= countOptLen(i, option)
			if x < 0 { // match
				r.SetValue(i)
				return true, nil
			}
			if x == 0 { // the character between two options
				return
			}
			x--
		}
		return // not found
	})
}
