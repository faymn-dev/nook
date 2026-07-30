package language_test

import (
	"testing"

	"github.com/faymn-dev/initiator/internals/language"
)

func TestLanguage(t *testing.T) {
	type Example struct {
		Name   string
		Input  string
		Output string
	}

	examples := []Example{
		{
			Name:   "text to text nodes",
			Input:  "This is **text** with an *italic* word and a `code block` and an ![obi wan image](https://i.imgur.com/fJRm4Vk.jpeg) and a [link](https://boot.dev)",
			Output: "<p>This is <strong>text</strong></p>",
		},
	}

	for _, example := range examples {
		output := language.Render(example.Input)
		if output != example.Output {
			t.Errorf("%s:\n\texpected %s\n\tgot %s", example.Name, example.Output, output)
		}
	}
}
