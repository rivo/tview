// +build ignore

// Copyright 2017 The TCell Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This command is used to generate suitable configuration files in either
// go syntax or in JSON.  It defaults to JSON output on stdout.  If no
// term values are specified on the command line, then $TERM is used.
//
// Usage is like this:
//
// mkinfo [-init] [-go file.go] [-json file.json] [-quiet] [-nofatal] [<term>...]
//
// -gzip     specifies output should be compressed (json only)
// -go       specifies Go output into the named file.  Use - for stdout.
// -json     specifies JSON output in the named file.  Use - for stdout
// -nofatal  indicates that errors loading definitions should not be fatal
//

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/terminfo"
)

type termcap struct {
	name    string
	desc    string
	aliases []string
	bools   map[string]bool
	nums    map[string]int
	strs    map[string]string
}

func (tc *termcap) getnum(s string) int {
	return (tc.nums[s])
}

func (tc *termcap) getflag(s string) bool {
	return (tc.bools[s])
}

func (tc *termcap) getstr(s string) string {
	return (tc.strs[s])
}

const (
	NONE = iota
	CTRL
	ESC
)

func unescape(s string) string {
	// Various escapes are in \x format.  Control codes are
	// encoded as ^M (carat followed by ASCII equivalent).
	// Escapes are: \e, \E - escape
	//  \0 NULL, \n \l \r \t \b \f \s for equivalent C escape.
	buf := &bytes.Buffer{}
	esc := NONE

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch esc {
		case NONE:
			switch c {
			case '\\':
				esc = ESC
			case '^':
				esc = CTRL
			default:
				buf.WriteByte(c)
			}
		case CTRL:
			buf.WriteByte(c - 0x40)
			esc = NONE
		case ESC:
			switch c {
			case 'E', 'e':
				buf.WriteByte(0x1b)
			case '0':
				buf.WriteByte(0)
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case 's':
				buf.WriteByte(' ')
			case 'l':
				panic("WTF: weird format: " + s)
			default:
				buf.WriteByte(c)
			}
			esc = NONE
		}
	}
	return (buf.String())
}

func (tc *termcap) setupterm(name string) error {
	cmd := exec.Command("infocmp", "-1", name)
	output := &bytes.Buffer{}
	cmd.Stdout = output

	tc.strs = make(map[string]string)
	tc.bools = make(map[string]bool)
	tc.nums = make(map[string]int)

	err := cmd.Run()
	if err != nil {
		return err
	}

	// Now parse the output.
	// We get comment lines (starting with "#"), followed by
	// a header line that looks like "<name>|<alias>|...|<desc>"
	// then capabilities, one per line, starting with a tab and ending
	// with a comma and newline.
	lines := strings.Split(output.String(), "\n")
	for len(lines) > 0 && strings.HasPrefix(lines[0], "#") {
		lines = lines[1:]
	}

	// Ditch trailing empty last line
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	header := lines[0]
	if strings.HasSuffix(header, ",") {
		header = header[:len(header)-1]
	}
	names := strings.Split(header, "|")
	tc.name = names[0]
	names = names[1:]
	if len(names) > 0 {
		tc.desc = names[len(names)-1]
		names = names[:len(names)-1]
	}
	tc.aliases = names
	for _, val := range lines[1:] {
		if (!strings.HasPrefix(val, "\t")) ||
			(!strings.HasSuffix(val, ",")) {
			return (errors.New("malformed infocmp: " + val))
		}

		val = val[1:]
		val = val[:len(val)-1]

		if k := strings.SplitN(val, "=", 2); len(k) == 2 {
			tc.strs[k[0]] = unescape(k[1])
		} else if k := strings.SplitN(val, "#", 2); len(k) == 2 {
			if u, err := strconv.ParseUint(k[1], 10, 0); err != nil {
				return (err)
			} else {
				tc.nums[k[0]] = int(u)
			}
		} else {
			tc.bools[val] = true
		}
	}
	return nil
}

