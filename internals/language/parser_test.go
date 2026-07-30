package language

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/faymn-dev/nook/internals/node"
)

func TestInlineParser(t *testing.T) {
	type Example struct {
		Name   string
		Input  []Token
		Output *node.HTMLNode
	}

	examples := []Example{
		{
			Name: "nothing",
			Input: []Token{
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(),
		},
		{
			Name: "italics",
			Input: []Token{
				{Variant: TokenStar},
				{Variant: TokenString, Value: "this should be in italics"},
				{Variant: TokenStar},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("em", nil, node.TextNode("this should be in italics")),
			),
		},
		{
			Name: "bolded",
			Input: []Token{
				{Variant: TokenDoubleStar},
				{Variant: TokenString, Value: "this should be bolded"},
				{Variant: TokenDoubleStar},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("strong", nil, node.TextNode("this should be bolded")),
			),
		},
		{
			Name: "italics and bolded",
			Input: []Token{
				{Variant: TokenTripleStar},
				{Variant: TokenString, Value: "this should be in italics and bolded"},
				{Variant: TokenTripleStar},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("strong", nil,
					node.NewHTMLNode("em", nil, node.TextNode("this should be in italics and bolded")),
				),
			),
		},
		{
			Name: "bolded inside of italics",
			Input: []Token{
				{Variant: TokenStar},
				{Variant: TokenString, Value: "this should be in italics "},
				{Variant: TokenDoubleStar},
				{Variant: TokenString, Value: "and this in bold"},
				{Variant: TokenDoubleStar},
				{Variant: TokenStar},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("em", nil, node.TextNode("this should be in italics "),
					node.NewHTMLNode("strong", nil,
						node.TextNode("and this in bold"),
					),
				),
			),
		},
		{
			Name: "simple link",
			Input: []Token{
				{Variant: TokenLBracket},
				{Variant: TokenString, Value: "Google"},
				{Variant: TokenRBracket},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "https://google.com"},
				{Variant: TokenRParen},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("a", map[string]string{"href": "https://google.com"},
					node.TextNode("Google"),
				),
			),
		},
		{
			Name: "formatted text inside link",
			Input: []Token{
				{Variant: TokenLBracket},
				{Variant: TokenStar},
				{Variant: TokenString, Value: "italic link"},
				{Variant: TokenStar},
				{Variant: TokenRBracket},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "https://example.com"},
				{Variant: TokenRParen},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("a", map[string]string{"href": "https://example.com"},
					node.NewHTMLNode("em", nil,
						node.TextNode("italic link"),
					),
				),
			),
		},
		{
			Name: "simple image",
			Input: []Token{
				{Variant: TokenBang},
				{Variant: TokenLBracket},
				{Variant: TokenString, Value: "Alt text"},
				{Variant: TokenRBracket},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "https://example.com/image.png"},
				{Variant: TokenRParen},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("img", map[string]string{
					"src": "https://example.com/image.png",
					"alt": "Alt text",
				}),
			),
		},
		{
			Name: "image inside link",
			Input: []Token{
				{Variant: TokenLBracket},
				{Variant: TokenBang},
				{Variant: TokenLBracket},
				{Variant: TokenString, Value: "Clickable Image"},
				{Variant: TokenRBracket},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "https://example.com/image.png"},
				{Variant: TokenRParen},
				{Variant: TokenRBracket},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "https://example.com"},
				{Variant: TokenRParen},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("a", map[string]string{"href": "https://example.com"},
					node.NewHTMLNode("img", map[string]string{
						"src": "https://example.com/image.png",
						"alt": "Clickable Image",
					}),
				),
			),
		},
		{
			Name: "image with empty alt text",
			Input: []Token{
				{Variant: TokenBang},
				{Variant: TokenLBracket},
				{Variant: TokenRBracket},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "https://example.com/photo.jpg"},
				{Variant: TokenRParen},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("img", map[string]string{
					"src": "https://example.com/photo.jpg",
					"alt": "",
				}),
			),
		},
	}

	for _, example := range examples {
		output := parseInline(example.Input)
		if !reflect.DeepEqual(output, example.Output) {
			formatted, _ := json.MarshalIndent(output, "", "  ")
			t.Errorf("%s: got %s", example.Name, string(formatted))
		}
	}
}

