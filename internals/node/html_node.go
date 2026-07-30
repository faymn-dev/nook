package node

import (
	"fmt"
	"slices"
	"strings"
)

type Renderer interface {
	ToHTML() string
	GetTagName() string
	GetChildren() []Renderer
}

const FragmentTagName = "fragment"

var selfTerminates = []string{"img", "br", "hr"}

type HTMLProps map[string]string

type HTMLNode struct {
	Tag      string
	Props    HTMLProps
	Children []Renderer
}

func NewHTMLNode(tag string, props HTMLProps, children ...Renderer) *HTMLNode {
	return &HTMLNode{
		Tag:      tag,
		Props:    props,
		Children: children,
	}
}

func NewHTMLFragment(children ...Renderer) *HTMLNode {
	return &HTMLNode{
		Tag:      FragmentTagName,
		Children: children,
	}
}

func (h *HTMLNode) ToHTML() string {
	if h == nil {
		return ""
	}

	var sb strings.Builder

	if h.Tag != FragmentTagName {
		fmt.Fprintf(&sb, "<%s", h.Tag)
		if h.Props != nil {
			keys := getKeys(h.Props)
			slices.Sort(keys)
			for _, key := range keys {
				fmt.Fprintf(&sb, " %s=\"%s\"", key, h.Props[key])
			}
		}

		if slices.Index(selfTerminates, h.Tag) != -1 { // self terminates
			sb.WriteString(" />")
			return sb.String()
		} else {
			sb.WriteString(">")
		}
	}

	for _, child := range h.Children {
		sb.WriteString(child.ToHTML())
	}

	if h.Tag != FragmentTagName {
		fmt.Fprintf(&sb, "</%s>", h.Tag)
	}

	return sb.String()
}

func (h *HTMLNode) GetTagName() string {
	return h.Tag
}

func (h *HTMLNode) GetChildren() []Renderer {
	return h.Children
}

func getKeys(m map[string]string) []string {
	result := make([]string, len(m))
	i := 0
	for key := range m {
		result[i] = key
		i++
	}
	return result
}
