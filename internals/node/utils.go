package node

import "strings"

// if the first token is a string, we'll strip the space of the front
// useful if this is a heading, for example
func trimStartingSpace(tokens []Token) []Token {
	if len(tokens) == 0 {
		return tokens
	}

	if tokens[0].Variant == TokenString {
		tokens[0].Value = strings.TrimLeft(tokens[0].Value, " ")
	}

	return tokens
}

// convert a list of tokens into a string
func stringifyTokens(tokens []Token) string {
	var sb strings.Builder
	for _, token := range tokens {
		sb.WriteString(token.String())
	}
	return sb.String()
}

// remove the first starting newline and last ending newline from a list of tokens
func trimNewlines(tokens []Token) []Token {
	if len(tokens) > 1 && tokens[0].Variant == TokenNewline {
		tokens = tokens[1:]
	}
	if len(tokens) > 1 && tokens[len(tokens)-1].Variant == TokenNewline {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens
}

func preprocessTokensNoop(tokens []Token) []Token {
	return tokens
}
