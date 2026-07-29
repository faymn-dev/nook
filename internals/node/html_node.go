package node

import (
	"fmt"
	"strings"
)

type HTMLProps map[string]string

type HTMLNode struct {
	Tag      string
	Props    HTMLProps
	Children []Renderer
}

const FragmentTag = "fragment"

func NewHTMLNode(tag string, props HTMLProps, children ...Renderer) *HTMLNode {
	return &HTMLNode{
		Tag:      tag,
		Props:    props,
		Children: children,
	}
}

func NewHTMLFragment(children ...Renderer) *HTMLNode {
	return &HTMLNode{
		Tag:      FragmentTag,
		Children: children,
	}
}

func (h *HTMLNode) ToHTML() string {
	if h == nil {
		return ""
	}

	var sb strings.Builder

	if h.Tag != FragmentTag {
		fmt.Fprintf(&sb, "<%s", h.Tag)
		if h.Props != nil {
			for key, value := range h.Props {
				fmt.Fprintf(&sb, " %s=\"%s\"", key, value)
			}
		}
		sb.WriteString(">")
	}

	for _, child := range h.Children {
		sb.WriteString(child.ToHTML())
	}

	if h.Tag != FragmentTag {
		fmt.Fprintf(&sb, "</%s>", h.Tag)
	}

	return sb.String()
}

func (h *HTMLNode) GetChildren() []Renderer {
	return h.Children
}
