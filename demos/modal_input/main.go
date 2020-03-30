// Demo code for the ModalInput primitive.
package main

import (
	"fmt"

	"github.com/rivo/tview"
)

func main() {
	response := ""
	app := tview.NewApplication()
	modal := tview.NewModalInput("Colour", "Pink").
		SetText("What's your favourite colour?").
		AddButtons([]string{"Ok", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel, answer string) {
			response = answer
			if buttonLabel == "Ok" {
				app.Stop()
			}
		})
	if err := app.SetRoot(modal, false).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
	fmt.Printf("User entered: %v\n", response)
}
