// Demo code for the TextView primitive.
package main

import (
	"github.com/rivo/tview"
)

const corporate = `			Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.
	В лесу	родилась 	ёлочка	.
\tBring to the table win-win survival strategies to ensure proactive domination. At the end of the day, going forward, a new normal that has evolved from generation X is on the runway heading towards a streamlined cloud solution. User generated content in real-time will have multiple touchpoints for offshoring.

	Capitalize on low hanging fruit to identify a ballpark value added activity to beta test. Override the digital divide with additional clickthroughs from DevOps. Nanotechnology immersion along the information highway will close the loop on focusing solely on the bottom line.

[yellow]Press Enter, then Tab/Backtab for word selections`

func main() {
	app := tview.NewApplication()
	textView := tview.NewTextView().
		SetEditable(true).
		SetWordWrap(false). // TODO: always wrap if editable
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetText(corporate)
	textView.SetBorder(true)
	if err := app.SetRoot(textView, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
