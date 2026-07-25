package node

import "testing"

func TestEquality(t *testing.T) {
	node := NewTextNode("This is a text node", TextNodeBold, "")
	node2 := NewTextNode("This is a text node", TextNodeBold, "")

	if !node.Equals(node2) {
		t.Error("expected text nodes to equal each other")
		return
	}
}
