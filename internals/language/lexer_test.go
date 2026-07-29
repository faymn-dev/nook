package language

import (
	"slices"
	"testing"
)

func TestTokenizer(t *testing.T) {
	type Example struct {
		Name   string
		Input  string
		Output []Token
	}

	examples := []Example{
		// stars
		{
			Name:  "stars",
			Input: "this should be *in italics* maybe",
			Output: []Token{
				{Variant: TokenString, Value: "this should be "},
				{Variant: TokenStar},
				{Variant: TokenString, Value: "in italics"},
				{Variant: TokenStar},
				{Variant: TokenString, Value: " maybe"},
				{Variant: TokenEOF},
			},
		},
		{
			Name:  "stars alone",
			Input: "*in italics*",
			Output: []Token{
				{Variant: TokenStar},
				{Variant: TokenString, Value: "in italics"},
				{Variant: TokenStar},
				{Variant: TokenEOF},
			},
		},
		{
			Name:  "stars at the end",
			Input: "this should be *in italics*",
			Output: []Token{
				{Variant: TokenString, Value: "this should be "},
				{Variant: TokenStar},
				{Variant: TokenString, Value: "in italics"},
				{Variant: TokenStar},
				{Variant: TokenEOF},
			},
		},
		{
			Name:  "stars with newlines",
			Input: "this should be *in\nitalics*",
			Output: []Token{
				{Variant: TokenString, Value: "this should be "},
				{Variant: TokenStar},
				{Variant: TokenString, Value: "in"},
				{Variant: TokenNewline},
				{Variant: TokenString, Value: "italics"},
				{Variant: TokenStar},
				{Variant: TokenEOF},
			},
		},

		// code blocks
		{
			Name:  "inline code",
			Input: "inline code `let x = 10`",
			Output: []Token{
				{Variant: TokenString, Value: "inline code "},
				{Variant: TokenCode},
				{Variant: TokenString, Value: "let x = 10"},
				{Variant: TokenCode},
				{Variant: TokenEOF},
			},
		},
		{
			Name:  "multiline code block",
			Input: "```js\nlet x = 10;\nprint(x)\n```",
			Output: []Token{
				{Variant: TokenCodeBlock},
				// maybe this is a little silly that we don't just smash it all into the value of TokenCodeBlock
				// imo the parser should handle everything, the lexer doesn't care about context
				{Variant: TokenString, Value: "js"},
				{Variant: TokenNewline},
				{Variant: TokenString, Value: "let x = 10;"},
				{Variant: TokenNewline},
				{Variant: TokenString, Value: "print"},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "x"},
				{Variant: TokenRParen},
				{Variant: TokenNewline},
				{Variant: TokenCodeBlock},
				{Variant: TokenEOF},
			},
		},

		// strike through
		{
			Name:  "basic strikethrough",
			Input: "~~hello~~",
			Output: []Token{
				{Variant: TokenStrikethrough},
				{Variant: TokenString, Value: "hello"},
				{Variant: TokenStrikethrough},
				{Variant: TokenEOF},
			},
		},

		// highlight
		{
			Name:  "basic highlight",
			Input: "==hello==",
			Output: []Token{
				{Variant: TokenHighlight},
				{Variant: TokenString, Value: "hello"},
				{Variant: TokenHighlight},
				{Variant: TokenEOF},
			},
		},

		// links and images
		{
			Name:  "link",
			Input: "[look at this really cool link](https://google.com)",
			Output: []Token{
				{Variant: TokenLBracket},
				{Variant: TokenString, Value: "look at this really cool link"},
				{Variant: TokenRBracket},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "https://google.com"},
				{Variant: TokenRParen},
				{Variant: TokenEOF},
			},
		},
		{
			Name:  "link with formatting",
			Input: "[**bold this guy**](https://google.com)",
			Output: []Token{
				{Variant: TokenLBracket},
				{Variant: TokenDoubleStar},
				{Variant: TokenString, Value: "bold this guy"},
				{Variant: TokenDoubleStar},
				{Variant: TokenRBracket},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "https://google.com"},
				{Variant: TokenRParen},
				{Variant: TokenEOF},
			},
		},
		{
			Name:  "image",
			Input: "![alt text](https://google.com/favicon.ico)",
			Output: []Token{
				{Variant: TokenBang},
				{Variant: TokenLBracket},
				{Variant: TokenString, Value: "alt text"},
				{Variant: TokenRBracket},
				{Variant: TokenLParen},
				{Variant: TokenString, Value: "https://google.com/favicon.ico"},
				{Variant: TokenRParen},
				{Variant: TokenEOF},
			},
		},

		// header
		{
			Name:  "basic header",
			Input: "# Heading 1\n## Heading 2\n### Heading 3\n#### Heading 4\n##### Heading 5\n###### Heading 6\n####### Heading 7",
			Output: []Token{
				{Variant: TokenHeading, Value: "#"},
				{Variant: TokenString, Value: " Heading 1"},
				{Variant: TokenNewline},
				{Variant: TokenHeading, Value: "##"},
				{Variant: TokenString, Value: " Heading 2"},
				{Variant: TokenNewline},
				{Variant: TokenHeading, Value: "###"},
				{Variant: TokenString, Value: " Heading 3"},
				{Variant: TokenNewline},
				{Variant: TokenHeading, Value: "####"},
				{Variant: TokenString, Value: " Heading 4"},
				{Variant: TokenNewline},
				{Variant: TokenHeading, Value: "#####"},
				{Variant: TokenString, Value: " Heading 5"},
				{Variant: TokenNewline},
				{Variant: TokenHeading, Value: "######"},
				{Variant: TokenString, Value: " Heading 6"},
				{Variant: TokenNewline},
				{Variant: TokenHeading, Value: "#######"},
				{Variant: TokenString, Value: " Heading 7"},
				{Variant: TokenEOF},
			},
		},

		// separators
		{
			Name:  "basic separator",
			Input: "---",
			Output: []Token{
				{Variant: TokenSeparator, Value: "---"},
				{Variant: TokenEOF},
			},
		},
		{
			Name:  "longer separator",
			Input: "---------",
			Output: []Token{
				{Variant: TokenSeparator, Value: "---------"},
				{Variant: TokenEOF},
			},
		},
		{
			Name:  "not a separator",
			Input: "--",
			Output: []Token{
				{Variant: TokenString, Value: "--"},
				{Variant: TokenEOF},
			},
		},

		// list items
		{
			Name:  "basic list",
			Input: "- get apples\n- cheezits\n  - nested list item",
			Output: []Token{
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " get apples"},
				{Variant: TokenNewline},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " cheezits"},
				{Variant: TokenNewline},
				{Variant: TokenIndent, Value: "  "},
				{Variant: TokenListItem},
				{Variant: TokenString, Value: " nested list item"},
				{Variant: TokenEOF},
			},
		},
		{
			Name:  "basic numbered list",
			Input: "1. get apples\n2. cheezits\n  1. nested list item",
			Output: []Token{
				{Variant: TokenNumberedListItem, Value: "1."},
				{Variant: TokenString, Value: " get apples"},
				{Variant: TokenNewline},
				{Variant: TokenNumberedListItem, Value: "2."},
				{Variant: TokenString, Value: " cheezits"},
				{Variant: TokenNewline},
				{Variant: TokenIndent, Value: "  "},
				{Variant: TokenNumberedListItem, Value: "1."},
				{Variant: TokenString, Value: " nested list item"},
				{Variant: TokenEOF},
			},
		},

		// escaped tokens
		{
			Name:  "escaped tokens",
			Input: `\*`,
			Output: []Token{
				{Variant: TokenEscape, Value: "*"},
				{Variant: TokenEOF},
			},
		},
	}

	for _, example := range examples {
		output := Tokenize(example.Input)
		if !slices.Equal(example.Output, output) {
			t.Errorf("%s: expected %v, got %v", example.Name, example.Output, output)
		}
	}
}
