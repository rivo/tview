package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const helloWorld = `[green]package[white] main

[green]import[white] (
    [red]"github.com/rivo/tview"[white]
)

[green]func[white] [yellow]main[white]() {
    box := tview.[yellow]NewBox[white]().
        [yellow]SetBorder[white](true).
        [yellow]SetTitle[white]([red]"Hello, world!"[white])
    tview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](box, true).
        [yellow]Run[white]()
}`

// HelloWorld shows a simple "Hello world" example.
func HelloWorld(nextSlide func()) (title string, content tview.Primitive) {
	// We use a text view because we want to capture keyboard input.
	textView := tview.NewTextView().SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	textView.SetBorder(true).SetTitle("Hello, world!")
	return "Hello, world", Code(textView, 30, 10, helloWorld)
}
