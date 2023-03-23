package tview

import (
	"math"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

const (
	AutocompletedNavigate = iota // The user navigated the autocomplete list (using the errow keys).
	AutocompletedTab             // The user selected an autocomplete entry using the tab key.
	AutocompletedEnter           // The user selected an autocomplete entry using the enter key.
	AutocompletedClick           // The user selected an autocomplete entry by clicking the mouse button on it.
)

// InputField is a one-line box (three lines if there is a title) where the
// user can enter text. Use [InputField.SetAcceptanceFunc] to accept or reject
// input, [InputField.SetChangedFunc] to listen for changes, and
// [InputField.SetMaskCharacter] to hide input from onlookers (e.g. for password
// input).
//
// The input field also has an optional autocomplete feature. It is initialized
// by the [InputField.SetAutocompleteFunc] function. For more control over the
// autocomplete drop-down's behavior, you can also set the
// [InputField.SetAutocompletedFunc].
//
// The following keys can be used for navigation and editing:
//
//   - Left arrow: Move left by one character.
//   - Right arrow: Move right by one character.
//   - Down arrow: Open the autocomplete drop-down.
//   - Tab, Enter: Select the current autocomplete entry.
//   - Home, Ctrl-A, Alt-a: Move to the beginning of the line.
//   - End, Ctrl-E, Alt-e: Move to the end of the line.
//   - Alt-left, Alt-b: Move left by one word.
//   - Alt-right, Alt-f: Move right by one word.
//   - Backspace: Delete the character before the cursor.
//   - Delete: Delete the character after the cursor.
//   - Ctrl-K: Delete from the cursor to the end of the line.
//   - Ctrl-W: Delete the last word before the cursor.
//   - Ctrl-U: Delete the entire line.
//
// See https://github.com/rivo/tview/wiki/InputField for an example.
type InputField struct {
	*Box

	// Whether or not this input field is disabled/read-only.
	disabled bool

	// The text that was entered.
	text string

	// The text to be displayed before the input area.
	label string

	// The text to be displayed in the input area when "text" is empty.
	placeholder string

	// The label style.
	labelStyle tcell.Style

	// The style of the input area with input text.
	fieldStyle tcell.Style

	// The style of the input area with placeholder text.
	placeholderStyle tcell.Style

	// The screen width of the label area. A value of 0 means use the width of
	// the label text.
	labelWidth int

	// The screen width of the input area. A value of 0 means extend as much as
	// possible.
	fieldWidth int

	// A character to mask entered text (useful for password fields). A value of 0
	// disables masking.
	maskCharacter rune

	// The cursor position as a byte index into the text string.
	cursorPos int

	// An optional autocomplete function which receives the current text of the
	// input field and returns a slice of strings to be displayed in a drop-down
	// selection.
	autocomplete func(text string) []string

	// The List object which shows the selectable autocomplete entries. If not
	// nil, the list's main texts represent the current autocomplete entries.
	autocompleteList      *List
	autocompleteListMutex sync.Mutex

	// The styles of the autocomplete entries.
	autocompleteStyles struct {
		main       tcell.Style
		selected   tcell.Style
		background tcell.Color
	}

	// An optional function which is called when the user selects an
	// autocomplete entry. The text and index of the selected entry (within the
	// list) is provided, as well as the user action causing the selection (one
	// of the "Autocompleted" values). The function should return true if the
	// autocomplete list should be closed. If nil, the input field will be
	// updated automatically when the user navigates the autocomplete list.
	autocompleted func(text string, index int, source int) bool

	// An optional function which may reject the last character that was entered.
	accept func(text string, ch rune) bool

	// An optional function which is called when the input has changed.
	changed func(text string)

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, enter, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)

	fieldX int // The x-coordinate of the input field as determined during the last call to Draw().
	offset int // The number of bytes of the text string skipped ahead while drawing.
}

// NewInputField returns a new input field.
func NewInputField() *InputField {
	i := &InputField{
		Box:              NewBox(),
		labelStyle:       tcell.StyleDefault.Foreground(Styles.SecondaryTextColor),
		fieldStyle:       tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
		placeholderStyle: tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.ContrastSecondaryTextColor),
	}
	i.autocompleteStyles.main = tcell.StyleDefault.Foreground(Styles.PrimitiveBackgroundColor)
	i.autocompleteStyles.selected = tcell.StyleDefault.Background(Styles.PrimaryTextColor).Foreground(Styles.PrimitiveBackgroundColor)
	i.autocompleteStyles.background = Styles.MoreContrastBackgroundColor
	return i
}

