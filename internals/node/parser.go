package node

func Parse(tokens []Token) (Renderer, error) {
	document := NewHTMLNode("div", nil)
	p := &parser{
		Stream: Stream[Token]{
			data: tokens,
		},
	}

	for p.HasData() {
		p.Next()
	}

	return document, nil
}

type parser struct {
	Stream[Token]

	line   int
	column int
}
