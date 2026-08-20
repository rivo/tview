package tview

import (
	"errors"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// TestQueueUpdateReturnsAfterRunStops verifies that queued updates do not wait
// forever after the event loop stops unexpectedly.
func TestQueueUpdateReturnsAfterRunStops(t *testing.T) {
	screen := tcell.NewSimulationScreen("")
	app := NewApplication().SetRoot(NewBox(), true).SetScreen(screen)
	runDone := make(chan error, 1)
	go func() {
		runDone <- app.Run()
	}()

	screen.PostEventWait(tcell.NewEventError(errors.New("screen error")))
	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("Run did not return")
	}

	updateDone := make(chan struct{})
	go func() {
		app.QueueUpdate(func() {})
		close(updateDone)
	}()
	select {
	case <-updateDone:
	case <-time.After(time.Second):
		t.Fatal("QueueUpdate did not return")
	}
}

func TestQueueUpdateReturnsWhenRunStopsWhileWaiting(t *testing.T) {
	app := NewApplication()
	updates := make(chan queuedUpdate)
	updatesDone := make(chan struct{})
	app.updates = updates
	app.updatesDone = updatesDone

	queued := make(chan struct{})
	go func() {
		<-updates
		close(queued)
	}()
	returned := make(chan struct{})
	go func() {
		app.QueueUpdate(func() {})
		close(returned)
	}()
	select {
	case <-queued:
	case <-time.After(time.Second):
		t.Fatal("QueueUpdate did not queue the update")
	}

	close(updatesDone)
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("QueueUpdate did not return")
	}
}
