package tview

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"

	"github.com/gdamore/tcell"
)

// Uncomment only for dtaugging
var debugBuffer bytes.Buffer

func log(args ...interface{}) {
	fmt.Fprintln(&debugBuffer, args...)
}

func init() {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%s", debugBuffer.String())
		})
		http.ListenAndServe(":9090", nil)
	}()
}

const (
	snapshotPath string = "screen"
)

var (
	// typical keys
	KeyUp     = tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
	KeyDown   = tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
	KeyLeft   = tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone)
	KeyRight  = tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone)
	KeyHome   = tcell.NewEventKey(tcell.KeyHome, ' ', tcell.ModNone)
	KeyEnd    = tcell.NewEventKey(tcell.KeyEnd, ' ', tcell.ModNone)
	KeyEnter  = tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone)
	KeyChar   = tcell.NewEventKey(0, 'Ъ', tcell.ModNone)
	KeyDelete = tcell.NewEventKey(tcell.KeyDelete, ' ', tcell.ModNone)
)

func screenShot(s tcell.SimulationScreen) string {
	cells, width, height := s.GetContents()
	var buf bytes.Buffer
	x, y, visible := s.GetCursor()
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			position := row*width + col
			if col == x && row == y && visible {
				fmt.Fprintf(&buf, "█")
			} else {
				fmt.Fprintf(&buf, "%s", string(cells[position].Runes))
			}
		}
		fmt.Fprintf(&buf, "\n")
	}
	fmt.Fprintf(&buf, "Cursor {x:%d,y:%d} %v\n", x, y, visible)
	return buf.String()
}

func repearKeys(ks []*tcell.EventKey, n int) (keys []*tcell.EventKey) {
	for i := 0; i < n; i++ {
		keys = append(keys, ks...)
	}
	return
}

type testCase struct {
	border  bool
	screenX int
	screenY int
	text    string
	keys    []*tcell.EventKey
}

func (tc testCase) Run(t *testing.T, simScreen tcell.SimulationScreen) (
	app *Application,
	ta *TextArea,
) {
	if simScreen == nil {
		t.Fatalf("simScreen is nil")
	}
	simScreen.Clear()
	if err := simScreen.Init(); err != nil {
		t.Fatalf("%v", err)
	}
	simScreen.SetSize(tc.screenX, tc.screenY)

	app = NewApplication()
	ta = NewTextArea().SetText(tc.text)
	b := ta.GetBox()
	b.SetBorder(tc.border)
	app.SetRoot(ta, false)

	go func() {
		app.SetScreen(simScreen)
		if err := app.Run(); err != nil {
			panic(err)
		}
	}()
	return
}

var tcs []testCase

type tkey = *tcell.EventKey

var (
	borders = []bool{false, true}
	texts   = []string{
		"",
		strings.Repeat("世界", 5),
		strings.Repeat("世界", 50),
		strings.Repeat(strings.Repeat("世界", 10)+"\n", 2),
		strings.Repeat(strings.Repeat("世界", 10)+"\n", 20),
	}
	movements = [][]*tcell.EventKey{
		// Do nothing
		[]*tcell.EventKey{},
		// List to down and up
		append(repearKeys([]tkey{KeyDown}, 40), repearKeys([]tkey{KeyUp}, 40)...),
		// Moving on each rune
		repearKeys(append(append(repearKeys([]tkey{KeyRight}, 40), repearKeys([]tkey{KeyLeft}, 40)...), KeyDown, KeyUp, KeyDown), 40),
		repearKeys(append(repearKeys([]tkey{KeyRight}, 40), KeyDown, KeyUp, KeyDown), 40),
		// Fast move by line
		repearKeys([]tkey{KeyHome, KeyEnd, KeyDown}, 10),
		// Delete rune, newline
		repearKeys([]tkey{KeyHome, KeyChar, KeyEnter, KeyDelete, KeyDelete, KeyEnd, KeyChar, KeyDelete, KeyDown}, 10),
		// single operations
		[]tkey{KeyEnter},
		[]tkey{KeyDelete},
		// Delete all
		repearKeys(append([]tkey{KeyEnd}, repearKeys([]tkey{KeyDelete}, 40)...), 20),
	}
	screenSizes = []int{5, 7, 40}
	mouseMove   = []struct{ x, y int }{
		{0, 0},
		{0, 100},
		{100, -1},
		{3, 3},
	}
)

