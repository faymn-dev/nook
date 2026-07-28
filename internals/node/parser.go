package node

import (
	"fmt"
	"strings"
)

func Parse(tokens []Token) (Renderer, error) {
	p := &parser{
		Stream: Stream[Token]{
			data: tokens,
		},
		document: NewHTMLNode("div", nil),
	}

	for p.HasData() {
		current := p.Current()

		switch current.Variant {
		case TokenEOF:
		case TokenNewline:
		case TokenHeading:
			textContent, err := p.expectTokenVariant(TokenString)
			if err != nil {
				return nil, err
			}

			level := len(current.Value)
			if level > 7 {
				return nil, fmt.Errorf("headings cannot be greater than level 7 %s", p.location())
			}

			p.appendChild(NewHTMLNode(fmt.Sprintf("h%d", level), nil, TextNode(strings.TrimSpace(textContent.Value))))
		default:
			p.addToParagraph(current)
		}

		p.Consume()
	}

	return p.document, nil
}

type parser struct {
	Stream[Token]

	document  *HTMLNode
	paragraph []Token

	line   int
	column int
}

func (p *parser) appendChild(node Renderer) {
	if len(p.paragraph) > 0 {
		var sb strings.Builder
		for _, token := range p.paragraph {
			sb.WriteString(token.String())
		}

		textContent := sb.String()
		p.document.Children = append(p.document.Children, NewHTMLNode("p", nil, TextNode(textContent)))
		if strings.HasSuffix(textContent, "  ") {
			p.appendChild(NewHTMLNode("br", nil))
		}

		p.paragraph = []Token{}
	}

	p.document.Children = append(p.document.Children, node)
}

func (p *parser) addToParagraph(token Token) {
	p.paragraph = append(p.paragraph, token)
}

func (p *parser) checkNewline(token Token) {
}

func (p *parser) Consume() Token {
	token := p.Current()
	if token.Variant == TokenNewline {
		p.line++
		p.column = 0
	} else {
		p.column++
	}

	p.Next()
	return p.Current()
}

func (p *parser) isNewline() bool {
	return p.column == 0
}

func (p *parser) expectTokenVariant(variant TokenVariant) (Token, error) {
	actualToken := p.Current()
	if actualToken.Variant == variant {
		p.Next()
		return actualToken, nil
	}
	return Token{}, fmt.Errorf("unexpected token %v %s", actualToken, p.location())
}

func (p *parser) location() string {
	return fmt.Sprintf("at line %d, column %d", p.line, p.column)
}
