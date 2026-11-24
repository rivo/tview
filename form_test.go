package tview

import "testing"

// TestFormRequestedFocusIsConsumed checks that SetFocus applies only once.
func TestFormRequestedFocusIsConsumed(t *testing.T) {
	for _, button := range []bool{false, true} {
		name := "item"
		if button {
			name = "button"
		}
		t.Run(name, func(t *testing.T) {
			form := NewForm().AddInputField("first", "", 0, nil, nil)
			var requested Primitive
			if button {
				form.AddButton("second", nil)
				requested = form.GetButton(0)
			} else {
				form.AddInputField("second", "", 0, nil, nil)
				requested = form.GetFormItem(1).(*InputField).textArea
			}
			form.SetFocus(1)
			app := NewApplication().SetRoot(form, true)
			if app.GetFocus() != requested {
				t.Fatal("requested element was not focused")
			}
			first := form.GetFormItem(0)
			app.SetFocus(first)
			var got Primitive
			form.Focus(func(p Primitive) { got = p })
			if got != first {
				t.Fatal("consumed request overrode the currently focused item")
			}
		})
	}
}

// TestFormRebuildAfterRequestedFocus checks that a completed request cannot
// prevent focus when a form is rebuilt with fewer elements.
func TestFormRebuildAfterRequestedFocus(t *testing.T) {
	form := NewForm().AddInputField("first", "", 0, nil, nil).AddButton("last", nil).SetFocus(1)
	app := NewApplication().SetRoot(form, true)
	if app.GetFocus() != form.GetButton(0) {
		t.Fatal("button did not receive requested focus")
	}
	form.Clear(true).AddInputField("replacement", "", 0, nil, nil)
	app.SetFocus(form)
	if !form.GetFormItem(0).HasFocus() {
		t.Fatal("stale requested index prevented focus on rebuilt form")
	}
}

// TestFormMissingRequestIsConsumed checks that an unavailable requested index
// is not carried into a later focus operation.
func TestFormMissingRequestIsConsumed(t *testing.T) {
	form := NewForm().AddInputField("first", "", 0, nil, nil).SetFocus(5)
	form.Focus(func(Primitive) { t.Fatal("absent item received focus") })
	var got Primitive
	form.Focus(func(p Primitive) { got = p })
	if got != form.GetFormItem(0) {
		t.Fatal("unused request was not consumed")
	}
}

// TestFormFocusRequestSetDuringDelegation checks that consuming an old request
// before delegation preserves a new request made by the delegate.
func TestFormFocusRequestSetDuringDelegation(t *testing.T) {
	form := NewForm().AddButton("first", nil).AddButton("second", nil).SetFocus(0)
	form.Focus(func(p Primitive) {
		if p != form.GetButton(0) {
			t.Fatal("first request was not honored")
		}
		form.SetFocus(1)
	})
	form.Focus(func(p Primitive) {
		if p != form.GetButton(1) {
			t.Fatal("request made during delegation was lost")
		}
	})
}
