package language

import (
	"fmt"

	"github.com/faymn-dev/initiator/internals/node"
)

const indentSize = 2

func Parse(tokens []Token) (node.Renderer, error) {
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
		case TokenNewline: // seeing a random newline should do nothing
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
			p.appendChild(node.NewHTMLNode(tagName, nil, textNode))
		case TokenCodeBlock:
			var language string
			if p.Peek().Variant == TokenString {
				language = p.consume().Value
			}

			_, err := p.expectNextToken(TokenNewline)
			if err != nil {
				return nil, err
			}

			codeTokens := p.collectUntil(TokenCodeBlock)
			code := stringifyTokens(trimTokens(codeTokens, TokenNewline))
			p.appendChild(node.NewHTMLNode("pre", node.HTMLProps{"data-language": language},
				node.NewHTMLNode("code", nil, node.TextNode(code)),
			))
		case TokenSeparator:
			p.appendChild(node.NewHTMLNode("br", nil))
		case TokenListItem:
			type listContext struct {
				level int
				list  *node.HTMLNode
			}

			parentList := node.NewHTMLNode("ul", nil)
			stack := []listContext{{level: 0, list: parentList}}

			// approach: given the parent (ctx), parse each incoming list item
			for {

				level := 0
				if p.Current().Variant == TokenIndent {
					level = len(p.Current().Value) / indentSize
					p.consume() // skip indentation
				}

				if _, err := p.expectCurrentToken(TokenListItem); err != nil {
					return nil, fmt.Errorf("expected list item token %s", p.errorAt())
				}
				p.consume() // skip list item

				contentNode, err := p.collectUntilThenInlineParse(TokenNewline, trimStartingSpace)
				if err != nil {
					return nil, err
				}
				p.consume() // skip new line

				liNode := node.NewHTMLNode("li", nil, contentNode.GetChildren()...)

				// decide where the list item (liNode) should go
				top := stack[len(stack)-1]
				if level > top.level { // deeper level, so add it to last li of parent
					var parentLiNode *node.HTMLNode
					if len(top.list.Children) > 0 {
						parentLiNode = top.list.Children[len(top.list.Children)-1].(*node.HTMLNode)
					} else {
						return nil, fmt.Errorf("malformed list %s", p.errorAt())
					}

					childListNode := node.NewHTMLNode("ul", nil)
					// childListNode.Children = append(childListNode.Children, liNode)
					parentLiNode.Children = append(parentLiNode.Children, childListNode)
					stack = append(stack, listContext{
						list:  childListNode,
						level: level,
					})
				} else if level < top.level { // pop until we find context at/above level
					for len(stack) > 1 && stack[len(stack)-1].level >= level {
						stack = stack[:len(stack)-1]
					}
				}

				currentList := stack[len(stack)-1].list
				currentList.Children = append(currentList.Children, liNode)

				if !(p.Current().Variant == TokenIndent || p.Current().Variant == TokenListItem) {
					break
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

func parseInline(line int, tokens []Token) (node.Renderer, error) {
	tokens = append(tokens, Token{Variant: TokenEOF})
	p := &parser{
		Stream: Stream[Token]{
			data: tokens,
		},
		document: node.NewHTMLFragment(),
		line:     line,
	}

loop:
	for p.HasData() {
		current := p.Current()

		var err error
		switch current.Variant {
		case TokenEOF:
			break loop
		case TokenStar:
			_, err = p.inlineParseToken(TokenStar, true, []string{"em"})
		case TokenDoubleStar:
			_, err = p.inlineParseToken(TokenDoubleStar, true, []string{"strong"})
		case TokenTripleStar:
			_, err = p.inlineParseToken(TokenTripleStar, true, []string{"strong", "em"})
		case TokenCode:
			_, err = p.inlineParseToken(TokenCode, false, []string{"code"})
		case TokenStrikethrough:
			_, err = p.inlineParseToken(TokenStrikethrough, true, []string{"s"})
		case TokenHighlight:
			_, err = p.inlineParseToken(TokenHighlight, true, []string{"mark"})
		case TokenBang:
		case TokenLBracket:
		default:
			p.appendToDocument(current)
		}

		if err != nil {
			return nil, err
		}
		p.consume()
	}

	return p.document, nil
}

type parser struct {
	Stream[Token]

	document      *node.HTMLNode
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
		document:  node.NewHTMLFragment(),
		isNewline: true,
	}
}

// actually parse stuff
func (p *parser) inlineParseToken(variant TokenVariant, inlineParse bool, parent []string) (node.Renderer, error) {
	p.consume() // skip current

	line := p.line
	tokens := p.collectUntil(variant)
	p.consume() // skip variant

	var result node.Renderer
	var err error
	if inlineParse {
		result, err = parseInline(line, tokens)
	} else {
		result = node.NewHTMLFragment(node.TextNode(stringifyTokens(tokens)))
	}

	if err != nil {
		// TODO more descriptive errors
		return nil, err
	}

	children := result
	for len(parent) > 0 {
		top := parent[len(parent)-1]
		parent = parent[:len(parent)-1]
		if children.GetTagName() == node.FragmentTagName {
			children = node.NewHTMLNode(top, nil, children.GetChildren()...)
		} else {
			children = node.NewHTMLNode(top, nil, children)
		}
	}

	p.document.Children = append(p.document.Children, children)

	return children, nil
}

func (p *parser) collectUntilThenInlineParse(tokenVariant TokenVariant, preprocessTokens func([]Token) []Token) (node.Renderer, error) {
	line := p.line
	tokens := p.collectUntil(tokenVariant)
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

		p.document.Children = append(p.document.Children, node.NewHTMLNode("p", nil, children.GetChildren()...))
		if p.paragraph[len(p.paragraph)-1].Variant == TokenIndent {
			p.document.Children = append(p.document.Children, node.NewHTMLNode("br", nil))
		}
		p.paragraph = nil
		p.paragraphLine = p.line
	}
	return nil
}

func (p *parser) appendChild(node node.Renderer) error {
	if err := p.flush(); err != nil {
		return err
	}
	p.document.Children = append(p.document.Children, node)
	return nil
}

// use this when parsing blocks
func (p *parser) appendToParagraph(token Token) {
	p.paragraph = append(p.paragraph, token)
}

// use this when parsing inline
func (p *parser) appendToDocument(tokens ...Token) {
	p.document.Children = append(p.document.Children, node.TextNode(stringifyTokens(tokens)))
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
func (p *parser) collectUntil(tokenVariant TokenVariant) []Token {
	result := []Token{}
	current := p.Current()
	for p.HasData() && current.Variant != tokenVariant && current.Variant != TokenEOF {
		result = append(result, current)
		current = p.consume()
	}
	// ends at the tokenVariant
	return result
}

func (p *parser) errorAt() string {
	return fmt.Sprintf("at line %d", p.line)
}
