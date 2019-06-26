package tview_test

import (
	"testing"

	"github.com/rivo/tview"
)

func TestSetCurrentItemCallsChangedFunc(t *testing.T) {
	var idx int
	list := tview.NewList().
		AddItem("List item 1", "Some explanatory text", 'a', nil).
		AddItem("List item 2", "Some explanatory text", 'b', nil)
	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		idx = index
	})

	const changeTo int = 1
	list.SetCurrentItem(changeTo)
	if idx != changeTo {
		t.Errorf("Set changed func was not called, or was called with the wrong index: want=%d, got=%d", changeTo, idx)
	}
}
