package tview

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gdamore/tcell"
)

const (
	randomDataSize        = 512
	appendSetupWriteCount = 64 // Write randomData 64 times before appending
)

var (
	randomData        = generateRandomData()
	textViewTestCases = generateTestCases()
)

type textViewTestCase struct {
	app      bool
	color    bool
	region   bool
	scroll   bool
	wrap     bool
	wordwrap bool
}

func (c *textViewTestCase) String() string {
	return fmt.Sprintf("App=%c/Color=%c/Region=%c/Scroll=%c/Wrap=%c/WordWrap=%c", cl(c.app), cl(c.color), cl(c.region), cl(c.scroll), cl(c.wrap), cl(c.wordwrap))
}

func TestTextViewWrite(t *testing.T) {
	t.Parallel()

	for _, c := range textViewTestCases {
		c := c // Capture

		t.Run(c.String(), func(t *testing.T) {
			t.Parallel()

			var (
				tv, _, err = testTextView(c)
				n          int
			)
			if err != nil {
				t.Error(err)
			}

			if c.app {
				err = prepareAppendTextView(tv)
				if err != nil {
					t.Error(err)
				}
			}

			n, err = tv.Write(randomData)
			if err != nil {
				t.Errorf("failed to write (successfully wrote %d) bytes: %s", n, err)
			} else if n != randomDataSize {
				t.Errorf("failed to write: expected to write %d bytes, wrote %d", randomDataSize, n)
			}

			tv.Clear()
		})
	}
}

func BenchmarkTextViewWrite(b *testing.B) {
	for _, c := range textViewTestCases {
		c := c // Capture

		b.Run(c.String(), func(b *testing.B) {
			var (
				tv, _, err = testTextView(c)
				n          int
			)
			if err != nil {
				b.Error(err)
			}

			if c.app {
				err = prepareAppendTextView(tv)
				if err != nil {
					b.Error(err)
				}
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				n, err = tv.Write(randomData)
				if err != nil {
					b.Errorf("failed to write (successfully wrote %d) bytes: %s", n, err)
				} else if n != randomDataSize {
					b.Errorf("failed to write: expected to write %d bytes, wrote %d", randomDataSize, n)
				}

				if !c.app {
					b.StopTimer()
					tv.Clear()
					b.StartTimer()
				}
			}
		})
	}
}

func TestTextViewDraw(t *testing.T) {
	t.Parallel()

	for _, c := range textViewTestCases {
		c := c // Capture

		t.Run(c.String(), func(t *testing.T) {
			t.Parallel()

			var (
				tv, sc, err = testTextView(c)
				n           int
			)
			if err != nil {
				t.Error(err)
			}

			if c.app {
				err = prepareAppendTextView(tv)
				if err != nil {
					t.Error(err)
				}

				tv.Draw(sc)
			}

			n, err = tv.Write(randomData)
			if err != nil {
				t.Errorf("failed to write (successfully wrote %d) bytes: %s", n, err)
			} else if n != randomDataSize {
				t.Errorf("failed to write: expected to write %d bytes, wrote %d", randomDataSize, n)
			}

			tv.Draw(sc)
		})
	}
}

func BenchmarkTextViewDraw(b *testing.B) {
	for _, c := range textViewTestCases {
		c := c // Capture

		b.Run(c.String(), func(b *testing.B) {
			var (
				tv, sc, err = testTextView(c)
				n           int
			)
			if err != nil {
				b.Error(err)
			}

			if c.app {
				err = prepareAppendTextView(tv)
				if err != nil {
					b.Error(err)
				}

				tv.Draw(sc)
			}

			n, err = tv.Write(randomData)
			if err != nil {
				b.Errorf("failed to write (successfully wrote %d) bytes: %s", n, err)
			} else if n != randomDataSize {
				b.Errorf("failed to write: expected to write %d bytes, wrote %d", randomDataSize, n)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tv.Draw(sc)
			}
		})
	}
}

func generateTestCases() []*textViewTestCase {
	var cases []*textViewTestCase

	app := false
	for i := 0; i < 2; i++ {
		if i == 1 {
			app = true
		}

		color := false
		for i := 0; i < 2; i++ {
			if i == 1 {
				color = true
			}

			region := false
			for i := 0; i < 2; i++ {
				if i == 1 {
					region = true
				}

				scroll := false
				for i := 0; i < 2; i++ {
					if i == 1 {
						scroll = true
					}

					wrap := false
					for i := 0; i < 2; i++ {
						if i == 1 {
							wrap = true
						}

						wordwrap := false
						for i := 0; i < 2; i++ {
							if i == 1 {
								if !wrap {
									// WordWrap requires Wrap
									continue
								}

								wordwrap = true
							}

							cases = append(cases, &textViewTestCase{app, color, region, scroll, wrap, wordwrap})
						}
					}
				}
			}
		}
	}

	return cases
}

func generateRandomData() []byte {
	var (
		b bytes.Buffer
		r = 33
	)

	for i := 0; i < randomDataSize; i++ {
		if i%80 == 0 && i <= 160 {
			b.WriteRune('\n')
		} else if i%7 == 0 {
			b.WriteRune(' ')
		} else {
			b.WriteRune(rune(r))
		}

		r++
		if r == 127 {
			r = 33
		}
	}

	return b.Bytes()
}

func tvc(tv *TextView, c *textViewTestCase) *TextView {
	return tv.SetDynamicColors(c.color).SetRegions(c.region).SetScrollable(c.scroll).SetWrap(c.wrap).SetWordWrap(c.wordwrap)
}

func cl(v bool) rune {
	if v {
		return 'T'
	}

	return 'F'
}

func testTextView(c *textViewTestCase) (*TextView, tcell.Screen, error) {
	tv := NewTextView()

	sc := tcell.NewSimulationScreen("UTF-8")
	sc.SetSize(80, 24)

	_ = NewApplication().SetRoot(tv, true).SetScreen(sc)

	return tvc(tv, c), sc, nil
}

func prepareAppendTextView(t *TextView) error {
	for i := 0; i < appendSetupWriteCount; i++ {
		n, err := t.Write(randomData)
		if err != nil {
			return fmt.Errorf("failed to write (successfully wrote %d) bytes: %s", n, err)
		} else if n != randomDataSize {
			return fmt.Errorf("failed to write: expected to write %d bytes, wrote %d", randomDataSize, n)
		}
	}

	return nil
}
