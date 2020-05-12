package tview

import (
	"bytes"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell"
)

const (
	snapshotPath string = "screen"
)

func TestEditBox(t *testing.T) {

	type testCase struct {
		border           bool
		screenX, screenY int
		text             string
		inp              []*tcell.EventKey
	}

	ch := make(chan testCase, 100)

	Repeat := func(ks []*tcell.EventKey, n int) (keys []*tcell.EventKey) {
		for i := 0; i < n; i++ {
			keys = append(keys, ks...)
		}
		return
	}

	var (
		up      = tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
		down    = tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
		left    = tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone)
		right   = tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone)
		char    = tcell.NewEventKey(0, 'й', tcell.ModNone)
		newline = tcell.NewEventKey(0, '\n', tcell.ModNone)
		delete  = tcell.NewEventKey(tcell.KeyDelete, '\n', tcell.ModNone)

		screenMin = 5
		screenMax = 10 // 5 // todo: do 10
	)

	_ = up
	_ = down
	_ = left
	_ = right
	_ = char
	_ = newline
	_ = delete

	go func() {
		for _, border := range []bool{false, true} {
			for screenX := screenMin; screenX <= screenMax; screenX++ {
				for screenY := screenMin; screenY <= screenMax; screenY++ {
					for _, text := range []string{
						"",
						" ",
						"\n",
						"\n\n",
						"\n \n",
						strings.Repeat(" ", 30),
						strings.Repeat("\n", 30),
						strings.Repeat("世界世界世界d世界\n", 5),
						"golang",
						strings.Repeat("golang\n", 20),
						strings.Repeat("golang\n\n", 10),
					} {
						for _, inp := range [][]*tcell.EventKey{
							// -
							[]*tcell.EventKey{},
							// -
							Repeat([]*tcell.EventKey{down}, 40),
							// -
							Repeat(append(append(append(
								Repeat([]*tcell.EventKey{right}, 10),
								Repeat([]*tcell.EventKey{left}, 2)...),
								down,
								up,
								down)), 40),
							// -
							Repeat(append(append(append(
								Repeat([]*tcell.EventKey{right}, 40),
								Repeat([]*tcell.EventKey{left}, 40)...),
								down,
								up,
								down)), 40),
							// -
						//	Repeat(append(Repeat([]*tcell.EventKey{right, char}, 40), left, down, up, down), 40),
						//	// -
						//	Repeat([]*tcell.EventKey{right, newline}, 40),
							// -
							Repeat([]*tcell.EventKey{delete}, 40),
							// -
							Repeat([]*tcell.EventKey{delete, right, up, down}, 40),
						} {
							ch <- testCase{border, screenX, screenY, text, inp}
						}
					}
				}
			}
		}
		close(ch)
	}()

	count := 0
	for tc := range ch {
		count++
		t.Run(fmt.Sprintf("%d", count), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("%#v", tc)
					str := debug.Stack()
					t.Fatalf("%v\n%v", r, string(str))
				}
			}()
			simScreen := tcell.NewSimulationScreen("UTF-8")
			simScreen.Init()
			simScreen.SetSize(tc.screenX, tc.screenY)

			app := NewApplication()
			defer func() {
				app.Stop()
			}()
			app.SetScreen(simScreen)
			eb := NewEditBox().SetText(tc.text)
			b := eb.GetBox()
			b.SetBorder(true)
			app.SetRoot(eb, true)

			go func() {
				if err := app.Run(); err != nil {
					panic(err)
				}
			}()

			for _, ek := range tc.inp {
				eb.InputHandler()(ek, nil)
				app.Draw()
			}
			time.Sleep(time.Millisecond) // for avoid terminal: too many tty

			cells, width, height := simScreen.GetContents()
			var buf bytes.Buffer
			for row := 0; row < height; row++ {
				for col := 0; col < width; col++ {
					position := row*width + col
					fmt.Fprintf(&buf, "%s", string(cells[position].Runes))
				}
				fmt.Fprintf(&buf, "\n")
			}

			fmt.Fprintf(os.Stdout, buf.String())

		})
	}

	// tcs := []struct {
	// 	name string
	// 	box  *Box
	// }{
	// 	{
	// 		name: "box.Simple",
	// 		box:  NewBox(),
	// 	},
	// 	{
	// 		name: "box.Bold.Border",
	// 		box:  NewBox().SetBorder(true).SetBorderAttributes(tcell.AttrBold).SetTitle("Hello"),
	// 	},
	// 	{
	// 		name: "box.AlignLeft",
	// 		box:  NewBox().SetBorder(true).SetTitle("Left").SetTitleAlign(AlignLeft),
	// 	},
	// 	{
	// 		name: "box.AlignRight",
	// 		box:  NewBox().SetBorder(true).SetTitle("Right").SetTitleAlign(AlignRight),
	// 	},
	// 	{
	// 		name: "box.AlignCenter",
	// 		box:  NewBox().SetBorder(true).SetTitle("Center").SetTitleAlign(AlignCenter),
	// 	},
	// }
	//
	// for i := range tcs {
	// 	i := i
	// 	t.Run(tcs[i].name, func(t *testing.T) {
	// 		t.Parallel()
	// 		simScreen := tcell.NewSimulationScreen("UTF-8")
	// 		simScreen.Init()
	// 		simScreen.SetSize(10, 5)
	//
	// 		app := NewApplication()
	// 		app.SetScreen(simScreen)
	// 		app.SetRoot(tcs[i].box, true)
	//
	// 		go func() {
	// 			if err := app.Run(); err != nil {
	// 				panic(err)
	// 			}
	// 		}()
	//
	// 		app.Draw()
	//
	// 		//	time.Sleep(time.Second)
	//
	// 		cells, width, height := simScreen.GetContents()
	// 		var buf bytes.Buffer
	// 		for row := 0; row < height; row++ {
	// 			for col := 0; col < width; col++ {
	// 				position := row*width + col
	// 				fmt.Fprintf(&buf, "%s", string(cells[position].Runes))
	// 			}
	// 			fmt.Fprintf(&buf, "\n")
	// 		}
	//
	// 		snapshotFilename := filepath.Join(".", snapshotPath, tcs[i].name)
	//
	// 		// for update test screens run in console:
	// 		// UPDATE=true go test
	// 		if os.Getenv("UPDATE") == "true" {
	// 			if err := ioutil.WriteFile(snapshotFilename, buf.Bytes(), 0644); err != nil {
	// 				t.Fatalf("Cannot write snapshot to file: %v", err)
	// 			}
	// 		}
	//
	// 		content, err := ioutil.ReadFile(snapshotFilename)
	// 		if err != nil {
	// 			t.Fatalf("Cannot read snapshot file: %v", err)
	// 		}
	//
	// 		if !bytes.Equal(buf.Bytes(), content) {
	// 			t.Errorf("Snapshots is not same")
	// 		}
	// 	})
	// }
}
