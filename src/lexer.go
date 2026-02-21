/**************************************************************/
/*
   lexer.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

type TokenType string

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"
	INT     TokenType = "INT"
	PLUS    TokenType = "PLUS"
)

type Token struct {
	Type    TokenType
	Literal string
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

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '+':
		tok = Token{Type: PLUS, Literal: string(l.ch)}
	case 0:
		tok = Token{Type: EOF, Literal: ""}
	default:
		if isDigit(l.ch) {
			tok.Type = INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch)}
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}
