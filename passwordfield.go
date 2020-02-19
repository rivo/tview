package tview

import "github.com/gdamore/tcell"

// NewPasswordField returns a new input field intended for use as a password.
// This is primarily here to make addition of a password field more readable
// when calling Form.AddFormItem(NewPasswordField(),&PasswordFieldArgs{...})
func NewPasswordField() *InputField {
	return NewInputField()
}

// PasswordFieldArgs provides a concise and readable way to initialize
// all the properties of the PasswordField struct by passing to the
// <form>.AddFormItem(item,args).
type PasswordFieldArgs struct {
	baseFormItemArgs

	// The text to be displayed before the input area.
	Label string

	// The text that was entered.
	Text string

	// A character to mask entered text (useful for password fields).
	// A value of 0 disables masking.
	MaskCharacter rune

	// An optional function which is called when the input has changed.
	ChangedFunc func(text string)

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

	// An optional function which is called before the box is drawn.
	ValidateFunc func(*InputField, *tcell.EventKey) bool
}

// applyPasswordFieldArgs applies the values from a PasswordFieldArgs{}
// struct to the associated properties of the PasswordField.
func (i *InputField) applyPasswordFieldArgs(args *PasswordFieldArgs) *InputField {
	i.SetLabel(args.Label)
	i.SetText(args.Text)
	i.SetFieldWidth(args.FieldWidth)
	i.SetMaskCharacter(args.MaskCharacter)
	i.SetChangedFunc(args.ChangedFunc)

	if args.LabelWidth > 0 {
		i.SetLabelWidth(args.LabelWidth)
	}
	if args.LabelColor != 0 {
		i.SetLabelColor(args.LabelColor)
	}
	if args.FieldBackgroundColor != 0 {
		i.SetFieldBackgroundColor(args.FieldBackgroundColor)
	}
	if args.FieldTextColor != 0 {
		i.SetFieldTextColor(args.FieldTextColor)
	}
	if args.BackgroundColor != 0 {
		i.SetBackgroundColor(args.BackgroundColor)
	}
	if args.DoneFunc != nil {
		i.SetDoneFunc(args.DoneFunc)
	}
	if args.DrawFunc != nil {
		i.SetDrawFunc(args.DrawFunc)
	}
	if args.FinishedFunc != nil {
		i.SetFinishedFunc(args.FinishedFunc)
	}
	if args.ValidateFunc != nil {
		i.SetValidateFunc(args.ValidateFunc)
	}
	if args.InputCaptureFunc != nil {
		i.SetInputCapture(args.InputCaptureFunc)
	}
	return i
}
