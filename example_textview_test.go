package tview_test

import (
	"fmt"

	"github.com/rivo/tview"
)

func ExampleTextView_BatchWriter() {
	tv := tview.NewTextView()

	w := tv.BatchWriter()
	defer w.Close()
	w.Clear()
	fmt.Fprintln(w, "To sit in solemn silence")
	fmt.Fprintln(w, "on a dull, dark, dock")
	fmt.Println(tv.GetText(false))
	// Output:
	// To sit in solemn silence
	// on a dull, dark, dock
}