// SetText sets the current text of the input field.
func (i *InputField) SetText(text string) *InputField {
	i.text = text
	i.cursorPos = len(text)
	if i.changed != nil {
		i.changed(text)
	}
	return i
}

// GetText returns the current text of the input field.
func (i *InputField) GetText() string {
	return i.text
}

// SetLabel sets the text to be displayed before the input area.
func (i *InputField) SetLabel(label string) *InputField {
	i.label = label
	return i
}

// GetLabel returns the text to be displayed before the input area.
func (i *InputField) GetLabel() string {
	return i.label
}

// SetLabelWidth sets the screen width of the label. A value of 0 will cause the
// primitive to use the width of the label string.
func (i *InputField) SetLabelWidth(width int) *InputField {
	i.labelWidth = width
	return i
}

// SetPlaceholder sets the text to be displayed when the input text is empty.
func (i *InputField) SetPlaceholder(text string) *InputField {
	i.placeholder = text
	return i
}

// SetLabelColor sets the text color of the label.
func (i *InputField) SetLabelColor(color tcell.Color) *InputField {
	i.labelStyle = i.labelStyle.Foreground(color)
	return i
}

// SetLabelStyle sets the style of the label.
func (i *InputField) SetLabelStyle(style tcell.Style) *InputField {
	i.labelStyle = style
	return i
}

// GetLabelStyle returns the style of the label.
func (i *InputField) GetLabelStyle() tcell.Style {
	return i.labelStyle
}

// SetFieldBackgroundColor sets the background color of the input area.
func (i *InputField) SetFieldBackgroundColor(color tcell.Color) *InputField {
	i.fieldStyle = i.fieldStyle.Background(color)
	return i
}

// SetFieldTextColor sets the text color of the input area.
func (i *InputField) SetFieldTextColor(color tcell.Color) *InputField {
	i.fieldStyle = i.fieldStyle.Foreground(color)
	return i
}

// SetFieldStyle sets the style of the input area (when no placeholder is
// shown).
func (i *InputField) SetFieldStyle(style tcell.Style) *InputField {
	i.fieldStyle = style
	return i
}

// GetFieldStyle returns the style of the input area (when no placeholder is
// shown).
func (i *InputField) GetFieldStyle() tcell.Style {
	return i.fieldStyle
}

// SetPlaceholderTextColor sets the text color of placeholder text.
func (i *InputField) SetPlaceholderTextColor(color tcell.Color) *InputField {
	i.placeholderStyle = i.placeholderStyle.Foreground(color)
	return i
}

// SetPlaceholderStyle sets the style of the input area (when a placeholder is
// shown).
func (i *InputField) SetPlaceholderStyle(style tcell.Style) *InputField {
	i.placeholderStyle = style
	return i
}

// GetPlaceholderStyle returns the style of the input area (when a placeholder
// is shown).
func (i *InputField) GetPlaceholderStyle() tcell.Style {
	return i.placeholderStyle
}

// SetAutocompleteStyles sets the colors and style of the autocomplete entries.
// For details, see List.SetMainTextStyle(), List.SetSelectedStyle(), and
// Box.SetBackgroundColor().
func (i *InputField) SetAutocompleteStyles(background tcell.Color, main, selected tcell.Style) *InputField {
	i.autocompleteStyles.background = background
	i.autocompleteStyles.main = main
	i.autocompleteStyles.selected = selected
	return i
}

// SetFormAttributes sets attributes shared by all form items.
func (i *InputField) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	i.labelWidth = labelWidth
	i.backgroundColor = bgColor
	i.SetLabelColor(labelColor).
		SetFieldTextColor(fieldTextColor).
		SetFieldBackgroundColor(fieldBgColor)
	return i
}

// SetFieldWidth sets the screen width of the input area. A value of 0 means
// extend as much as possible.
func (i *InputField) SetFieldWidth(width int) *InputField {
	i.fieldWidth = width
	return i
}

