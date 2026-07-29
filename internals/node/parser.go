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
			type listContext struct {
				parent *listContext
				list   *HTMLNode // list we are currently inside
				level  int
			}

			parentList := NewHTMLNode("ul", nil)
			ctx := &listContext{list: parentList, level: 0}

			for ctx != nil {
				// calculate indentation
				indents, err := p.collectUntil(TokenListItem)
				level := 0
				for _, token := range indents {
					if token.Variant != TokenIndent {
						return nil, fmt.Errorf("expected indent token %s", p.errorAt())
					}
					level++
				}
				p.consume() // skip list item

				contentNode, err := p.collectUntilThenInlineParse(TokenNewline, trimStartingSpace)
				if err != nil {
					return nil, err
				}
				p.consume() // skip new line

				liNode := NewHTMLNode("li", nil, contentNode.GetChildren()...)

				if level == ctx.level {
					ctx.list.Children = append(ctx.list.Children, liNode)
				} else if level > ctx.level {
					childListNode := NewHTMLNode("ul", nil)
					childListNode.Children = append(childListNode.Children, liNode)
					lastLiNode := ctx.list.Children[len(ctx.list.Children)-1].(*HTMLNode)
					lastLiNode.Children = append(lastLiNode.Children, childListNode)
					ctx = &listContext{
						parent: ctx,
						list:   childListNode,
						level:  level,
					}
				} else {
					curr := ctx
					for curr.parent != nil && curr.level >= level {
						curr = curr.parent
					}

					if curr == nil { // climbed out of the root list
						parentList.Children = append(parentList.Children, liNode)
						ctx = &listContext{list: parentList}
					} else {
						curr.list.Children = append(curr.list.Children, liNode)
						ctx = curr
					}
				}

				if !(p.Current().Variant == TokenIndent || p.Current().Variant == TokenListItem) {
					ctx = nil
				}
			}

			p.appendChild(parentList)
			continue
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

		p.document.Children = append(p.document.Children, NewHTMLNode("p", nil, children.GetChildren()...))
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
