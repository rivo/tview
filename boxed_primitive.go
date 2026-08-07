package tview

// BoxedPrimitive defines a primitive with a Box
// It's implemented by Box and any type that has *Box field
type BoxedPrimitive interface {
	Primitive
	GetBox() *Box
}