// This program is used to collect data from the system's terminfo library,
// and write it into Go source code.  That is, we maintain our terminfo
// capabilities encoded in the program.  It should never need to be run by
// an end user, but developers can use this to add codes for additional
// terminal types.
//
// If a terminal name ending with -truecolor is given, and we cannot find
// one, we will try to fabricate one from either the -256color (if present)
// or the unadorned base name, adding the XTerm specific 24-bit color
// escapes.  We believe that all 24-bit capable terminals use the same
// escape sequences, and terminfo has yet to evolve to support this.
func getinfo(name string) (*terminfo.Terminfo, string, error) {
	var tc termcap
	addTrueColor := false
	if err := tc.setupterm(name); err != nil {
		if strings.HasSuffix(name, "-truecolor") {
			base := name[:len(name)-len("-truecolor")]
			// Probably -256color is closest to what we want
			if err = tc.setupterm(base + "-256color"); err != nil {
				err = tc.setupterm(base)
			}
			if err == nil {
				addTrueColor = true
			}
			tc.name = name
		}
		if err != nil {
			return nil, "", err
		}
	}
	t := &terminfo.Terminfo{}
	// If this is an alias record, then just emit the alias
	t.Name = tc.name
	if t.Name != name {
		return t, "", nil
	}
	t.Aliases = tc.aliases
	t.Colors = tc.getnum("colors")
	t.Columns = tc.getnum("cols")
	t.Lines = tc.getnum("lines")
	t.Bell = tc.getstr("bel")
	t.Clear = tc.getstr("clear")
	t.EnterCA = tc.getstr("smcup")
	t.ExitCA = tc.getstr("rmcup")
	t.ShowCursor = tc.getstr("cnorm")
	t.HideCursor = tc.getstr("civis")
	t.AttrOff = tc.getstr("sgr0")
	t.Underline = tc.getstr("smul")
	t.Bold = tc.getstr("bold")
	t.Blink = tc.getstr("blink")
	t.Dim = tc.getstr("dim")
	t.Reverse = tc.getstr("rev")
	t.EnterKeypad = tc.getstr("smkx")
	t.ExitKeypad = tc.getstr("rmkx")
	t.SetFg = tc.getstr("setaf")
	t.SetBg = tc.getstr("setab")
	t.SetCursor = tc.getstr("cup")
	t.CursorBack1 = tc.getstr("cub1")
	t.CursorUp1 = tc.getstr("cuu1")
	t.KeyF1 = tc.getstr("kf1")
	t.KeyF2 = tc.getstr("kf2")
	t.KeyF3 = tc.getstr("kf3")
	t.KeyF4 = tc.getstr("kf4")
	t.KeyF5 = tc.getstr("kf5")
	t.KeyF6 = tc.getstr("kf6")
	t.KeyF7 = tc.getstr("kf7")
	t.KeyF8 = tc.getstr("kf8")
	t.KeyF9 = tc.getstr("kf9")
	t.KeyF10 = tc.getstr("kf10")
	t.KeyF11 = tc.getstr("kf11")
	t.KeyF12 = tc.getstr("kf12")
	t.KeyF13 = tc.getstr("kf13")
	t.KeyF14 = tc.getstr("kf14")
	t.KeyF15 = tc.getstr("kf15")
	t.KeyF16 = tc.getstr("kf16")
	t.KeyF17 = tc.getstr("kf17")
	t.KeyF18 = tc.getstr("kf18")
	t.KeyF19 = tc.getstr("kf19")
	t.KeyF20 = tc.getstr("kf20")
	t.KeyF21 = tc.getstr("kf21")
	t.KeyF22 = tc.getstr("kf22")
	t.KeyF23 = tc.getstr("kf23")
	t.KeyF24 = tc.getstr("kf24")
	t.KeyF25 = tc.getstr("kf25")
	t.KeyF26 = tc.getstr("kf26")
	t.KeyF27 = tc.getstr("kf27")
	t.KeyF28 = tc.getstr("kf28")
	t.KeyF29 = tc.getstr("kf29")
	t.KeyF30 = tc.getstr("kf30")
	t.KeyF31 = tc.getstr("kf31")
	t.KeyF32 = tc.getstr("kf32")
	t.KeyF33 = tc.getstr("kf33")
	t.KeyF34 = tc.getstr("kf34")
	t.KeyF35 = tc.getstr("kf35")
	t.KeyF36 = tc.getstr("kf36")
	t.KeyF37 = tc.getstr("kf37")
	t.KeyF38 = tc.getstr("kf38")
	t.KeyF39 = tc.getstr("kf39")
	t.KeyF40 = tc.getstr("kf40")
	t.KeyF41 = tc.getstr("kf41")
	t.KeyF42 = tc.getstr("kf42")
	t.KeyF43 = tc.getstr("kf43")
	t.KeyF44 = tc.getstr("kf44")
	t.KeyF45 = tc.getstr("kf45")
	t.KeyF46 = tc.getstr("kf46")
	t.KeyF47 = tc.getstr("kf47")
	t.KeyF48 = tc.getstr("kf48")
	t.KeyF49 = tc.getstr("kf49")
	t.KeyF50 = tc.getstr("kf50")
	t.KeyF51 = tc.getstr("kf51")
	t.KeyF52 = tc.getstr("kf52")
	t.KeyF53 = tc.getstr("kf53")
	t.KeyF54 = tc.getstr("kf54")
	t.KeyF55 = tc.getstr("kf55")
	t.KeyF56 = tc.getstr("kf56")
	t.KeyF57 = tc.getstr("kf57")
	t.KeyF58 = tc.getstr("kf58")
	t.KeyF59 = tc.getstr("kf59")
	t.KeyF60 = tc.getstr("kf60")
	t.KeyF61 = tc.getstr("kf61")
	t.KeyF62 = tc.getstr("kf62")
	t.KeyF63 = tc.getstr("kf63")
	t.KeyF64 = tc.getstr("kf64")
	t.KeyInsert = tc.getstr("kich1")
	t.KeyDelete = tc.getstr("kdch1")
	t.KeyBackspace = tc.getstr("kbs")
	t.KeyHome = tc.getstr("khome")
	t.KeyEnd = tc.getstr("kend")
	t.KeyUp = tc.getstr("kcuu1")
	t.KeyDown = tc.getstr("kcud1")
	t.KeyRight = tc.getstr("kcuf1")
	t.KeyLeft = tc.getstr("kcub1")
	t.KeyPgDn = tc.getstr("knp")
	t.KeyPgUp = tc.getstr("kpp")
	t.KeyBacktab = tc.getstr("kcbt")
	t.KeyExit = tc.getstr("kext")
	t.KeyCancel = tc.getstr("kcan")
	t.KeyPrint = tc.getstr("kprt")
	t.KeyHelp = tc.getstr("khlp")
	t.KeyClear = tc.getstr("kclr")
	t.AltChars = tc.getstr("acsc")
	t.EnterAcs = tc.getstr("smacs")
	t.ExitAcs = tc.getstr("rmacs")
	t.EnableAcs = tc.getstr("enacs")
	t.Mouse = tc.getstr("kmous")
	t.KeyShfRight = tc.getstr("kRIT")
	t.KeyShfLeft = tc.getstr("kLFT")
	t.KeyShfHome = tc.getstr("kHOM")
	t.KeyShfEnd = tc.getstr("kEND")

	// Terminfo lacks descriptions for a bunch of modified keys,
	// but modern XTerm and emulators often have them.  Let's add them,
	// if the shifted right and left arrows are defined.
	if t.KeyShfRight == "\x1b[1;2C" && t.KeyShfLeft == "\x1b[1;2D" {
		t.KeyShfUp = "\x1b[1;2A"
		t.KeyShfDown = "\x1b[1;2B"
		t.KeyMetaUp = "\x1b[1;9A"
		t.KeyMetaDown = "\x1b[1;9B"
		t.KeyMetaRight = "\x1b[1;9C"
		t.KeyMetaLeft = "\x1b[1;9D"
		t.KeyAltUp = "\x1b[1;3A"
		t.KeyAltDown = "\x1b[1;3B"
		t.KeyAltRight = "\x1b[1;3C"
		t.KeyAltLeft = "\x1b[1;3D"
		t.KeyCtrlUp = "\x1b[1;5A"
		t.KeyCtrlDown = "\x1b[1;5B"
		t.KeyCtrlRight = "\x1b[1;5C"
		t.KeyCtrlLeft = "\x1b[1;5D"
		t.KeyAltShfUp = "\x1b[1;4A"
		t.KeyAltShfDown = "\x1b[1;4B"
		t.KeyAltShfRight = "\x1b[1;4C"
		t.KeyAltShfLeft = "\x1b[1;4D"

		t.KeyMetaShfUp = "\x1b[1;10A"
		t.KeyMetaShfDown = "\x1b[1;10B"
		t.KeyMetaShfRight = "\x1b[1;10C"
		t.KeyMetaShfLeft = "\x1b[1;10D"

		t.KeyCtrlShfUp = "\x1b[1;6A"
		t.KeyCtrlShfDown = "\x1b[1;6B"
		t.KeyCtrlShfRight = "\x1b[1;6C"
		t.KeyCtrlShfLeft = "\x1b[1;6D"
	}
	// And also for Home and End
	if t.KeyShfHome == "\x1b[1;2H" && t.KeyShfEnd == "\x1b[1;2F" {
		t.KeyCtrlHome = "\x1b[1;5H"
		t.KeyCtrlEnd = "\x1b[1;5F"
		t.KeyAltHome = "\x1b[1;9H"
		t.KeyAltEnd = "\x1b[1;9F"
		t.KeyCtrlShfHome = "\x1b[1;6H"
		t.KeyCtrlShfEnd = "\x1b[1;6F"
		t.KeyAltShfHome = "\x1b[1;4H"
		t.KeyAltShfEnd = "\x1b[1;4F"
		t.KeyMetaShfHome = "\x1b[1;10H"
		t.KeyMetaShfEnd = "\x1b[1;10F"
	}

	// And the same thing for rxvt and workalikes (Eterm, aterm, etc.)
	// It seems that urxvt at least send ESC as ALT prefix for these,
	// although some places seem to indicate a separate ALT key sesquence.
	if t.KeyShfRight == "\x1b[c" && t.KeyShfLeft == "\x1b[d" {
		t.KeyShfUp = "\x1b[a"
		t.KeyShfDown = "\x1b[b"
		t.KeyCtrlUp = "\x1b[Oa"
		t.KeyCtrlDown = "\x1b[Ob"
		t.KeyCtrlRight = "\x1b[Oc"
		t.KeyCtrlLeft = "\x1b[Od"
	}
	if t.KeyShfHome == "\x1b[7$" && t.KeyShfEnd == "\x1b[8$" {
		t.KeyCtrlHome = "\x1b[7^"
		t.KeyCtrlEnd = "\x1b[8^"
	}

	// If the kmous entry is present, then we need to record the
	// the codes to enter and exit mouse mode.  Sadly, this is not
	// part of the terminfo databases anywhere that I've found, but
	// is an extension.  The escape codes are documented in the XTerm
	// manual, and all terminals that have kmous are expected to
	// use these same codes, unless explicitly configured otherwise
	// vi XM.  Note that in any event, we only known how to parse either
	// x11 or SGR mouse events -- if your terminal doesn't support one
	// of these two forms, you maybe out of luck.
	t.MouseMode = tc.getstr("XM")
	if t.Mouse != "" && t.MouseMode == "" {
		// we anticipate that all xterm mouse tracking compatible
		// terminals understand mouse tracking (1000), but we hope
		// that those that don't understand any-event tracking (1003)
		// will at least ignore it.  Likewise we hope that terminals
		// that don't understand SGR reporting (1006) just ignore it.
		t.MouseMode = "%?%p1%{1}%=%t%'h'%Pa%e%'l'%Pa%;" +
			"\x1b[?1000%ga%c\x1b[?1002%ga%c\x1b[?1003%ga%c\x1b[?1006%ga%c"
	}

	// We only support colors in ANSI 8 or 256 color mode.
	if t.Colors < 8 || t.SetFg == "" {
		t.Colors = 0
	}
	if t.SetCursor == "" {
		return nil, "", errors.New("terminal not cursor addressable")
	}

	// For padding, we lookup the pad char.  If that isn't present,
	// and npc is *not* set, then we assume a null byte.
	t.PadChar = tc.getstr("pad")
	if t.PadChar == "" {
		if !tc.getflag("npc") {
			t.PadChar = "\u0000"
		}
	}

	// For some terminals we fabricate a -truecolor entry, that may
	// not exist in terminfo.
	if addTrueColor {
		t.SetFgRGB = "\x1b[38;2;%p1%d;%p2%d;%p3%dm"
		t.SetBgRGB = "\x1b[48;2;%p1%d;%p2%d;%p3%dm"
		t.SetFgBgRGB = "\x1b[38;2;%p1%d;%p2%d;%p3%d;" +
			"48;2;%p4%d;%p5%d;%p6%dm"
	}

	// For terminals that use "standard" SGR sequences, lets combine the
	// foreground and background together.
	if strings.HasPrefix(t.SetFg, "\x1b[") &&
		strings.HasPrefix(t.SetBg, "\x1b[") &&
		strings.HasSuffix(t.SetFg, "m") &&
		strings.HasSuffix(t.SetBg, "m") {
		fg := t.SetFg[:len(t.SetFg)-1]
		r := regexp.MustCompile("%p1")
		bg := r.ReplaceAllString(t.SetBg[2:], "%p2")
		t.SetFgBg = fg + ";" + bg
	}

	return t, tc.desc, nil
}

