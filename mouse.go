package tview

import (
	"time"
)

// MouseAction indicates one of the actions the mouse is logically doing.
type MouseAction int16

const (
	MouseMove MouseAction = iota
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
