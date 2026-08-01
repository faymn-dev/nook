package language

import (
	"fmt"

	"github.com/faymn-dev/nook/internals/node"
)

const indentSize = 2

func Parse(tokens []Token) (node.Renderer, error) {
	p := newParser(tokens)

loop:
	for p.HasData() {
		current := p.Current()

		if current.Variant == TokenNewline && p.Peek().Variant == TokenNewline { // double newline
			p.flush()
			p.Next() // skip newline
			p.Next() // skip newline
			continue
		} else if !(p.Previous().Variant == TokenNewline || p.Previous().Variant == TokenEOF) { // not a newline before, so don't process as block
			p.appendToParagraph(current)
			p.Next()
			continue
		}

		var result node.Renderer

		// this are "blocks", meaning the previous token must be a newline
		switch current.Variant {
		case TokenEOF:
			break loop
		case TokenNewline:
		case TokenHeading:
			result = p.tryParseHeading()
		case TokenCodeBlock:
			result = p.parseCodeBlock()
		case TokenSeparator:
			result = node.NewHTMLNode("br", nil)
		case TokenListItem:
			result = p.tryParseList("ul", TokenListItem)
		case TokenNumberedListItem:
			result = p.tryParseList("ol", TokenNumberedListItem)
		case TokenBlockquote:
			result = p.tryParseBlockquote()
		}

		if result != nil {
			p.appendChild(result)
		} else {
			p.appendToParagraph(current)
		}

		p.Next()
	}

	p.flush()
	return p.document, nil
}

func inlineParse(tokens []Token) node.Renderer {
	tokens = append(tokens, Token{Variant: TokenEOF})
	p := &parser{
		Stream: Stream[Token]{
			data: tokens,
		},
	}
	inlineFragment := node.NewHTMLFragment()

loop:
	for p.HasData() {
		current := p.Current()

		var result node.Renderer
		switch current.Variant {
		case TokenEOF:
			break loop
		case TokenStar:
			result = p.trySymmetricParse(TokenStar, true, []string{"em"})
		case TokenDoubleStar:
			result = p.trySymmetricParse(TokenDoubleStar, true, []string{"strong"})
		case TokenTripleStar:
			result = p.trySymmetricParse(TokenTripleStar, true, []string{"strong", "em"})
		case TokenCode:
			result = p.trySymmetricParse(TokenCode, false, []string{"code"})
		case TokenStrikethrough:
			result = p.trySymmetricParse(TokenStrikethrough, true, []string{"s"})
		case TokenHighlight:
			result = p.trySymmetricParse(TokenHighlight, true, []string{"mark"})
		case TokenBang:
			result = p.tryParseImage()
		case TokenLBracket:
			result = p.tryParseLink()
		}

		if result == nil {
			result = node.TextNode(current.String())
		}
		inlineFragment.Children = append(inlineFragment.Children, result)
		p.Next()
	}

	return inlineFragment
}

type parser struct {
	Stream[Token]

	document  *node.HTMLNode
	paragraph []Token
}

func newParser(tokens []Token) *parser {
	return &parser{
		Stream: Stream[Token]{
			data: tokens,
		},
		document: node.NewHTMLFragment(),
	}
}

func (p *parser) tryParseHeading() node.Renderer {
	current := p.Current()
	if current.Variant != TokenHeading {
		return nil
	}

	level := min(len(current.Value), 7)
	p.Next() // skip heading

	textNode := p.collectUntilThenInlineParse(TokenNewline, trimStartingSpace)
	tagName := fmt.Sprintf("h%d", level)
	return node.NewHTMLNode(tagName, nil, textNode)
}

func (p *parser) parseCodeBlock() node.Renderer {
	// TODO should this behavior change?
	// atm this does actually process tokens
	// the other methods don't until the very end when a result is made
	var language string
	if p.Peek().Variant == TokenString {
		language = p.Next().Value
	}

	p.Next() // skip code block
	if p.Current().Variant != TokenNewline {
		return nil
	}
	p.Next()

	codeTokens := p.collectUntil(TokenCodeBlock)
	code := stringifyTokens(trimTokens(codeTokens, TokenNewline))
	return node.NewHTMLNode("pre", node.HTMLProps{"data-language": language},
		node.NewHTMLNode("code", nil, node.TextNode(code)),
	)
}

func (p *parser) tryParseImage() node.Renderer {
	if p.Current().Variant != TokenBang {
		return nil
	}
	p.Next() // skip !

	parsedBrackets := p.tryParseBrackets()
	if parsedBrackets == nil {
		return nil
	}

	alt := stringifyTokens(parsedBrackets.label)
	src := stringifyTokens(parsedBrackets.url)

	return node.NewHTMLNode("img", node.HTMLProps{"src": src, "alt": alt})
}

