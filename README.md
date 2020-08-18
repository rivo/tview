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

- [password manager](https://github.com/7onetella/password)
- [CLI bookmark manager](https://github.com/Endi1/drawer)
- [A caving database interface written in Go](https://github.com/IdlePhysicist/cave-logger)
- [App for rental of electic bikes](https://github.com/MrDienns/bike-commerce)
- [interactive file browse and exec any command.](https://github.com/bannzai/itree)
- [A simple CRM](https://github.com/broadcastle/crm)
- [Terminal UI for todist](https://github.com/cyberdummy/todoista)
- [graphical kubectl wrapper](https://github.com/dcaiafa/kpick)
- [Decred Decentralized Exchange ](https://github.com/decred/dcrdex)
- [Kubernetes CLI To Manage Your Clusters In Style! ](https://github.com/derailed/k9s)
- [A CLI file browser for Raspberry PI](https://github.com/destinmoulton/pixi)
- [A tool to manage projects.](https://github.com/divramod/dp)
- [A simple app for BMI monitoring](https://github.com/erleene/go-bmi)
- [Stream TIDAL from command line](https://github.com/godsic/vibe)
- [Secure solution for fully decentralized password management](https://github.com/guillaumemichel/passtor/)
- [A growing collection of convenient little tools to work with systemd services](https://github.com/muesli/service-tools/)
- [A terminal based browser for Redis written in Go](https://github.com/nitishm/redis-terminal)
- [First project for the Computer Networks course.](https://github.com/pablogadhi/XMPPClient)
- [CLI tool build in Golang for managing ssh connection](https://github.com/patilsuraj767/connection-master)
- [Test your typing speed in the terminal!](https://github.com/shilangyu/typer-go)
- [TUI Client for Docker](https://github.com/skanehira/docui)
- [SSH client using certificates signed by HashiCorp Vault](https://github.com/stephane-martin/vssh)
- [A go terminal based pos software.](https://github.com/thebmw/y2kpos)
- [VMware vCenter Text UI](https://github.com/thebsdbox/vctui)
- [Bookmarks on terminal](https://github.com/tryffel/bookmarker)
- [A UDP testing utility](https://github.com/vaelen/udp-tester)
- [A simple Kanban board for your terminal](https://github.com/witchard/toukan)
- [The personal information dashboard for your terminal. ](https://github.com/wtfutil/wtf)
- [MySQL database to Golang struct](https://github.com/xxjwxc/gormt)
- [Cryptowatch Go SDK](https://github.com/y3sh/cw-sdk-go)
- [Discord, TUI and SIXEL.](https://gitlab.com/diamondburned/6cord)

## Documentation

Refer to https://pkg.go.dev/github.com/rivo/tview for the package's documentation.

## Dependencies

This package is based on [github.com/gdamore/tcell](https://github.com/gdamore/tcell) (and its dependencies) as well as on [github.com/rivo/uniseg](https://github.com/rivo/uniseg).

## Your Feedback

Add your issue here on GitHub. Feel free to get in touch if you have any questions.
