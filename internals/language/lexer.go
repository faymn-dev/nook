package language

import (
	"unicode"
)

type TokenVariant int

const (
	TokenEOF TokenVariant = iota

	TokenIndent
	TokenNewline
	TokenString

	// block tokens
	TokenHeading          // #
	TokenCodeBlock        // ```
	TokenSeparator        // ---
	TokenListItem         // -
	TokenNumberedListItem // 1. or any number followed by a period
	TokenBlockquote       // >

	// inline tokens
	TokenStar          // *
	TokenDoubleStar    // **
	TokenTripleStar    // ***
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
	Variant TokenVariant
	Value   string
}

var tokenVariantToString = map[TokenVariant]string{
	TokenNewline:       "\n",
	TokenCodeBlock:     "```",
	TokenListItem:      "-",
	TokenStar:          "*",
	TokenDoubleStar:    "**",
	TokenTripleStar:    "***",
	TokenCode:          "`",
	TokenStrikethrough: "~~",
	TokenHighlight:     "==",
	TokenBang:          "!",
	TokenLParen:        "(",
	TokenRParen:        ")",
	TokenLBracket:      "[",
	TokenRBracket:      "]",
}

func (t Token) String() string {
	// yet another layer of ducktape to make this work
	// ideally, the value would just be on the token
	// default to retrieve from map, otherwise use token value
	value, ok := tokenVariantToString[t.Variant]
	if len(t.Value) == 0 && ok {
		return value
	}
	return t.Value
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

	for l.HasData() {
		current := l.Current()

		switch current {
		case '\n':
			l.commitToken(Token{Variant: TokenNewline})
		case ' ':
			if (l.Previous() == '\n' || l.Previous() == 0) && l.Peek() == ' ' { // anything more than two spaces at the start of the line is considered indentation
				l.commitToken(Token{Variant: TokenIndent, Value: string(l.collectWhile(' '))})
				continue
			} else {
				l.addToWord(current)
			}
		case '*':
			if l.Peek() == '*' && l.PeekOffset(2) == '*' {
				l.commitToken(Token{Variant: TokenTripleStar})
				l.Next() // skip current *
				l.Next() // skip peek *
			} else if l.Peek() == '*' {
				l.commitToken(Token{Variant: TokenDoubleStar})
				l.Next() // skip current *
			} else {
				l.commitToken(Token{Variant: TokenStar})
			}
		case '>':
			l.commitToken(Token{Variant: TokenBlockquote, Value: string(l.collectWhile('>'))})
			// anytime we use collectWhile, we end on a different token
			// we don't want to autoskip it, so next iteration
			continue
		case '#':
			l.commitToken(Token{Variant: TokenHeading, Value: string(l.collectWhile('#'))})
			continue
		case '-':
			separator := l.collectWhile('-')
			if len(separator) == 1 {
				l.commitToken(Token{Variant: TokenListItem})
			} else if len(separator) == 2 {
				// just two dashes is just a word
				l.addToWord('-')
				l.addToWord('-')
			} else {
				l.commitToken(Token{Variant: TokenSeparator, Value: string(separator)})
			}
			continue
		case '`':
			if l.matchAhead('`', '`') {
				// TODO should we care about 4 backtiks in a row?
				// we can potentially reuse the collectWhile method to ensure it's a length of three?
				l.commitToken(Token{Variant: TokenCodeBlock})
				l.Next() // skip current `
				l.Next() // skip peek `
			} else {
				l.commitToken(Token{Variant: TokenCode})
			}
		case '~':
			if l.Peek() == '~' {
				l.commitToken(Token{Variant: TokenStrikethrough})
				l.Next() // skip current ~
			} else { // not a strike through, so add it to the word normally
				l.addToWord(current)
			}
		case '=':
			if l.Peek() == '=' {
				l.commitToken(Token{Variant: TokenHighlight})
				l.Next() // skip current =
			} else {
				l.addToWord(current)
			}
		case '!':
			l.commitToken(Token{Variant: TokenBang})
		case '(':
			l.commitToken(Token{Variant: TokenLParen})
		case ')':
			l.commitToken(Token{Variant: TokenRParen})
		case '[':
			l.commitToken(Token{Variant: TokenLBracket})
		case ']':
			l.commitToken(Token{Variant: TokenRBracket})
		case '\\':
			if l.HasNext() {
				l.Next()
				l.commitToken(Token{Variant: TokenEscape, Value: string(l.Current())})
			} else { // you escaped... nothing, because nothing is left
				l.addToWord(current)
			}
		default:
			if unicode.IsDigit(current) {
				digits := l.collectWhileDigit()
				if l.Current() == '.' {
					l.commitToken(Token{Variant: TokenNumberedListItem, Value: string(digits) + "."})
					l.Next() // skip current .
				} else { // it's not a list item, so just shove it into the word
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

	l.commitToken(Token{Variant: TokenEOF})

	return l.result
}

func (l *lexer) commitToken(t Token) {
	if len(l.word) > 0 {
		l.result = append(l.result, Token{Variant: TokenString, Value: string(l.word)})
	}
	l.result = append(l.result, t)
	l.word = []rune{}
}

func (l *lexer) addToWord(char rune) {
	l.word = append(l.word, char)
}

// when using this method, you probably don't want to call .Next()
// because the last token you'll be on is the token AFTER target
func (l *lexer) collectWhileDigit() []rune {
	result := []rune{}
	for l.HasData() && unicode.IsDigit(l.Current()) {
		result = append(result, l.Current())
		l.Next()
	}
	return result
}

// when using this method, you probably don't want to call .Next()
// because the last token you'll be on is the token AFTER target
func (l *lexer) collectWhile(target rune) []rune {
	result := []rune{}
	for l.HasData() && l.Current() == target {
		result = append(result, target)
		l.Next()
	}
	return result
}

func (l *lexer) matchAhead(targets ...rune) bool {
	for i, target := range targets {
		if target != l.PeekOffset(i+1) {
			return false
		}
	}
	return true
}