func (p *parser) tryParseLink() node.Renderer {
	parsedBrackets := p.tryParseBrackets()
	if parsedBrackets == nil {
		return nil
	}

	label := inlineParse(parsedBrackets.label)
	href := stringifyTokens(parsedBrackets.url)

	return node.NewHTMLNode("a", node.HTMLProps{"href": href}, label.GetChildren()...)
}

type parsedBrackets struct {
	label []Token
	url   []Token
}

// utility to parse stuff in the format of [label](url)
// I separated this because links and images basically look the same
func (p *parser) tryParseBrackets() *parsedBrackets {
	bracketDepth := 0
	closeBracketIndex := -1

	// find matching bracket
	for i := range len(p.data) {
		token := p.data[i]
		switch token.Variant {
		case TokenLBracket:
			bracketDepth += 1
		case TokenRBracket:
			bracketDepth -= 1
		}

		if bracketDepth == 0 {
			closeBracketIndex = i
			break
		}
	}

	openParenIndex := closeBracketIndex + 1
	closeParenIndex := -1

	// bracket never closed or open paren does not follow ]
	if closeBracketIndex == -1 || p.PeekOffset(openParenIndex).Variant != TokenLParen {
		return nil
	}

	for i := openParenIndex; i < len(p.data); i++ {
		token := p.data[i]
		if token.Variant == TokenRParen {
			closeParenIndex = i
			break
		}
	}

	if closeParenIndex == -1 {
		return nil
	}

	result := &parsedBrackets{
		label: p.data[1:closeBracketIndex],
		url:   p.data[openParenIndex+1 : closeParenIndex],
	}

	p.data = p.data[closeParenIndex:]

	return result
}

// badly modified tryParseList method
func (p *parser) tryParseBlockquote() node.Renderer {
	// try to figure out where the list ends
	currentlyInBlockquote := true
	endBlockquoteIndex := -1
	for i := range len(p.data) - 1 {
		token := p.data[i]
		nextToken := p.data[i+1]

		switch token.Variant {
		case TokenNewline:
			currentlyInBlockquote = false
		case TokenBlockquote:
			// note: indents are guaranteed to be at the start of the line
			currentlyInBlockquote = true
		}

		if !currentlyInBlockquote && nextToken.Variant != TokenBlockquote {
			endBlockquoteIndex = i
			break
		}
	}

	if endBlockquoteIndex == -1 {
		return nil
	}

	type listContext struct {
		level int
		list  *node.HTMLNode
	}

	tokens := p.data[:endBlockquoteIndex:endBlockquoteIndex]
	tokens = append(tokens, Token{Variant: TokenEOF})
	bqParser := &parser{
		Stream: Stream[Token]{
			data: tokens,
		},
	}

	parentBqNode := node.NewHTMLNode("blockquote", nil)
	stack := []listContext{{level: 1, list: parentBqNode}}
	// why is this level 1 whereas in the list, we start at level 0?
	// intuitively, this is because a Token of ">" is guaranteed to be of level 1, whereas a list item with no identation is at level 0

	// given the parent (ctx), parse each incoming list item
	for bqParser.Current().Variant == TokenBlockquote {
		level := 0
		if bqParser.Current().Variant == TokenBlockquote {
			level = len(bqParser.Current().Value)
			bqParser.Next() // skip blockquote
		}

		contentNode := bqParser.collectUntilThenInlineParse(TokenNewline, trimStartingSpace)
		bqParser.Next() // skip new line
		pNode := node.NewHTMLNode("p", nil, contentNode.GetChildren()...)

		// decide where the pNode should go
		top := stack[len(stack)-1]
		if level > top.level { // deeper level, so add it to top
			childBqNode := node.NewHTMLNode("blockquote", nil)
			top.list.Children = append(top.list.Children, childBqNode)
			stack = append(stack, listContext{
				list:  childBqNode,
				level: level,
			})
		} else if level < top.level { // pop until we find context at/above level
			for len(stack) > 1 && stack[len(stack)-1].level >= level {
				stack = stack[:len(stack)-1]
			}
		}

		currentBq := stack[len(stack)-1].list
		currentBq.Children = append(currentBq.Children, pNode)
	}

	p.data = p.data[endBlockquoteIndex:]

	return parentBqNode
}

