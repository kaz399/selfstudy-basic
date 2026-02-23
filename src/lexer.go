/**************************************************************/
/*
   lexer.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

import (
	"strings"
	"unicode"
)

type TokenType string

const (
	// special
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Literals
	IDENT  TokenType = "IDENT"
	NUMBER TokenType = "NUMBER"
	STRING TokenType = "STRING"

	// operators
	ASSIGN TokenType = "="
	PLUS   TokenType = "+"
	MINUS  TokenType = "-"
	ASTER  TokenType = "*"
	SLASH  TokenType = "/"

	EQ  TokenType = "=" // same lexeme as ASSIGN; parser decides context
	NEQ TokenType = "<>"
	LT  TokenType = "<"
	LTE TokenType = "<="
	GT  TokenType = ">"
	GTE TokenType = ">="

	// delimters
	LPAREN TokenType = "("
	RPAREN TokenType = ")"
	COMMA  TokenType = ","

	// keywords
	REM   TokenType = "REM"
	LET   TokenType = "LET"
	PRINT TokenType = "PRINT"
	INPUT TokenType = "INPUT"
	IF    TokenType = "IF"
	THEN  TokenType = "THEN"
	GOTO  TokenType = "GOTO"
	END   TokenType = "END"

	// REPL comamnds
	RUN  TokenType = "RUN"
	LIST TokenType = "LIST"
	NEW  TokenType = "NEW"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"REM":   REM,
	"LET":   LET,
	"PRINT": PRINT,
	"INPUT": INPUT,
	"IF":    IF,
	"THEN":  THEN,
	"GOTO":  GOTO,
	"END":   END,
	"RUN":   RUN,
	"LIST":  LIST,
	"NEW":   NEW,
}

func LookupIdent(s string) TokenType {
	u := strings.ToUpper(s)
	if tok, ok := keywords[u]; ok {
		return tok
	}
	return IDENT
}

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	switch l.ch {
	case 0:
		return Token{Type: EOF, Literal: ""}
	case '+':
		tok := Token{Type: PLUS, Literal: "+"}
		l.readChar()
		return tok
	case '-':
		tok := Token{Type: MINUS, Literal: "-"}
		l.readChar()
		return tok
	case '*':
		tok := Token{Type: ASTER, Literal: "*"}
		l.readChar()
		return tok
	case '/':
		tok := Token{Type: SLASH, Literal: "/"}
		l.readChar()
		return tok
	case '(':
		tok := Token{Type: LPAREN, Literal: "("}
		l.readChar()
		return tok
	case ')':
		tok := Token{Type: RPAREN, Literal: ")"}
		l.readChar()
		return tok
	case ',':
		tok := Token{Type: COMMA, Literal: ","}
		l.readChar()
		return tok
	case '=':
		tok := Token{Type: ASSIGN, Literal: "="}
		l.readChar()
		return tok
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: LTE, Literal: "<="}
		}
		if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			return Token{Type: NEQ, Literal: "<>"}
		}
		tok := Token{Type: LT, Literal: "<"}
		l.readChar()
		return tok
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return Token{Type: GTE, Literal: ">="}
		}
		tok := Token{Type: GT, Literal: ">"}
		l.readChar()
		return tok
	case '"':
		s, ok := l.readString()
		if !ok {
			return Token{Type: ILLEGAL, Literal: "unterminated string"}
		}
		return Token{Type: STRING, Literal: s}
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			upper := strings.ToUpper(ident)
			return Token{Type: LookupIdent(ident), Literal: upper}
		}
		if isDigit(l.ch) {
			num := l.readNumber()
			return Token{Type: NUMBER, Literal: num}
		}
		illegal := Token{Type: ILLEGAL, Literal: string(l.ch)}
		l.readChar()
		return illegal
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '$' {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position
	hasDot := false
	for isDigit(l.ch) || (!hasDot && l.ch == '.') {
		if l.ch == '.' {
			hasDot = true
		}
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readString() (string, bool) {
	// current char is '"'
	l.readChar()
	start := l.position
	for l.ch != '"' && l.ch != 0 && l.ch != '\n' {
		l.readChar()
	}
	if l.ch != '"' {
		return "", false
	}
	s := l.input[start:l.position]
	l.readChar() // consume closing '"'
	return s, true
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
