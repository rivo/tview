package tview

import (
	"strings"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	Checkmark   = '✘'
	Uncheckmark = '❏'
)

var (
	lineStyle = tcell.StyleDefault.Foreground(Styles.PrimaryTextColor)
)

type Checklist struct {
	*Box
	hasFocus       bool
	deleteHasFocus bool
	contents       []string
	checked        []bool
	selected       int
	input          *InputField
	deleteButton   *Button
	UseChecks      bool
}

func NewChecklist(contents ...string) *Checklist {
	contents = append(contents, "")
	c := &Checklist{
		Box:          NewBox(),
		contents:     contents,
		checked:      make([]bool, len(contents)),
		selected:     len(contents) - 2,
		input:        NewInputField(),
		deleteButton: NewButton("X"),
		UseChecks:    true,
	}

	c.deleteButton.
		SetSelectedFunc(func() {
			c.removeCurrent()
		}).
		SetLabelColor(Styles.SecondaryTextColor).
		SetBackgroundColor(Styles.PrimitiveBackgroundColor)

	c.input.
		SetChangedFunc(func(changed string) {
			if c.selected >= len(c.contents) {
				return
			}

			c.contents[c.selected] = c.input.GetText()
			lastIndex := len(c.contents) - 1

			//If last line is no longer empty add a new empty line
			if c.selected == lastIndex && changed != "" {
				c.contents = append(c.contents, "")
				c.checked = append(c.checked, false)
			}
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyUp:
				c.selected--
				if c.selected < 0 {
					c.selected = len(c.contents) - 1
				}
			case tcell.KeyDown:
				c.selected++
				if c.selected >= len(c.contents) {
					c.selected = 0
				}
			case tcell.KeyCtrlD:
				c.removeCurrent()
			case tcell.KeyCtrlA:
				c.contents = append(
					c.contents[:c.selected],
					append([]string{""}, c.contents[c.selected:]...)...,
				)
				c.checked = append(c.checked, false)
			}
			return event
		})

	return c
}

func (c *Checklist) removeCurrent() {
	if c.selected == len(c.contents)-1 && c.contents[c.selected] == "" {
		return
	}
	c.contents = append(c.contents[:c.selected], c.contents[c.selected+1:]...)
	c.checked = append(c.checked[:c.selected], c.checked[c.selected+1:]...)
	if c.selected >= len(c.contents) {
		c.selected--
	}
}

func (c *Checklist) Draw(screen tcell.Screen) {
	c.Box.Draw(screen)
	x, y, w, h := c.GetInnerRect()

	var maxI, maxStringLen int
	for i, s := range c.contents {
		if i >= h {
			break
		}

		if l := runewidth.StringWidth(s); l > maxStringLen {
			maxStringLen = l
		}
		maxI = i
	}

	yy := maxI + 1
	var sb strings.Builder
	for i, s := range c.contents[0:yy] {
		sb.Reset()

		if c.UseChecks {
			check := Uncheckmark
			if i == len(c.contents)-1 {
				check = '❍'
			} else if c.checked[i] {
				check = Checkmark
			}

			sb.WriteString(runewidth.FillRight(string(check), 2))
		}

		padded := runewidth.FillRight(s, maxStringLen)
		sb.WriteString(padded)

		contentLine := sb.String()

		printWithStyle(
			screen, contentLine,
			x, y+i,
			w, AlignLeft,
			lineStyle,
		)
	}
	text := ""
	if c.selected < len(c.contents) {
		text = c.contents[c.selected]
	}
	inputOffset := 0
	if c.UseChecks {
		inputOffset += 2
	}
	c.input.SetText(text)
	c.input.SetRect(x+inputOffset, y+c.selected, w-4, 1)
	c.input.Draw(screen)

	if c.selected != len(c.contents)-1 {
		c.deleteButton.SetRect(x+w-1, y+c.selected, 1, 1)
		c.deleteButton.Draw(screen)
	}
}

// Focus is called by the application when the primitive receives focus.
func (c *Checklist) Focus(delegate func(p Primitive)) {
	if len(c.contents) == 0 {
		c.hasFocus = true
		return
	}
	c.hasFocus = false

	doneHandler := func(key tcell.Key) {
		lastIndex := len(c.contents) - 1
		onLastIndex := c.selected == lastIndex

		switch key {
		case tcell.KeyTab:
			if onLastIndex {
				c.deleteHasFocus = false
			} else {
				c.deleteHasFocus = !c.deleteHasFocus
			}

			if !c.deleteHasFocus {
				c.selected++
				if c.selected >= len(c.contents) {
					c.selected = 0
				}
			}
			c.Focus(delegate)
		case tcell.KeyBacktab:
			if onLastIndex {
				c.deleteHasFocus = false
			} else {
				c.deleteHasFocus = !c.deleteHasFocus
			}

			if !c.deleteHasFocus {
				c.selected--
				if c.selected < 0 {
					c.selected = lastIndex
				}
			}
			c.Focus(delegate)
		case tcell.KeyEnter:
			if c.UseChecks {
				c.checked[c.selected] = !c.checked[c.selected]
			}
		}
	}
	c.input.SetDoneFunc(doneHandler)
	c.deleteButton.SetBlurFunc(doneHandler)

	if c.deleteHasFocus {
		delegate(c.deleteButton)
	} else {
		delegate(c.input)
	}
}

// HasFocus returns whether or not this primitive has focus.
func (c *Checklist) HasFocus() bool {
	if c.hasFocus {
		return true
	}

	if c.input.HasFocus() || c.deleteButton.HasFocus() {
		return true
	}

	return false
}

// Contents return the non-empty strings
func (c *Checklist) Contents() []string {
	nonEmpty := make([]string, 0, len(c.contents))
	for _, s := range c.contents {
		if t := strings.TrimSpace(s); len(t) == 0 {
			continue
		}
		nonEmpty = append(nonEmpty, s)
	}
	return nonEmpty
}
