package language_test

import (
	"testing"

	"github.com/faymn-dev/nook/internals/language"
)

func TestLanguage(t *testing.T) {
	type Example struct {
		Name   string
		Input  string
		Output string
	}

	examples := []Example{
		{
			Name:   "nothing",
			Input:  "",
			Output: "",
		},
		{
			Name:   "text to text nodes",
			Input:  "This is **text** with an *italic* word and a `code block` and an ![obi wan image](https://i.imgur.com/fJRm4Vk.jpeg) and a [link](https://boot.dev)",
			Output: `<p>This is <strong>text</strong> with an <em>italic</em> word and a <code>code block</code> and an <img alt="obi wan image" src="https://i.imgur.com/fJRm4Vk.jpeg" /> and a <a href="https://boot.dev">link</a></p>`,
		},
		{
			Name:   "paragraphs",
			Input:  "This is **bolded** paragraph\ntext in a p\ntag here\n\nThis is another paragraph with *italic* text and `code` here",
			Output: "<p>This is <strong>bolded</strong> paragraph\ntext in a p\ntag here</p><p>This is another paragraph with <em>italic</em> text and <code>code</code> here</p>",
		},
		{
			Name:   "codeblocks",
			Input:  "```\nThis is text that _should_ remain\nthe **same** even with inline stuff```",
			Output: "<pre data-language=\"\"><code>This is text that _should_ remain\nthe **same** even with inline stuff</code></pre>",
		},
	}

	for _, example := range examples {
		output := language.Render(example.Input)
		if output != example.Output {
			t.Errorf("%s:\n  expected %s\n  got %s", example.Name, example.Output, output)
		}
	}
}
