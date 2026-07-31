package node

import (
	"strings"
)

const DocumentTagName = "document"

type Document struct {
	children []Renderer
}

func NewDocument(children ...Renderer) *Document {
	return &Document{children: children}
}

func (d *Document) ToHTML() string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>")
	sb.WriteString("<html lang=\"en\">")
	for _, child := range d.children {
		sb.WriteString(child.ToHTML())
	}
	sb.WriteString("</html>")

	return sb.String()
}

func (d *Document) GetTagName() string {
	return DocumentTagName
}

func (d *Document) GetChildren() []Renderer {
	return d.children
}
