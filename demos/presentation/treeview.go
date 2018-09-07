package main

import (
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const treeAllCode = `[green]package[white] main

[green]import[white] [red]"github.com/rivo/tview"[white]

[green]func[white] [yellow]main[white]() {
	$$$

	root := tview.[yellow]NewTreeNode[white]([red]"Root"[white]).
		[yellow]AddChild[white](tview.[yellow]NewTreeNode[white]([red]"First child"[white]).
			[yellow]AddChild[white](tview.[yellow]NewTreeNode[white]([red]"Grandchild A"[white])).
			[yellow]AddChild[white](tview.[yellow]NewTreeNode[white]([red]"Grandchild B"[white]))).
		[yellow]AddChild[white](tview.[yellow]NewTreeNode[white]([red]"Second child"[white]).
			[yellow]AddChild[white](tview.[yellow]NewTreeNode[white]([red]"Grandchild C"[white])).
			[yellow]AddChild[white](tview.[yellow]NewTreeNode[white]([red]"Grandchild D"[white]))).
		[yellow]AddChild[white](tview.[yellow]NewTreeNode[white]([red]"Third child"[white]))

	tree.[yellow]SetRoot[white](root).
		[yellow]SetCurrentNode[white](root)

	tview.[yellow]NewApplication[white]().
		[yellow]SetRoot[white](tree, true).
		[yellow]Run[white]()
}`

const treeBasicCode = `tree := tview.[yellow]NewTreeView[white]()`

const treeTopLevelCode = `tree := tview.[yellow]NewTreeView[white]().
		[yellow]SetTopLevel[white]([red]1[white])`

const treeAlignCode = `tree := tview.[yellow]NewTreeView[white]().
		[yellow]SetAlign[white](true)`

const treePrefixCode = `tree := tview.[yellow]NewTreeView[white]().
		[yellow]SetGraphics[white](false).
		[yellow]SetTopLevel[white]([red]1[white]).
		[yellow]SetPrefixes[white]([][green]string[white]{
			[red]"[red[]* "[white],
			[red]"[darkcyan[]- "[white],
			[red]"[darkmagenta[]- "[white],
		})`

type node struct {
	text     string
	expand   bool
	selected func()
	children []*node
}

var (
	tree          = tview.NewTreeView()
	treeNextSlide func()
	treeCode      = tview.NewTextView().SetWrap(false).SetDynamicColors(true)
)

var rootNode = &node{
	text: "Root",
	children: []*node{
		{text: "Expand all", selected: func() { tree.GetRoot().ExpandAll() }},
		{text: "Collapse all", selected: func() {
			for _, child := range tree.GetRoot().GetChildren() {
				child.CollapseAll()
			}
		}},
		{text: "Hide root node", expand: true, children: []*node{
			{text: "Tree list starts one level down"},
			{text: "Works better for lists where no top node is needed"},
			{text: "Switch to this layout", selected: func() {
				tree.SetAlign(false).SetTopLevel(1).SetGraphics(true).SetPrefixes(nil)
				treeCode.SetText(strings.Replace(treeAllCode, "$$$", treeTopLevelCode, -1))
			}},
		}},
		{text: "Align node text", expand: true, children: []*node{
			{text: "For trees that are similar to lists"},
			{text: "Hierarchy shown only in line drawings"},
			{text: "Switch to this layout", selected: func() {
				tree.SetAlign(true).SetTopLevel(0).SetGraphics(true).SetPrefixes(nil)
				treeCode.SetText(strings.Replace(treeAllCode, "$$$", treeAlignCode, -1))
			}},
		}},
		{text: "Prefixes", expand: true, children: []*node{
			{text: "Best for hierarchical bullet point lists"},
			{text: "You can define your own prefixes per level"},
			{text: "Switch to this layout", selected: func() {
				tree.SetAlign(false).SetTopLevel(1).SetGraphics(false).SetPrefixes([]string{"[red]* ", "[darkcyan]- ", "[darkmagenta]- "})
				treeCode.SetText(strings.Replace(treeAllCode, "$$$", treePrefixCode, -1))
			}},
		}},
		{text: "Basic tree with graphics", expand: true, children: []*node{
			{text: "Lines illustrate hierarchy"},
			{text: "Basic indentation"},
			{text: "Switch to this layout", selected: func() {
				tree.SetAlign(false).SetTopLevel(0).SetGraphics(true).SetPrefixes(nil)
				treeCode.SetText(strings.Replace(treeAllCode, "$$$", treeBasicCode, -1))
			}},
		}},
		{text: "Next slide", selected: func() { treeNextSlide() }},
	}}

// TreeView demonstrates the tree view.
func TreeView(nextSlide func()) (title string, content tview.Primitive) {
	treeNextSlide = nextSlide
	tree.SetBorder(true).
		SetTitle("TreeView")

	// Add nodes.
	var add func(target *node) *tview.TreeNode
	add = func(target *node) *tview.TreeNode {
		node := tview.NewTreeNode(target.text).
			SetSelectable(target.expand || target.selected != nil).
			SetExpanded(target == rootNode).
			SetReference(target)
		if target.expand {
			node.SetColor(tcell.ColorGreen)
		} else if target.selected != nil {
			node.SetColor(tcell.ColorRed)
		}
		for _, child := range target.children {
			node.AddChild(add(child))
		}
		return node
	}
	root := add(rootNode)
	tree.SetRoot(root).
		SetCurrentNode(root).
		SetSelectedFunc(func(n *tview.TreeNode) {
			original := n.GetReference().(*node)
			if original.expand {
				n.SetExpanded(!n.IsExpanded())
			} else if original.selected != nil {
				original.selected()
			}
		})

	treeCode.SetText(strings.Replace(treeAllCode, "$$$", treeBasicCode, -1)).
		SetBorderPadding(1, 1, 2, 0)

	return "Tree", tview.NewFlex().
		AddItem(tree, 0, 1, true).
		AddItem(treeCode, codeWidth, 1, false)
}
