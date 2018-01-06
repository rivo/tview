package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const inputField = `[green]package[white] main

[green]import[white] (
    [red]"strconv"[white]

    [red]"github.com/gdamore/tcell"[white]
    [red]"github.com/rivo/tview"[white]
)

[green]func[white] [yellow]main[white]() {
    input := tview.[yellow]NewInputField[white]().
        [yellow]SetLabel[white]([red]"Enter a number: "[white]).
        [yellow]SetAcceptanceFunc[white](
            tview.InputFieldInteger,
        ).[yellow]SetDoneFunc[white]([yellow]func[white](key tcell.Key) {
            text := input.[yellow]GetText[white]()
            n, _ := strconv.[yellow]Atoi[white](text)
            [blue]// We have a number.[white]
        })
    tview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](input, true).
        [yellow]Run[white]()
}`

// InputField demonstrates the InputField.
func InputField(nextSlide func()) (title string, content tview.Primitive) {
	input := tview.NewInputField().
		SetLabel("Enter a number: ").
		SetAcceptanceFunc(tview.InputFieldInteger).SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	return "Input", Code(input, 30, 1, inputField)
}
