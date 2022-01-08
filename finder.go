package tview

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"regexp"
	"sort"
	"strings"
)

const (
	finderLabelDefault                = " "
	finderSelectedLabelDefault        = "->"
	finderSelectedLabelPaddingDefault = 1

	finderInputPlaceholderDefault = "Type here..."
)

// MatcherFunction is called in order to check if parts of an item cam be matched
// to the passed filter string. It returns a slice of all successive matches, the
// sore to sort items according to their similarity and a boolean value indicating
// if there was a match at all. [2]int represents the start position (inclusive)
// and end position (exclusive) within the item string.
type MatcherFunction func(item string, filter string) ([][2]int, int, bool)

// ItemNameProviderFunction is called in order to retrieve the name of the item
// at the passed index. The string value returned from this function will be
// displayed in the finder list.
type ItemNameProviderFunction func(index int) string

// matched is used internally to describe matches between an Item and filter string
type matched struct {
	// idx is the index of an item of the original slice which was used to
	// search matched strings.
	idx int
	// matches is a slice of all successive matches
	// [2]int represents an interval of the match position.
	matches [][2]int
	// score indicates how similar the match is to the actual item. This
	// can be used to sort matches.
	score int
}

type Finder struct {
	*Box

	// The total number of items available.
	itemCount int

	// The function to be invoked when matching items against the entered updateMatches text.
	matcherFunction MatcherFunction

	// The text to be displayed before the input area.
	inputLabel string

	// The right padding of the inputLabel (space between label and input field)
	inputLabelPadding int

	// The inputLabel style.
	inputLabelStyle tcell.Style

	// The text to be displayed in the input area when "text" is empty.
	placeholder string

	// The style of the input area with placeholder text.
	placeholderStyle tcell.Style

	// The style of the input area with input text.
	fieldStyle tcell.Style

	// The text to be displayed before an item which is not selected.
	itemLabel string

	// The right padding of the itemLabel (space between label and item text)
	itemLabelPadding int

	// The style of a default list item which is not selected.
	itemStyle tcell.Style

	// The style of the label to be displayed before a (not selected) item.
	itemLabelStyle tcell.Style

	// The style of list item which is selected.
	selectedItemStyle tcell.Style

	// The text to be displayed before a selected item.
	selectedItemLabel string

	// The right padding of the selectedItemLabel (space between label and item text)
	selectedItemLabelPadding int

	// The style of the label to be displayed before a selected item.
	selectedItemLabelStyle tcell.Style

	// The style of the counter.
	counterStyle tcell.Style

	// The style used to highlight matches between items and the updateMatches text
	highlightMatchStyle tcell.Style

	// If false, the background color of highlightMatchStyle will be applied.
	highlightMatchMaintainBackgroundColor bool

	// If true, the entire row is highlighted when selected.
	highlightFullLine bool

	// If true, the selection is only shown when the finder primitive has focus.
	selectedFocusOnly bool

	// Whether navigating the list will wrap around.
	wrapAround bool

	// The index of the currently selected item in matched.
	selectedIndex int

	// The currently selected matched entry.
	selectedMatch *matched

	// An optional function which is called when the user has navigated to a new item.
	changed func(index int)

	// An optional function which is called when the user indicated that they
	// are done entering text. The index of the item which was selected most recently is provided.
	done func(index int)

	// internal properties
	// ----------------------------

	// All items described with each item mapped to a string by means of a ItemNameProviderFunction.
	rawItems []string

	// All available matchable items.
	allItems []matched

	// All currently matched items. Equals allItems if no filterText is provided.
	matched []matched

	// The number of list items skipped at the top before the first item is  drawn.
	itemOffset int

	// The text that was entered to updateMatches the entries.
	filterText string

	// The cursor position as a byte index into the text string.
	cursorPos int

	// The number of bytes of the text string skipped ahead while drawing.
	offset int
}

