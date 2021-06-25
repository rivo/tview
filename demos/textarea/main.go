// Demo code for the TextView primitive.
package main

import (
	"github.com/rivo/tview"
)

const corporate = `See: [yellow]https://ru.wikipedia.org/wiki/Go

	Go (часто также Golang) — компилируемый многопоточный язык программирования, разработанный внутри компании Google[4]. Разработка Go началась в сентябре 2007 года, его непосредственным проектированием занимались Роберт Гризмер, Роб Пайк и Кен Томпсон[5], занимавшиеся до этого проектом разработки операционной системы Inferno. Официально язык был представлен в ноябре 2009 года.
	На данный момент поддержка официального компилятора, разрабатываемого создателями языка, осуществляется для операционных систем FreeBSD, OpenBSD, Linux, macOS, Windows, DragonFly BSD, Plan 9, Solaris, Android, AIX.[6].
	
See: [yellow]https://golang.org/

[green]// You can edit this code!
[green]// Click here and start typing.
package main
import "fmt"
func main() {
	// 012345689012345678901234567890
	fmt.Println("[red]Hello, 世界[white]")
}

You will see:
世界
`

func main() {
	app := tview.NewApplication()
	textArea := tview.NewTextArea().
		SetText(corporate)
	b := textArea.GetBox()
	b.SetBorder(true)
	if err := app.SetRoot(textArea, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
