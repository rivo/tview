// Demo code for a simple chat application using TextView regions.
package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MarkovChain implements a simple order of one Markov chain for text
// generation.
type MarkovChain struct {
	Words  []string      // All unique words.
	Starts []int         // Indexes of starting words in Words.
	Chain  map[int][]int // Maps word index to possible next word indexes in Words. Ordered by frequency (highest first).
}

var (
	//go:embed chain.txt
	chainData []byte

	// The Markov chain.
	chain MarkovChain
)

func main() {
	// Load the Markov chain and generate chat lines.
	loadChain()
	ch := generateChat()

	app := tview.NewApplication()

	// Create a chat view.
	chat := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		ScrollToEnd()

	// The users will be displayed separately.
	users := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight).
		ScrollToEnd()

	// Syncing the scrolling is left to the student as an exercise.
	chat.SetInputCapture(func(*tcell.EventKey) *tcell.EventKey {
		return nil
	}).SetMouseCapture(func(tview.MouseAction, *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		return tview.MouseConsumed, nil
	})
	users.SetInputCapture(func(*tcell.EventKey) *tcell.EventKey {
		return nil
	}).SetMouseCapture(func(tview.MouseAction, *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		return tview.MouseConsumed, nil
	})

	// Add chat lines as they are generated.
	go func() {
		var (
			message     int
			userColors  = [...]string{"red", "green", "yellow", "blue", "magenta", "cyan"}
			userToColor = make(map[string]string)
		)
		for line := range ch {
			sleep := time.Second
			if message < 20 {
				sleep = 0
			}
			time.Sleep(sleep)

			// Add the line to the chat view.
			userName := line[0]
			if _, ok := userToColor[userName]; !ok {
				userToColor[userName] = userColors[len(userToColor)%len(userColors)]
			}
			line = line[1:] // Don't display the user name here.
			fmt.Fprintf(chat, "[\"%d\"]%s\n\n", message, strings.Join(line, " "))

			// Add the user name to the users view. Then redraw everything.
			app.QueueUpdateDraw(func() {
				lines := users.GetWrappedLineCount()
				regions := chat.GetRegions(lines, true)
				if len(regions) > 0 {
					lastRegion := regions[len(regions)-1]
					fmt.Fprintf(users, "[%s]%s", userToColor[userName], userName)
					for index := lastRegion.StartRow; index <= lastRegion.EndRow; index++ {
						fmt.Fprintln(users)
					}
				}
			})

			message++
		}
	}()

	// Make a layout and start the application.
	grid := tview.NewGrid().
		SetGap(0, 1).
		SetColumns(0, 8, 40, 0).
		AddItem(users, 0, 1, 1, 1, 0, 0, false).
		AddItem(chat, 0, 2, 1, 1, 0, 0, true)
	grid.SetBorder(true).SetTitle("Chat")

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// loadChain loads the Markov chain from the embedded base64 data.
func loadChain() {
	b := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(chainData))
	zr, err := gzip.NewReader(b)
	if err != nil {
		panic(err)
	}
	defer zr.Close()
	g := gob.NewDecoder(zr)
	if err := g.Decode(&chain); err != nil {
		panic(err)
	}
}

// generateChat generates chat lines based on the Markov chain and sends them
// to a channel. The first word of each line is always the name of a user,
// followed by a colon.
func generateChat() <-chan []string {
	ch := make(chan []string)

	go func() {
		defer close(ch)

		var lastStart int
		for {
			var line []string

			// Pick a random starting word different from the previous one.
			var word int
			for {
				start := chain.Starts[rand.Intn(len(chain.Starts))]
				if start != lastStart {
					word = start
					lastStart = start
					line = append(line, chain.Words[word])
					break
				}
			}

			// Generate the rest of the line.
			for {
				// Pick a random next word.
				next := chain.Chain[word][rand.Intn(len(chain.Chain[word]))]

				// If we hit the end token, emit the line.
				if chain.Words[next] == "$" {
					ch <- line
					break
				}

				// Otherwise, just add the word and continue.
				line = append(line, chain.Words[next])
				word = next
			}
		}
	}()

	return ch
}
