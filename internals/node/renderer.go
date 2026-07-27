package node

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Renderer interface {
	ToHTML() string
}

type blockHandler struct {
	// an empty or nil prefix means match against everything
	prefixes []string

	// returning a nil Renderer & nil error indicates that the block handler should match against something else
	handle func(block string) (Renderer, error)
}

var blockHandlers = []blockHandler{
	{
		prefixes: []string{"#"},
		handle:   handleHeading,
	},
	{
		prefixes: []string{"- ", "* ", "+ "},
		handle: func(block string) (Renderer, error) {
			return handleList("ul", strings.Split(block, "\n"))
		},
	},
	{
		prefixes: []string{"1. "},
		handle: func(block string) (Renderer, error) {
			return handleList("ol", strings.Split(block, "\n"))
		},
	},
	{
		prefixes: []string{"***", "___", "---"},
		handle:   handleSeparator,
	},
	{
		prefixes: []string{"> "},
		handle: func(block string) (Renderer, error) {
			return handleBlockQuote(strings.Split(block, "\n"))
		},
	},
	{
		prefixes: []string{"```"},
		handle:   handleCodeBlock,
	},
	{
		prefixes: nil,
		handle:   handleText,
	},
}

// we assume that the markdown is formatted
// blocks are separated by two newlines, for example
func MarkdownToHTMLNodes(markdown string) ([]HTMLNode, error) {
	blocksSeq := strings.SplitSeq(markdown, "\n\n")
	nodes := []Renderer{}

blockLoop:
	for block := range blocksSeq {
		for _, handler := range blockHandlers {
			if hasPrefix(block, handler.prefixes) {
				node, err := handler.handle(block)
				if err != nil {
					return nil, err
				} else if node == nil {
					continue
				}

				nodes = append(nodes, node)

				if strings.HasSuffix(block, "  ") {
					// if paragraph has two spaces at the end, we add a line break
					nodes = append(nodes, NewHTMLNode("br", nil))
				}

				continue blockLoop
			}
		}
	}

	return nil, nil
}

func handleHeading(block string) (Renderer, error) {
	headingCount := 0
	for char := range block {
		if char == '#' {
			headingCount += 1
		} else {
			break
		}
	}
	if headingCount > 6 {
		// validate heading count
		return nil, fmt.Errorf("invalid heading count - %d", headingCount)
	}

	headingText := strings.TrimSpace(block[headingCount:])
	return NewHTMLNode(fmt.Sprintf("h%d", headingCount), nil, TextNode(headingText)), nil
}

func handleList(tag string, lines []string) (Renderer, error) {
	list := NewHTMLNode(tag, nil)
	indent := "  "
	for len(lines) > 0 {
		line := lines[0]

		if strings.HasPrefix(line, indent) {
			// collect all lines that start with a space
			var childLines []string
			for len(lines) > 0 && strings.HasPrefix(line, indent) {
				// add child line, but strip first two spaces
				childLines = append(childLines, strings.TrimPrefix(lines[0], indent))
				lines = lines[1:]
			}

			// convert child list into html node
			childList, err := handleList(tag, childLines)
			if err != nil {
				return nil, err
			}

			list.Children = append(list.Children, childList)
		} else {
			// remove stuff before the first space
			// for example "1. hello world" -> "hello world"
			line = strings.TrimSpace(line)
			line = line[strings.Index(line, " ")+1:]

			// standard list item
			line, err := handleText(line)
			if err != nil {
				return nil, err
			}

			list.Children = append(list.Children, NewHTMLNode("li", nil, line))
		}
	}

	return list, nil
}

func handleSeparator(block string) (Renderer, error) {
	// validate block and ensure it's all a single character
	targetChar, _ := utf8.DecodeRuneInString(block)
	for _, char := range block {
		if targetChar != char {
			return nil, fmt.Errorf("invalid separator - expected %q, got %q", targetChar, char)
		}
	}

	return NewHTMLNode("hr", nil), nil
}

func handleBlockQuote(lines []string) (Renderer, error) {
	blockQuote := NewHTMLNode("blockquote", nil)

	for len(lines) > 0 {
		line := lines[0]
		if strings.HasPrefix(line, ">> ") {
			var childLines []string
			for len(lines) > 0 && strings.HasPrefix(line, ">>") {
				childLines = append(childLines, lines[0])
				lines = lines[1:]
			}

			childBlockQuote, err := handleBlockQuote(childLines)
			if err != nil {
				return nil, err
			}

			blockQuote.Children = append(blockQuote.Children, nil, childBlockQuote)
		} else {
			blockQuote.Children = append(blockQuote.Children, nil, TextNode(line))
		}

		lines = lines[1:]
	}

	return blockQuote, nil
}

func handleCodeBlock(block string) (Renderer, error) {
	if !strings.HasSuffix(block, "```") {
		return nil, fmt.Errorf("invalid code block - expected closing backticks")
	}

	block = strings.TrimPrefix(block, "```")

	var language string
	index := strings.Index(block, "\n")
	if index > -1 {
		// block has more than one newline, so we can assign a code language
		language = block[:index]
		block = block[index+1:]
	}

	codeBlock := NewHTMLNode("pre", nil,
		NewHTMLNode("pre", HTMLProps{"data-language": language}, TextNode(block)),
	)
	return codeBlock, nil
}

func handleText(text string) (Renderer, error) {
	result := NewHTMLNode("span", nil)

	delimiterStack := []rune{}

	for len(text) > 0 {
		char, size := utf8.DecodeRuneInString(text)

		// switch get(2) {
		// case "**":
		// case "__":
		// case "==":
		// }
		//
		// switch get(1) {
		// case "\\":
		// case "*":
		// case "_":
		// case "!":
		// case "[":
		// case "<":
		// case "`":
		// }

		text = text[size:]
	}

	return result, nil
}
