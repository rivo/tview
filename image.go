package tview

import (
	"image"
	"math"
)

// Types of dithering applied to images.
const (
	ImageDitheringNone           = iota // No dithering.
	ImageDitheringThreshold             // Grey scale thresholding at 50%.
	ImageDitheringFloydSteinberg        // Floyd-Steinberg dithering (the default).
)

// The number of colors supported by true color terminals (R*G*B = 256*256*256).
const TrueColor = 16777216

// Image implements a widget that displays one image. The original image
// (specified with [SetImage]) is resized according to the widget's size (see
// [SetSize]), using the colors available in the terminal (see [SetColors]),
// applying dithering if necessary (see [SetDithering]).
//
// Images are approximated by graphical characters in the terminal. The
// resolution is therefore limited by the number of characters that can be drawn
// in the terminal and the colors available in the terminal.
//
// Don't rely on the exact pixels drawn by this widget. The image drawing
// algorithm may change in the future to improve the appearance of the image.
type Image struct {
	*Box

	// The image to be displayed. If nil, the widget will be empty.
	image image.Image

	// The size of the image. If a value is 0, the corresponding size is chosen
	// automatically based on the other size while preserving the image's aspect
	// ratio. If both are 0, the image uses as much space as possible. A
	// negative value represents a percentage, e.g. -50 means 50% of the
	// available space.
	width, height int

	// The number of colors to use. If 0, the number of colors is chosen based
	// on the terminal's capabilities.
	colors int

	// The dithering algorithm to use, one of the constants starting with
	// "ImageDithering".
	dithering int

	// The background color to use (RGB) for transparent pixels.
	backgroundColor [3]int8

	// The width of a terminal's cell divided by its height.
	aspectRatio float64

	// Horizontal and vertical alignment, one of the "Align" constants.
	alignHorizontal, alignVertical int

	// The actual image size (in cells) when it was drawn the last time.
	lastWidth, lastHeight int

	// The actual image (in cells) when it was drawn the last time. The size of
	// this slice is 4 * lastWidth * lastHeight (with a factor of 4 because we
	// can draw four pixels per cell), indexed by y*lastWidth*2 + x. Each pixel
	// is an RGB value (0-255).
	pixels [][3]int
}

// NewImage returns a new image widget with an empty image (use [SetImage] to
// specify the image to be displayed). The image will use the widget's entire
// available space. The dithering algorithm is set to Floyd-Steinberg dithering.
// The terminal's cell aspect ratio is set to 1.
func NewImage() *Image {
	return &Image{
		Box:             NewBox(),
		dithering:       ImageDitheringFloydSteinberg,
		aspectRatio:     1,
		alignHorizontal: AlignCenter,
		alignVertical:   AlignCenter,
	}
}

// SetImage sets the image to be displayed. If nil, the widget will be empty.
func (i *Image) SetImage(image image.Image) *Image {
	i.image = image
	i.lastWidth, i.lastHeight = 0, 0
	return i
}

// SetSize sets the size of the image. Positive values refer to cells in the
// terminal. Negative values refer to a percentage of the available space (e.g.
// -50 means 50%). A value of 0 means that the corresponding size is chosen
// automatically based on the other size while preserving the image's aspect
// ratio. If both are 0, the image uses as much space as possible while still
// preserving the aspect ratio.
func (i *Image) SetSize(rows, columns int) *Image {
	i.width = columns
	i.height = rows
	return i
}

// SetColors sets the number of colors to use. This should be the number of
// colors supported by the terminal. If 0, the number of colors is chosen based
// on the $TERM environment variable (which may or may not be reliable).
//
// Only the values 0, 2, 8, 256, and 16777216 ([TrueColor]) are supported. Other
// values will be rounded up to the next supported value, to a maximum of
// 16777216.
//
// The effect of using more colors than supported by the terminal is undefined.
func (i *Image) SetColors(colors int) *Image {
	i.colors = colors
	i.lastWidth, i.lastHeight = 0, 0
	return i
}

// SetDithering sets the dithering algorithm to use, one of the constants
// starting with "ImageDithering", for example [ImageDitheringFloydSteinberg].
func (i *Image) SetDithering(dithering int) *Image {
	i.dithering = dithering
	i.lastWidth, i.lastHeight = 0, 0
	return i
}

// SetBackgroundColor sets the background color to use (RGB) for transparent
// pixels in the original image. The default is black (0, 0, 0).
func (i *Image) SetBackgroundColor(r, g, b int8) *Image {
	i.backgroundColor = [3]int8{r, g, b}
	i.lastWidth, i.lastHeight = 0, 0
	return i
}

// SetAspectRatio sets the width of a terminal's cell divided by its height.
// You may change the default of 1 if your terminal uses a different aspect
// ratio. This is used to calculate the size of the image if one of the sizes
// is 0. The function will panic if the aspect ratio is 0 or less.
func (i *Image) SetAspectRatio(aspectRatio float64) *Image {
	if aspectRatio <= 0 {
		panic("aspect ratio must be greater than 0")
	}
	i.aspectRatio = aspectRatio
	i.lastWidth, i.lastHeight = 0, 0
	return i
}