// GetFieldWidth returns this primitive's field width.
func (i *InputField) GetFieldWidth() int {
	return i.fieldWidth
}

// GetFieldHeight returns this primitive's field height.
func (i *InputField) GetFieldHeight() int {
	return 1
}

// SetDisabled sets whether or not the item is disabled / read-only.
func (i *InputField) SetDisabled(disabled bool) FormItem {
	i.disabled = disabled
	if i.finished != nil {
		i.finished(-1)
	}
	return i
}

// SetMaskCharacter sets a character that masks user input on a screen. A value
// of 0 disables masking.
func (i *InputField) SetMaskCharacter(mask rune) *InputField {
	i.maskCharacter = mask
	return i
}

// SetAutocompleteFunc sets an autocomplete callback function which may return
// strings to be selected from a drop-down based on the current text of the
// input field. The drop-down appears only if len(entries) > 0. The callback is
// invoked in this function and whenever the current text changes or when
// Autocomplete() is called. Entries are cleared when the user selects an entry
// or presses Escape.
func (i *InputField) SetAutocompleteFunc(callback func(currentText string) (entries []string)) *InputField {
	i.autocomplete = callback
	i.Autocomplete()
	return i
}

// SetAutocompletedFunc sets a callback function which is invoked when the user
// selects an entry from the autocomplete drop-down list. The function is passed
// the text of the selected entry (stripped of any color tags), the index of the
// entry, and the user action that caused the selection, e.g.
// [AutocompletedNavigate]. It returns true if the autocomplete drop-down should
// be closed after the callback returns or false if it should remain open, in
// which case [InputField.Autocomplete] is called to update the drop-down's
// contents.
//
// If no such callback is set (or nil is provided), the input field will be
// updated with the selection any time the user navigates the autocomplete
// drop-down list. So this function essentially gives you more control over the
// autocomplete functionality.
func (i *InputField) SetAutocompletedFunc(autocompleted func(text string, index int, source int) bool) *InputField {
	i.autocompleted = autocompleted
	return i
}

// Autocomplete invokes the autocomplete callback (if there is one). If the
// length of the returned autocomplete entries slice is greater than 0, the
// input field will present the user with a corresponding drop-down list the
// next time the input field is drawn.
//
// It is safe to call this function from any goroutine. Note that the input
// field is not redrawn automatically unless called from the main goroutine
// (e.g. in response to events).
func (i *InputField) Autocomplete() *InputField {
	i.autocompleteListMutex.Lock()
	defer i.autocompleteListMutex.Unlock()
	if i.autocomplete == nil {
		return i
	}

	// Do we have any autocomplete entries?
	entries := i.autocomplete(i.text)
	if len(entries) == 0 {
		// No entries, no list.
		i.autocompleteList = nil
		return i
	}

	// Make a list if we have none.
	if i.autocompleteList == nil {
		i.autocompleteList = NewList()
		i.autocompleteList.ShowSecondaryText(false).
			SetMainTextStyle(i.autocompleteStyles.main).
			SetSelectedStyle(i.autocompleteStyles.selected).
			SetHighlightFullLine(true).
			SetBackgroundColor(i.autocompleteStyles.background)
	}

	// Fill it with the entries.
	currentEntry := -1
	suffixLength := 9999 // I'm just waiting for the day somebody opens an issue with this number being too small.
	i.autocompleteList.Clear()
	for index, entry := range entries {
		i.autocompleteList.AddItem(entry, "", 0, nil)
		if strings.HasPrefix(entry, i.text) && len(entry)-len(i.text) < suffixLength {
			currentEntry = index
			suffixLength = len(i.text) - len(entry)
		}
	}

	// Set the selection if we have one.
	if currentEntry >= 0 {
		i.autocompleteList.SetCurrentItem(currentEntry)
	}

	return i
}

// SetAcceptanceFunc sets a handler which may reject the last character that was
// entered (by returning false).
//
// This package defines a number of variables prefixed with InputField which may
// be used for common input (e.g. numbers, maximum text length).
func (i *InputField) SetAcceptanceFunc(handler func(textToCheck string, lastChar rune) bool) *InputField {
	i.accept = handler
	return i
}

