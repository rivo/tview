package tview

import "github.com/gdamore/tcell"

// Styles defines various colors used when primitives are initialized. These
// may be changed to accommodate a different look and feel.
//
// The default is for applications with a black background and basic colors:
// black, white, yellow, green, and blue.
var Styles = struct {
	ModalBackgroundColor        tcell.Color
	LabelTextColor              tcell.Color
	FieldBackgroundColor        tcell.Color
	FieldTextColor              tcell.Color
	FieldDisableBackgroundColor tcell.Color
	FieldDisableTextColor       tcell.Color
	ButtonBackgroundColor       tcell.Color
	ButtonTextColor             tcell.Color
	PrimitiveBackgroundColor    tcell.Color // Main background color for primitives.
	ContrastBackgroundColor     tcell.Color // Background color for contrasting elements.
	MoreContrastBackgroundColor tcell.Color // Background color for even more contrasting elements.
	BorderColor                 tcell.Color // Box borders.
	TitleColor                  tcell.Color // Box titles.
	GraphicsColor               tcell.Color // Graphics.
	PrimaryTextColor            tcell.Color // Primary text.
	SecondaryTextColor          tcell.Color // Secondary text (e.g. labels).
	TertiaryTextColor           tcell.Color // Tertiary text (e.g. subtitles, notes).
	InverseTextColor            tcell.Color // Text on primary-colored backgrounds.
	ContrastSecondaryTextColor  tcell.Color // Secondary text on ContrastBackgroundColor-colored backgrounds.

	// Semigraphical runes.
	GraphicsHoriBar             rune
	GraphicsVertBar             rune
	GraphicsTopLeftCorner       rune
	GraphicsTopRightCorner      rune
	GraphicsBottomLeftCorner    rune
	GraphicsBottomRightCorner   rune
	GraphicsLeftT               rune
	GraphicsRightT              rune
	GraphicsTopT                rune
	GraphicsBottomT             rune
	GraphicsCross               rune
	GraphicsDbVertBar           rune
	GraphicsDbHorBar            rune
	GraphicsDbTopLeftCorner     rune
	GraphicsDbTopRightCorner    rune
	GraphicsDbBottomRightCorner rune
	GraphicsDbBottomLeftCorner  rune
	GraphicsEllipsis            rune

	GraphicsRadioChecked   string
	GraphicsRadioUnchecked string

	GraphicsCheckboxChecked   string
	GraphicsCheckboxUnchecked string
}{
	ModalBackgroundColor:        tcell.ColorBlack,
	LabelTextColor:              tcell.ColorWhite,
	FieldBackgroundColor:        tcell.ColorGrey,
	FieldTextColor:              tcell.ColorBlack,
	FieldDisableBackgroundColor: tcell.ColorBlack,
	FieldDisableTextColor:       tcell.ColorWhite,
	ButtonBackgroundColor:       tcell.ColorBlack,
	ButtonTextColor:             tcell.ColorWhite,
	PrimitiveBackgroundColor:    tcell.ColorBlack,
	ContrastBackgroundColor:     tcell.ColorBlue,
	MoreContrastBackgroundColor: tcell.ColorGreen,
	BorderColor:                 tcell.ColorWhite,
	TitleColor:                  tcell.ColorWhite,
	GraphicsColor:               tcell.ColorWhite,
	PrimaryTextColor:            tcell.ColorWhite,
	SecondaryTextColor:          tcell.ColorYellow,
	TertiaryTextColor:           tcell.ColorGreen,
	InverseTextColor:            tcell.ColorBlue,
	ContrastSecondaryTextColor:  tcell.ColorDarkCyan,

	GraphicsHoriBar:             '\u2500',
	GraphicsVertBar:             '\u2502',
	GraphicsTopLeftCorner:       '\u250c',
	GraphicsTopRightCorner:      '\u2510',
	GraphicsBottomLeftCorner:    '\u2514',
	GraphicsBottomRightCorner:   '\u2518',
	GraphicsLeftT:               '\u251c',
	GraphicsRightT:              '\u2524',
	GraphicsTopT:                '\u252c',
	GraphicsBottomT:             '\u2534',
	GraphicsCross:               '\u253c',
	GraphicsDbVertBar:           '\u2550',
	GraphicsDbHorBar:            '\u2551',
	GraphicsDbTopLeftCorner:     '\u2554',
	GraphicsDbTopRightCorner:    '\u2557',
	GraphicsDbBottomRightCorner: '\u255d',
	GraphicsDbBottomLeftCorner:  '\u255a',
	GraphicsEllipsis:            '\u2026',

	GraphicsRadioChecked:   "\u25c9",
	GraphicsRadioUnchecked: "\u25ef",

	GraphicsCheckboxChecked:   "X",
	GraphicsCheckboxUnchecked: " ",
}
