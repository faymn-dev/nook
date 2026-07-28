package node

import (
	"slices"
	"testing"
)

func TestLexer(t *testing.T) {
	type Example struct {
		Name   string
		Input  string
		Output []token
	}

	examples := []Example{
		// stars
		{
			Name:  "stars",
			Input: "this should be *in italics* maybe",
			Output: []token{
				{variant: TokenString, value: "this should be "},
				{variant: TokenStar},
				{variant: TokenString, value: "in italics"},
				{variant: TokenStar},
				{variant: TokenString, value: " maybe"},
				{variant: TokenEOF},
			},
		},
		{
			Name:  "stars alone",
			Input: "*in italics*",
			Output: []token{
				{variant: TokenStar},
				{variant: TokenString, value: "in italics"},
				{variant: TokenStar},
				{variant: TokenEOF},
			},
		},
		{
			Name:  "stars at the end",
			Input: "this should be *in italics*",
			Output: []token{
				{variant: TokenString, value: "this should be "},
				{variant: TokenStar},
				{variant: TokenString, value: "in italics"},
				{variant: TokenStar},
				{variant: TokenEOF},
			},
		},
		{
			Name:  "stars with newlines",
			Input: "this should be *in\nitalics*",
			Output: []token{
				{variant: TokenString, value: "this should be "},
				{variant: TokenStar},
				{variant: TokenString, value: "in"},
				{variant: TokenNewline},
				{variant: TokenString, value: "italics"},
				{variant: TokenStar},
				{variant: TokenEOF},
			},
		},

		// code blocks
		{
			Name:  "inline code",
			Input: "inline code `let x = 10`",
			Output: []token{
				{variant: TokenString, value: "inline code "},
				{variant: TokenCode},
				{variant: TokenString, value: "let x = 10"},
				{variant: TokenCode},
				{variant: TokenEOF},
			},
		},
		{
			Name:  "multiline code block",
			Input: "```js\nlet x = 10;\nprint(x)\n```",
			Output: []token{
				{variant: TokenCodeBlock},
				{variant: TokenString, value: "js"},
				{variant: TokenNewline},
				{variant: TokenString, value: "let x = 10;"},
				{variant: TokenNewline},
				{variant: TokenString, value: "print(x)"},
				{variant: TokenNewline},
				{variant: TokenCodeBlock},
				{variant: TokenEOF},
			},
		},
	}

	for _, example := range examples {
		output, err := Tokenize(example.Input)
		if err != nil {
			t.Errorf("%s: failed to process input: %v", example.Name, err)
			continue
		}

		if !slices.Equal(example.Output, output) {
			t.Errorf("%s: expected %v, got %v", example.Name, example.Output, output)
		}
	}
}
