package node

import "testing"

func TestEquality(t *testing.T) {
	type Example struct {
		InputNodeA *TextNode
		InputNodeB *TextNode
		Output     bool
	}

	examples := []Example{
		{
			InputNodeA: NewTextNode("This is a text node", TextNodeBold, ""),
			InputNodeB: NewTextNode("This is a text node", TextNodeBold, ""),
			Output:     true,
		},
		{
			InputNodeA: NewTextNode("This is a text node", TextNodeBold, ""),
			InputNodeB: NewTextNode("This is a different node", TextNodeBold, ""),
			Output:     false,
		},
		{
			InputNodeA: NewTextNode("This is a text node", TextNodeItalic, ""),
			InputNodeB: NewTextNode("This is a different node", TextNodeBold, ""),
			Output:     false,
		},
	}

	for i, example := range examples {
		output := example.InputNodeA.Equals(example.InputNodeB)
		if output != example.Output {
			t.Errorf("example %d - expected %v, got %v", i, example.Output, output)
			return
		}
	}
}
