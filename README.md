# Rich Interactive Widgets for Terminal UIs

[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/rivo/tview)
[![Go Report](https://img.shields.io/badge/go%20report-A%2B-brightgreen.svg)](https://goreportcard.com/report/github.com/rivo/tview)

This Go package provides commonly needed components for terminal based user interfaces.

![Screenshot](tview.gif)

Among these components are:

- __Input forms__ (include __input/password fields__, __drop-down selections__, __checkboxes__, and __buttons__)
- Navigable multi-color __text views__
- Sophisticated navigable __table views__
- Flexible __tree views__
- Selectable __lists__
- __Grid__, __Flexbox__ and __page layouts__
- Modal __message windows__
- An __application__ wrapper

They come with lots of customization options and can be easily extended to fit your needs.

## Installation

```bash
go get github.com/rivo/tview
```

## Hello World

This basic example creates a box titled "Hello, World!" and displays it in your terminal:

```go
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
```

Check out the [GitHub Wiki](https://github.com/rivo/tview/wiki) for more examples along with screenshots. Or try the examples in the "demos" subdirectory.

For a presentation highlighting this package, compile and run the program found in the "demos/presentation" subdirectory.

## Projects using tview

- https://github.com/7onetella/password
- https://github.com/Endi1/drawer
- https://github.com/IdlePhysicist/cave-logger
- https://github.com/MrDienns/bike-commerce
- https://github.com/bannzai/itree
- https://github.com/bluemon0919/todotool
- https://github.com/broadcastle/crm
- https://github.com/cyberdummy/todoista
- https://github.com/dcaiafa/kpick
- https://github.com/decred/dcrdex
- https://github.com/derailed/k9s
- https://github.com/destinmoulton/pixi
- https://github.com/divramod/dp
- https://github.com/epiclabs-io/winman
- https://github.com/erleene/go-bmi
- https://github.com/godsic/vibe
- https://github.com/guillaumemichel/passtor/tree/master/client
- https://github.com/muesli/service-tools/
- https://github.com/nitishm/redis-terminal
- https://github.com/pablogadhi/XMPPClient
- https://github.com/patilsuraj767/connection-master
- https://github.com/shilangyu/typer-go
- https://github.com/skanehira/docuix
- https://github.com/stephane-martin/vssh
- https://github.com/thebmw/y2kpos
- https://github.com/thebsdbox/vctui
- https://github.com/tryffel/bookmarker
- https://github.com/vaelen/udp-tester
- https://github.com/witchard/toukan
- https://github.com/wtfutil/wtf
- https://github.com/xxjwxc/gormt
- https://github.com/y3sh/cw-sdk-go
- https://gitlab.com/diamondburned/6cord

## Documentation

Refer to https://pkg.go.dev/github.com/rivo/tview for the package's documentation.

## Dependencies

This package is based on [github.com/gdamore/tcell](https://github.com/gdamore/tcell) (and its dependencies) as well as on [github.com/rivo/uniseg](https://github.com/rivo/uniseg).

## Your Feedback

Add your issue here on GitHub. Feel free to get in touch if you have any questions.
