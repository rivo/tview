package tview

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// focusFormItem records focus the way an application does: blur the previous
// primitive, then focus the next one.
func focusFormItem(current *Primitive) func(Primitive) {
	return func(p Primitive) {
		if *current != nil && *current != p {
			(*current).Blur()
		}
		*current = p
		p.Focus(func(Primitive) {})
	}
}

func TestFormScrollableTextViewScrolls(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)
	screen.SetSize(80, 24)

	form := NewForm().
		AddCheckbox("Agree", false, nil).
		AddTextView("Notes", "one\ntwo\nthree\nfour\nfive", 40, 3, false, true).
		AddButton("Save", nil)
	form.SetRect(0, 0, 80, 24)
	form.Draw(screen)

	notes := form.GetFormItem(1).(*TextView)
	var focused Primitive
	setFocus := focusFormItem(&focused)
	form.Focus(setFocus)

	form.InputHandler()(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), setFocus)
	if focused != notes {
		t.Fatal("tab skipped the scrollable text view")
	}

	form.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), setFocus)
	form.Draw(screen)
	row, _ := notes.GetScrollOffset()
	if row != 1 {
		t.Fatalf("scroll offset = %d, want 1", row)
	}
}

func TestFormStaticTextViewIsSkipped(t *testing.T) {
	form := NewForm().
		AddTextView("Notes", "one\ntwo\nthree", 40, 3, false, false).
		AddButton("Save", nil)
	notes := form.GetFormItem(0)
	if !notes.GetDisabled() {
		t.Fatal("a text view that cannot scroll should not take focus")
	}

	var focused Primitive
	form.Focus(focusFormItem(&focused))
	if focused == notes {
		t.Fatal("form focused a text view that cannot scroll")
	}
}
