// Package treeprint provides a simple ASCII tree composing tool.
package treeprint

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// NodeVisitor function type for iterating over nodes
type NodeVisitor func(item *Node)

// Tree represents a tree structure with leaf-nodes and branch-nodes.
type Tree interface {
	// AddNode adds a new Node to a branch.
	AddNode(a any) Tree
	// AddNodef adds a new Node to a branch.
	AddNodef(format string, a ...any) Tree
	// AddMetaNode adds a new Node with meta value provided to a branch.
	AddMetaNode(meta any, a any) Tree
	// AddMetaNodef adds a new Node with meta value provided to a branch.
	AddMetaNodef(meta any, format string, a ...any) Tree
	// AddBranch adds a new branch Node (a level deeper).
	AddBranch(a any) Tree
	// AddBranchf adds a new branch Node (a level deeper).
	AddBranchf(format string, a ...any) Tree
	// AddMetaBranch adds a new branch Node (a level deeper) with meta value provided.
	AddMetaBranch(meta any, a any) Tree
	// AddMetaBranchf adds a new branch Node (a level deeper) with meta value provided.
	AddMetaBranchf(meta any, format string, a ...any) Tree
	// Branch converts a leaf-Node to a branch-Node,
	// applying this on a branch-Node does no effect.
	Branch() Tree
	// FindByMeta finds a Node whose meta value matches the provided one by reflect.DeepEqual,
	// returns nil if not found.
	FindByMeta(meta any) Tree
	// FindByValue finds a Node whose value matches the provided one by reflect.DeepEqual,
	// returns nil if not found.
	FindByValue(value any) Tree
	//  returns the last Node of a tree
	FindLastNode() Tree
	// String renders the tree or subtree as a string.
	String() string
	// Bytes renders the tree or subtree as byteslice.
	Bytes() []byte
	// Writer renders the tree or subtree as byteslice.
	Writer(w io.Writer)

	SetValue(value any)
	SetValuef(format string, a ...any)

	SetMetaValue(meta any)

	// VisitAll iterates over the tree, branches and nodes.
	// If need to iterate over the whole tree, use the root Node.
	// Note this method uses a breadth-first approach.
	VisitAll(fn NodeVisitor)
}

type Node struct {
	Root  *Node
	Meta  any
	Value any
	Nodes []*Node
}

func (n *Node) FindLastNode() Tree {
	var ns = n.Nodes
	if len(ns) == 0 {
		return nil
	}
	return ns[len(ns)-1]
}

func (n *Node) AddNode(a any) Tree {
	n.Nodes = append(n.Nodes, &Node{
		Root:  n,
		Value: a,
	})
	return n
}

func (n *Node) AddNodef(format string, a ...any) Tree {
	return n.AddNode(fmt.Sprintf(format, a...))
}

func (n *Node) AddMetaNode(meta any, a any) Tree {
	n.Nodes = append(n.Nodes, &Node{
		Root:  n,
		Meta:  meta,
		Value: a,
	})
	return n
}

func (n *Node) AddMetaNodef(meta any, format string, a ...any) Tree {
	return n.AddMetaNode(meta, fmt.Sprintf(format, a...))
}

func (n *Node) AddBranch(a any) Tree {
	var branch = &Node{
		Root:  n,
		Value: a,
	}
	n.Nodes = append(n.Nodes, branch)
	return branch
}

func (n *Node) AddBranchf(format string, a ...any) Tree {
	return n.AddBranch(fmt.Sprintf(format, a...))
}

func (n *Node) AddMetaBranch(meta any, a any) Tree {
	var branch = &Node{
		Root:  n,
		Meta:  meta,
		Value: a,
	}
	n.Nodes = append(n.Nodes, branch)
	return branch
}

func (n *Node) AddMetaBranchf(meta any, format string, a ...any) Tree {
	return n.AddMetaBranch(meta, fmt.Sprintf(format, a...))
}

func (n *Node) Branch() Tree {
	n.Root = nil
	return n
}

func (n *Node) FindByMeta(meta any) Tree {
	for _, node := range n.Nodes {
		if reflect.DeepEqual(node.Meta, meta) {
			return node
		}
		if v := node.FindByMeta(meta); v != nil {
			return v
		}
	}
	return nil
}

func (n *Node) FindByValue(value any) Tree {
	for _, node := range n.Nodes {
		if reflect.DeepEqual(node.Value, value) {
			return node
		}
		if v := node.FindByMeta(value); v != nil {
			return v
		}
	}
	return nil
}