// NewFinder returns a new finder.
func NewFinder() *Finder {

	search := &Finder{
		Box:                      NewBox(),
		matcherFunction:          defaultMatcher,
		inputLabelStyle:          tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.SecondaryTextColor),
		placeholderStyle:         tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.ContrastSecondaryTextColor),
		fieldStyle:               tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
		itemLabelStyle:           tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.SecondaryTextColor),
		itemStyle:                tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.PrimaryTextColor),
		selectedItemStyle:        tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.PrimitiveBackgroundColor),
		selectedItemLabelStyle:   tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.TertiaryTextColor),
		counterStyle:             tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.SecondaryTextColor),
		highlightMatchStyle:      tcell.StyleDefault.Background(Styles.PrimitiveBackgroundColor).Foreground(Styles.TertiaryTextColor).Bold(true),
		inputLabel:               finderSelectedLabelDefault,
		selectedItemLabel:        finderSelectedLabelDefault,
		itemLabel:                finderLabelDefault,
		placeholder:              finderInputPlaceholderDefault,
		inputLabelPadding:        finderSelectedLabelPaddingDefault,
		itemLabelPadding:         len(finderSelectedLabelDefault) + finderSelectedLabelPaddingDefault - len(finderLabelDefault),
		selectedItemLabelPadding: finderSelectedLabelPaddingDefault,
		selectedIndex:            -1,
		highlightFullLine:        false,
		selectedFocusOnly:        true,
		wrapAround:               true,
	}

	return search
}

// SetItems sets the available number of items and a function to provide a name of a specific item by index.
func (f *Finder) SetItems(itemCount int, provider ItemNameProviderFunction) *Finder {

	f.itemCount = itemCount
	f.allItems = make([]matched, itemCount)
	f.rawItems = make([]string, itemCount)

	for i := 0; i < itemCount; i++ {
		f.allItems[i] = matched{idx: i}
		f.rawItems[i] = provider(i)
	}

	f.matched = f.allItems

	f.SetCurrentItem(0)

	return f
}

// SetInputLabel sets the text to be displayed before the input area.
func (f *Finder) SetInputLabel(inputLabel string) *Finder {
	f.inputLabel = inputLabel
	return f
}

// SetInputPadding sets the right padding of the inputLabel (space between label and input field),
func (f *Finder) SetInputPadding(padding int) *Finder {
	f.inputLabelPadding = padding
	return f
}

// SetInputLabelStyle sets the style of the input label.
func (f *Finder) SetInputLabelStyle(style tcell.Style) *Finder {
	f.inputLabelStyle = style
	return f
}

// SetPlaceholder sets the text to be displayed when the input text is empty.
func (f *Finder) SetPlaceholder(placeholder string) *Finder {
	f.placeholder = placeholder
	return f
}

// SetPlaceholderStyle sets the style of the input area (when a placeholder is
// shown).
func (f *Finder) SetPlaceholderStyle(style tcell.Style) *Finder {
	f.placeholderStyle = style
	return f
}

// SetFieldStyle sets the style of the input area (when no placeholder is
// shown).
func (f *Finder) SetFieldStyle(style tcell.Style) *Finder {
	f.fieldStyle = style
	return f
}

// SetItemLabel sets the text to be displayed before an item which is not selected.
func (f *Finder) SetItemLabel(label string) *Finder {
	f.itemLabel = label
	return f
}

// SetItemLabelPadding sets the right padding of the itemLabel (space between label and item text),
func (f *Finder) SetItemLabelPadding(itemLabelPadding int) *Finder {
	f.itemLabelPadding = itemLabelPadding
	return f
}

// SetItemLabelStyle sets the style of the label to be displayed before a (not selected) item.
func (f *Finder) SetItemLabelStyle(style tcell.Style) *Finder {
	f.itemLabelStyle = style
	return f
}

// SetItemStyle sets the style of a list item when not selected.
func (f *Finder) SetItemStyle(style tcell.Style) *Finder {
	f.itemStyle = style
	return f
}

// SetSelectedItemStyle sets the style of a list item when selected.
func (f *Finder) SetSelectedItemStyle(style tcell.Style) *Finder {
	f.selectedItemStyle = style
	return f
}

// SetSelectedItemLabel sets the text to be displayed before the currently selected item.
func (f *Finder) SetSelectedItemLabel(label string) *Finder {
	f.selectedItemLabel = label
	return f
}

// SetSelectedItemLabelPadding sets the right padding of the selectedItemLabel (space between label and item text),
func (f *Finder) SetSelectedItemLabelPadding(padding int) *Finder {
	f.selectedItemLabelPadding = padding
	return f
}

// SetSelectedItemLabelStyle sets the style of the label to be displayed before a selected item.
func (f *Finder) SetSelectedItemLabelStyle(style tcell.Style) *Finder {
	f.selectedItemLabelStyle = style
	return f
}

// SetCounterStyle sets the style of the counter.
func (f *Finder) SetCounterStyle(style tcell.Style) *Finder {
	f.counterStyle = style
	return f
}

