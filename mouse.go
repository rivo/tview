package tview

import (
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

// Does not set MouseMove or *Click actions.
func getMouseButtonAction(lastBtn, btn tcell.ButtonMask) MouseAction {
	btnDiff := btn ^ lastBtn
	var action MouseAction

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
func getMouseClickAction(lastAct, action MouseAction) MouseAction {
	if action&MouseMove == 0 {
		if action&MouseLeftUp != 0 {
			if lastAct&(MouseLeftClick&MouseLeftDoubleClick) == 0 {
				action |= MouseLeftClick
			} else if lastAct&MouseLeftDoubleClick == 0 {
				action |= MouseLeftDoubleClick
			}
		}
		if action&MouseMiddleUp != 0 {
			if lastAct&(MouseMiddleClick&MouseMiddleDoubleClick) == 0 {
				action |= MouseMiddleClick
			} else if lastAct&MouseMiddleDoubleClick == 0 {
				action |= MouseMiddleDoubleClick
			}
		}
		if action&MouseRightUp != 0 {
			if lastAct&(MouseRightClick&MouseRightDoubleClick) == 0 {
				action |= MouseRightClick
			} else if lastAct&MouseRightDoubleClick == 0 {
				action |= MouseRightDoubleClick
			}
		}
	}
	return action
}
