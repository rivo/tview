package tview

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

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
		inp              []tcell.KeyEvent
	}

	ch := make(chan testCase, 0)

	Repeat := func(ks []tcell.Key, n int) (keys []tcell.Key) {
		for i := 0; i < n; i++ {
			keys = append(keys, ks...)
		}
		return
	}

	go func() {
		for _, border := range []bool{false, true} {
			for screenX := 5; screenX <= 10; screenX++ {
				for screenY := 5; screenY <= 10; screenY++ {
					for _, text := range []string{
						"",
						strings.Repeat("世界世界世界d世界\n", 5),
						"golang",
						strings.Repeat("golang\n", 20),
						strings.Repeat("golang\n\n", 10),
					} {
						for _, inp := range [][]tcell.Key{
							//  []tcell.Key{},
							Repeat([]tcell.Key{tcell.KeyDown}, 40),
							// 	Repeat(append(append(append(
							// 		Repeat([]tcell.Key{tcell.KeyRight}, 10),
							// 		Repeat([]tcell.Key{tcell.KeyLeft}, 2)...),
							// 		tcell.KeyDown,
							// 		tcell.KeyUp,
							// 		tcell.KeyDown)), 40),
							// 	Repeat(append(append(append(
							// 		Repeat([]tcell.Key{tcell.KeyRight}, 40),
							// 		Repeat([]tcell.Key{tcell.KeyLeft}, 40)...),
							// 		tcell.KeyDown,
							// 		tcell.KeyUp,
							// 		tcell.KeyDown)), 40),
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
			t.Parallel()
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

			// 			defer func() {
			// 				if r := recover(); r!=nil{
			// 					app.Stop()
			// 					t.Fatalf("%v", r)
			// 				}
			// 			}()

			for _, k := range tc.inp {
				event := tcell.NewEventKey(k, ' ', tcell.ModNone)
				eb.InputHandler()(event, nil)
				app.Draw()
			}

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