// SetChangedFunc sets a handler which is called whenever the text of the input
// field has changed. It receives the current text (after the change).
func (i *InputField) SetChangedFunc(handler func(text string)) *InputField {
	i.changed = handler
	return i
}

// SetDoneFunc sets a handler which is called when the user is done entering
// text. The callback function is provided with the key that was pressed, which
// is one of the following:
//
//   - KeyEnter: Done entering text.
//   - KeyEscape: Abort text input.
//   - KeyTab: Move to the next field.
//   - KeyBacktab: Move to the previous field.
func (i *InputField) SetDoneFunc(handler func(key tcell.Key)) *InputField {
	i.done = handler
	return i
}

// SetFinishedFunc sets a callback invoked when the user leaves this form item.
func (i *InputField) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	i.finished = handler
	return i
}

// Focus is called when this primitive receives focus.
func (i *InputField) Focus(delegate func(p Primitive)) {
	// If we're part of a form and this item is disabled, there's nothing the
	// user can do here so we're finished.
	if i.finished != nil && i.disabled {
		i.finished(-1)
		return
	}

	i.Box.Focus(delegate)
}

// Blur is called when this primitive loses focus.
func (i *InputField) Blur() {
	i.Box.Blur()
	i.autocompleteList = nil // Hide the autocomplete drop-down.
}

// Draw draws this primitive onto the screen.
func (i *InputField) Draw(screen tcell.Screen) {
	i.Box.DrawForSubclass(screen, i)

	// Prepare
	x, y, width, height := i.GetInnerRect()
	rightLimit := x + width
	if height < 1 || rightLimit <= x {
		return
	}

	// Draw label.
	_, labelBg, _ := i.labelStyle.Decompose()
	if i.labelWidth > 0 {
		labelWidth := i.labelWidth
		if labelWidth > width {
			labelWidth = width
		}
		printWithStyle(screen, i.label, x, y, 0, labelWidth, AlignLeft, i.labelStyle, labelBg == tcell.ColorDefault)
		x += labelWidth
	} else {
		_, drawnWidth, _, _ := printWithStyle(screen, i.label, x, y, 0, width, AlignLeft, i.labelStyle, labelBg == tcell.ColorDefault)
		x += drawnWidth
	}

	// Draw input area.
	i.fieldX = x
	fieldWidth := i.fieldWidth
	text := i.text
	inputStyle := i.fieldStyle
	placeholder := text == "" && i.placeholder != ""
	if placeholder {
		inputStyle = i.placeholderStyle
	}
	_, inputBg, _ := inputStyle.Decompose()
	if fieldWidth == 0 {
		fieldWidth = math.MaxInt32
	}
	if rightLimit-x < fieldWidth {
		fieldWidth = rightLimit - x
	}
	if i.disabled {
		inputStyle = inputStyle.Background(i.backgroundColor)
	}
	if inputBg != tcell.ColorDefault {
		for index := 0; index < fieldWidth; index++ {
			screen.SetContent(x+index, y, ' ', nil, inputStyle)
		}
	}

	// Text.
	var cursorScreenPos int
	if placeholder {
		// Draw placeholder text.
		printWithStyle(screen, Escape(i.placeholder), x, y, 0, fieldWidth, AlignLeft, i.placeholderStyle, true)
		i.offset = 0
	} else {
		// Draw entered text.
		if i.maskCharacter > 0 {
			text = strings.Repeat(string(i.maskCharacter), utf8.RuneCountInString(i.text))
		}
		if fieldWidth >= uniseg.StringWidth(text) {
			// We have enough space for the full text.
			printWithStyle(screen, Escape(text), x, y, 0, fieldWidth, AlignLeft, i.fieldStyle, true)
			i.offset = 0
			iterateString(text, func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth, boundaries int) bool {
				if textPos >= i.cursorPos {
					return true
				}
				cursorScreenPos += screenWidth
				return false
			})
		} else {
			// The text doesn't fit. Where is the cursor?
			if i.cursorPos < 0 {
				i.cursorPos = 0
			} else if i.cursorPos > len(text) {
				i.cursorPos = len(text)
			}
			// Shift the text so the cursor is inside the field.
			var shiftLeft int
			if i.offset > i.cursorPos {
				i.offset = i.cursorPos
			} else if subWidth := uniseg.StringWidth(text[i.offset:i.cursorPos]); subWidth > fieldWidth-1 {
				shiftLeft = subWidth - fieldWidth + 1
			}
			currentOffset := i.offset
			iterateString(text, func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth, boundaries int) bool {
				if textPos >= currentOffset {
					if shiftLeft > 0 {
						i.offset = textPos + textWidth
						shiftLeft -= screenWidth
					} else {
						if textPos+textWidth > i.cursorPos {
							return true
						}
						cursorScreenPos += screenWidth
					}
				}
				return false
			})
			printWithStyle(screen, Escape(text[i.offset:]), x, y, 0, fieldWidth, AlignLeft, i.fieldStyle, true)
		}
	}

	// Draw autocomplete list.
	i.autocompleteListMutex.Lock()
	defer i.autocompleteListMutex.Unlock()
	if i.autocompleteList != nil {
		// How much space do we need?
		lheight := i.autocompleteList.GetItemCount()
		lwidth := 0
		for index := 0; index < lheight; index++ {
			entry, _ := i.autocompleteList.GetItemText(index)
			width := TaggedStringWidth(entry)
			if width > lwidth {
				lwidth = width
			}
		}

		// We prefer to drop down but if there is no space, maybe drop up?
		lx := x
		ly := y + 1
		_, sheight := screen.Size()
		if ly+lheight >= sheight && ly-2 > lheight-ly {
			ly = y - lheight
			if ly < 0 {
				ly = 0
			}
		}
		if ly+lheight >= sheight {
			lheight = sheight - ly
		}
		i.autocompleteList.SetRect(lx, ly, lwidth, lheight)
		i.autocompleteList.Draw(screen)
	}

	// Set cursor.
	if i.HasFocus() {
		screen.ShowCursor(x+cursorScreenPos, y)
	}
}

