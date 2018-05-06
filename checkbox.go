package tview

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
)

// Checkbox implements a simple box for boolean values which can be checked and
// unchecked.
//
// See https://github.com/rivo/tview/wiki/Checkbox for an example.
type Checkbox struct {
	*Box

	align int

	labelFiller string

	lockColors bool

	// Whether or not this box is checked.
	checked bool

	// The text to be displayed before the input area.
	subLabel string

	// The item sub label color.
	subLabelColor tcell.Color

	// The text to be displayed before the input area.
	label string

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// An optional function which is called when the user changes the checked
	// state of this checkbox.
	changed func(checked bool)

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
}

// NewCheckbox returns a new input field.
func NewCheckbox() *Checkbox {
	checkbox := &Checkbox{
		Box:                  NewBox(),
		labelColor:           Styles.LabelTextColor,
		fieldBackgroundColor: Styles.ButtonBackgroundColor,
		fieldTextColor:       Styles.ButtonTextColor,
		align:                AlignLeft,
		labelFiller:          " ",
	}
	checkbox.height = 1
	return checkbox
}

// SetChecked sets the state of the checkbox.
func (c *Checkbox) SetChecked(checked bool) *Checkbox {
	c.checked = checked
	return c
}

// IsChecked returns whether or not the box is checked.
func (c *Checkbox) IsChecked() bool {
	return c.checked
}

// SetLabel sets the text to be displayed before the input area.
func (c *Checkbox) SetLabel(label string) *Checkbox {
	if !strings.Contains(label, "%s") {
		label += "%s"
	}
	c.label = label
	return c
}

// GetLabel returns the text to be displayed before the input area.
func (c *Checkbox) GetLabel() string {
	return c.label
}

// GetLabelWidth returns label width.
func (c *Checkbox) GetLabelWidth() int {
	return StringWidth(strings.Replace(c.subLabel+c.label, "%s", "", -1))
}

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (c *Checkbox) SetLabelWidth(width int) *Checkbox {
	c.labelWidth = width
	return c
}

// SetLabelColor sets the color of the label.
func (c *Checkbox) SetLabelColor(color tcell.Color) *Checkbox {
	c.labelColor = color
	return c
}

// SetFieldAlign sets the input alignment within the checkbox box. This must be
// either AlignLeft, AlignCenter, or AlignRight.
func (c *Checkbox) SetFieldAlign(align int) FormItem {
	c.align = align
	return c
}

// GetFieldAlign returns the input alignment within the checkbox box.
func (c *Checkbox) GetFieldAlign() (align int) {
	return c.align
}

// SetLabelFiller sets a sign which will be fill the label when this one need to stretch
func (c *Checkbox) SetLabelFiller(Filler string) FormItem {
	c.labelFiller = Filler
	return c
}

// SetFieldBackgroundColor sets the background color of the input area.
func (c *Checkbox) SetFieldBackgroundColor(color tcell.Color) *Checkbox {
	c.fieldBackgroundColor = color
	return c
}

// SetFieldTextColor sets the text color of the input area.
func (c *Checkbox) SetFieldTextColor(color tcell.Color) *Checkbox {
	c.fieldTextColor = color
	return c
}

// SetSubLabel sets the text to be displayed before the input area.
func (c *Checkbox) SetSubLabel(label string) *Checkbox {
	c.subLabel = label
	return c
}

// SetSubLabelColor sets the color of the subLabel.
func (c *Checkbox) SetSubLabelColor(color tcell.Color) *Checkbox {
	c.subLabelColor = color
	return c
}

// SetLockColors locks the change of colors by form
func (c *Checkbox) SetLockColors(lock bool) *Checkbox {
	c.lockColors = lock
	return c
}

// SetFormAttributes sets attributes shared by all form items.
func (c *Checkbox) SetFormAttributes(labelWidth, fieldWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	if c.labelWidth == 0 {
		c.labelWidth = labelWidth
	}
	if !c.lockColors {
		c.labelColor = labelColor
		c.backgroundColor = bgColor
		c.fieldTextColor = fieldBgColor
		c.fieldBackgroundColor = fieldTextColor
	}
	return c
}

// GetFieldWidth returns this primitive's field width.
func (c *Checkbox) GetFieldWidth() int {
	return StringWidth(Styles.GraphicsCheckboxUnchecked)
}

// SetChangedFunc sets a handler which is called when the checked state of this
// checkbox was changed by the user. The handler function receives the new
// state.
func (c *Checkbox) SetChangedFunc(handler func(checked bool)) *Checkbox {
	c.changed = handler
	return c
}

// SetDoneFunc sets a handler which is called when the user is done entering
// text. The callback function is provided with the key that was pressed, which
// is one of the following:
//
//   - KeyEscape: Abort text input.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (c *Checkbox) SetDoneFunc(handler func(key tcell.Key)) *Checkbox {
	c.done = handler
	return c
}

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (c *Checkbox) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	c.finished = handler
	return c
}

// GetFinishedFunc returns SetDoneFunc().
func (c *Checkbox) GetFinishedFunc() func(key tcell.Key) {
	return c.finished
}

// Draw draws this primitive onto the screen.
func (c *Checkbox) Draw(screen tcell.Screen) {
	c.Box.Draw(screen)

	// Prepare
	x, y, width, height := c.GetInnerRect()
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}

	// Draw label.
	var labels = []struct {
		text  string
		color tcell.Color
	}{{
		text:  c.subLabel,
		color: c.subLabelColor,
	}, {
		text:  c.label,
		color: c.labelColor,
	}}

	if len(labels) > 0 {
		labelWidth := c.labelWidth
		if labelWidth > rightLimit-x {
			labelWidth = rightLimit - x
		}

		addCount := labelWidth - c.GetLabelWidth()

		for _, label := range labels {
			if addCount > 0 && strings.Contains(label.text, "%s") {
				label.text = fmt.Sprintf(label.text, strings.Repeat(c.labelFiller, addCount))
				addCount = 0
			} else {
				label.text = strings.Replace(label.text, "%s", "", -1)
			}

			labelWidth = StringWidth(label.text)
			Print(screen, label.text, x, y, labelWidth, AlignLeft, label.color)
			x += labelWidth
		}
	}

	// Draw checkbox.
	fieldStyle := tcell.StyleDefault.Background(c.fieldBackgroundColor).Foreground(c.fieldTextColor)
	if c.focus.HasFocus() {
		fieldStyle = fieldStyle.Background(c.fieldTextColor).Foreground(c.fieldBackgroundColor)
	}
	line := Styles.GraphicsCheckboxChecked
	if !c.checked {
		line = Styles.GraphicsCheckboxUnchecked
	}
	width = c.GetFieldWidth()

	for i := 0; i < width; i++ {
		screen.SetContent(x+i, y, rune(line[i]), nil, fieldStyle)
	}
}

// InputHandler returns the handler for this primitive.
func (c *Checkbox) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		// Process key event.
		switch key := event.Key(); key {
		case tcell.KeyRune, tcell.KeyEnter: // Check.
			if key == tcell.KeyRune && event.Rune() != ' ' {
				break
			}
			c.checked = !c.checked
			if c.changed != nil {
				c.changed(c.checked)
			}
		case tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEscape: // We're done.
			if c.done != nil {
				c.done(key)
			}
			if c.finished != nil {
				c.finished(key)
			}
		}
	})
}