// SetHighlightMatchStyle sets the style used to highlight matches between items and the filter text
func (f *Finder) SetHighlightMatchStyle(style tcell.Style) *Finder {
	f.highlightMatchStyle = style
	return f
}

// SetHighlightMatchMaintainBackgroundColor defines whether the background color of highlightMatchStyle
// will be applied. If set to true, the background color of itemStyle or selectedItemStyle will be
// maintained. If false, the background color of highlightMatchStyle will be applied.
func (f *Finder) SetHighlightMatchMaintainBackgroundColor(maintainBackgroundColor bool) *Finder {
	f.highlightMatchMaintainBackgroundColor = maintainBackgroundColor
	return f
}

// SetHighlightFullLine sets a flag which determines whether the colored
// background of selected items spans the entire width of the view. If set to
// true, the highlight spans the entire view. If set to false, only the text of
// the selected item from beginning to end is highlighted.
func (f *Finder) SetHighlightFullLine(highlight bool) *Finder {
	f.highlightFullLine = highlight
	return f
}

// SetSelectedFocusOnly sets a flag which determines when the currently selected
// finder item is highlighted. If set to true, selected items are only highlighted
// when the finder has focus. If set to false, they are always highlighted.
func (f *Finder) SetSelectedFocusOnly(focusOnly bool) *Finder {
	f.selectedFocusOnly = focusOnly
	return f
}

// SetWrapAround sets the flag that determines whether navigating the finder list will
// wrap around. That is, navigating downwards on the last item will move the
// selection to the first item (similarly in the other direction). If set to
// false, the selection won't change when navigating downwards on the last item
// or navigating upwards on the first item.
func (f *Finder) SetWrapAround(wrapAround bool) *Finder {
	f.wrapAround = wrapAround
	return f
}

// SetCurrentItem sets the currently selected item by its index, starting at 0
// for the first item. If a negative index is provided, items are referred to
// from the back (-1 = last item, -2 = second-to-last item, and so on). Out of
// range indices are clamped to the beginning/end.
//
// Calling this function triggers a "changed" event if the selection changes.
func (f *Finder) SetCurrentItem(index int) *Finder {
	if index < 0 {
		index = f.itemCount + index
	}
	if index >= len(f.matched) {
		index = len(f.matched) - 1
	}
	if index < 0 && index > len(f.matched) {
		index = 0
	}

	var selected *matched
	if index >= 0 && index < len(f.matched) {
		selected = &f.matched[index]
	}
	prevSelected := f.selectedMatch

	if selected != prevSelected && f.changed != nil {
		if index >= 0 {
			f.selectedMatch = selected
			f.changed(f.matched[index].idx)
		} else {
			f.selectedMatch = nil
			f.changed(-1)
		}
	}

	f.selectedIndex = index
	f.selectedMatch = selected
	return f
}

// GetCurrentItem returns the index of the currently selected finder item,
// starting at 0 for the first item. Returns -1 if no item is selected (this
// only occurs when the list is empty).
func (f *Finder) GetCurrentItem() int {
	return f.selectedIndex
}

// SetMatcherFunc sets the function which is called in order to match the
// filter string with all available items.
func (f *Finder) SetMatcherFunc(matcher MatcherFunction) *Finder {
	f.matcherFunction = matcher
	return f
}

// SetChangedFunc sets the function which is called when the user navigates to
// an item. The function receives the item's index in the list of items
// (starting with 0, -1 when no item is selected).
func (f *Finder) SetChangedFunc(handler func(index int)) *Finder {
	f.changed = handler
	if f.changed != nil {
		f.changed(f.matched[f.selectedIndex].idx)
	}
	return f
}

// SetDoneFunc sets a handler which is called when the user is done entering
// text. The callback function is provided with the index of the most recently
// selected item. The index is -1 when no item was selected.
func (f *Finder) SetDoneFunc(handler func(index int)) *Finder {
	f.done = handler
	return f
}

// updateMatches refreshes matched according to the current filterText.
func (f *Finder) updateMatches() {
	scoresProvided := false
	if f.filterText != "" {
		var newMatched []matched
		for i := range f.rawItems {
			if matches, score, ok := f.matcherFunction(f.rawItems[i], f.filterText); ok && len(matches) > 0 {
				newMatched = append(newMatched, matched{idx: i, score: score, matches: matches})
				if !scoresProvided && score > 0 {
					scoresProvided = true
				}
			}
		}

		if scoresProvided {
			sort.Slice(newMatched, func(i, j int) bool {
				return newMatched[i].score > newMatched[j].score
			})
		}

		f.matched = newMatched
	} else {
		f.matched = f.allItems
	}

	if f.selectedIndex < 0 && len(f.matched) > 0 {
		f.SetCurrentItem(0)
	} else if s := len(f.matched) - 1; s < f.selectedIndex {
		f.SetCurrentItem(s)
	} else {
		f.SetCurrentItem(f.selectedIndex)
	}
}

