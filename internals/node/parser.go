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
		document: NewHTMLFragment(),
	}

loop:
	for p.HasData() {
		current := p.Current()

		switch current.Variant {
		case TokenEOF:
			break loop
		case TokenNewline:
		case TokenHeading:
			if !p.isNewline() {
				p.addToParagraph(current)
			} else {
				textContent, err := p.expectNextToken(TokenString)
				if err != nil {
					return nil, err
				}
				level := len(current.Value)
				if level > 7 {
					return nil, fmt.Errorf("headings cannot be greater than level 7 %s", p.errorAt())
				}
				p.appendChild(NewHTMLNode(fmt.Sprintf("h%d", level), nil, TextNode(strings.TrimSpace(textContent.Value))))
			}
		case TokenCodeBlock:
			if !p.isNewline() {
				p.addToParagraph(current)
			} else {
				var language string
				if p.Peek().Variant == TokenString {
					language = p.Consume().Value
				}

				_, err := p.expectNextToken(TokenNewline)
				if err != nil {
					return nil, err
				}

				code, err := p.collectUntil(TokenCodeBlock)
				if err != nil {
					return nil, err
				}

				p.appendChild(NewHTMLNode("pre", HTMLProps{"data-language": language},
					NewHTMLNode("code", nil, TextNode(strings.TrimSpace(stringifyTokens(code)))),
				))
			}
		default:
			p.addToParagraph(current)
		}

		p.Consume()
	}

	p.flush()
	return p.document, nil
}

func stringifyTokens(tokens []Token) string {
	var sb strings.Builder
	for _, token := range tokens {
		sb.WriteString(token.String())
	}
	return sb.String()
}

type parser struct {
	Stream[Token]

	document  *HTMLNode
	paragraph []Token

	line   int
	column int
}

func (p *parser) flush() {
	if len(p.paragraph) > 0 {
		textContent := stringifyTokens(p.paragraph)
		p.document.Children = append(p.document.Children, NewHTMLNode("p", nil, TextNode(textContent)))
		if strings.HasSuffix(textContent, "  ") {
			p.appendChild(NewHTMLNode("br", nil))
		}

		p.paragraph = []Token{}
	}
}

func (p *parser) appendChild(node Renderer) {
	p.flush()
	p.document.Children = append(p.document.Children, node)
}

func (p *parser) addToParagraph(token Token) {
	p.paragraph = append(p.paragraph, token)
}

func (p *parser) Consume() Token {
	token := p.Current()
	if token.Variant == TokenNewline {
		p.line++
		p.column = 0
	} else {
		p.column++
	}
	return p.Next()
}

func (p *parser) isNewline() bool {
	return p.column == 0
}

func (p *parser) expectNextToken(tokenVariant TokenVariant) (Token, error) {
	actualToken := p.Consume()
	switch actualToken.Variant {
	case tokenVariant:
		return actualToken, nil
	case TokenEOF:
		return Token{}, fmt.Errorf("unexpected end of input %s", p.errorAt())
	default:
		return Token{}, fmt.Errorf("unexpected token %v %s", actualToken, p.errorAt())
	}
}

// not the same as collectWhile in the lexer
// this ends at the target token, rather than after (so skipping it at the end is required over continue)
func (p *parser) collectUntil(tokenVariant TokenVariant) ([]Token, error) {
	result := []Token{}
	current := p.Current()
	for p.HasData() && current.Variant != tokenVariant {
		if current.Variant == TokenEOF {
			return nil, fmt.Errorf("unexpected end of input %s", p.errorAt())
		}

		result = append(result, current)
		current = p.Consume()
	}
	// ends at the tokenVariant
	return result, nil
}

func (p *parser) errorAt() string {
	return fmt.Sprintf("at line %d, column %d", p.line, p.column)
}