// InputHandler returns the handler for this primitive.
func (i *InputField) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return i.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if i.disabled {
			return
		}

		// Trigger changed events.
		currentText := i.text
		defer func() {
			if i.text != currentText {
				i.Autocomplete()
				if i.changed != nil {
					i.changed(i.text)
				}
			}
		}()

		// Movement functions.
		home := func() { i.cursorPos = 0 }
		end := func() { i.cursorPos = len(i.text) }
		moveLeft := func() {
			iterateStringReverse(i.text[:i.cursorPos], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				i.cursorPos -= textWidth
				return true
			})
		}
		moveRight := func() {
			iterateString(i.text[i.cursorPos:], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth, boundaries int) bool {
				i.cursorPos += textWidth
				return true
			})
		}
		moveWordLeft := func() {
			i.cursorPos = len(regexp.MustCompile(`\S+\s*$`).ReplaceAllString(i.text[:i.cursorPos], ""))
		}
		moveWordRight := func() {
			i.cursorPos = len(i.text) - len(regexp.MustCompile(`^\s*\S+\s*`).ReplaceAllString(i.text[i.cursorPos:], ""))
		}

		// Add character function. Returns whether or not the rune character is
		// accepted.
		add := func(r rune) bool {
			newText := i.text[:i.cursorPos] + string(r) + i.text[i.cursorPos:]
			if i.accept != nil && !i.accept(newText, r) {
				return false
			}
			i.text = newText
			i.cursorPos += len(string(r))
			return true
		}

		// Finish up.
		finish := func(key tcell.Key) {
			if i.done != nil {
				i.done(key)
			}
			if i.finished != nil {
				i.finished(key)
			}
		}

		// If we have an autocomplete list, there are certain keys we will
		// forward to it.
		i.autocompleteListMutex.Lock()
		defer i.autocompleteListMutex.Unlock()
		if i.autocompleteList != nil {
			i.autocompleteList.SetChangedFunc(nil)
			switch key := event.Key(); key {
			case tcell.KeyEscape: // Close the list.
				i.autocompleteList = nil
				return
			case tcell.KeyEnter, tcell.KeyTab: // Intentional selection.
				if i.autocompleted != nil {
					index := i.autocompleteList.GetCurrentItem()
					text, _ := i.autocompleteList.GetItemText(index)
					source := AutocompletedEnter
					if key == tcell.KeyTab {
						source = AutocompletedTab
					}
					if i.autocompleted(stripTags(text), index, source) {
						i.autocompleteList = nil
						currentText = i.GetText()
					}
				} else {
					i.autocompleteList = nil
				}
				return
			case tcell.KeyDown, tcell.KeyUp, tcell.KeyPgDn, tcell.KeyPgUp:
				i.autocompleteList.SetChangedFunc(func(index int, text, secondaryText string, shortcut rune) {
					text = stripTags(text)
					if i.autocompleted != nil {
						if i.autocompleted(text, index, AutocompletedNavigate) {
							i.autocompleteList = nil
							currentText = i.GetText()
						}
					} else {
						i.SetText(text)
						currentText = stripTags(text) // We want to keep the autocomplete list open and unchanged.
					}
				})
				i.autocompleteList.InputHandler()(event, setFocus)
				return
			}
		}

		// Process key event for the input field.
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
			i.text = ""
			i.cursorPos = 0
		case tcell.KeyCtrlK: // Delete until the end of the line.
			i.text = i.text[:i.cursorPos]
		case tcell.KeyCtrlW: // Delete last word.
			lastWord := regexp.MustCompile(`\S+\s*$`)
			newText := lastWord.ReplaceAllString(i.text[:i.cursorPos], "") + i.text[i.cursorPos:]
			i.cursorPos -= len(i.text) - len(newText)
			i.text = newText
		case tcell.KeyBackspace, tcell.KeyBackspace2: // Delete character before the cursor.
			iterateStringReverse(i.text[:i.cursorPos], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth int) bool {
				i.text = i.text[:textPos] + i.text[textPos+textWidth:]
				i.cursorPos -= textWidth
				return true
			})
			if i.offset >= i.cursorPos {
				i.offset = 0
			}
		case tcell.KeyDelete, tcell.KeyCtrlD: // Delete character after the cursor.
			iterateString(i.text[i.cursorPos:], func(main rune, comb []rune, textPos, textWidth, screenPos, screenWidth, boundaries int) bool {
				i.text = i.text[:i.cursorPos] + i.text[i.cursorPos+textWidth:]
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
		case tcell.KeyDown:
			i.autocompleteListMutex.Unlock() // We're still holding a lock.
			i.Autocomplete()
			i.autocompleteListMutex.Lock()
		case tcell.KeyEnter, tcell.KeyEscape, tcell.KeyTab, tcell.KeyBacktab:
			finish(key)
		}
	})
}