func dotGoAddInt(w io.Writer, n string, i int) {
	if i == 0 {
		// initialized to 0, ignore
		return
	}
	fmt.Fprintf(w, "\t\t%-13s %d,\n", n+":", i)
}
func dotGoAddStr(w io.Writer, n string, s string) {
	if s == "" {
		return
	}
	fmt.Fprintf(w, "\t\t%-13s %q,\n", n+":", s)
}

func dotGoAddArr(w io.Writer, n string, a []string) {
	if len(a) == 0 {
		return
	}
	fmt.Fprintf(w, "\t\t%-13s []string{", n+":")
	did := false
	for _, b := range a {
		if did {
			fmt.Fprint(w, ", ")
		}
		did = true
		fmt.Fprintf(w, "%q", b)
	}
	fmt.Fprintln(w, "},")
}

func dotGoHeader(w io.Writer, packname string) {
	fmt.Fprintln(w, "// Generated automatically.  DO NOT HAND-EDIT.")
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "package %s\n", packname)
	fmt.Fprintln(w, "")
}

func dotGoTrailer(w io.Writer) {
}

func dotGoInfo(w io.Writer, t *terminfo.Terminfo, desc string) {

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "func init() {")
	fmt.Fprintf(w, "\t// %s\n", desc)
	fmt.Fprintln(w, "\tAddTerminfo(&Terminfo{")
	dotGoAddStr(w, "Name", t.Name)
	dotGoAddArr(w, "Aliases", t.Aliases)
	dotGoAddInt(w, "Columns", t.Columns)
	dotGoAddInt(w, "Lines", t.Lines)
	dotGoAddInt(w, "Colors", t.Colors)
	dotGoAddStr(w, "Bell", t.Bell)
	dotGoAddStr(w, "Clear", t.Clear)
	dotGoAddStr(w, "EnterCA", t.EnterCA)
	dotGoAddStr(w, "ExitCA", t.ExitCA)
	dotGoAddStr(w, "ShowCursor", t.ShowCursor)
	dotGoAddStr(w, "HideCursor", t.HideCursor)
	dotGoAddStr(w, "AttrOff", t.AttrOff)
	dotGoAddStr(w, "Underline", t.Underline)
	dotGoAddStr(w, "Bold", t.Bold)
	dotGoAddStr(w, "Dim", t.Dim)
	dotGoAddStr(w, "Blink", t.Blink)
	dotGoAddStr(w, "Reverse", t.Reverse)
	dotGoAddStr(w, "EnterKeypad", t.EnterKeypad)
	dotGoAddStr(w, "ExitKeypad", t.ExitKeypad)
	dotGoAddStr(w, "SetFg", t.SetFg)
	dotGoAddStr(w, "SetBg", t.SetBg)
	dotGoAddStr(w, "SetFgBg", t.SetFgBg)
	dotGoAddStr(w, "PadChar", t.PadChar)
	dotGoAddStr(w, "AltChars", t.AltChars)
	dotGoAddStr(w, "EnterAcs", t.EnterAcs)
	dotGoAddStr(w, "ExitAcs", t.ExitAcs)
	dotGoAddStr(w, "EnableAcs", t.EnableAcs)
	dotGoAddStr(w, "SetFgRGB", t.SetFgRGB)
	dotGoAddStr(w, "SetBgRGB", t.SetBgRGB)
	dotGoAddStr(w, "SetFgBgRGB", t.SetFgBgRGB)
	dotGoAddStr(w, "Mouse", t.Mouse)
	dotGoAddStr(w, "MouseMode", t.MouseMode)
	dotGoAddStr(w, "SetCursor", t.SetCursor)
	dotGoAddStr(w, "CursorBack1", t.CursorBack1)
	dotGoAddStr(w, "CursorUp1", t.CursorUp1)
	dotGoAddStr(w, "KeyUp", t.KeyUp)
	dotGoAddStr(w, "KeyDown", t.KeyDown)
	dotGoAddStr(w, "KeyRight", t.KeyRight)
	dotGoAddStr(w, "KeyLeft", t.KeyLeft)
	dotGoAddStr(w, "KeyInsert", t.KeyInsert)
	dotGoAddStr(w, "KeyDelete", t.KeyDelete)
	dotGoAddStr(w, "KeyBackspace", t.KeyBackspace)
	dotGoAddStr(w, "KeyHome", t.KeyHome)
	dotGoAddStr(w, "KeyEnd", t.KeyEnd)
	dotGoAddStr(w, "KeyPgUp", t.KeyPgUp)
	dotGoAddStr(w, "KeyPgDn", t.KeyPgDn)
	dotGoAddStr(w, "KeyF1", t.KeyF1)
	dotGoAddStr(w, "KeyF2", t.KeyF2)
	dotGoAddStr(w, "KeyF3", t.KeyF3)
	dotGoAddStr(w, "KeyF4", t.KeyF4)
	dotGoAddStr(w, "KeyF5", t.KeyF5)
	dotGoAddStr(w, "KeyF6", t.KeyF6)
	dotGoAddStr(w, "KeyF7", t.KeyF7)
	dotGoAddStr(w, "KeyF8", t.KeyF8)
	dotGoAddStr(w, "KeyF9", t.KeyF9)
	dotGoAddStr(w, "KeyF10", t.KeyF10)
	dotGoAddStr(w, "KeyF11", t.KeyF11)
	dotGoAddStr(w, "KeyF12", t.KeyF12)
	dotGoAddStr(w, "KeyF13", t.KeyF13)
	dotGoAddStr(w, "KeyF14", t.KeyF14)
	dotGoAddStr(w, "KeyF15", t.KeyF15)
	dotGoAddStr(w, "KeyF16", t.KeyF16)
	dotGoAddStr(w, "KeyF17", t.KeyF17)
	dotGoAddStr(w, "KeyF18", t.KeyF18)
	dotGoAddStr(w, "KeyF19", t.KeyF19)
	dotGoAddStr(w, "KeyF20", t.KeyF20)
	dotGoAddStr(w, "KeyF21", t.KeyF21)
	dotGoAddStr(w, "KeyF22", t.KeyF22)
	dotGoAddStr(w, "KeyF23", t.KeyF23)
	dotGoAddStr(w, "KeyF24", t.KeyF24)
	dotGoAddStr(w, "KeyF25", t.KeyF25)
	dotGoAddStr(w, "KeyF26", t.KeyF26)
	dotGoAddStr(w, "KeyF27", t.KeyF27)
	dotGoAddStr(w, "KeyF28", t.KeyF28)
	dotGoAddStr(w, "KeyF29", t.KeyF29)
	dotGoAddStr(w, "KeyF30", t.KeyF30)
	dotGoAddStr(w, "KeyF31", t.KeyF31)
	dotGoAddStr(w, "KeyF32", t.KeyF32)
	dotGoAddStr(w, "KeyF33", t.KeyF33)
	dotGoAddStr(w, "KeyF34", t.KeyF34)
	dotGoAddStr(w, "KeyF35", t.KeyF35)
	dotGoAddStr(w, "KeyF36", t.KeyF36)
	dotGoAddStr(w, "KeyF37", t.KeyF37)
	dotGoAddStr(w, "KeyF38", t.KeyF38)
	dotGoAddStr(w, "KeyF39", t.KeyF39)
	dotGoAddStr(w, "KeyF40", t.KeyF40)
	dotGoAddStr(w, "KeyF41", t.KeyF41)
	dotGoAddStr(w, "KeyF42", t.KeyF42)
	dotGoAddStr(w, "KeyF43", t.KeyF43)
	dotGoAddStr(w, "KeyF44", t.KeyF44)
	dotGoAddStr(w, "KeyF45", t.KeyF45)
	dotGoAddStr(w, "KeyF46", t.KeyF46)
	dotGoAddStr(w, "KeyF47", t.KeyF47)
	dotGoAddStr(w, "KeyF48", t.KeyF48)
	dotGoAddStr(w, "KeyF49", t.KeyF49)
	dotGoAddStr(w, "KeyF50", t.KeyF50)
	dotGoAddStr(w, "KeyF51", t.KeyF51)
	dotGoAddStr(w, "KeyF52", t.KeyF52)
	dotGoAddStr(w, "KeyF53", t.KeyF53)
	dotGoAddStr(w, "KeyF54", t.KeyF54)
	dotGoAddStr(w, "KeyF55", t.KeyF55)
	dotGoAddStr(w, "KeyF56", t.KeyF56)
	dotGoAddStr(w, "KeyF57", t.KeyF57)
	dotGoAddStr(w, "KeyF58", t.KeyF58)
	dotGoAddStr(w, "KeyF59", t.KeyF59)
	dotGoAddStr(w, "KeyF60", t.KeyF60)
	dotGoAddStr(w, "KeyF61", t.KeyF61)
	dotGoAddStr(w, "KeyF62", t.KeyF62)
	dotGoAddStr(w, "KeyF63", t.KeyF63)
	dotGoAddStr(w, "KeyF64", t.KeyF64)
	dotGoAddStr(w, "KeyCancel", t.KeyCancel)
	dotGoAddStr(w, "KeyPrint", t.KeyPrint)
	dotGoAddStr(w, "KeyExit", t.KeyExit)
	dotGoAddStr(w, "KeyHelp", t.KeyHelp)
	dotGoAddStr(w, "KeyClear", t.KeyClear)
	dotGoAddStr(w, "KeyBacktab", t.KeyBacktab)
	dotGoAddStr(w, "KeyShfLeft", t.KeyShfLeft)
	dotGoAddStr(w, "KeyShfRight", t.KeyShfRight)
	dotGoAddStr(w, "KeyShfUp", t.KeyShfUp)
	dotGoAddStr(w, "KeyShfDown", t.KeyShfDown)
	dotGoAddStr(w, "KeyCtrlLeft", t.KeyCtrlLeft)
	dotGoAddStr(w, "KeyCtrlRight", t.KeyCtrlRight)
	dotGoAddStr(w, "KeyCtrlUp", t.KeyCtrlUp)
	dotGoAddStr(w, "KeyCtrlDown", t.KeyCtrlDown)
	dotGoAddStr(w, "KeyMetaLeft", t.KeyMetaLeft)
	dotGoAddStr(w, "KeyMetaRight", t.KeyMetaRight)
	dotGoAddStr(w, "KeyMetaUp", t.KeyMetaUp)
	dotGoAddStr(w, "KeyMetaDown", t.KeyMetaDown)
	dotGoAddStr(w, "KeyAltLeft", t.KeyAltLeft)
	dotGoAddStr(w, "KeyAltRight", t.KeyAltRight)
	dotGoAddStr(w, "KeyAltUp", t.KeyAltUp)
	dotGoAddStr(w, "KeyAltDown", t.KeyAltDown)
	dotGoAddStr(w, "KeyAltShfLeft", t.KeyAltShfLeft)
	dotGoAddStr(w, "KeyAltShfRight", t.KeyAltShfRight)
	dotGoAddStr(w, "KeyAltShfUp", t.KeyAltShfUp)
	dotGoAddStr(w, "KeyAltShfDown", t.KeyAltShfDown)
	dotGoAddStr(w, "KeyMetaShfLeft", t.KeyMetaShfLeft)
	dotGoAddStr(w, "KeyMetaShfRight", t.KeyMetaShfRight)
	dotGoAddStr(w, "KeyMetaShfUp", t.KeyMetaShfUp)
	dotGoAddStr(w, "KeyMetaShfDown", t.KeyMetaShfDown)
	dotGoAddStr(w, "KeyCtrlShfLeft", t.KeyCtrlShfLeft)
	dotGoAddStr(w, "KeyCtrlShfRight", t.KeyCtrlShfRight)
	dotGoAddStr(w, "KeyCtrlShfUp", t.KeyCtrlShfUp)
	dotGoAddStr(w, "KeyCtrlShfDown", t.KeyCtrlShfDown)
	dotGoAddStr(w, "KeyShfHome", t.KeyShfHome)
	dotGoAddStr(w, "KeyShfEnd", t.KeyShfEnd)
	dotGoAddStr(w, "KeyCtrlHome", t.KeyCtrlHome)
	dotGoAddStr(w, "KeyCtrlEnd", t.KeyCtrlEnd)
	dotGoAddStr(w, "KeyMetaHome", t.KeyMetaHome)
	dotGoAddStr(w, "KeyMetaEnd", t.KeyMetaEnd)
	dotGoAddStr(w, "KeyAltHome", t.KeyAltHome)
	dotGoAddStr(w, "KeyAltEnd", t.KeyAltEnd)
	dotGoAddStr(w, "KeyCtrlShfHome", t.KeyCtrlShfHome)
	dotGoAddStr(w, "KeyCtrlShfEnd", t.KeyCtrlShfEnd)
	dotGoAddStr(w, "KeyMetaShfHome", t.KeyMetaShfHome)
	dotGoAddStr(w, "KeyMetaShfEnd", t.KeyMetaShfEnd)
	dotGoAddStr(w, "KeyAltShfHome", t.KeyAltShfHome)
	dotGoAddStr(w, "KeyAltShfEnd", t.KeyAltShfEnd)
	fmt.Fprintln(w, "\t})")
	fmt.Fprintln(w, "}")
}

