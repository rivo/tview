package tview

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// newDropDownTestScreen returns a screen whose lifecycle belongs to the test.
func newDropDownTestScreen(t *testing.T) tcell.SimulationScreen {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)
	screen.SetSize(40, 10)
	return screen
}

// dropDownTestRow reads a row, skipping wide characters' trailing cells.
func dropDownTestRow(screen tcell.Screen, x, y, width int) string {
	var text strings.Builder
	for column := x; column < x+width; column++ {
		main, combining, _, cellWidth := screen.GetContent(column, y)
		text.WriteRune(main)
		for _, r := range combining {
			text.WriteRune(r)
		}
		if cellWidth > 1 {
			column += cellWidth - 1
		}
	}
	return text.String()
}

// TestDropDownArrow checks field sizing, clipping, and styled Unicode text.
func TestDropDownArrow(t *testing.T) {
	for _, test := range []struct {
		name       string
		text       string
		fieldWidth int
		rectWidth  int
		configure  func(*DropDown)
		wantWidth  int
		want       string
	}{
		{name: "automatic", text: "alpha", rectWidth: 7, wantWidth: 7, want: "alpha ▾"},
		{name: "explicit", text: "alpha", fieldWidth: 5, rectWidth: 5, wantWidth: 5, want: "alp ▾"},
		{name: "clipped", text: "alpha", rectWidth: 3, wantWidth: 7, want: "a ▾"},
		{name: "one cell", text: "alpha", fieldWidth: 1, rectWidth: 1, wantWidth: 1, want: "▾"},
		{name: "two cells", text: "alpha", fieldWidth: 2, rectWidth: 2, wantWidth: 2, want: " ▾"},
		{name: "no space", text: "alpha", wantWidth: 7},
		{name: "empty", rectWidth: 2, configure: func(d *DropDown) {
			d.SetOptions(nil, nil)
		}, wantWidth: 2, want: " ▾"},
		{name: "placeholder", rectWidth: 10, configure: func(d *DropDown) {
			d.SetCurrentOption(-1).SetTextOptions("", "", "", "", "Pick one")
		}, wantWidth: 10, want: "Pick one ▾"},
		{name: "decorated", text: "abc", rectWidth: 7, configure: func(d *DropDown) {
			d.SetTextOptions("(", ")", "<", ">", "")
		}, wantWidth: 7, want: "<abc> ▾"},
		{name: "styled", text: "[red]界[-]", rectWidth: 4, wantWidth: 4, want: "界 ▾"},
		{name: "literal tags", text: "[red]", rectWidth: 7, configure: func(d *DropDown) {
			d.SetUseStyleTags(false)
		}, wantWidth: 7, want: "[red] ▾"},
		{name: "literal placeholder", rectWidth: 7, configure: func(d *DropDown) {
			d.SetUseStyleTags(false).SetCurrentOption(-1).SetTextOptions("", "", "", "", "[red]")
		}, wantWidth: 7, want: "[red] ▾"},
		{name: "wide clipped", text: "界", fieldWidth: 3, rectWidth: 3, wantWidth: 3, want: "  ▾"},
		{name: "combining", text: "e\u0301x", fieldWidth: 3, rectWidth: 3, wantWidth: 3, want: "e\u0301 ▾"},
		{name: "emoji", text: "👩‍💻!", fieldWidth: 4, rectWidth: 4, wantWidth: 4, want: "👩‍💻 ▾"},
		{name: "label only", text: "alpha", rectWidth: 5, configure: func(d *DropDown) {
			d.SetLabel("Label")
		}, wantWidth: 7, want: "Label"},
	} {
		t.Run(test.name, func(t *testing.T) {
			screen := newDropDownTestScreen(t)
			d := NewDropDown().SetOptions([]string{test.text}, nil).
				SetCurrentOption(0).SetFieldWidth(test.fieldWidth).ShowArrow(true)
			if test.configure != nil {
				test.configure(d)
			}
			d.SetRect(1, 1, test.rectWidth, 1)
			screen.SetContent(0, 1, '#', nil, tcell.StyleDefault)
			screen.SetContent(test.rectWidth+1, 1, '#', nil, tcell.StyleDefault)
			d.Draw(screen)
			if got := d.GetFieldWidth(); got != test.wantWidth {
				t.Errorf("field width = %d, want %d", got, test.wantWidth)
			}
			if got := dropDownTestRow(screen, 1, 1, test.rectWidth); got != test.want {
				t.Errorf("field = %q, want %q", got, test.want)
			}
			if strings.HasSuffix(test.want, "▾") {
				if arrow, _, _, width := screen.GetContent(test.rectWidth, 1); arrow != '▾' || width != 1 {
					t.Errorf("arrow cell = %q with width %d, want a one-cell arrow", arrow, width)
				}
			}
			for _, column := range []int{0, test.rectWidth + 1} {
				if main, _, _, _ := screen.GetContent(column, 1); main != '#' {
					t.Errorf("cell outside the field at %d was overwritten: %q", column, main)
				}
			}
		})
	}
}

