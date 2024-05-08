package tview

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestApplication_deadlock_check(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Errorf("screen.Init() error = %v, want nil", err)
	}

	app := NewApplication()
	app.SetScreen(screen)
	app.SetRoot(NewBox().SetTitle("Hello, world!"), true)

	panicCh := make(chan bool)
	go func() {
		defer func() {
			panicCh <- recover() != nil
		}()
		if err := app.Run(); err != nil {
			t.Errorf("Application.Run() error = %v, want nil", err)
		}
	}()

	go func() {
		app.QueueUpdate(func() {
			app.QueueUpdate(func() {
				t.Errorf("impossible case")
			})
		})
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock detected")
	case panicked := <-panicCh:
		if panicked {
			t.Log("panic detected, deadlock avoided")
		} else {
			t.Log("impossible case where deadlock did not occur, but things are working fine :)")
		}
	}
}
