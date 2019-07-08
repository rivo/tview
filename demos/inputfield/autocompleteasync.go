package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type company struct {
	Name string `json:name`
}

func main() {
	app := tview.NewApplication()
	inputField := tview.NewInputField().
		SetLabel("Enter a company name: ").
		SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			app.Stop()
		})

	// Set up autocomplete function.
	var mutex sync.Mutex
	prefixMap := make(map[string][]string)
	inputField.SetAutocompleteFunc(func(currentText string) []string {
		// Ignore empty text.
		prefix := strings.TrimSpace(strings.ToLower(currentText))
		if prefix == "" {
			return nil
		}

		// Do we have entries for this text already?
		mutex.Lock()
		defer mutex.Unlock()
		entries, ok := prefixMap[prefix]
		if ok {
			return entries
		}

		// No entries yet. Issue a request to the API in a goroutine.
		go func() {
			// Ignore errors in this demo.
			url := "https://autocomplete.clearbit.com/v1/companies/suggest?query=" + url.QueryEscape(prefix)
			res, err := http.Get(url)
			if err != nil {
				return
			}

			// Store the result in the prefix map.
			var companies []*company
			dec := json.NewDecoder(res.Body)
			if err := dec.Decode(&companies); err != nil {
				return
			}
			entries := make([]string, 0, len(companies))
			for _, c := range companies {
				entries = append(entries, c.Name)
			}
			mutex.Lock()
			prefixMap[prefix] = entries
			mutex.Unlock()

			// Trigger an update to the input field.
			inputField.Autocomplete()

			// Also redraw the screen.
			app.Draw()
		}()

		return nil
	})

	if err := app.SetRoot(inputField, true).Run(); err != nil {
		panic(err)
	}
}
