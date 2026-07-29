package node

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	type Example struct {
		Name   string
		Input  []Token
		Output *HTMLNode
	}

	examples := []Example{
		{
			Name: "nothing",
			Input: []Token{
				{Variant: TokenEOF},
			},
			Output: NewHTMLFragment(),
		},

		{
			Name: "headings",
			Input: []Token{
				{Variant: TokenHeading, Value: "#"},
				{Variant: TokenString, Value: " Lorem Ipsum"},
				{Variant: TokenNewline},
				{Variant: TokenEOF},
			},
			Output: NewHTMLFragment(NewHTMLNode("h1", nil, NewHTMLFragment(TextNode("Lorem Ipsum")))),
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
			Output: NewHTMLFragment(NewHTMLNode("pre", HTMLProps{"data-language": ""}, NewHTMLNode("code", nil, TextNode("let x = 10;")))),
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
			Output: NewHTMLFragment(NewHTMLNode("pre", HTMLProps{"data-language": "js"}, NewHTMLNode("code", nil, TextNode("let x = 10;")))),
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
			Output: NewHTMLFragment(
				NewHTMLNode("ul", nil,
					NewHTMLNode("li", nil, TextNode("list item 1")),
					NewHTMLNode("li", nil, TextNode("list item 2")),
					NewHTMLNode("li", nil, TextNode("list item 3")),
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
			Output: NewHTMLFragment(
				NewHTMLNode("ul", nil,
					NewHTMLNode("li", nil, TextNode("list item 1")),
					NewHTMLNode("li", nil, TextNode("list item 2"),
						NewHTMLNode("ul", nil,
							NewHTMLNode("li", nil, TextNode("list item 2a")),
							NewHTMLNode("li", nil, TextNode("list item 2b")),
						),
					),
					NewHTMLNode("li", nil, TextNode("list item 3")),
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
			Output: NewHTMLFragment(
				NewHTMLNode("ul", nil,
					NewHTMLNode("li", nil, TextNode("hello world")),
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
			Output: NewHTMLFragment(
				NewHTMLNode("ul", nil,
					NewHTMLNode("li", nil, TextNode("item 1")),
					NewHTMLNode("li", nil, TextNode("item 2")),
					NewHTMLNode("li", nil, TextNode("item 3")),
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
			Output: NewHTMLFragment(
				NewHTMLNode("ul", nil,
					NewHTMLNode("li", nil, TextNode("level 1"),
						NewHTMLNode("ul", nil,
							NewHTMLNode("li", nil, TextNode("level 2"),
								NewHTMLNode("ul", nil,
									NewHTMLNode("li", nil, TextNode("level 3")),
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
			Output: NewHTMLFragment(
				NewHTMLNode("ul", nil,
					NewHTMLNode("li", nil),
					NewHTMLNode("li", nil, TextNode("not empty")),
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
			Output: NewHTMLFragment(
				NewHTMLNode("ul", nil,
					NewHTMLNode("li", nil, TextNode("parent"),
						NewHTMLNode("ul", nil,
							NewHTMLNode("li", nil, TextNode("child 1")),
							NewHTMLNode("li", nil, TextNode("child 2")),
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
