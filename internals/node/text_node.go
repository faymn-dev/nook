package node

type TextNode string

func (t TextNode) ToHTML() string {
	return string(t)
}

func (t TextNode) GetTagName() string {
	return ""
}

func (t TextNode) GetChildren() []Renderer {
	return nil
}
