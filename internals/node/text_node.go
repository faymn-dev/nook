package node

type TextNode string

func (t TextNode) ToHTML() string {
	return string(t)
}
