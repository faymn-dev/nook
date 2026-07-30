package node_test

import (
	"testing"

	"github.com/faymn-dev/nook/internals/node"
)

func TestHTMLNodeToHTML(t *testing.T) {
	type Example struct {
		Name   string
		Input  node.Renderer
		Output string
	}

	examples := []Example{
		{
			Name:   "basic paragraph",
			Input:  node.NewHTMLNode("p", node.HTMLProps{"data-test": "example"}, node.TextNode("hello world")),
			Output: "<p data-test=\"example\">hello world</p>",
		},
		{
			Name: "many nested divs",
			Input: node.NewHTMLNode("div", node.HTMLProps{},
				node.NewHTMLNode("div", node.HTMLProps{},
					node.NewHTMLNode("div", node.HTMLProps{}),
				),
			),
			Output: "<div><div><div></div></div></div>",
		},
		{
			Name:   "",
			Input:  node.NewHTMLNode("p", node.HTMLProps{"data-test": "example"}, node.TextNode("hello world")),
			Output: "<p data-test=\"example\">hello world</p>",
		},
	}

	for _, example := range examples {
		output := example.Input.ToHTML()
		if output != example.Output {
			t.Errorf("%s: expected %q, got %q", example.Name, example.Output, output)
		}
	}
}
