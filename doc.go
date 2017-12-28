/*
Package tview implements primitives for terminal based applications. It uses
github.com/gdamore/tcell.

Hello World

Here is a very basic example showing a box with the title "Hello, world!":

  package main

  import (
  	"github.com/rivo/tview"
  )

  func main() {
  	box := tview.NewBox().SetBorder(true).SetTitle("Hello, world!")
  	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
  		panic(err)
  	}
  }

First, we create a box primitive with a border and a title. Then we create an
application, set the box as its root primitive, and run the event loop. It
exits when the application's Stop() function is called or when Ctrl-C is
pressed.

If we have a primitive which consumes key presses, we call the application's
SetFocus() function to redirect all key presses to that primitive. Most
primitives then offer ways to install handlers that allow you to react to any
actions performed on them.

No mouse input (yet).
*/
package tview