// TestDropDownArrowDefault checks that enabling the arrow is reversible.
func TestDropDownArrowDefault(t *testing.T) {
	screen := newDropDownTestScreen(t)
	d := NewDropDown().SetOptions([]string{"alpha"}, nil).SetCurrentOption(0)
	d.SetRect(0, 0, 7, 1)
	for _, test := range []struct {
		name      string
		configure func()
		wantWidth int
		want      string
	}{
		{name: "default", wantWidth: 5, want: "alpha  "},
		{name: "enabled", configure: func() { d.ShowArrow(true) }, wantWidth: 7, want: "alpha ▾"},
		{name: "disabled again", configure: func() { d.ShowArrow(false) }, wantWidth: 5, want: "alpha  "},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.configure != nil {
				test.configure()
			}
			d.Draw(screen)
			if got := d.GetFieldWidth(); got != test.wantWidth {
				t.Errorf("field width = %d, want %d", got, test.wantWidth)
			}
			if got := dropDownTestRow(screen, 0, 0, 7); got != test.want {
				t.Errorf("field = %q, want %q", got, test.want)
			}
		})
	}
}

// TestDropDownArrowForm checks the automatic width used by horizontal forms.
func TestDropDownArrowForm(t *testing.T) {
	screen := newDropDownTestScreen(t)
	d := NewDropDown().SetLabel("Choice:").SetOptions([]string{"界"}, nil).
		SetCurrentOption(0).SetTextOptions("", "", "<", ">", "").ShowArrow(true)
	form := NewForm().SetHorizontal(true).AddFormItem(d)
	form.SetRect(0, 0, 40, 5)
	form.Draw(screen)
	x, y, width, _ := d.GetInnerRect()
	if got, want := dropDownTestRow(screen, x, y, width), "Choice: <界> ▾"; got != want {
		t.Errorf("form field = %q, want %q", got, want)
	}
}

// TestDropDownArrowInteraction checks keyboard input, mouse input, and styles.
func TestDropDownArrowInteraction(t *testing.T) {
	screen := newDropDownTestScreen(t)
	fieldStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)
	focusedStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	disabledStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)
	selected := ""
	d := NewDropDown().SetOptions([]string{"alpha", "界面"}, func(text string, index int) {
		selected = text
	}).SetCurrentOption(0).SetFieldWidth(5).ShowArrow(true).
		SetFieldStyle(fieldStyle).SetFocusedStyle(focusedStyle).SetDisabledStyle(disabledStyle)
	d.SetRect(1, 1, 5, 1)
	d.Draw(screen)
	if _, _, style, _ := screen.GetContent(5, 1); style != fieldStyle {
		t.Errorf("unfocused arrow style = %v, want %v", style, fieldStyle)
	}
	app := NewApplication().SetRoot(d, false)
	setFocus := func(p Primitive) { app.SetFocus(p) }
	d.Draw(screen)
	if _, _, style, _ := screen.GetContent(5, 1); style != focusedStyle {
		t.Errorf("focused arrow style = %v, want %v", style, focusedStyle)
	}
	d.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), setFocus)
	d.InputHandler()(tcell.NewEventKey(tcell.KeyRune, '界', tcell.ModNone), setFocus)
	d.Draw(screen)
	if got, want := dropDownTestRow(screen, 1, 1, 5), "界  ▾"; got != want {
		t.Errorf("open field with typed prefix = %q, want %q", got, want)
	}
	if _, _, style, _ := screen.GetContent(5, 1); style != fieldStyle {
		t.Errorf("open arrow style = %v, want %v", style, fieldStyle)
	}
	d.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), setFocus)
	if index, text := d.GetCurrentOption(); index != 1 || text != "界面" || selected != text || d.IsOpen() {
		t.Fatalf("selection after keyboard input: index=%d, text=%q, callback=%q, open=%v", index, text, selected, d.IsOpen())
	}
	d.MouseHandler()(MouseLeftDown, tcell.NewEventMouse(5, 1, tcell.Button1, tcell.ModNone), setFocus)
	if !d.IsOpen() {
		t.Fatal("clicking the arrow did not open the list")
	}
	d.InputHandler()(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone), setFocus)
	d.SetDisabled(true)
	d.Draw(screen)
	if _, _, style, _ := screen.GetContent(5, 1); style != disabledStyle {
		t.Errorf("disabled arrow style = %v, want %v", style, disabledStyle)
	}
	d.MouseHandler()(MouseLeftDown, tcell.NewEventMouse(5, 1, tcell.Button1, tcell.ModNone), setFocus)
	if d.IsOpen() {
		t.Fatal("clicking a disabled arrow opened the list")
	}
}