func (n *Node) Writer(w io.Writer) {
	var level = 0
	var levelsEnded []int
	if n.Root == nil {
		if n.Meta != nil {
			fmt.Fprintf(w, "[%v]  %v", n.Meta, n.Value)
		} else {
			fmt.Fprintf(w, "%v", n.Value)
		}
		io.WriteString(w, "\n")
	} else {
		var edge = EdgeTypeMid
		if len(n.Nodes) == 0 {
			edge = EdgeTypeEnd
			levelsEnded = append(levelsEnded, level)
		}
		printValues(w, 0, levelsEnded, edge, n)
	}
	if len(n.Nodes) > 0 {
		printNodes(w, level, levelsEnded, n.Nodes)
	}
}

func (n *Node) Bytes() []byte {
	var buf = &bytes.Buffer{}
	n.Writer(buf)
	return buf.Bytes()
}

func (n *Node) String() string {
	return string(n.Bytes())
}

func (n *Node) SetValue(value any) {
	n.Value = value
}

func (n *Node) SetValuef(format string, a ...any) {
	n.Value = fmt.Sprintf(format, a...)
}

func (n *Node) SetMetaValue(meta any) {
	n.Meta = meta
}

func (n *Node) VisitAll(fn NodeVisitor) {
	for _, node := range n.Nodes {
		fn(node)

		if len(node.Nodes) > 0 {
			node.VisitAll(fn)
			continue
		}
	}
}

func printNodes(wr io.Writer, level int, levelsEnded []int, nodes []*Node) {
	for i, node := range nodes {
		var edge = EdgeTypeMid
		if i == len(nodes)-1 {
			levelsEnded = append(levelsEnded, level)
			edge = EdgeTypeEnd
		}
		printValues(wr, level, levelsEnded, edge, node)
		if len(node.Nodes) > 0 {
			printNodes(wr, level+1, levelsEnded, node.Nodes)
		}
	}
}

func printValues(wr io.Writer, level int, levelsEnded []int, edge EdgeType, node *Node) {
	for i := 0; i < level; i++ {
		if isEnded(levelsEnded, i) {
			fmt.Fprint(wr, strings.Repeat(" ", IndentSize+1))
			continue
		}
		fmt.Fprintf(wr, "%s%s", EdgeTypeLink, strings.Repeat(" ", IndentSize))
	}

	var val = renderValue(level, node)
	var meta = node.Meta

	if meta != nil {
		fmt.Fprintf(wr, "%s [%v]  %v\n", edge, meta, val)
		return
	}
	fmt.Fprintf(wr, "%s %v\n", edge, val)
}

func isEnded(levelsEnded []int, level int) bool {
	for _, l := range levelsEnded {
		if l == level {
			return true
		}
	}
	return false
}

func renderValue(level int, node *Node) any {
	var lines = strings.Split(fmt.Sprintf("%v", node.Value), "\n")

	// If value does not contain multiple lines, return itself.
	if len(lines) < 2 {
		return node.Value
	}

	// If value contains multiple lines,
	// generate a padding and prefix each line with it.
	var pad = padding(level, node)

	for i := 1; i < len(lines); i++ {
		lines[i] = fmt.Sprintf("%s%s", pad, lines[i])
	}

	return strings.Join(lines, "\n")
}

// padding returns a padding for the multiline values with correctly placed link edges.
// It is generated by traversing the tree upwards (from leaf to the root of the tree)
// and, on each level, checking if the Node the last one of its siblings.
// If a Node is the last one, the padding on that level should be empty (there's nothing to link to below it).
// If a Node is not the last one, the padding on that level should be the link edge so the sibling below is correctly connected.
func padding(level int, node *Node) string {
	var links = make([]string, level+1)

	for node.Root != nil {
		if isLast(node) {
			links[level] = strings.Repeat(" ", IndentSize+1)
		} else {
			links[level] = fmt.Sprintf("%s%s", EdgeTypeLink, strings.Repeat(" ", IndentSize))
		}
		level--
		node = node.Root
	}

	return strings.Join(links, "")
}

// isLast checks if the Node is the last one in the slice of its parent children
func isLast(n *Node) bool {
	return n == n.Root.FindLastNode()
}

type EdgeType string

var (
	EdgeTypeLink EdgeType = "│"
	EdgeTypeMid  EdgeType = "├─"
	EdgeTypeEnd  EdgeType = "└─"
)

// IndentSize is the number of spaces per tree level.
var IndentSize = 3

// New Generates new tree
func New() Tree {
	return &Node{Value: "."}
}

// NewWithRoot Generates new tree with the given root value
func NewWithRoot(root any) Tree {
	return &Node{Value: root}
}

// NewWithRoot Generates new tree with the given root value
func NewWithRootf(format string, a ...any) Tree {
	return &Node{Value: fmt.Sprintf(format, a...)}
}
