// Generated automatically.  DO NOT HAND-EDIT.

package terminfo

func init() {
	// Sun Microsystems Inc. workstation console
	AddTerminfo(&Terminfo{
		Name:         "sun",
		Aliases:      []string{"sun1", "sun2"},
		Columns:      80,
		Lines:        34,
		Bell:         "\a",
		Clear:        "\f",
		AttrOff:      "\x1b[m",
		Reverse:      "\x1b[7m",
		PadChar:      "\x00",
		SetCursor:    "\x1b[%i%p1%d;%p2%dH",
		CursorBack1:  "\b",
		CursorUp1:    "\x1b[A",
		KeyUp:        "\x1b[A",
		KeyDown:      "\x1b[B",
		KeyRight:     "\x1b[C",
		KeyLeft:      "\x1b[D",
		KeyDelete:    "177",
		KeyBackspace: "\b",
		KeyHome:      "\x1b[214z",
		KeyEnd:       "\x1b[220z",
		KeyPgUp:      "\x1b[216z",
		KeyPgDn:      "\x1b[222z",
		KeyF1:        "\x1b[224z",
		KeyF2:        "\x1b[225z",
		KeyF3:        "\x1b[226z",
		KeyF4:        "\x1b[227z",
		KeyF5:        "\x1b[228z",
		KeyF6:        "\x1b[229z",
		KeyF7:        "\x1b[230z",
		KeyF8:        "\x1b[231z",
		KeyF9:        "\x1b[232z",
		KeyF10:       "\x1b[233z",
		KeyF11:       "\x1b[234z",
		KeyF12:       "\x1b[235z",
	})
}
