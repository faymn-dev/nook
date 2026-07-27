package node

import (
	"slices"
	"testing"
)

func TestStars(t *testing.T) {
	type Example struct {
		Name   string
		Input  string
		Output []token
	}

	examples := []Example{
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
	}

	for _, example := range examples {
		output, err := lex(example.Input)
		if err != nil {
			t.Errorf("%s: failed to process input: %v", example.Name, err)
			continue
		}

		if !slices.Equal(example.Output, output) {
			t.Errorf("%s: expected %v, got %v", example.Name, example.Output, output)
		}
	}
}
