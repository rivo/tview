package tview

import (
	"testing"
)

func Test_GetBox(t *testing.T) {
	for _, tt := range []struct {
		name   string
		getBox func() (BoxedPrimitive, *Box)
	}{
		{name: "Box", getBox: func() (BoxedPrimitive, *Box) {
			box := NewBox()
			return box, box
		}},
		{name: "Button", getBox: func() (BoxedPrimitive, *Box) {
			p := NewButton("Click me")
			return p, p.Box
		}},
		{name: "Checkbox", getBox: func() (BoxedPrimitive, *Box) {
			p := NewCheckbox()
			return p, p.Box
		}},
		{name: "DropDown", getBox: func() (BoxedPrimitive, *Box) {
			p := NewDropDown()
			return p, p.Box
		}},
		{name: "Flex", getBox: func() (BoxedPrimitive, *Box) {
			p := NewFlex()
			return p, p.Box
		}},
		{name: "Form", getBox: func() (BoxedPrimitive, *Box) {
			p := NewForm()
			return p, p.Box
		}},
		{name: "Frame", getBox: func() (BoxedPrimitive, *Box) {
			p := NewFrame(NewTextView())
			return p, p.Box
		}},
		{name: "Grid", getBox: func() (BoxedPrimitive, *Box) {
			p := NewGrid()
			return p, p.Box
		}},
		{name: "Image", getBox: func() (BoxedPrimitive, *Box) {
			p := NewImage()
			return p, p.Box
		}},
		{name: "InputField", getBox: func() (BoxedPrimitive, *Box) {
			p := NewInputField()
			return p, p.Box
		}},
		{name: "List", getBox: func() (BoxedPrimitive, *Box) {
			p := NewList()
			return p, p.Box
		}},
		{name: "Modal", getBox: func() (BoxedPrimitive, *Box) {
			p := NewModal()
			return p, p.Box
		}},
		{name: "Pages", getBox: func() (BoxedPrimitive, *Box) {
			p := NewPages()
			return p, p.Box
		}},
		{name: "Table", getBox: func() (BoxedPrimitive, *Box) {
			p := NewTable()
			return p, p.Box
		}},
		{name: "TextArea()", getBox: func() (BoxedPrimitive, *Box) {
			p := NewTextArea()
			return p, p.Box
		}},
		{name: "TextView", getBox: func() (BoxedPrimitive, *Box) {
			p := NewTextView()
			return p, p.Box
		}},
		{name: "TreeView", getBox: func() (BoxedPrimitive, *Box) {
			p := NewTextView()
			return p, p.Box
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			primitive, box := tt.getBox()
			actualBox := primitive.GetBox()
			if actualBox != box {
				t.Errorf("GetBox() got = %+v\nwant %+v", actualBox, box)
			}
		})
	}
}