// defaultMatcher is a simple default implementation for MatcherFunction. It just
// returns the first substring indices of text matching filter (case-insensitive). Score is
// always 0.
func defaultMatcher(text string, filter string) ([][2]int, int, bool) {
	if index := strings.Index(strings.ToLower(text), strings.ToLower(filter)); index >= 0 {
		return [][2]int{{index, index + len(filter)}}, 0, true
	}
	return [][2]int{}, 0, false
}

// Draw draws this primitive onto the screen.
func (f *Finder) Draw(screen tcell.Screen) {
	f.Box.DrawForSubclass(screen, f)

	// Determine the dimensions.
	x, y, width, height := f.GetInnerRect()
	if _, totalHeight := screen.Size(); totalHeight < height {
		height = totalHeight
	}

	maxX := x + width - 1
	// Start drawing at the bottom of the scree
	currentY := y + height - 1
	drawnLines := 0

	// Draw the input field & decrease currentY by 1 since the input field was draw right at the bottom
	f.drawInputField(screen)
	currentY--
	drawnLines += 1

	// Draw the counter
	printWithStyle(
		screen,
		fmt.Sprintf("%d/%d", len(f.matched), f.itemCount),
		f.inputOffset(), currentY, 0, width, AlignLeft, f.counterStyle, true,
	)
	drawnLines += 1
	currentY--

	if f.itemOffset > len(f.matched) {
		// the matched items have been updated while scrolling
		// adjust the offset so that the selected items are still in view
		f.itemOffset = len(f.matched) - 1
	}

	// Adjust offset to keep the current selection in view.
	availableSlots := height - drawnLines             // number of items that can be drawn
	maxItemIndex := f.itemOffset + availableSlots - 1 // max item inde

	if f.selectedIndex >= maxItemIndex {
		f.itemOffset = f.selectedIndex - availableSlots + 1
	} else if f.selectedIndex < f.itemOffset {
		f.itemOffset--
	}

	maxItemIndex = f.itemOffset + availableSlots - 1

	// Draw the list items.
	for index, matchItem := range f.matched {

		item := f.rawItems[matchItem.idx]
		if currentY < 0 {
			continue
		}

		if index < f.itemOffset || index > maxItemIndex {
			continue
		}

		truncate := func(text string, maxWidth int) string {
			if len(text) > maxWidth {
				return fmt.Sprintf("%s...", text[0:maxWidth-3])
			}
			return text
		}

		label := f.itemLabel
		labelStyle := f.itemLabelStyle
		itemStyle := f.itemStyle
		labelPadding := f.itemLabelPadding
		selected := false
		if index == f.selectedIndex && (!f.selectedFocusOnly || f.HasFocus()) {
			label = f.selectedItemLabel
			labelStyle = f.selectedItemLabelStyle
			itemStyle = f.selectedItemStyle
			labelPadding = f.selectedItemLabelPadding
			selected = true
		}

		currentX := x

		// print label
		printWithStyle(
			screen, label, x, currentY, 0, width, AlignLeft, labelStyle, false,
		)
		currentX += len(label)

		if selected {
			printWithStyle(
				screen, " ", currentX, currentY, 0, width, AlignLeft, labelStyle, false,
			)
		}

		// set a gap between label and item title
		currentX += labelPadding

		maxTitleWidth := width - currentX - 1

		// print title && don't increase currentX any further since we still need to draw the background
		itemTitle := truncate(item, maxTitleWidth)
		itemTitleLen := TaggedStringWidth(itemTitle)
		printWithStyle(
			screen,
			itemTitle,
			currentX,
			currentY, 0, maxTitleWidth, AlignLeft, itemStyle, true,
		)

		// print background
		textWidth := width
		if !f.highlightFullLine {
			if tw := TaggedStringWidth(item); tw < textWidth {
				textWidth = tw
			}
		}
		for bx := currentX; bx < minInt(currentX+itemTitleLen, maxX); bx++ {
			m, c, _, _ := screen.GetContent(bx, currentY)
			screen.SetContent(bx, currentY, m, c, itemStyle)
		}

		// draw match highlighting
		if f.filterText != "" {
			for _, m := range matchItem.matches {
				printWithStyle(
					screen,
					item[m[0]:m[1]],
					currentX+m[0],
					currentY,
					0,
					width, AlignLeft,
					f.highlightMatchStyle,
					f.highlightMatchMaintainBackgroundColor,
				)
			}

		}

		currentY--
	}
}

