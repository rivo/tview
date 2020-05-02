package tview

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdamore/tcell"
)

const (
	snapshotPath string = "screen"
)

func TestBox(t *testing.T) {
	tcs := []struct {
		name string
		box  *Box
	}{
		{
			name: "box.Simple",
			box:  NewBox(),
		},
		{
			name: "box.Bold.Border",
			box:  NewBox().SetBorder(true).SetBorderAttributes(tcell.AttrBold).SetTitle("Hello"),
		},
		{
			name: "box.AlignLeft",
			box:  NewBox().SetBorder(true).SetTitle("Left").SetTitleAlign(AlignLeft),
		},
		{
			name: "box.AlignRight",
			box:  NewBox().SetBorder(true).SetTitle("Right").SetTitleAlign(AlignRight),
		},
		{
			name: "box.AlignCenter",
			box:  NewBox().SetBorder(true).SetTitle("Center").SetTitleAlign(AlignCenter),
		},
	}

	for i := range tcs {
		i := i
		t.Run(tcs[i].name, func(t *testing.T) {
			t.Parallel()
			simScreen := tcell.NewSimulationScreen("UTF-8")
			simScreen.Init()
			simScreen.SetSize(10, 5)

			app := NewApplication()
			app.SetScreen(simScreen)
			app.SetRoot(tcs[i].box, true)

			go func() {
				if err := app.Run(); err != nil {
					panic(err)
				}
			}()

			time.Sleep(time.Second)

			cells, width, height := simScreen.GetContents()
			var buf bytes.Buffer
			for row := 0; row < height; row++ {
				for col := 0; col < width; col++ {
					position := row*width + col
					fmt.Fprintf(&buf, "%s", string(cells[position].Runes))
				}
				fmt.Fprintf(&buf, "\n")
			}

			snapshotFilename := filepath.Join(".", snapshotPath, tcs[i].name)

			if os.Getenv("UPDATE") == "true" {
				if err := ioutil.WriteFile(snapshotFilename, buf.Bytes(), 0644); err != nil {
					t.Fatalf("Cannot write snapshot to file: %v", err)
				}
			}

			content, err := ioutil.ReadFile(snapshotFilename)
			if err != nil {
				t.Fatalf("Cannot read snapshot file: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), content) {
				t.Errorf("Snapshots is not same")
			}
		})
	}
}