// SetAlign sets the vertical and horizontal alignment of the image within the
// widget's space. The possible values are [AlignTop], [AlignCenter], and
// [AlignBottom] for vertical alignment and [AlignLeft], [AlignCenter], and
// [AlignRight] for horizontal alignment. The default is [AlignCenter] for both.
func (i *Image) SetAlign(vertical, horizontal int) *Image {
	i.alignHorizontal = horizontal
	i.alignVertical = vertical
	return i
}

// render re-populates the [Image.pixels] slice besed on the current settings,
// if [Image.lastWidth] and [Image.lastHeight] don't match the current image's
// size. It also sets the new image size in these two variables.
func (i *Image) render() {
	// If there is no image, there are no pixels.
	if i.image == nil {
		i.pixels = nil
		return
	}

	// Calculate the new (terminal-space) image size.
	bounds := i.image.Bounds()
	imageWidth, imageHeight := bounds.Dx(), bounds.Dy()
	if i.aspectRatio != 1.0 {
		imageWidth = int(float64(imageWidth) / i.aspectRatio)
	}
	width, height := i.width, i.height
	_, _, innerWidth, innerHeight := i.GetInnerRect()
	if width == 0 && height == 0 {
		// Use all available space.
		width, height = innerWidth, innerHeight
		if adjustedWidth := imageWidth * height / imageHeight; adjustedWidth < width {
			width = adjustedWidth
		} else {
			height = imageHeight * width / imageWidth
		}
	} else {
		// Turn percentages into absolute values.
		if width < 0 {
			width = innerWidth * -width / 100
		}
		if height < 0 {
			height = innerHeight * -height / 100
		}
		if width == 0 {
			// Adjust the width.
			width = imageWidth * height / imageHeight
		} else if height == 0 {
			// Adjust the height.
			height = imageHeight * width / imageWidth
		}
	}
	if width <= 0 || height <= 0 {
		i.pixels = nil
		return
	}

	// If nothing has changed, we're done.
	if i.lastWidth == width && i.lastHeight == height {
		return
	}
	i.lastWidth, i.lastHeight = width, height // This could still be larger than the available space but that's ok for now.

	// Generate the initial pixels by resizing the image.
	i.resize()
}

// resize resizes the image to the current size and stores the result in
// [Image.pixels]. It is assumed that [Image.lastWidth] and [Image.lastHeight]
// are positive values.
func (i *Image) resize() {
	// Because most of the time, we will be downsizing the image, we don't even
	// attempt to do any fancy interpolation. For each target pixel, we
	// calculate a weighted average of the source pixels using their coverage
	// area.

	bounds := i.image.Bounds()
	srcWidth, srcHeight := bounds.Dx(), bounds.Dy()
	tgtWidth, tgtHeight := i.lastWidth*2, i.lastHeight*2
	coverageWidth, coverageHeight := float64(srcWidth)/float64(tgtWidth), float64(srcHeight)/float64(tgtHeight)
	i.pixels = make([][3]int, tgtWidth*tgtHeight)
	weights := make([]float64, tgtWidth*tgtHeight)
	for srcY := bounds.Min.Y; srcY < bounds.Max.Y; srcY++ {
		for srcX := bounds.Min.X; srcX < bounds.Max.X; srcX++ {
			r32, g32, b32, _ := i.image.At(srcX, srcY).RGBA()
			r, g, b := int(r32>>8), int(g32>>8), int(b32>>8)

			// Iterate over all target pixels. Outer loop is Y.
			startY := float64(srcY-bounds.Min.Y) * coverageHeight
			endY := startY + coverageHeight
			fromY, toY := int(startY), int(endY)
			for tgtY := fromY; tgtY <= toY && tgtY < tgtHeight; tgtY++ {
				coverageY := 1.0
				if tgtY == fromY {
					coverageY -= math.Mod(startY, 1.0)
				}
				if tgtY == toY {
					coverageY -= 1.0 - math.Mod(endY, 1.0)
				}

				// Inner loop is X.
				startX := float64(srcX-bounds.Min.X) * coverageWidth
				endX := startX + coverageWidth
				fromX, toX := int(startX), int(endX)
				for tgtX := fromX; tgtX <= toX && tgtX < tgtWidth; tgtX++ {
					coverageX := 1.0
					if tgtX == fromX {
						coverageX -= math.Mod(startX, 1.0)
					}
					if tgtX == toX {
						coverageX -= 1.0 - math.Mod(endX, 1.0)
					}

					// Add a weighted contribution to the target pixel.
					index := tgtY*tgtWidth + tgtX
					i.pixels[index][0] += r
					i.pixels[index][1] += g
					i.pixels[index][2] += b
					weights[index] += coverageX * coverageY
				}
			}
		}
	}

	// Normalize the pixels.
	for index, weight := range weights {
		if weight > 0 {
			i.pixels[index][0] = int(float64(i.pixels[index][0]) / weight)
			i.pixels[index][1] = int(float64(i.pixels[index][1]) / weight)
			i.pixels[index][2] = int(float64(i.pixels[index][2]) / weight)
		}
	}
}
