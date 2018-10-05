package ptool

import (
	"fmt"
	"strings"
)

// TNode -
type TNode struct {
	Type  int
	Links []*TNode
	Value string
	Data  interface{}
	Pos   TPos
}

// NewNode -
func NewNode(id int, val string, nodes ...*TNode) *TNode {
	return &TNode{Type: id, Value: val, Links: nodes}
}

// Depth -
func Depth(node *TNode) int {
	if node == nil || len(node.Links) == 0 {
		return 0
	}
	ret := 0
	for i := range node.Links {
		d := Depth(node.Links[i])
		if d > ret {
			ret = d
		}
	}
	return ret + 1
}

const (
	sCorner   = "└─"
	sThrough  = "├─"
	sVert     = "│ "
	sAlign    = "──"
	sDownLink = "┬─"
	sSpace    = "  "
)

func nodeToStr(node *TNode, links []string, current int, isLast bool, fn func(*TNode) string) string {
	for current > len(links)-1 {
		links = append(links, sAlign)
	}
	if isLast {
		links[current] = sCorner
	}
	if links[current] == sVert {
		links[current] = sThrough
	}
	for i := current + 1; i < len(links); i++ {
		links[i] = sAlign
	}
	l := len(node.Links)
	if l > 0 && len(links)-1 > current {
		links[current+1] = sDownLink
	}
	prefix := strings.Join(links, "")
	ret := fmt.Sprintf("%v%v\n", prefix, fn(node))
	links[current] = sVert
	if isLast {
		links[current] = sSpace
	}
	if len(links)-1 > current && links[current+1] == sDownLink {
		links[current+1] = sVert
	}
	for i := range node.Links {
		ret += nodeToStr(node.Links[i], links, current+1, i == l-1, fn)
	}
	return ret
}

// CustomString -
func (o *TNode) CustomString(fn func(*TNode) string) string {
	depth := 0
	depth = Depth(o)
	links := []string{}
	for i := 0; i < depth+1; i++ {
		links = append(links, sAlign)
	}
	return nodeToStr(o, links, 0, true, fn)
}

// String -
func (o *TNode) String() string {
	return o.CustomString(func(node *TNode) string {
		return fmt.Sprintf("%v %v %q", node.Pos, node.Type, node.Value)
	})
}
