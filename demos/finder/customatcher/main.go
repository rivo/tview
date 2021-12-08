package main

import (
	"github.com/rivo/tview"
	"regexp"
)

func main() {

	app := tview.NewApplication()

	items := []string{
		"foo",
		"bar",
		"foo bar foo",
		"baz",
		"baz two",
		"lorem ipsum",
	}

	// Don't use the default matcher (which will just return the first substring indices and uses a fixed score of 0).
	// This custom implementation uses a regex expression and returns all matching substring indices.
	// Moreover, it uses the number of matches for each item as score value (the default matcher assumes score 0 for
	// every item). Thus, items with more matches will be sorted before items with fewer matches (e.g. when searching
	// for "foo", item "foo bar foo" will be sorted before "foo".
	matcher := func(item string, filter string) ([][2]int, int, bool) {

		re := regexp.MustCompile(filter)

		matchedIndices := re.FindAllStringSubmatchIndex(item, -1)
		if matchCount := len(matchedIndices); matchCount > 0 {
			matches := make([][2]int, matchCount)
			for i := range matchedIndices {
				matches[i] = [2]int{matchedIndices[i][0], matchedIndices[i][1]}
			}
			return matches, len(matchedIndices), true
		}

		// no match
		return [][2]int{}, 0, false
	}

	finder := tview.NewFinder().
		SetItems(len(items), func(index int) string {
			return items[index]
		}).
		SetDoneFunc(func(index int) {
			app.Stop()
		}).
		SetMatcherFunc(matcher)

	if err := app.SetRoot(finder, true).Run(); err != nil {
		panic(err)
	}
}
