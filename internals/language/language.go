package language

import "github.com/faymn-dev/nook/internals/node"

func Render(markdown string) string {
	// TODO parse should never throw an error
	result, _ := Parse(Tokenize(markdown))
	return result.ToHTML()
}

func RenderDocument(markdown string) string {
	// TODO parse should never throw an error
	result, _ := Parse(Tokenize(markdown))
	return node.NewDocument(result).ToHTML()
}