func main() {
	gofile := ""
	jsonfile := ""
	packname := "terminfo"
	nofatal := false
	quiet := false
	dogzip := false

	flag.StringVar(&gofile, "go", "", "generate go source in named file")
	flag.StringVar(&jsonfile, "json", "", "generate json in named file")
	flag.StringVar(&packname, "P", packname, "package name (go source)")
	flag.BoolVar(&nofatal, "nofatal", false, "errors are not fatal")
	flag.BoolVar(&quiet, "quiet", false, "suppress error messages")
	flag.BoolVar(&dogzip, "gzip", false, "compress json output")
	flag.Parse()
	var e error
	js := []byte{}

	args := flag.Args()
	if len(args) == 0 {
		args = []string{os.Getenv("TERM")}
	}

	tdata := make(map[string]*terminfo.Terminfo)
	descs := make(map[string]string)

	for _, term := range args {
		if t, desc, e := getinfo(term); e != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr,
					"Failed loading %s: %v\n", term, e)
			}
			if !nofatal {
				os.Exit(1)
			}
		} else {
			tdata[term] = t
			descs[term] = desc
		}
	}

	if len(tdata) == 0 {
		// No data.
		os.Exit(0)
	}
	if gofile != "" {
		w := os.Stdout
		if gofile != "-" {
			if w, e = os.Create(gofile); e != nil {
				fmt.Fprintf(os.Stderr, "Failed: %v", e)
				os.Exit(1)
			}
		}
		dotGoHeader(w, packname)
		for term, t := range tdata {
			if t.Name == term {
				dotGoInfo(w, t, descs[term])
			}
		}
		dotGoTrailer(w)
		if w != os.Stdout {
			w.Close()
		}
	} else {
		o := os.Stdout
		if jsonfile != "-" && jsonfile != "" {
			if o, e = os.Create(jsonfile); e != nil {
				fmt.Fprintf(os.Stderr, "Failed: %v", e)
			}
		}
		var w io.WriteCloser
		w = o
		if dogzip {
			w = gzip.NewWriter(o)
		}
		for _, term := range args {
			if t := tdata[term]; t != nil {
				js, e = json.Marshal(t)
				fmt.Fprintln(w, string(js))
			}
			// arguably if there is more than one term, this
			// should be a javascript array, but that's not how
			// we load it.  We marshal objects one at a time from
			// the file.
		}
		if e != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v", e)
			os.Exit(1)
		}
		w.Close()
		if w != o {
			o.Close()
		}
	}
}
