package node

type TokenVariant int

const (
	TokenEOF TokenVariant = iota

	TokenNewline
	TokenString

	// inline tokens
	TokenStar
	TokenDoubleStar
)

type token struct {
	variant TokenVariant
	value   string
}

type lexer struct {
	data   []rune
	cursor int

	result []token
	word   []rune
}

func newLexer(markdown string) *lexer {
	return &lexer{
		data: []rune(markdown),
	}
}

func lex(markdown string) ([]token, error) {
	l := newLexer(markdown)
	// return nil, fmt.Errorf("unexpected character %q at line %d, column %d", l.current(), l.line, l.column)

	for l.hasCurrent() {
		d := l.current()

		if d == '\n' {
			l.commitToken(token{variant: TokenNewline})
		} else if d == '*' {
			if l.peek() == '*' {
				l.commitToken(token{variant: TokenDoubleStar})
				l.next() // skip over current *
			} else {
				l.commitToken(token{variant: TokenStar})
			}
		} else {
			// add to word
			l.word = append(l.word, d)
		}

		l.next()
	}

	l.commitToken(token{variant: TokenEOF})

	return l.result, nil
}

func (l *lexer) commitToken(t token) {
	if len(l.word) > 0 {
		l.result = append(l.result, token{variant: TokenString, value: string(l.word)})
	}
	l.result = append(l.result, t)
	l.word = []rune{}
}

// ============================================================
// utilities for moving the cursor around
// ============================================================

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
	if l.hasNext() {
		return l.data[l.cursor+1]
	}
	return 0
}
