package main

import (
	"github.com/rivo/tview"
)

const form = `[green]package[white] main

[green]import[white] (
    [red]"github.com/rivo/tview"[white]
)

[green]func[white] [yellow]main[white]() {
    form := tview.[yellow]NewForm[white]().
        [yellow]AddInputField[white]([red]"First name:"[white], [red]""[white], [red]20[white], nil).
        [yellow]AddInputField[white]([red]"Last name:"[white], [red]""[white], [red]20[white], nil).
        [yellow]AddDropDown[white]([red]"Role:"[white], [][green]string[white]{
            [red]"Engineer"[white],
            [red]"Manager"[white],
            [red]"Administration"[white],
        }, [red]0[white], nil).
        [yellow]AddCheckbox[white]([red]"On vacation:"[white], false, nil).
        [yellow]AddButton[white]([red]"Save"[white], [yellow]func[white]() { [blue]/* Save data */[white] }).
        [yellow]AddButton[white]([red]"Cancel"[white], [yellow]func[white]() { [blue]/* Cancel */[white] })
    tview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](form, true).
        [yellow]SetFocus[white](form).
        [yellow]Run[white]()
}`

// Form demonstrates forms.
func Form(nextSlide func()) (title string, content tview.Primitive) {
	f := tview.NewForm().
		AddInputField("First name:", "", 20, nil).
		AddInputField("Last name:", "", 20, nil).
		AddDropDown("Role:", []string{"Engineer", "Manager", "Administration"}, 0, nil).
		AddCheckbox("On vacation:", false, nil).
		AddButton("Save", nextSlide).
		AddButton("Cancel", nextSlide)
	f.SetBorder(true).SetTitle("Employee Information")
	return "Forms", Code(f, 36, 13, form)
}
