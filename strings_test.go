package tview

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestStyleTagAttributeReset checks that reset tags restore the complete initial
// attributes without changing colors, regions, or hyperlinks.
func TestStyleTagAttributeReset(t *testing.T) {
	styles := []struct {
		name  string
		style tcell.Style
	}{
		{"default", tcell.StyleDefault},
		{"bold", tcell.StyleDefault.Bold(true)},
		{"solid", tcell.StyleDefault.Underline(true)},
		{"double", tcell.StyleDefault.Underline(tcell.UnderlineStyleDouble, tcell.ColorYellow)},
		{"curly", tcell.StyleDefault.Underline(tcell.UnderlineStyleCurly, tcell.ColorRed)},
		{"attribute only", tcell.StyleDefault.Attributes(tcell.AttrUnderline)},
	}
	for _, test := range styles {
		for _, reset := range []string{"[::-]", "[-:-:-]"} {
			t.Run(test.name+reset, func(t *testing.T) {
				initial := test.style.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
				state := &stepState{unisegState: -1, style: initial}
				cluster, rest, state := step("[red:blue:bu]A"+reset+"B[::U]C"+reset+"D", state, stepOptionsStyle)
				if cluster != "A" || state.Style() == initial {
					t.Fatalf("styled control: cluster=%q style=%#v", cluster, state.Style())
				}
				for _, want := range []string{"B", "C", "D"} {
					cluster, rest, state = step(rest, state, stepOptionsStyle)
					if cluster != want {
						t.Fatalf("cluster=%q, want %q", cluster, want)
					}
					if want == "C" {
						continue // The explicit flag is checked by its following reset.
					}
					wantStyle := initial
					if reset == "[::-]" {
						wantStyle = wantStyle.Foreground(tcell.ColorRed).Background(tcell.ColorBlue)
					}
					if state.Style() != wantStyle {
						t.Errorf("%s style=%#v, want %#v", want, state.Style(), wantStyle)
					}
				}
			})
		}
	}
	t.Run("hyperlink and region", func(t *testing.T) {
		state := &stepState{unisegState: -1, style: tcell.StyleDefault}
		_, rest, state := step(`[red:blue:bu:https://example.com]["art"]A[::-]B`, state, stepOptionsStyle|stepOptionsRegion)
		before := state.Style()
		cluster, _, state := step(rest, state, stepOptionsStyle|stepOptionsRegion)
		want := before.Underline(false).Attributes(0)
		if cluster != "B" || state.Style() != want || state.region != "art" {
			t.Errorf("reset: cluster=%q style=%#v region=%q, want B %#v art", cluster, state.Style(), state.region, want)
		}
	})
}

// TestTextViewAttributeReset checks the rendered cells across Unicode clusters,
// line breaks, and repeated draws using both plain and custom initial styles.
func TestTextViewAttributeReset(t *testing.T) {
	for _, initial := range []tcell.Style{tcell.StyleDefault, tcell.StyleDefault.Underline(tcell.UnderlineStyleCurly, tcell.ColorYellow)} {
		for _, reset := range []string{"[::-]", "[-:-:-]"} {
			screen := tcell.NewSimulationScreen("UTF-8")
			if err := screen.Init(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(screen.Fini)
			screen.SetSize(4, 2)
			initial = initial.Foreground(tcell.ColorGreen).Background(tcell.ColorBlack)
			view := NewTextView().SetDynamicColors(true).SetTextStyle(initial).
				SetText("[::bu]⣿" + reset + "界e\u0301\n▒")
			view.SetRect(0, 0, 4, 2)
			for draw := 0; draw < 2; draw++ {
				view.Draw(screen)
				ch, _, style, _ := screen.GetContent(0, 0)
				_, _, attrs := style.Decompose()
				if ch != '⣿' || attrs&tcell.AttrBold == 0 || style == style.Underline(false).Attributes(attrs) {
					t.Fatalf("styled control: rune=%q style=%#v", ch, style)
				}
				for _, cell := range []struct {
					x, y int
					ch   rune
				}{{1, 0, '界'}, {3, 0, 'e'}, {0, 1, '▒'}} {
					ch, combining, style, _ := screen.GetContent(cell.x, cell.y)
					if ch != cell.ch || style != initial {
						t.Errorf("%s draw=%d (%d,%d): rune=%q style=%#v, want %q %#v", reset, draw, cell.x, cell.y, ch, style, cell.ch, initial)
					}
					if ch == 'e' && (len(combining) != 1 || combining[0] != '\u0301') {
						t.Errorf("combining mark=%q, want acute accent", combining)
					}
				}
			}
		}
	}
}

// TestPrintAttributeReset checks the shared printing path when alignment clips
// off the start of the text before the reset tag is reached.
func TestPrintAttributeReset(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(2, 1)
	initial := tcell.StyleDefault.Foreground(tcell.ColorGreen).
		Underline(tcell.UnderlineStyleDouble, tcell.ColorYellow)
	_, _, width := printWithStyle(screen, "ZZ[::u]A[::-]B", 0, 0, 0, 2, AlignRight, initial, false)
	ch, _, style, _ := screen.GetContent(1, 0)
	if width != 2 || ch != 'B' || style != initial {
		t.Errorf("width=%d rune=%q style=%#v, want 2 B %#v", width, ch, style, initial)
	}
}