func init() {
	for _, border := range []bool{false, true} {
		for _, screenX := range screenSizes {
			for _, screenY := range screenSizes {
				for _, text := range texts {
					for _, keys := range movements {
						tcs = append(tcs, testCase{
							border:  border,
							screenX: screenX,
							screenY: screenY,
							text:    text,
							keys:    keys,
						})
					}
				}
			}
		}
	}
}

func TestTextAreaCrash(t *testing.T) {
	if testing.Short() {
		t.Skip("run this test only for finding the crash")
	}
	count := 0
	for _, tc := range tcs {
		tc := tc
		count++
		t.Run(fmt.Sprintf("%d", count), func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r != nil {
					t.Logf("%#v", tc)
					str := debug.Stack()
					t.Fatalf("%v\n%v", r, string(str))
				}
			}()

			simScreen := tcell.NewSimulationScreen("UTF-8")
			app, ta := tc.Run(t, simScreen)
			defer func() {
				app.Stop()
			}()

			isChanged := false
			for _, ek := range tc.keys {
				if ek.Key() == tcell.KeyDelete ||
					ek.Key() == tcell.KeyEnter ||
					ek.Key() == KeyChar.Key() ||
					ek.Key() == tcell.KeyBackspace {
					isChanged = true
				}
				ta.InputHandler()(ek, nil)
				app.Draw()
			}

			for i := range mouseMove {
				ta.MouseHandler()(
					MouseLeftClick,
					tcell.NewEventMouse(mouseMove[i].x, mouseMove[i].y, tcell.Button1, tcell.ModNone),
					func(p Primitive) {})
			}

			if !isChanged || (isChanged && len(ta.GetText()) == len(tc.text)) {
				if ta.GetText() != tc.text {
					t.Errorf("text is not same")
				}
			}
		})
	}
}

func TestTextArea(t *testing.T) {
	tcs := []testCase{
		{
			border:  true,
			screenX: 19,
			screenY: 20,
			text: `// You can edit this code!
// Click here and start typing.
package main

import "fmt"

func main() {
fmt.Println("Hello, 世界")
}`,

			keys: []tkey{
				KeyRight, KeyLeft,
				KeyDown, KeyUp,
				KeyEnd, KeyHome,
				KeyUp, KeyLeft,
				KeyDown, KeyDown, KeyDown, KeyDown, KeyDown, KeyDown, KeyDown, KeyDown, KeyDown, KeyDown,
				KeyRight,
				KeyUp, KeyEnd, KeyRight,
				KeyEnter, KeyChar, KeyDelete, KeyDelete,
			},
		},
	}

	for index, tc := range tcs {
		t.Run(fmt.Sprintf("%d", index), func(t *testing.T) {
			simScreen := tcell.NewSimulationScreen("UTF-8")
			app, ta := tc.Run(t, simScreen)
			defer func() {
				app.Stop()
			}()

			// filename
			prefix := strings.Replace(t.Name(), string(filepath.Separator), "_", -1)

			for step, ek := range tc.keys {
				ta.InputHandler()(ek, nil)
				app.Draw()

				// screenshot comparing
				ss := screenShot(simScreen)
				snapshotFilename := filepath.Join(".", snapshotPath, prefix +"_"+ strconv.Itoa(step))

				// for update test screens run in console:
				// UPDATE=true go test
				if os.Getenv("UPDATE") == "true" {
					if err := ioutil.WriteFile(snapshotFilename, []byte(ss), 0644); err != nil {
						t.Fatalf("Cannot write snapshot to file: %v", err)
					}
				}

				content, err := ioutil.ReadFile(snapshotFilename)
				if err != nil {
					t.Fatalf("Cannot read snapshot file: %v", err)
				}

				if !bytes.Equal([]byte(ss), content) {
					t.Errorf("Snapshots is not same:\n%s\n%s", ss, string(content))
				}
			}

			if b := ta.GetText(); b != tc.text {
				t.Errorf("text is not same:\n%s\n%s",b,tc.text)
			}
		})
	}

}
