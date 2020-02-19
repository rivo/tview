package tview

import (
	"github.com/gdamore/tcell"
)

// Checkbox implements a simple box for boolean values which can be checked and
// unchecked.
//
// See https://github.com/rivo/tview/wiki/Checkbox for an example.
type Checkbox struct {
	*Box

	// Whether or not this box is checked.
	checked bool

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

	// A callback function to be called when one of the field exit keys — Enter,
	// Tab, Backtab, or Escape — is used. If this callback returns false it will
	// bypass the input handler and leave the focus on the non-valid field.
	valid func(*Checkbox, *tcell.EventKey) bool
}

// NewCheckbox returns a new input field.
func NewCheckbox() *Checkbox {
	return &Checkbox{
		Box:                  NewBox(),
		labelColor:           Styles.SecondaryTextColor,
		fieldBackgroundColor: Styles.ContrastBackgroundColor,
		fieldTextColor:       Styles.PrimaryTextColor,
	}
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
	c.label = label
	return c
}

// GetLabel returns the text to be displayed before the input area.
func (c *Checkbox) GetLabel() string {
	return c.label
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

// SetFormAttributes sets attributes shared by all form items.
func (c *Checkbox) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	c.labelWidth = labelWidth
	c.labelColor = labelColor
	c.backgroundColor = bgColor
	c.fieldTextColor = fieldTextColor
	c.fieldBackgroundColor = fieldBgColor
	return c
}

// GetFieldWidth returns this primitive's field width.
func (c *Checkbox) GetFieldWidth() int {
	return 1
}

// SetChangedFunc sets a handler which is called when the checked state of this
// checkbox was changed by the user. The handler function receives the new
// state.
func (c *Checkbox) SetChangedFunc(handler func(checked bool)) *Checkbox {
	c.changed = handler
	return c
}

// SetDoneFunc sets a handler which is called when the user is done using the
// checkbox. The callback function is provided with the key that was pressed,
// which is one of the following:
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

// SetValidateFunc sets a callback to be called when one of the field exit keys —
// Enter, Tab, Backtab, or Escape — is used. If this callback returns false it
// will stop the input handler from moving focus to a different field.
func (c *Checkbox) SetValidateFunc(handler func(*Checkbox, *tcell.EventKey) bool) *Checkbox {
	c.valid = handler
	return c
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
	if c.labelWidth > 0 {
		labelWidth := c.labelWidth
		if labelWidth > rightLimit-x {
			labelWidth = rightLimit - x
		}
		Print(screen, c.label, x, y, labelWidth, AlignLeft, c.labelColor)
		x += labelWidth
	} else {
		_, drawnWidth := Print(screen, c.label, x, y, rightLimit-x, AlignLeft, c.labelColor)
		x += drawnWidth
	}

	// Draw checkbox.
	fieldStyle := tcell.StyleDefault.Background(c.fieldBackgroundColor).Foreground(c.fieldTextColor)
	if c.focus.HasFocus() {
		fieldStyle = fieldStyle.Background(c.fieldTextColor).Foreground(c.fieldBackgroundColor)
	}
	checkedRune := 'X'
	if !c.checked {
		checkedRune = ' '
	}
	screen.SetContent(x, y, checkedRune, nil, fieldStyle)
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

// CheckboxArgs provides a concise and readable way to initialize
// all the properties of the Checkbox struct by passing to the
// <form>.AddFormItem(item,args).
type CheckboxArgs struct {
	baseFormItemArgs

	// The text to be displayed before the input area.
	Label string

	// Whether or not this box is checked.
	Checked bool

	// An optional function which is called when the user changes
	// the checked state of this checkbox.
	ChangedFunc func(checked bool)

	// The screen width of the input area. A value of 0 means extend
	// as much as possible.
	FieldWidth int

	// The screen width of the label area. A value of 0 means use
	// the width of the label text.
	LabelWidth int

	// The label color.
	LabelColor tcell.Color

	// The background color of the input area.
	FieldBackgroundColor tcell.Color

	// The text color of the input area.
	FieldTextColor tcell.Color

	// The background color of the input area.
	BackgroundColor tcell.Color

	// An optional function which is called when the user indicated
	// that they are done entering text. The key which was pressed
	// is provided (tab, shift-tab, enter, or escape).
	DoneFunc func(key tcell.Key)

	// A callback function set by the Form class and called when
	// the user leaves this form item.
	FinishedFunc func(key tcell.Key)

	// An optional function which is called before the box is drawn.
	DrawFunc func(screen tcell.Screen, x, y, width, height int) (int, int, int, int)

	// An optional capture function which receives a key event and
	// returns the event to be forwarded to the primitive's default
	// input handler (nil if nothing should be forwarded).
	InputCaptureFunc func(event *tcell.EventKey) *tcell.EventKey

	// A callback function to be called when one of the field exit keys — Enter,
	// Tab, Backtab, or Escape — is used. If this callback returns false it will
	// bypass the input handler and leave the focus on the non-valid field.
	ValidateFunc func(*Checkbox, *tcell.EventKey) bool
}

// ApplyArgs applies the values from a CheckboxArgs{} struct to the
// associated properties of the Checkbox.
func (c *Checkbox) ApplyArgs(args *CheckboxArgs) *Checkbox {

	c.SetLabel(args.Label)
	c.SetChecked(args.Checked)
	c.SetChangedFunc(args.ChangedFunc)

	if args.LabelWidth > 0 {
		c.SetLabelWidth(args.LabelWidth)
	}
	if args.LabelColor != 0 {
		c.SetLabelColor(args.LabelColor)
	}
	if args.FieldBackgroundColor != 0 {
		c.SetFieldBackgroundColor(args.FieldBackgroundColor)
	}
	if args.FieldTextColor != 0 {
		c.SetFieldTextColor(args.FieldTextColor)
	}
	if args.BackgroundColor != 0 {
		c.SetBackgroundColor(args.BackgroundColor)
	}
	if args.DoneFunc != nil {
		c.SetDoneFunc(args.DoneFunc)
	}
	if args.DrawFunc != nil {
		c.SetDrawFunc(args.DrawFunc)
	}
	if args.FinishedFunc != nil {
		c.SetFinishedFunc(args.FinishedFunc)
	}
	if args.ValidateFunc != nil {
		c.SetValidateFunc(args.ValidateFunc)
	}
	if args.InputCaptureFunc != nil {
		c.SetInputCapture(args.InputCaptureFunc)
	}
	return c
}
