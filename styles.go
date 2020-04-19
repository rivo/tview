package tview

import "github.com/gdamore/tcell"

// Theme defines the colors used when primitives are initialized.
type Theme struct {
	// Title, border and other lines
	TitleColor    tcell.Color // Box titles.
	BorderColor   tcell.Color // Box borders.
	GraphicsColor tcell.Color // Graphics.

	// Text
	PrimaryTextColor           tcell.Color // Primary text.
	SecondaryTextColor         tcell.Color // Secondary text (e.g. labels).
	TertiaryTextColor          tcell.Color // Tertiary text (e.g. subtitles, notes).
	InverseTextColor           tcell.Color // Text on primary-colored backgrounds.
	ContrastSecondaryTextColor tcell.Color // Secondary text on ContrastBackgroundColor-colored backgrounds.

	// Background
	PrimitiveBackgroundColor    tcell.Color // Main background color for primitives.
	ContrastBackgroundColor     tcell.Color // Background color for contrasting elements.
	MoreContrastBackgroundColor tcell.Color // Background color for even more contrasting elements.

	// Context menu
	ContextMenuPaddingTop    int // Top padding.
	ContextMenuPaddingBottom int // Bottom padding.
	ContextMenuPaddingLeft   int // Left padding.
	ContextMenuPaddingRight  int // Right padding.
}

// Styles defines the theme for applications. The default is for a black
// background and some basic colors: black, white, yellow, green, cyan, and
// blue.
var Styles = Theme{
	TitleColor:    tcell.ColorWhite,
	BorderColor:   tcell.ColorWhite,
	GraphicsColor: tcell.ColorWhite,

	PrimaryTextColor:           tcell.ColorWhite,
	SecondaryTextColor:         tcell.ColorYellow,
	TertiaryTextColor:          tcell.ColorGreen,
	InverseTextColor:           tcell.ColorBlue,
	ContrastSecondaryTextColor: tcell.ColorDarkCyan,

	PrimitiveBackgroundColor:    tcell.ColorBlack,
	ContrastBackgroundColor:     tcell.ColorBlue,
	MoreContrastBackgroundColor: tcell.ColorGreen,

	ContextMenuPaddingTop:    0,
	ContextMenuPaddingBottom: 0,
	ContextMenuPaddingLeft:   1,
	ContextMenuPaddingRight:  1,
}
