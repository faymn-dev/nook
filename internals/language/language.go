package language

func Render(markdown string) string {
	// TODO parse should never throw an error
	result, _ := Parse(Tokenize(markdown))
	return result.ToHTML()
}
