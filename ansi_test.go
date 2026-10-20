package tview_test

import (
	"io"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestANSIAttributes checks translated attributes through the TextView renderer.
func TestANSIAttributes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []tcell.AttrMask
	}{
		{"bold and dim", "\x1b[1;2mA\x1b[22mB\x1b[1mC", []tcell.AttrMask{tcell.AttrBold | tcell.AttrDim, 0, tcell.AttrBold}},
		{"italic", "\x1b[3mA\x1b[23mB\x1b[3mC", []tcell.AttrMask{tcell.AttrItalic, 0, tcell.AttrItalic}},
		{"underline", "\x1b[4mA\x1b[24mB\x1b[4mC", []tcell.AttrMask{tcell.AttrUnderline, 0, tcell.AttrUnderline}},
		{"blink", "\x1b[5mA\x1b[25mB\x1b[5mC", []tcell.AttrMask{tcell.AttrBlink, 0, tcell.AttrBlink}},
		{"reverse", "\x1b[7mA\x1b[27mB\x1b[7mC", []tcell.AttrMask{tcell.AttrReverse, 0, tcell.AttrReverse}},
		{"strike", "\x1b[9mA\x1b[29mB\x1b[9mC", []tcell.AttrMask{tcell.AttrStrikeThrough, 0, tcell.AttrStrikeThrough}},
		{"stack trace", "\x1b[4m\x1b[1mA\x1b[24mB\x1b[22mC", []tcell.AttrMask{tcell.AttrUnderline | tcell.AttrBold, tcell.AttrBold, 0}},
		{"reset then enable", "\x1b[1mA\x1b[0;3mB\x1b[0mC", []tcell.AttrMask{tcell.AttrBold, tcell.AttrItalic, 0}},
		{"ordered parameters", "\x1b[1;22mA\x1b[22;1mB\x1b[1;0mC", []tcell.AttrMask{0, tcell.AttrBold, 0}},
		{"terminal reset", "\x1b[1mA\x1bc\x1b[31mB\x1b[3mC", []tcell.AttrMask{tcell.AttrBold, 0, tcell.AttrItalic}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, chunked := range []bool{false, true} {
				screen := tcell.NewSimulationScreen("UTF-8")
				if err := screen.Init(); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(screen.Fini)
				screen.SetSize(10, 2)
				view := tview.NewTextView().SetDynamicColors(true)
				view.SetRect(0, 0, 10, 2)
				if chunked {
					writer := tview.ANSIWriter(view)
					for _, b := range []byte(test.input) {
						if _, err := writer.Write([]byte{b}); err != nil {
							t.Fatal(err)
						}
					}
				} else {
					if _, err := io.WriteString(view, tview.TranslateANSI(test.input)); err != nil {
						t.Fatal(err)
					}
				}
				view.Draw(screen)
				for x, want := range test.want {
					ch, _, style, _ := screen.GetContent(x, 0)
					_, _, got := style.Decompose()
					if ch != rune('A'+x) || got != want {
						t.Errorf("chunked=%v cell=%d: rune=%q attributes=%v, want %v", chunked, x, ch, got, want)
					}
				}
			}
		})
	}
}

// TestANSIColorAndLinePersistence checks that attribute changes preserve colors
// and that a reset restores the TextView's default color across line boundaries.
func TestANSIColorAndLinePersistence(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(10, 3)
	view := tview.NewTextView().SetDynamicColors(true).SetTextColor(tcell.ColorGreen)
	view.SetRect(0, 0, 10, 3)
	writer := tview.ANSIWriter(view)
	for _, text := range []string{"\x1b[1;31mA\n", "\x1b[22mB\n", "\x1b[0;2mC"} {
		if _, err := io.WriteString(writer, text); err != nil {
			t.Fatal(err)
		}
	}
	view.Draw(screen)
	for y, want := range []struct {
		color tcell.Color
		attrs tcell.AttrMask
	}{{tcell.ColorMaroon, tcell.AttrBold}, {tcell.ColorMaroon, 0}, {tcell.ColorGreen, tcell.AttrDim}} {
		ch, _, style, _ := screen.GetContent(0, y)
		color, _, attrs := style.Decompose()
		if ch != rune('A'+y) || color != want.color || attrs != want.attrs {
			t.Errorf("line %d: rune=%q color=%v attributes=%v, want color=%v attributes=%v", y, ch, color, attrs, want.color, want.attrs)
		}
	}
}
