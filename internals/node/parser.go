package node

import (
	"fmt"
)

func Parse(tokens []Token) (Renderer, error) {
	p := newParser(tokens)

loop:
	for p.HasData() {
		current := p.Current()
		if !p.isNewline && current.Variant != TokenNewline {
			p.appendToParagraph(current)
			p.consume()
			continue
		}

		switch current.Variant {
		case TokenEOF:
			break loop
		case TokenNewline:
			// seeing a random newline should do nothing
		case TokenHeading:
			level := len(current.Value)
			if level > 7 {
				return nil, fmt.Errorf("headings cannot be greater than level 7 %s", p.errorAt())
			}
			p.consume() // skip heading

			textNode, err := p.collectUntilThenInlineParse(TokenNewline, trimStartingSpace)
			if err != nil {
				return nil, err
			}

			tagName := fmt.Sprintf("h%d", level)
			p.appendChild(NewHTMLNode(tagName, nil, textNode))
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
		case TokenSeparator:
			p.appendChild(NewHTMLNode("br", nil))
		case TokenListItem:
			type stackItem struct {
				list  *HTMLNode
				level int
			}

			parentList := NewHTMLNode("ul", nil)
			stack := []stackItem{
				{list: parentList, level: 0},
			}

			for len(stack) > 0 {
				if _, err := p.expectCurrentToken(TokenListItem); err != nil {
					return nil, fmt.Errorf("expected list item token %s", p.errorAt())
				}
				p.consume()

				top := &stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				listItemNode, err := p.collectUntilThenInlineParse(TokenNewline, preprocessTokensNoop)
				if err != nil {
					return nil, err
				}
				parentList.Children = append(parentList.Children, NewHTMLNode("li", nil, listItemNode.GetChildren()...))

				switch p.Peek().Variant {
				case TokenListItem:
					stack = append(stack, *top)
					p.consume() // skip current \n
				case TokenIndent:
					indents, err := p.collectUntil(TokenListItem)
					if err != nil {
						return nil, err
					}
					p.consume() // skip current \n

					level := len(indents)
					if level > top.level {
						childList := NewHTMLNode("ul", nil)
						top.list.Children = append(top.list.Children, NewHTMLNode("li", nil, childList))
						stack = append(stack, stackItem{
							list:  childList,
							level: level,
						})
					} else {
						stack = append(stack, *top)
					}
				}
			}

			p.appendChild(parentList)
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
		default:
			p.document.Children = append(p.document.Children, TextNode(current.String()))
		}

		p.consume()
	}

	return p.document, nil
}

type parser struct {
	Stream[Token]

	document      *HTMLNode
	paragraph     []Token
	paragraphLine int // line the starts off the paragraph
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

func (p *parser) collectUntilThenInlineParse(tokenVariant TokenVariant, preprocessTokens func([]Token) []Token) (Renderer, error) {
	line := p.line
	tokens, err := p.collectUntil(tokenVariant)
	if err != nil {
		return nil, err
	}

	node, err := parseInline(line, preprocessTokens(tokens))
	if err != nil {
		return nil, err
	}

	return node, nil
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

func (p *parser) expectCurrentToken(tokenVariant TokenVariant) (Token, error) {
	actualToken := p.Current()
	switch actualToken.Variant {
	case tokenVariant:
		return actualToken, nil
	case TokenEOF:
		return Token{}, fmt.Errorf("unexpected end of input %s", p.errorAt())
	default:
		return Token{}, fmt.Errorf("unexpected token %v %s", actualToken, p.errorAt())
	}

}

func (p *parser) expectNextToken(tokenVariant TokenVariant) (Token, error) {
	p.consume()
	return p.expectCurrentToken(tokenVariant)
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
