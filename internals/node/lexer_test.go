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
				// maybe this is a little silly that we don't just smash it all into the value of TokenCodeBlock
				// imo the parser should handle everything, the lexer doesn't care about context
				{variant: TokenString, value: "js"},
				{variant: TokenNewline},
				{variant: TokenString, value: "let x = 10;"},
				{variant: TokenNewline},
				{variant: TokenString, value: "print"},
				{variant: TokenLParen},
				{variant: TokenString, value: "x"},
				{variant: TokenRParen},
				{variant: TokenNewline},
				{variant: TokenCodeBlock},
				{variant: TokenEOF},
			},
		},

		// strike through
		{
			Name:  "basic strikethrough",
			Input: "~~hello~~",
			Output: []token{
				{variant: TokenStrikethrough},
				{variant: TokenString, value: "hello"},
				{variant: TokenStrikethrough},
				{variant: TokenEOF},
			},
		},

		// highlight
		{
			Name:  "basic highlight",
			Input: "==hello==",
			Output: []token{
				{variant: TokenHighlight},
				{variant: TokenString, value: "hello"},
				{variant: TokenHighlight},
				{variant: TokenEOF},
			},
		},

		// links and images
		{
			Name:  "link",
			Input: "[look at this really cool link](https://google.com)",
			Output: []token{
				{variant: TokenLBracket},
				{variant: TokenString, value: "look at this really cool link"},
				{variant: TokenRBracket},
				{variant: TokenLParen},
				{variant: TokenString, value: "https://google.com"},
				{variant: TokenRParen},
				{variant: TokenEOF},
			},
		},
		{
			Name:  "link with formatting",
			Input: "[**bold this guy**](https://google.com)",
			Output: []token{
				{variant: TokenLBracket},
				{variant: TokenDoubleStar},
				{variant: TokenString, value: "bold this guy"},
				{variant: TokenDoubleStar},
				{variant: TokenRBracket},
				{variant: TokenLParen},
				{variant: TokenString, value: "https://google.com"},
				{variant: TokenRParen},
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
