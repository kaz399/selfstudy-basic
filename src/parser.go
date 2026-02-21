/**************************************************************/
/*
   parser.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

import (
	"strconv"
)

type Node interface {
	// for debug
	TokenLiteral() string
}

type Expression interface {
	Node
	// dummy method for distinguishing between expression and statement
	expressionNode()
}

// INT token
type IntegerLiteral struct {
	Token Token
	Value int64
}

func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) expressionNode()      {}

// InfixExpression is a node representing an infix expression such as “left operand + right operand”
type InfixExpression struct {
	Token    Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) expressionNode()      {}

type Parser struct {
	l         *Lexer
	curToken  Token // current token
	peekToken Token // next token (for lookahead)
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l}
	// set peekToken
	p.nextToken()
	// set currentToken
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseExpression() Expression {
	leftNode := &IntegerLiteral{Token: p.curToken}
	leftNode.Value, _ = strconv.ParseInt(p.curToken.Literal, 0, 64)

	if p.peekToken.Type == PLUS {
		p.nextToken()

		expression := &InfixExpression{
			Token:    p.curToken,
			Operator: p.curToken.Literal,
			Left:     leftNode,
		}

		p.nextToken()

		rightNode := &IntegerLiteral{Token: p.curToken}
		rightNode.Value, _ = strconv.ParseInt(p.curToken.Literal, 0, 64)
		expression.Right = rightNode

		return expression
	}

	return leftNode
}