// MouseHandler returns the mouse handler for this primitive.
func (i *InputField) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return i.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		if i.disabled {
			return false, nil
		}

		currentText := i.GetText()
		defer func() {
			if i.GetText() != currentText {
				i.Autocomplete()
				if i.changed != nil {
					i.changed(i.text)
				}
			}
		}()

		// If we have an autocomplete list, forward the mouse event to it.
		i.autocompleteListMutex.Lock()
		defer i.autocompleteListMutex.Unlock()
		if i.autocompleteList != nil {
			i.autocompleteList.SetChangedFunc(func(index int, text, secondaryText string, shortcut rune) {
				text = stripTags(text)
				if i.autocompleted != nil {
					if i.autocompleted(text, index, AutocompletedClick) {
						i.autocompleteList = nil
						currentText = i.GetText()
					}
					return
				}
				i.SetText(text)
				i.autocompleteList = nil
			})
			if consumed, _ = i.autocompleteList.MouseHandler()(action, event, setFocus); consumed {
				setFocus(i)
				return
			}
		}

		// Is mouse event within the input field?
		x, y := event.Position()
		_, rectY, _, _ := i.GetInnerRect()
		if !i.InRect(x, y) {
			return false, nil
		}

		// Process mouse event.
		if y == rectY {
			if action == MouseLeftDown {
				setFocus(i)
				consumed = true
			} else if action == MouseLeftClick {
				// Determine where to place the cursor.
				if x >= i.fieldX {
					if !iterateString(i.text[i.offset:], func(main rune, comb []rune, textPos int, textWidth int, screenPos int, screenWidth, boundaries int) bool {
						if x-i.fieldX < screenPos+screenWidth {
							i.cursorPos = textPos + i.offset
							return true
						}
						return false
					}) {
						i.cursorPos = len(i.text)
					}
				}
				consumed = true
			}
		}

		return
	})
}
