package tview

import "github.com/gdamore/tcell"

// page represents one page of a Pages object.
type page struct {
	Name    string    // The page's name.
	Item    Primitive // The page's primitive.
	Visible bool      // Whether or not this page is visible.
}

// Pages is a container for other primitives often used as the application's
// root primitive. It allows to easily switch the visibility of the contained
// primitives.
type Pages struct {
	*Box

	// The contained pages.
	pages []*page

	// An optional handler which is called whenever the visibility or the order of
	// pages changes.
	changed func()
}

// NewPages returns a new Pages object.
func NewPages() *Pages {
	p := &Pages{
		Box: NewBox(),
	}
	p.focus = p
	return p
}

// SetChangedFunc sets a handler which is called whenever the visibility or the
// order of any visible pages changes. This can be used to redraw the pages.
func (p *Pages) SetChangedFunc(handler func()) *Pages {
	p.changed = handler
	return p
}

// AddPage adds a new page with the given name and primitive. Leaving the name
// empty or using the same name for multiple items may cause conflicts in other
// functions.
//
// Visible pages will be drawn in the order they were added (unless that order
// was changed in one of the other functions).
func (p *Pages) AddPage(name string, item Primitive, visible bool) *Pages {
	p.pages = append(p.pages, &page{Item: item, Name: name, Visible: true})
	if p.changed != nil {
		p.changed()
	}
	return p
}

// RemovePage removes the page with the given name.
func (p *Pages) RemovePage(name string) *Pages {
	for index, page := range p.pages {
		if page.Name == name {
			p.pages = append(p.pages[:index], p.pages[index+1:]...)
			if page.Visible && p.changed != nil {
				p.changed()
			}
			break
		}
	}
	return p
}

// ShowPage sets a page's visibility to "true" (in addition to any other pages
// which are already visible).
func (p *Pages) ShowPage(name string) *Pages {
	for _, page := range p.pages {
		if page.Name == name {
			page.Visible = true
			if p.changed != nil {
				p.changed()
			}
			break
		}
	}
	return p
}

// HidePage sets a page's visibility to "false".
func (p *Pages) HidePage(name string) *Pages {
	for _, page := range p.pages {
		if page.Name == name {
			page.Visible = false
			if p.changed != nil {
				p.changed()
			}
			break
		}
	}
	return p
}

// SwitchToPage sets a page's visibility to "true" and all other pages'
// visibility to "false".
func (p *Pages) SwitchToPage(name string) *Pages {
	for _, page := range p.pages {
		if page.Name == name {
			page.Visible = true
		} else {
			page.Visible = false
		}
	}
	if p.changed != nil {
		p.changed()
	}
	return p
}

// SendToFront changes the order of the pages such that the page with the given
// name comes last, causing it to be drawn last with the next update (if
// visible).
func (p *Pages) SendToFront(name string) *Pages {
	for index, page := range p.pages {
		if page.Name == name {
			if index < len(p.pages)-1 {
				p.pages = append(append(p.pages[:index], p.pages[index+1:]...), page)
			}
			if page.Visible && p.changed != nil {
				p.changed()
			}
			break
		}
	}
	return p
}

// SendToBack changes the order of the pages such that the page with the given
// name comes first, causing it to be drawn first with the next update (if
// visible).
func (p *Pages) SendToBack(name string) *Pages {
	for index, pg := range p.pages {
		if pg.Name == name {
			if index > 0 {
				p.pages = append(append([]*page{pg}, p.pages[:index]...), p.pages[index+1:]...)
			}
			if pg.Visible && p.changed != nil {
				p.changed()
			}
			break
		}
	}
	return p
}

// HasFocus returns whether or not this primitive has focus.
func (p *Pages) HasFocus() bool {
	for _, page := range p.pages {
		if page.Item.GetFocusable().HasFocus() {
			return true
		}
	}
	return false
}

// Draw draws this primitive onto the screen.
func (p *Pages) Draw(screen tcell.Screen) {
	for _, page := range p.pages {
		if !page.Visible {
			continue
		}
		page.Item.Draw(screen)
	}
}
