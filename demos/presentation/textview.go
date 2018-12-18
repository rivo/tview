package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const textView1 = `[green]func[white] [yellow]main[white]() {
	app := tview.[yellow]NewApplication[white]()
    textView := tview.[yellow]NewTextView[white]().
        [yellow]SetTextColor[white](tcell.ColorYellow).
        [yellow]SetScrollable[white](false).
        [yellow]SetChangedFunc[white]([yellow]func[white]() {
            app.[yellow]Draw[white]()
        })
    [green]go[white] [yellow]func[white]() {
        [green]var[white] n [green]int
[white]        [yellow]for[white] {
            n++
            fmt.[yellow]Fprintf[white](textView, [red]"%d "[white], n)
            time.[yellow]Sleep[white]([red]200[white] * time.Millisecond)
        }
    }()
    app.[yellow]SetRoot[white](textView, true).
        [yellow]Run[white]()
}`

// TextView1 demonstrates the basic text view.
func TextView1(nextSlide func()) (title string, content tview.Primitive) {
	textView := tview.NewTextView().
		SetTextColor(tcell.ColorYellow).
		SetScrollable(false).
		SetDoneFunc(func(key tcell.Key) {
			nextSlide()
		})
	textView.SetChangedFunc(func() {
		if textView.HasFocus() {
			app.Draw()
		}
	})
	go func() {
		var n int
		for {
			n++
			fmt.Fprintf(textView, "%d ", n)
			time.Sleep(200 * time.Millisecond)
		}
	}()
	textView.SetBorder(true).SetTitle("TextView implements io.Writer")
	return "Text 1", Code(textView, 36, 13, textView1)
}

const textView2 = `[green]package[white] main

[green]import[white] (
    [red]"strconv"[white]

    [red]"github.com/gdamore/tcell"[white]
    [red]"github.com/rivo/tview"[white]
)

[green]func[white] [yellow]main[white]() {
    ["0"]textView[""] := tview.[yellow]NewTextView[white]()
    ["1"]textView[""].[yellow]SetDynamicColors[white](true).
        [yellow]SetWrap[white](false).
        [yellow]SetRegions[white](true).
        [yellow]SetDoneFunc[white]([yellow]func[white](key tcell.Key) {
            highlights := ["2"]textView[""].[yellow]GetHighlights[white]()
            hasHighlights := [yellow]len[white](highlights) > [red]0
            [yellow]switch[white] key {
            [yellow]case[white] tcell.KeyEnter:
                [yellow]if[white] hasHighlights {
                    ["3"]textView[""].[yellow]Highlight[white]()
                } [yellow]else[white] {
                    ["4"]textView[""].[yellow]Highlight[white]([red]"0"[white]).
                        [yellow]ScrollToHighlight[white]()
                }
            [yellow]case[white] tcell.KeyTab:
                [yellow]if[white] hasHighlights {
                    current, _ := strconv.[yellow]Atoi[white](highlights[[red]0[white]])
                    next := (current + [red]1[white]) % [red]9
                    ["5"]textView[""].[yellow]Highlight[white](strconv.[yellow]Itoa[white](next)).
                        [yellow]ScrollToHighlight[white]()
                }
            [yellow]case[white] tcell.KeyBacktab:
                [yellow]if[white] hasHighlights {
                    current, _ := strconv.[yellow]Atoi[white](highlights[[red]0[white]])
                    next := (current - [red]1[white] + [red]9[white]) % [red]9
                    ["6"]textView[""].[yellow]Highlight[white](strconv.[yellow]Itoa[white](next)).
                        [yellow]ScrollToHighlight[white]()
                }
            }
        })
    fmt.[yellow]Fprint[white](["7"]textView[""], content)
    tview.[yellow]NewApplication[white]().
        [yellow]SetRoot[white](["8"]textView[""], true).
        [yellow]Run[white]()
}`

// TextView2 demonstrates the extended text view.
func TextView2(nextSlide func()) (title string, content tview.Primitive) {
	codeView := tview.NewTextView().
		SetWrap(false)
	fmt.Fprint(codeView, textView2)
	codeView.SetBorder(true).SetTitle("Buffer content")

	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetWrap(false).
		SetRegions(true).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				nextSlide()
				return
			}
			highlights := textView.GetHighlights()
			hasHighlights := len(highlights) > 0
			switch key {
			case tcell.KeyEnter:
				if hasHighlights {
					textView.Highlight()
				} else {
					textView.Highlight("0").
						ScrollToHighlight()
				}
			case tcell.KeyTab:
				if hasHighlights {
					current, _ := strconv.Atoi(highlights[0])
					next := (current + 1) % 9
					textView.Highlight(strconv.Itoa(next)).
						ScrollToHighlight()
				}
			case tcell.KeyBacktab:
				if hasHighlights {
					current, _ := strconv.Atoi(highlights[0])
					next := (current - 1 + 9) % 9
					textView.Highlight(strconv.Itoa(next)).
						ScrollToHighlight()
				}
			}
		})
	fmt.Fprint(textView, textView2)
	textView.SetBorder(true).SetTitle("TextView output")
	return "Text 2", tview.NewFlex().
		AddItem(textView, 0, 1, true).
		AddItem(codeView, 0, 1, false)
}
