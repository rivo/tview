package tview

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestQueueUpdateDraw_ConcurrentRunErrorExit(t *testing.T) {
	for i := 0; i < 200; i++ {
		screen := tcell.NewSimulationScreen("")
		if err := screen.Init(); err != nil {
			t.Fatalf("failed to init simulation screen: %v", err)
		}

		app := NewApplication()
		app.SetScreen(screen)

		runErrCh := make(chan error, 1)
		go func() {
			runErrCh <- app.Run()
		}()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.QueueUpdateDraw(func() {})
		}()

		screen.PostEvent(tcell.NewEventError(errors.New("screen failed")))

		select {
		case <-runErrCh:
		case <-time.After(2 * time.Second):
			t.Fatalf("iteration %d: app.Run did not return in time", i)
		}

		doneCh := make(chan struct{})
		go func() {
			wg.Wait()
			close(doneCh)
		}()

		select {
		case <-doneCh:
		case <-time.After(2 * time.Second):
			t.Fatalf("iteration %d: QueueUpdateDraw blocked after Run exited", i)
		}
	}
}

func TestQueueUpdateDraw_ConcurrentStopExit(t *testing.T) {
	for i := 0; i < 200; i++ {
		screen := tcell.NewSimulationScreen("")
		if err := screen.Init(); err != nil {
			t.Fatalf("failed to init simulation screen: %v", err)
		}

		app := NewApplication()
		app.SetScreen(screen)

		runErrCh := make(chan error, 1)
		go func() {
			runErrCh <- app.Run()
		}()

		var wg sync.WaitGroup
		for j := 0; j < 5; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				app.QueueUpdateDraw(func() {})
			}()
		}

		app.Stop()

		select {
		case <-runErrCh:
		case <-time.After(2 * time.Second):
			t.Fatalf("iteration %d: app.Run did not return in time after Stop", i)
		}

		doneCh := make(chan struct{})
		go func() {
			wg.Wait()
			close(doneCh)
		}()

		select {
		case <-doneCh:
		case <-time.After(2 * time.Second):
			t.Fatalf("iteration %d: QueueUpdateDraw blocked after Stop", i)
		}
	}
}

func TestQueueUpdateDraw_AfterApplicationStopped(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}

	app := NewApplication()
	app.SetScreen(screen)

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- app.Run()
	}()

	app.Stop()

	select {
	case <-runErrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("app.Run did not return in time after Stop")
	}

	doneCh := make(chan struct{})
	go func() {
		app.QueueUpdateDraw(func() {})
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(1 * time.Second):
		t.Fatal("QueueUpdateDraw blocked after application stopped")
	}
}

func TestQueueUpdateDraw_ChannelFullAfterStop(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}

	app := NewApplication()
	app.SetScreen(screen)

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- app.Run()
	}()

	app.Stop()

	select {
	case <-runErrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("app.Run did not return in time after Stop")
	}

	doneCh := make(chan struct{})
	go func() {
		for i := 0; i < queueSize*2; i++ {
			app.QueueUpdateDraw(func() {})
		}
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(2 * time.Second):
		t.Fatal("QueueUpdateDraw blocked when updates channel was saturated after stop")
	}
}

func TestQueueUpdateDraw_NormalExecution(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}

	app := NewApplication()
	app.SetScreen(screen)

	box := NewBox()
	app.SetRoot(box, true)

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- app.Run()
	}()

	executed := false
	doneCh := make(chan struct{})
	go func() {
		app.QueueUpdateDraw(func() {
			executed = true
		})
		close(doneCh)
	}()

	select {
	case <-doneCh:
		if !executed {
			t.Fatal("expected update closure to have executed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("QueueUpdateDraw blocked during normal execution")
	}

	app.Stop()

	select {
	case <-runErrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("app.Run did not return in time after Stop")
	}
}

func TestQueueUpdate_RestartApplication(t *testing.T) {
	app := NewApplication()

	for round := 0; round < 2; round++ {
		screen := tcell.NewSimulationScreen("")
		if err := screen.Init(); err != nil {
			t.Fatalf("round %d: failed to init simulation screen: %v", round, err)
		}

		app.SetScreen(screen)

		runErrCh := make(chan error, 1)
		go func() {
			runErrCh <- app.Run()
		}()

		time.Sleep(20 * time.Millisecond)

		executed := false
		doneCh := make(chan struct{})
		go func() {
			app.QueueUpdate(func() {
				executed = true
			})
			close(doneCh)
		}()

		select {
		case <-doneCh:
			if !executed {
				t.Fatalf("round %d: expected update to execute", round)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("round %d: QueueUpdate timed out", round)
		}

		app.Stop()

		select {
		case <-runErrCh:
		case <-time.After(2 * time.Second):
			t.Fatalf("round %d: app.Run did not return in time after Stop", round)
		}
	}
}

func TestQueueUpdateDraw_NilClosure(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}

	app := NewApplication()
	app.SetScreen(screen)

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- app.Run()
	}()

	doneCh := make(chan struct{})
	go func() {
		app.QueueUpdateDraw(nil)
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(2 * time.Second):
		t.Fatal("QueueUpdateDraw(nil) blocked")
	}

	app.Stop()

	select {
	case <-runErrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("app.Run did not return in time after Stop")
	}
}

func TestQueueUpdateDraw_HighConcurrencyShutdown(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}

	app := NewApplication()
	app.SetScreen(screen)

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- app.Run()
	}()

	var wg sync.WaitGroup
	stopCalling := make(chan struct{})

	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopCalling:
					return
				default:
					app.QueueUpdateDraw(func() {})
				}
			}
		}()
	}

	time.Sleep(50 * time.Millisecond)
	close(stopCalling)
	app.Stop()

	select {
	case <-runErrCh:
	case <-time.After(3 * time.Second):
		t.Fatal("app.Run did not return in time after Stop")
	}

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(3 * time.Second):
		t.Fatal("concurrent QueueUpdateDraw goroutines did not exit in time")
	}
}

func TestSync_ConcurrentShutdown(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}

	app := NewApplication()
	app.SetScreen(screen)

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- app.Run()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			app.Sync()
		}()
	}

	app.Stop()

	select {
	case <-runErrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("app.Run did not return in time after Stop")
	}

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-time.After(2 * time.Second):
		t.Fatal("Sync goroutines did not return in time")
	}
}
