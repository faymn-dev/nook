package node

import "fmt"

type TextNodeType int

const (
	TextNodePlain TextNodeType = iota
	TextNodeBold
	TextNodeItalic
	TextNodeCode
	TextNodeAnchor
	TextNodeImage
)

type TextNode struct {
	Type    TextNodeType
	Content string
	URL     string
}

type Option func(textNode *TextNode)

func NewTextNode(content string, options ...Option) *TextNode {
	textNode := &TextNode{
		Type:    TextNodePlain,
		Content: content,
	}

	for _, option := range options {
		option(textNode)
	}

	return textNode
}

func WithType(textNode *TextNode, textNodeType TextNodeType) {
	textNode.Type = textNodeType
}

func WithURL(textNode *TextNode, url string) {
	textNode.URL = url
}

// determine if two text nodes are equivalent
// note: nil and the empty string "" are considered the same in this scenario
func (t *TextNode) Equals(other *TextNode) bool {
	if t == nil || other == nil {
		return t == other
	}

	return t.Type == other.Type && t.Content == other.Content && t.URL == other.URL
}

func (t *TextNode) ToHTML() string {
	switch t.Type {
	case TextNodePlain:
		return t.Content
	case TextNodeBold:
		return fmt.Sprintf("<b>%s</b>", t.Content)
	case TextNodeItalic:
		return fmt.Sprintf("<i>%s</i>", t.Content)
	case TextNodeCode:
		return fmt.Sprintf("<code>%s</code>", t.Content)
	case TextNodeAnchor:
		return fmt.Sprintf("<a href=\"%s\">%s</a>", t.URL, t.Content)
	case TextNodeImage:
		return fmt.Sprintf("<img href=\"%s\" alt=\"%s\" />", t.URL, t.Content)
	}
	return ""
}
