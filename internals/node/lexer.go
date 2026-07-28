package node

import "unicode"

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
)

type Token struct {
	variant TokenVariant
	value   string
}

type lexer struct {
	data   []rune
	cursor int

	result []Token
	word   []rune
}

func newLexer(markdown string) *lexer {
	return &lexer{
		data: []rune(markdown),
	}
}

func Tokenize(markdown string) ([]Token, error) {
	l := newLexer(markdown)
	// return nil, fmt.Errorf("unexpected character %q at line %d, column %d", l.current(), l.line, l.column)

	for l.hasCurrent() {
		d := l.current()

		switch d {
		case '\n':
			l.commitToken(Token{variant: TokenNewline})
		case '*':
			if l.peek() == '*' {
				l.commitToken(Token{variant: TokenDoubleStar})
				l.next() // skip current *
			} else {
				l.commitToken(Token{variant: TokenStar})
			}
		case '#':
			value := string(l.collectWhile('#'))
			l.commitToken(Token{variant: TokenHeading, value: value})
			// anytime we use collectWhile, we end on a different token
			// we don't want to autoskip it, so next iteration
			continue
		case '-':
			separator := l.collectWhile('-')
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
			if l.peek() == '`' && l.peekOffset(2) == '`' {
				// TODO should we care about 4 backtiks in a row?
				// we can potentially reuse the collectWhile method to ensure it's a length of three?
				l.commitToken(Token{variant: TokenCodeBlock})
				l.next() // skip current `
				l.next() // skip peek `
			} else {
				l.commitToken(Token{variant: TokenCode})
			}
		case '~':
			if l.peek() == '~' {
				l.commitToken(Token{variant: TokenStrikethrough})
				l.next() // skip current ~
			} else {
				// not a strike through, so add it to the word normally
				l.addToWord(d)
			}
		case '=':
			if l.peek() == '=' {
				l.commitToken(Token{variant: TokenHighlight})
				l.next() // skip current =
			} else {
				l.addToWord(d)
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
		default:
			if unicode.IsDigit(d) {
				digits := l.collectWhileDigit()
				if l.current() == '.' {
					l.commitToken(Token{variant: TokenNumberedListItem})
					l.next() // skip current .
				} else {
					// it's not a list item, so just shove it into the word
					for _, d := range digits {
						l.addToWord(d)
					}
				}
				continue
			} else {
				l.addToWord(d)
			}
		}

		l.next()
	}

	l.commitToken(Token{variant: TokenEOF})

	return l.result, nil
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

// ============================================================
// utilities for moving the cursor around
// ============================================================

func (l *lexer) collectWhile(target rune) []rune {
	result := []rune{}
	for l.hasCurrent() && l.current() == target {
		result = append(result, target)
		l.next()
	}
	return result
}

func (l *lexer) collectWhileDigit() []rune {
	result := []rune{}
	for l.hasCurrent() && unicode.IsDigit(l.current()) {
		result = append(result, l.current())
		l.next()
	}
	return result
}

func (l *lexer) hasNext() bool {
	return l.cursor+1 < len(l.data)
}

func (l *lexer) hasCurrent() bool {
	return l.cursor < len(l.data)
}

func (l *lexer) current() rune {
	if l.hasCurrent() {
		return l.data[l.cursor]
	}
	return 0
}

func (l *lexer) next() rune {
	l.cursor++
	return l.current()
}

func (l *lexer) peek() rune {
	return l.peekOffset(1)
}

func (l *lexer) peekOffset(offset int) rune {
	if l.hasNext() {
		return l.data[l.cursor+offset]
	}
	return 0
}
