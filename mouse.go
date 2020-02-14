package tview

import (
	"time"

	"github.com/gdamore/tcell"
)

// MouseAction are bit flags indicating what the mouse is logically doing.
type MouseAction int32

const (
	MouseMove MouseAction = 1 << iota
	MouseLeftDown
	MouseLeftUp
	MouseLeftClick
	MouseLeftDoubleClick
	MouseMiddleDown
	MouseMiddleUp
	MouseMiddleClick
	MouseMiddleDoubleClick
	MouseRightDown
	MouseRightUp
	MouseRightClick
	MouseRightDoubleClick
	WheelUp
	WheelDown
	WheelLeft
	WheelRight
)

var DoubleClickInterval = 500 * time.Millisecond

// Does not set MouseMove or *Click actions.
func (action MouseAction) getMouseButtonAction(lastBtn, btn tcell.ButtonMask) MouseAction {
	btnDiff := btn ^ lastBtn

	if btnDiff&tcell.Button1 != 0 {
		if btn&tcell.Button1 != 0 {
			action |= MouseLeftDown
		} else {
			action |= MouseLeftUp
		}
	}

	if btnDiff&tcell.Button2 != 0 {
		if btn&tcell.Button2 != 0 {
			action |= MouseMiddleDown
		} else {
			action |= MouseMiddleUp
		}
	}

	if btnDiff&tcell.Button3 != 0 {
		if btn&tcell.Button3 != 0 {
			action |= MouseRightDown
		} else {
			action |= MouseRightUp
		}
	}

	if btn&tcell.WheelUp != 0 {
		action |= WheelUp
	}
	if btn&tcell.WheelDown != 0 {
		action |= WheelDown
	}
	if btn&tcell.WheelLeft != 0 {
		action |= WheelLeft
	}
	if btn&tcell.WheelRight != 0 {
		action |= WheelRight
	}

	return action
}

// Do not call if the mouse moved.
// Sets the *Click, including *DoubleClick.
// This should be called last, after setting all the other flags.
func (action MouseAction) getMouseClickAction(lastAct MouseAction, lastClickTime *time.Time) MouseAction {
	if action&MouseMove == 0 {
		if action&MouseLeftUp != 0 {
			if (*lastClickTime).Add(DoubleClickInterval).Before(time.Now()) {
				action |= MouseLeftClick
				*lastClickTime = time.Now()
			} else {
				action |= MouseLeftDoubleClick
				*lastClickTime = time.Time{} // reset
			}
		}
		if action&MouseMiddleUp != 0 {
			if (*lastClickTime).Add(DoubleClickInterval).Before(time.Now()) {
				action |= MouseMiddleClick
				*lastClickTime = time.Now()
			} else {
				action |= MouseMiddleDoubleClick
				*lastClickTime = time.Time{} // reset
			}
		}
		if action&MouseRightUp != 0 {
			if (*lastClickTime).Add(DoubleClickInterval).Before(time.Now()) {
				action |= MouseRightClick
				*lastClickTime = time.Now()
			} else {
				action |= MouseRightDoubleClick
				*lastClickTime = time.Time{} // reset
			}
		}
	}
	return action
}