// drawInputField draws the input field.
func (f *Finder) drawInputField(screen tcell.Screen) {

	// Prepare
	x, currentY, width, height := f.GetInnerRect()
	currentY = currentY + height - 1
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}

	// Draw inputLabel.
	_, drawnWidth, _, _ := printWithStyle(screen, f.inputLabel, x, currentY, 0, rightLimit-x, AlignLeft, f.inputLabelStyle, false)
	x += drawnWidth
	x += f.inputLabelPadding

	// Draw input area.
	fieldWidth := width - TaggedStringWidth(f.inputLabel) - f.inputLabelPadding
	text := f.filterText
	inputStyle := f.fieldStyle
	showPlaceholder := text == ""
	_, inputBg, _ := inputStyle.Decompose()

	if inputBg != tcell.ColorDefault {
		for index := 0; index < fieldWidth; index++ {
			screen.SetContent(x+index, currentY, ' ', nil, inputStyle)
		}
	}

	// Text.
	var cursorScreenPos int
	if showPlaceholder {
		// Draw showPlaceholder text.
		printWithStyle(screen, Escape(f.placeholder), x, currentY, 0, fieldWidth, AlignLeft, f.placeholderStyle, true)
		f.offset = 0
	} else {
		// Draw entered text.
		if fieldWidth >= stringWidth(text) {
			// We have enough space for the full text.
			printWithStyle(screen, Escape(text), x, currentY, 0, fieldWidth, AlignLeft, inputStyle, true)
			f.offset = 0
			iterateString(text, func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				if textPos >= f.cursorPos {
					return true
				}
				cursorScreenPos += screenWidth
				return false
			})
		} else {
			// The text doesn't fit. Where is the cursor?
			if f.cursorPos < 0 {
				f.cursorPos = 0
			} else if f.cursorPos > len(text) {
				f.cursorPos = len(text)
			}
			// Shift the text so the cursor is inside the field.
			var shiftLeft int
			if f.offset > f.cursorPos {
				f.offset = f.cursorPos
			} else if subWidth := stringWidth(text[f.offset:f.cursorPos]); subWidth > fieldWidth-1 {
				shiftLeft = subWidth - fieldWidth + 1
			}
			currentOffset := f.offset
			iterateString(text, func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				if textPos >= currentOffset {
					if shiftLeft > 0 {
						f.offset = textPos + textWidth
						shiftLeft -= screenWidth
					} else {
						if textPos+textWidth > f.cursorPos {
							return true
						}
						cursorScreenPos += screenWidth
					}
				}
				return false
			})
			printWithStyle(screen, Escape(text[f.offset:]), x, currentY, 0, fieldWidth, AlignLeft, inputStyle, true)
		}
	}

	// Set cursor.
	if f.HasFocus() {
		screen.ShowCursor(x+cursorScreenPos, currentY)
	}
}

// InputHandler returns the handler for this primitive.
func (f *Finder) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return f.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if event.Key() == tcell.KeyEscape {
			if f.done != nil {
				f.done(-1)
			}
			return
		}

		f.handleInputList(event)
		f.handleInputTextField(event)
	})
}

// handleInputList handles the input events for the list (moving the selection up and down)
func (f *Finder) handleInputList(event *tcell.EventKey) {

	newSelectedIndex := f.selectedIndex
	switch key := event.Key(); key {
	case tcell.KeyDown:
		newSelectedIndex--
	case tcell.KeyUp:
		newSelectedIndex++
	}

	if newSelectedIndex < 0 {
		if f.wrapAround {
			newSelectedIndex = len(f.matched) - 1
		} else {
			newSelectedIndex = 0
			f.itemOffset = 0
		}
	} else if newSelectedIndex >= len(f.matched) {
		if f.wrapAround {
			newSelectedIndex = 0
			f.itemOffset = 0
		} else {
			newSelectedIndex = len(f.matched) - 1
		}
	}

	f.SetCurrentItem(newSelectedIndex)
}

