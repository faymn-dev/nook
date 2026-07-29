package node

import (
	"fmt"
)

func Parse(tokens []Token) (Renderer, error) {
	p := newParser(tokens)

loop:
	for p.HasData() {
		current := p.Current()
		if !p.isNewline {
			p.appendToParagraph(current)
			p.consume()
			continue
		}

		switch current.Variant {
		case TokenEOF:
			break loop
		case TokenNewline:
		case TokenHeading:
			level := len(current.Value)
			if level > 7 {
				return nil, fmt.Errorf("headings cannot be greater than level 7 %s", p.errorAt())
			}
			p.consume() // skip heading

			prevLine := p.line
			headerTokens, err := p.collectUntil(TokenNewline)
			if err != nil {
				return nil, err
			}

			textContent, err := parseInline(prevLine, trimStartingSpace(headerTokens))
			if err != nil {
				return nil, err
			}

			tagName := fmt.Sprintf("h%d", level)
			p.appendChild(NewHTMLNode(tagName, nil, textContent.GetChildren()...))

		case TokenCodeBlock:
			var language string
			if p.Peek().Variant == TokenString {
				language = p.consume().Value
			}

			_, err := p.expectNextToken(TokenNewline)
			if err != nil {
				return nil, err
			}

			codeTokens, err := p.collectUntil(TokenCodeBlock)
			if err != nil {
				return nil, err
			}
			code := stringifyTokens(trimNewlines(codeTokens))
			p.appendChild(NewHTMLNode("pre", HTMLProps{"data-language": language},
				NewHTMLNode("code", nil, TextNode(code)),
			))
		default:
			p.appendToParagraph(current)
		}

		p.consume()
	}

	p.flush()
	return p.document, nil
}

func parseInline(line int, tokens []Token) (Renderer, error) {
	tokens = append(tokens, Token{Variant: TokenEOF})
	p := &parser{
		Stream: Stream[Token]{
			data: tokens,
		},
		document: NewHTMLFragment(),
		line:     line,
	}

loop:
	for p.HasData() {
		current := p.Current()

		switch current.Variant {
		case TokenEOF:
			break loop
		case TokenNewline:
		case TokenString:
			p.document.Children = append(p.document.Children, TextNode(current.Value))
		default:
			return nil, fmt.Errorf("unexpected token %s", p.errorAt())
		}

		p.consume()
	}

	return p.document, nil
}

type parser struct {
	Stream[Token]

	document      *HTMLNode
	paragraph     []Token
	paragraphLine int
	line          int
	isNewline     bool
}

func newParser(tokens []Token) *parser {
	return &parser{
		Stream: Stream[Token]{
			data: tokens,
		},
		document:  NewHTMLFragment(),
		isNewline: true,
	}
}

// manage tokens that get accumulated into the paragraph (not a block basically)

func (p *parser) flush() error {
	if len(p.paragraph) > 0 {
		children, err := parseInline(p.paragraphLine, p.paragraph)
		if err != nil {
			return err
		}

		p.document.Children = append(p.document.Children, children.GetChildren()...)
		if p.paragraph[len(p.paragraph)-1].Variant == TokenIndent {
			p.document.Children = append(p.document.Children, NewHTMLNode("br", nil))
		}
		p.paragraph = nil
		p.paragraphLine = p.line
	}
	return nil
}

func (p *parser) appendChild(node Renderer) error {
	if err := p.flush(); err != nil {
		return err
	}
	p.document.Children = append(p.document.Children, node)
	return nil
}

func (p *parser) appendToParagraph(token Token) {
	p.paragraph = append(p.paragraph, token)
}

// use instead of Next() to track line numbers
func (p *parser) consume() Token {
	token := p.Current()
	if token.Variant == TokenNewline {
		p.line++
		p.isNewline = true
	} else {
		p.isNewline = false
	}
	return p.Next()
}

func (p *parser) expectNextToken(tokenVariant TokenVariant) (Token, error) {
	actualToken := p.consume()
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
	for p.HasData() && current.Variant != tokenVariant && current.Variant != TokenEOF {
		result = append(result, current)
		current = p.consume()
	}
	// ends at the tokenVariant
	return result, nil
}

func (p *parser) errorAt() string {
	return fmt.Sprintf("at line %d", p.line)
}
