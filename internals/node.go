package node

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

func NewTextNode(content string, nodeType TextNodeType, url string) *TextNode {
	return &TextNode{
		Type:    nodeType,
		Content: content,
		URL:     url,
	}
}

// determine if two text nodes are equivalent
// note: nil and the empty string "" are considered the same in this scenario
func (t *TextNode) Equals(other *TextNode) bool {
	if t == nil || other == nil {
		return t == other
	}

	return t.Type == other.Type && t.Content == other.Content && t.URL == other.URL
}
