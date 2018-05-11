package tview

import (
	"github.com/gdamore/tcell"
)

//Cursor represents the path to an item in the TreeView
type Cursor []int

//treeItem represents one item in the TreeView
type treeItem struct {
	Text     string      //text of the treeItem
	parent   *treeItem   //pointer to the parent item
	expanded bool        //if the item is expanded and it's children are visible
	children []*treeItem //children of the item
}

//NextSibling retuns the next sibling of the item
//if the item is the last child of it's parent retuns nil
func (item *treeItem) NextSibling() *treeItem {
	if item.parent == nil {
		return nil
	}
	itemIndex := 0
	for _, child := range item.parent.children {
		if item == child {
			break
		}
		itemIndex++
	}
	if itemIndex == len(item.parent.children)-1 {
		return nil
	}
	return item.parent.children[itemIndex+1]
}

//PreviousSibling returns the previous sibling of the item
//if the item is the first child of it's parent retuns nil
func (item *treeItem) PreviousSibling() *treeItem {
	if item.parent == nil {
		return nil
	}
	itemIndex := 0
	for _, child := range item.parent.children {
		if item == child {
			break
		}
		itemIndex++
	}
	if itemIndex == 0 {
		return nil
	}
	return item.parent.children[itemIndex-1]
}

//TreeView displays an tree of items which are expandable
type TreeView struct {
	*Box

	//the root item (isn't displayed)
	root treeItem

	//the selected item
	selected *treeItem
}

//NewTreeView creates a new TreeView
func NewTreeView() *TreeView {
	return &TreeView{
		Box: NewBox(),
	}
}

//selectNext selects the next item
func (tv *TreeView) selectNext(item *treeItem) *treeItem {
	sibling := item.NextSibling()
	if sibling != nil {
		tv.selected = sibling
		return sibling
	}
	if item.parent == nil {
		tv.selected = item
		return item
	}
	c := tv.selectNext(item.parent)
	if c == nil || c == &tv.root {
		tv.selected = item
	}
	return c
}

//selectPrev selects the previous item
func (tv *TreeView) selectPrev(item *treeItem) *treeItem {
	sibling := item.PreviousSibling()
	if sibling != nil {
		if sibling.expanded && len(sibling.children) > 0 {
			tv.selected = sibling.children[len(sibling.children)-1]
			return tv.selected
		}
		tv.selected = sibling
		return tv.selected
	}
	if item.parent == nil || item.parent == &tv.root {
		tv.selected = item
		return tv.selected
	}
	tv.selected = item.parent
	return item.parent
}

// InputHandler returns the handler for this primitive.
func (tv *TreeView) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return tv.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		switch key := event.Key(); key {
		case tcell.KeyDown:
			if tv.selected == nil && len(tv.root.children) > 0 {
				tv.selected = tv.root.children[0]
				return
			}
			if tv.selected.expanded && len(tv.selected.children) > 0 {
				tv.selected = tv.selected.children[0]
				return
			}
			tv.selectNext(tv.selected)
		case tcell.KeyUp:
			if tv.selected == nil && len(tv.root.children) > 0 {
				tv.selected = tv.root.children[0]
				return
			}
			tv.selectPrev(tv.selected)
		case tcell.KeyRune:
			if event.Rune() != ' ' {
				return
			}
			if tv.selected == nil {
				return
			}
			tv.selected.expanded = !tv.selected.expanded
		}
	})
}

//AddItem appends a new item to the an item with the parent cursor
//returns a cursor to the new item
func (tv *TreeView) AddItem(parentCur Cursor, text string) Cursor {
	parent := tv.getItem(parentCur)
	parent.children = append(parent.children, &treeItem{parent: parent, Text: text})
	return append(Cursor(nil), append(parentCur, len(parent.children)-1)...)
}

//RemoveItem removes an item from TreeView
//panics if the cursor is out of range
func (tv *TreeView) RemoveItem(cur Cursor) {
	if len(cur) == 0 {
		panic("empty cursor")
	}
	item := tv.getItem(cur)
	itemIndex := cur[len(cur)-1]
	item.parent.children = append(item.parent.children[:itemIndex], item.parent.children[itemIndex+1:]...)
}

//getItem return treeItem pointer from a cursor
func (tv *TreeView) getItem(cur Cursor) *treeItem {
	item := &tv.root
	for _, i := range cur {
		item = item.children[i]
	}
	return item
}

//ToggleItem toggles the expansion of children of an item defined by cursor
//panics if the cursor is out of range
func (tv *TreeView) ToggleItem(cur Cursor) {
	if len(cur) == 0 {
		panic("empty cursor")
	}
	item := tv.getItem(cur)
	item.expanded = !item.expanded
}

//SetItemText set's an items text
//panics if the cursor is out of range
func (tv *TreeView) SetItemText(cur Cursor, text string) {
	if len(cur) == 0 {
		panic("empty cursor")
	}
	item := tv.getItem(cur)
	item.Text = text
}

//Draw draws this primitive onto the screen.
func (tv *TreeView) Draw(screen tcell.Screen) {
	tv.Box.Draw(screen)
	x, y, width, height := tv.GetInnerRect()
	for _, i := range tv.root.children {
		tv.drawItem(i, screen, x, &y, width, height)
	}
}

//drawItem draws a single item and it's children
func (tv *TreeView) drawItem(item *treeItem, screen tcell.Screen, x int, y *int, width, height int) {
	switch {
	case len(item.children) == 0:
		Print(screen, ">", x, *y, width, AlignLeft, Styles.PrimaryTextColor)
	case item.expanded:
		Print(screen, "-", x, *y, width, AlignLeft, Styles.PrimaryTextColor)
	case !item.expanded:
		Print(screen, "+", x, *y, width, AlignLeft, Styles.PrimaryTextColor)
	}
	if tv.selected == item {
		Print(screen, item.Text, x+2, *y, width, AlignLeft, Styles.SecondaryTextColor)
	} else {
		Print(screen, item.Text, x+2, *y, width, AlignLeft, Styles.PrimaryTextColor)
	}
	*y++
	if !item.expanded {
		return
	}
	bottomLimit := *y + height
	for _, i := range item.children {
		if *y >= bottomLimit {
			break
		}
		tv.drawItem(i, screen, x+2, y, width-3, height)
	}
}
