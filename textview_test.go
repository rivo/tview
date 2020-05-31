package tview_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rivo/tview"
)

func TestTextViewMaxLines(t *testing.T) {
	tv := tview.NewTextView()

	// append 100 lines with no limit set:
	for i := 0; i < 100; i++ {
		_, err := tv.Write([]byte(fmt.Sprintf("L%d\n", i)))
		if err != nil {
			t.Errorf("failed to write to TextView: %s", err)
		}
	}

	// retrieve the total text and see we have the 100 lines:
	count := strings.Count(tv.GetText(true), "\n")
	if count != 100 {
		t.Errorf("expected 100 lines, got %d", count)
	}

	// now set the maximum lines to 20, this should clip the buffer:
	tv.SetMaxLines(20)
	// verify buffer was clipped:
	count = len(strings.Split(tv.GetText(true), "\n"))
	if count != 20 {
		t.Errorf("expected 20 lines, got %d", count)
	}

	// append 100 more lines:
	for i := 100; i < 200; i++ {
		_, err := tv.Write([]byte(fmt.Sprintf("L%d\n", i)))
		if err != nil {
			t.Errorf("failed to write to TextView: %s", err)
		}
	}

	// Since max lines is set to 20, we should still get 20 lines:
	txt := tv.GetText(true)
	lines := strings.Split(txt, "\n")
	count = len(lines)
	if count != 20 {
		t.Errorf("expected 20 lines, got %d", count)
	}

	// and those 20 lines should be the last ones:
	if lines[0] != "L181" {
		t.Errorf("expected to get L181, got %s", lines[0])
	}

}