func TestParser(t *testing.T) {
	type Example struct {
		Name   string
		Input  []Token
		Output *node.HTMLNode
	}

	examples := []Example{
		{
			Name: "nothing",
			Input: []Token{
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(),
		},

		{
			Name: "headings",
			Input: []Token{
				{Variant: TokenHeading, Value: "#"},
				{Variant: TokenString, Value: " Lorem Ipsum"},
				{Variant: TokenNewline},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(node.NewHTMLNode("h1", nil, node.NewHTMLFragment(node.TextNode("Lorem Ipsum")))),
		},

		// code blocks
		{
			Name: "code block",
			Input: []Token{
				{Variant: TokenCodeBlock},
				{Variant: TokenNewline},
				{Variant: TokenString, Value: "let x = 10;"},
				{Variant: TokenNewline},
				{Variant: TokenCodeBlock},
				{Variant: TokenNewline},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(node.NewHTMLNode("pre", node.HTMLProps{"data-language": ""}, node.NewHTMLNode("code", nil, node.TextNode("let x = 10;")))),
		},
		{
			Name: "code block with language",
			Input: []Token{
				{Variant: TokenCodeBlock},
				{Variant: TokenString, Value: "js"},
				{Variant: TokenNewline},
				{Variant: TokenString, Value: "let x = 10;"},
				{Variant: TokenNewline},
				{Variant: TokenCodeBlock},
				{Variant: TokenNewline},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(node.NewHTMLNode("pre", node.HTMLProps{"data-language": "js"}, node.NewHTMLNode("code", nil, node.TextNode("let x = 10;")))),
		},

		{
			Name: "basic list",
			Input: []Token{
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " list item 1"},
				{Variant: TokenNewline},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " list item 2"},
				{Variant: TokenNewline},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " list item 3"},
				{Variant: TokenNewline},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("ul", nil,
					node.NewHTMLNode("li", nil, node.TextNode("list item 1")),
					node.NewHTMLNode("li", nil, node.TextNode("list item 2")),
					node.NewHTMLNode("li", nil, node.TextNode("list item 3")),
				),
			),
		},
		{
			Name: "nested list",
			Input: []Token{
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " list item 1"},
				{Variant: TokenNewline},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " list item 2"},
				{Variant: TokenNewline},

				{Variant: TokenIndent, Value: "  "},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " list item 2a"},
				{Variant: TokenNewline},
				{Variant: TokenIndent, Value: "  "},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " list item 2b"},
				{Variant: TokenNewline},

				{Variant: TokenListItem},
				{Variant: TokenString, Value: " list item 3"},
				{Variant: TokenNewline},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("ul", nil,
					node.NewHTMLNode("li", nil, node.TextNode("list item 1")),
					node.NewHTMLNode("li", nil, node.TextNode("list item 2"),
						node.NewHTMLNode("ul", nil,
							node.NewHTMLNode("li", nil, node.TextNode("list item 2a")),
							node.NewHTMLNode("li", nil, node.TextNode("list item 2b")),
						),
					),
					node.NewHTMLNode("li", nil, node.TextNode("list item 3")),
				),
			),
		},
		{
			Name: "single item list",
			Input: []Token{
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "hello world"},
				{Variant: TokenNewline},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("ul", nil,
					node.NewHTMLNode("li", nil, node.TextNode("hello world")),
				),
			),
		},
		{
			Name: "flat list",
			Input: []Token{
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "item 1"},
				{Variant: TokenNewline},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "item 2"},
				{Variant: TokenNewline},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "item 3"},
				{Variant: TokenNewline},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("ul", nil,
					node.NewHTMLNode("li", nil, node.TextNode("item 1")),
					node.NewHTMLNode("li", nil, node.TextNode("item 2")),
					node.NewHTMLNode("li", nil, node.TextNode("item 3")),
				),
			),
		},
		{
			Name: "deeply nested list",
			Input: []Token{
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "level 1"},
				{Variant: TokenNewline},

				{Variant: TokenIndent, Value: "  "},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "level 2"},
				{Variant: TokenNewline},

				{Variant: TokenIndent, Value: "    "},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "level 3"},
				{Variant: TokenNewline},

				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("ul", nil,
					node.NewHTMLNode("li", nil, node.TextNode("level 1"),
						node.NewHTMLNode("ul", nil,
							node.NewHTMLNode("li", nil, node.TextNode("level 2"),
								node.NewHTMLNode("ul", nil,
									node.NewHTMLNode("li", nil, node.TextNode("level 3")),
								),
							),
						),
					),
				),
			),
		},
		{
			Name: "list with empty items",
			Input: []Token{
				{Variant: TokenListItem},
				{Variant: TokenNewline},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "not empty"},
				{Variant: TokenNewline},
				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("ul", nil,
					node.NewHTMLNode("li", nil),
					node.NewHTMLNode("li", nil, node.TextNode("not empty")),
				),
			),
		},
		{
			Name: "multiple sub-lists under one item",
			Input: []Token{
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "parent"},
				{Variant: TokenNewline},

				{Variant: TokenIndent, Value: "  "},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "child 1"},
				{Variant: TokenNewline},

				{Variant: TokenIndent, Value: "  "},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: "child 2"},
				{Variant: TokenNewline},

				{Variant: TokenEOF},
			},
			Output: node.NewHTMLFragment(
				node.NewHTMLNode("ul", nil,
					node.NewHTMLNode("li", nil, node.TextNode("parent"),
						node.NewHTMLNode("ul", nil,
							node.NewHTMLNode("li", nil, node.TextNode("child 1")),
							node.NewHTMLNode("li", nil, node.TextNode("child 2")),
						),
					),
				),
			),
		},
	}

	for _, example := range examples {
		output, err := Parse(example.Input)
		if err != nil {
			t.Errorf("%s: %v", example.Name, err)
			continue
		}

		if !reflect.DeepEqual(output, example.Output) {
			formatted, _ := json.MarshalIndent(output, "", "  ")
			t.Errorf("%s: got %s", example.Name, string(formatted))
		}
	}
}

func TestParserErrors(t *testing.T) {
	type Example struct {
		Name  string
		Input []Token
	}

	examples := []Example{
		{
			Name: "bad code block",
			Input: []Token{
				{Variant: TokenCodeBlock},
				{Variant: TokenString, Value: "let x = 10;"},
				{Variant: TokenEOF},
			},
		},
	}

	for _, example := range examples {
		_, err := Parse(example.Input)
		if err == nil {
			t.Errorf("%s: expected error", example.Name)
		}
	}
}
