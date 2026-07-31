package node

import (
	"strings"
)

const DocumentTagName = "document"

type Document struct {
	head []Renderer
	body []Renderer
}

func NewDocument(head []Renderer, body []Renderer) *Document {
	return &Document{head: head, body: body}
}

func (d *Document) ToHTML() string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>")
	sb.WriteString("<html lang=\"en\">")
	sb.WriteString("<head>")
	for _, child := range d.head {
		sb.WriteString(child.ToHTML())
	}
	sb.WriteString("</head>")
	sb.WriteString("<body>")
	for _, child := range d.body {
		sb.WriteString(child.ToHTML())
	}
	sb.WriteString("</body>")
	sb.WriteString("</html>")
	return sb.String()
}

func (d *Document) GetTagName() string {
	return DocumentTagName
}

func (d *Document) GetChildren() []Renderer {
	return d.body
}