// TODO refactor this method
// I've just wrapped the old logic by just creating another internal parser just for lists
// there is undoubtedly a better way using just indexes (use the other "try" methods for reference)
func (p *parser) tryParseList(tagName string, listItemVariant TokenVariant) node.Renderer {
	// try to figure out where the list ends
	currentlyInList := true
	endListIndex := -1
	for i := range len(p.data) - 1 {
		token := p.data[i]
		nextToken := p.data[i+1]

		switch token.Variant {
		case TokenNewline:
			currentlyInList = false
		case listItemVariant, TokenIndent:
			// note: indents are guaranteed to be at the start of the line
			currentlyInList = true
		}

		if !currentlyInList && nextToken.Variant != listItemVariant && nextToken.Variant != TokenIndent {
			endListIndex = i
			break
		}
	}

	if endListIndex == -1 {
		return nil
	}

	type listContext struct {
		level int
		list  *node.HTMLNode
	}

	tokens := p.data[:endListIndex:endListIndex]
	tokens = append(tokens, Token{Variant: TokenEOF})
	listParser := &parser{
		Stream: Stream[Token]{
			data: tokens,
		},
	}

	parentList := node.NewHTMLNode(tagName, nil)
	stack := []listContext{{level: 0, list: parentList}}

	// given the parent (ctx), parse each incoming list item
	for listParser.Current().Variant == TokenIndent || listParser.Current().Variant == listItemVariant {
		level := 0
		if listParser.Current().Variant == TokenIndent {
			level = len(listParser.Current().Value) / indentSize
			listParser.Next() // skip indentation
		}

		if _, err := listParser.expectCurrentToken(listItemVariant); err != nil {
			return nil
		}
		listParser.Next() // skip list item

		contentNode := listParser.collectUntilThenInlineParse(TokenNewline, trimStartingSpace)

		listParser.Next() // skip new line

		liNode := node.NewHTMLNode("li", nil, contentNode.GetChildren()...)

		// decide where the list item (liNode) should go
		top := stack[len(stack)-1]
		if level > top.level { // deeper level, so add it to last li of parent
			var parentLiNode *node.HTMLNode
			if len(top.list.Children) > 0 {
				parentLiNode = top.list.Children[len(top.list.Children)-1].(*node.HTMLNode)
			} else {
				return nil
			}

			childListNode := node.NewHTMLNode(tagName, nil)
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

		// add list item to current list
		currentList := stack[len(stack)-1].list
		currentList.Children = append(currentList.Children, liNode)
	}

	p.data = p.data[endListIndex:]

	return parentList
}

// utility to parse up until a matching token variant and stuff it inside some parent
// useful for simple, symmetric tokens, like **bold text**
//
// trySymmetricParse(TokenDoubleStar, true, []string{"strong"})
func (p *parser) trySymmetricParse(variant TokenVariant, useInlineParse bool, parent []string) node.Renderer {
	if p.Current().Variant != variant {
		return nil
	}
	p.Next() // skip current

	closeVariantIndex := -1
	for i := range len(p.data) {
		token := p.data[i]
		if token.Variant == variant {
			closeVariantIndex = i
			break
		}
	}

	if closeVariantIndex == -1 {
		return nil
	}

	tokens := p.data[:closeVariantIndex]
	var result node.Renderer
	if useInlineParse {
		result = inlineParse(tokens)
	} else {
		result = node.NewHTMLFragment(node.TextNode(stringifyTokens(tokens)))
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

	p.data = p.data[closeVariantIndex:]

	return children
}

func (p *parser) collectUntilThenInlineParse(tokenVariant TokenVariant, preprocessTokens func([]Token) []Token) node.Renderer {
	tokens := p.collectUntil(tokenVariant)
	node := inlineParse(preprocessTokens(tokens))
	return node
}

// manage tokens that get accumulated into the paragraph (not a block basically)

func (p *parser) flush() {
	if !isEmpty(p.paragraph) {
		children := inlineParse(p.paragraph)
		p.document.Children = append(p.document.Children, node.NewHTMLNode("p", nil, children.GetChildren()...))
		if p.paragraph[len(p.paragraph)-1].Variant == TokenIndent {
			p.document.Children = append(p.document.Children, node.NewHTMLNode("br", nil))
		}
		p.paragraph = nil
	}
}

func (p *parser) appendChild(node node.Renderer) {
	if node == nil {
		return
	}

	p.flush()
	p.document.Children = append(p.document.Children, node)
}

// use this when parsing blocks
func (p *parser) appendToParagraph(token Token) {
	p.paragraph = append(p.paragraph, token)
}

func (p *parser) expectCurrentToken(tokenVariant TokenVariant) (Token, error) {
	actualToken := p.Current()
	switch actualToken.Variant {
	case tokenVariant:
		return actualToken, nil
	case TokenEOF:
		return Token{}, fmt.Errorf("unexpected end of input ")
	default:
		return Token{}, fmt.Errorf("unexpected token %v", actualToken)
	}
}

// not the same as collectWhile in the lexer
// this ends at the target token, rather than after (so skipping it at the end is required over continue)
func (p *parser) collectUntil(tokenVariant TokenVariant) []Token {
	result := []Token{}
	current := p.Current()
	for p.HasData() && current.Variant != tokenVariant && current.Variant != TokenEOF {
		result = append(result, current)
		current = p.Next()
	}
	// ends at the tokenVariant
	return result
}
