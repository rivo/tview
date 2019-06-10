package tview

// FormItemArgs is a collection of interfaces and structs to effectively
// provide a concise and readable method to initialize all the properties
// of the various FormItem structs provided by tview, e.g. InputFieldArgs,
// PasswordFieldArgs, DropDownArgs and CheckboxArgs.
//
// This is an alternate to calling the <form>.Add___() methods like
// AddInputField(), AddPasswordField(),  AddDropDown() and  AddCheckbox().
//
// To use these structs pass them to <form>.AddFormItem(item,args)
//
// See https://github.com/rivo/tview/wiki/FormItemArgs for an example.

// FormItemArgsImplementor is an interface that indicates to Go
// that the implementing struct should be allowed to be used as
// a valid `args` parameter in the collection of `Apply___Args()`
// methods.
// Someone implementing their own FormItem struct may also want
// to implement their own associated "Args" struct and implement
// this interface to allow uses of their FormItem to fully
// instantiate it using <form>.AddFormItem(item,args)
type FormItemArgsImplementor interface {
	ImplementsFormItemArgs()
}

// FormItemArgsApplier is an interface that allows someone to
// implement their an ApplyArgs(arg) method for their own custom
// FormItem so that <form>.ApplyFormItemArgs(item,args) can be
// called by the user of the FormItem in the same manner that
// the tview-builtin FormItem types can be used.
type FormItemArgsApplier interface {
	ApplyArgs(FormItemArgsImplementor) FormItem
}

// baseFormItemArgs is merely a convenience struct that can be
// embedded in another struct — e.g. InputFieldArgs, PasswordFieldArgs,
// DropDownArgs and CheckboxArgs — as a shortcut to having to
// add an ImplementsFormItemArgs() method in order to implement
// the FormItemArgsImplementor interface
type baseFormItemArgs struct{}

// ImplementsFormItemArgs implements the FormItemArgsImplementor
// interface in the baseFormItemArgs struct so the struct can be
// embedded in other structs — e.g. InputFieldArgs, PasswordFieldArgs,
// DropDownArgs and CheckboxArgs — as a shortcut to them having
// to each add their own ImplementsFormItemArgs() method.
func (baseFormItemArgs) ImplementsFormItemArgs() {}
