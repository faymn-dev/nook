package node

import (
	"unicode"
)

type TokenVariant int

const (
	TokenEOF TokenVariant = iota

	TokenNewline
	TokenString

	// block tokens
	TokenHeading          // #
	TokenCodeBlock        // ```
	TokenSeparator        // ---
	TokenListItem         // -
	TokenNumberedListItem // 1. or any number followed by a period

	// inline tokens
	TokenStar          // *
	TokenDoubleStar    // **
	TokenCode          // `
	TokenStrikethrough // ~~
	TokenHighlight     // ==
	TokenBang          // !
	TokenLParen        // (
	TokenRParen        // )
	TokenLBracket      // [
	TokenRBracket      // ]
	TokenEscape        // \whatever
)

type Token struct {
	variant TokenVariant
	value   string
}

type lexer struct {
	Stream[rune]

	result []Token
	word   []rune
}

func Tokenize(markdown string) []Token {
	l := &lexer{
		Stream: Stream[rune]{
			data: []rune(markdown),
		},
	}
	// return nil, fmt.Errorf("unexpected character %q at line %d, column %d", l.current(), l.line, l.column)

	for l.HasData() {
		current := l.Current()

		switch current {
		case '\n':
			l.commitToken(Token{variant: TokenNewline})
		case '*':
			if l.Peek() == '*' {
				l.commitToken(Token{variant: TokenDoubleStar})
				l.Next() // skip current *
			} else {
				l.commitToken(Token{variant: TokenStar})
			}
		case '#':
			l.commitToken(Token{variant: TokenHeading, value: string(l.CollectWhile('#'))})
			// anytime we use collectWhile, we end on a different token
			// we don't want to autoskip it, so next iteration
			continue
		case '-':
			separator := l.CollectWhile('-')
			if len(separator) == 1 {
				l.commitToken(Token{variant: TokenListItem})
			} else if len(separator) == 2 {
				// just two dashes is just a word
				l.addToWord('-')
				l.addToWord('-')
			} else {
				l.commitToken(Token{variant: TokenSeparator})
			}
			continue
		case '`':
			if l.MatchAhead('`', '`') {
				// TODO should we care about 4 backtiks in a row?
				// we can potentially reuse the collectWhile method to ensure it's a length of three?
				l.commitToken(Token{variant: TokenCodeBlock})
				l.Next() // skip current `
				l.Next() // skip peek `
			} else {
				l.commitToken(Token{variant: TokenCode})
			}
		case '~':
			if l.Peek() == '~' {
				l.commitToken(Token{variant: TokenStrikethrough})
				l.Next() // skip current ~
			} else {
				// not a strike through, so add it to the word normally
				l.addToWord(current)
			}
		case '=':
			if l.Peek() == '=' {
				l.commitToken(Token{variant: TokenHighlight})
				l.Next() // skip current =
			} else {
				l.addToWord(current)
			}
		case '!':
			l.commitToken(Token{variant: TokenBang})
		case '(':
			l.commitToken(Token{variant: TokenLParen})
		case ')':
			l.commitToken(Token{variant: TokenRParen})
		case '[':
			l.commitToken(Token{variant: TokenLBracket})
		case ']':
			l.commitToken(Token{variant: TokenRBracket})
		case '\\':
			if l.HasNext() {
				l.Next()
				l.commitToken(Token{variant: TokenEscape, value: string(l.Current())})
			} else {
				// you escaped... nothing, because nothing is left
				l.addToWord(current)
			}
		default:
			if unicode.IsDigit(current) {
				digits := l.collectWhileDigit()
				if l.Current() == '.' {
					l.commitToken(Token{variant: TokenNumberedListItem})
					l.Next() // skip current .
				} else {
					// it's not a list item, so just shove it into the word
					for _, digit := range digits {
						l.addToWord(digit)
					}
				}
				continue
			} else {
				l.addToWord(current)
			}
		}

		l.Next()
	}

	l.commitToken(Token{variant: TokenEOF})

	return l.result
}

func (l *lexer) commitToken(t Token) {
	if len(l.word) > 0 {
		l.result = append(l.result, Token{variant: TokenString, value: string(l.word)})
	}
	l.result = append(l.result, t)
	l.word = []rune{}
}

func (l *lexer) addToWord(char rune) {
	l.word = append(l.word, char)
}

func (l *lexer) collectWhileDigit() []rune {
	result := []rune{}
	for l.HasData() && unicode.IsDigit(l.Current()) {
		result = append(result, l.Current())
		l.Next()
	}
	return result
}
