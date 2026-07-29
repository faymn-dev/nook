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
			Name:   "nothing",
			Input:  []Token{},
			Output: NewHTMLFragment(),
		},
		{
			Name: "headings",
			Input: []Token{
				{Variant: TokenHeading, Value: "#"},
				{Variant: TokenString, Value: " Lorem Ipsum"},
				{Variant: TokenEOF},
			},
			Output: NewHTMLFragment(NewHTMLNode("h1", nil, TextNode("Lorem Ipsum"))),
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
			Name: "bad headings",
			Input: []Token{
				{Variant: TokenHeading, Value: "#"},
				{Variant: TokenEOF},
			},
		},
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