// handleInputTextField handles the input events for the input field.
func (f *Finder) handleInputTextField(event *tcell.EventKey) {

	// Trigger changed events.
	currentText := f.filterText
	defer func() {
		if f.filterText != currentText {
			f.updateMatches()
		}
	}()

	// Movement functions.
	home := func() { f.cursorPos = 0 }
	end := func() { f.cursorPos = len(f.filterText) }
	moveLeft := func() {
		iterateStringReverse(f.filterText[:f.cursorPos], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
			f.cursorPos -= textWidth
			return true
		})
	}
	moveRight := func() {
		iterateString(f.filterText[f.cursorPos:], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
			f.cursorPos += textWidth
			return true
		})
	}
	moveWordLeft := func() {
		f.cursorPos = len(regexp.MustCompile(`\S+\s*$`).ReplaceAllString(f.filterText[:f.cursorPos], ""))
	}
	moveWordRight := func() {
		f.cursorPos = len(f.filterText) - len(regexp.MustCompile(`^\s*\S+\s*`).ReplaceAllString(f.filterText[f.cursorPos:], ""))
	}

	// Add character function. Returns whether or not the rune character is
	// accepted.
	add := func(r rune) bool {
		newText := f.filterText[:f.cursorPos] + string(r) + f.filterText[f.cursorPos:]
		f.filterText = newText
		f.cursorPos += len(string(r))
		return true
	}

	// Finish up.
	finish := func(key tcell.Key) {
		if f.done != nil {
			f.done(f.matched[f.selectedIndex].idx)
		}
	}

	switch key := event.Key(); key {
	case tcell.KeyRune: // Regular character.
		if event.Modifiers()&tcell.ModAlt > 0 {
			// We accept some Alt- key combinations.
			switch event.Rune() {
			case 'a': // Home.
				home()
			case 'e': // End.
				end()
			case 'b': // Move word left.
				moveWordLeft()
			case 'f': // Move word right.
				moveWordRight()
			default:
				if !add(event.Rune()) {
					return
				}
			}
		} else {
			// Other keys are simply accepted as regular characters.
			if !add(event.Rune()) {
				return
			}
		}
	case tcell.KeyCtrlU: // Delete all.
		f.filterText = ""
		f.cursorPos = 0
	case tcell.KeyCtrlK: // Delete until the end of the line.
		f.filterText = f.filterText[:f.cursorPos]
	case tcell.KeyCtrlW: // Delete last word.
		lastWord := regexp.MustCompile(`\S+\s*$`)
		newText := lastWord.ReplaceAllString(f.filterText[:f.cursorPos], "") + f.filterText[f.cursorPos:]
		f.cursorPos -= len(f.filterText) - len(newText)
		f.filterText = newText
	case tcell.KeyBackspace, tcell.KeyBackspace2: // Delete character before the cursor.
		iterateStringReverse(f.filterText[:f.cursorPos], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
			f.filterText = f.filterText[:textPos] + f.filterText[textPos+textWidth:]
			f.cursorPos -= textWidth
			return true
		})
		if f.offset >= f.cursorPos {
			f.offset = 0
		}
	case tcell.KeyDelete, tcell.KeyCtrlD: // Delete character after the cursor.
		iterateString(f.filterText[f.cursorPos:], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
			f.filterText = f.filterText[:f.cursorPos] + f.filterText[f.cursorPos+textWidth:]
			return true
		})
	case tcell.KeyLeft:
		if event.Modifiers()&tcell.ModAlt > 0 {
			moveWordLeft()
		} else {
			moveLeft()
		}
	case tcell.KeyCtrlB:
		moveLeft()
	case tcell.KeyRight:
		if event.Modifiers()&tcell.ModAlt > 0 {
			moveWordRight()
		} else {
			moveRight()
		}
	case tcell.KeyCtrlF:
		moveRight()
	case tcell.KeyHome, tcell.KeyCtrlA:
		home()
	case tcell.KeyEnd, tcell.KeyCtrlE:
		end()
	case tcell.KeyEnter:
		finish(key)
	}
}

func (f *Finder) inputOffset() int {
	labelPaddingMax := f.itemLabelPadding
	if f.selectedItemLabelPadding > labelPaddingMax {
		labelPaddingMax = f.selectedItemLabelPadding
	}
	labelWidthMax := len(f.itemLabel)
	if labelWidthMax < len(f.selectedItemLabel) {
		labelWidthMax = len(f.selectedItemLabel)
	}

	return labelWidthMax + labelPaddingMax
}
