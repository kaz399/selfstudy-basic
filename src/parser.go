/**************************************************************/
/*
   parser.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

import (
	"fmt"
	"strconv"
)

type precedence int

const (
	_ precedence = iota
	LOWEST
	COMPARE // = <> < <= > >=
	SUM     // + -
	PRODUCT // * /
	PREFIX  // unary + -
)

// precedence map
var precedences = map[TokenType]precedence{
	ASSIGN: COMPARE,
	NEQ:    COMPARE,
	LT:     COMPARE,
	LTE:    COMPARE,
	GT:     COMPARE,
	GTE:    COMPARE,
	PLUS:   SUM,
	MINUS:  SUM,
	ASTER:  PRODUCT,
	SLASH:  PRODUCT,
}

type Parser struct {
	l       *Lexer
	curTok  Token // current token
	peekTok Token // next token (for lookahead)
	errors  []string
}

func NewParser(src string) *Parser {
	p := &Parser{l: NewLexer(src)}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	p.peekTok = p.l.NextToken()
}

func (p *Parser) Errors() []string { return p.errors }

func (p *Parser) addErr(format string, a ...any) {
	p.errors = append(p.errors, fmt.Sprintf(format, a...))
}

func (p *Parser) ParseStatement() Stmt {
	switch p.curTok.Type {
	case REM:
		return &RemStmt{}
	case LET:
		return p.parseLetStmt(true)
	case IDENT:
		if p.peekTok.Type == ASSIGN {
			return p.parseLetStmt(false)
		}
		p.addErr("unexpected identifier %q", p.curTok.Literal)
		return nil
	case PRINT:
		return p.parsePrintStmt()
	case INPUT:
		return p.parseInputStmt()
	case IF:
		return p.parseIfStmt()
	case GOTO:
		return p.parseGotoStmt()
	case END:
		return &EndStmt{}
	default:
		p.addErr("unexpected token %s", p.curTok.Type)
		return nil
	}
}

func (p *Parser) parseLetStmt(hasLET bool) Stmt {
	var name string
	if hasLET {
		p.nextToken() // move to IDENT
		if p.curTok.Type != IDENT {
			p.addErr("LET requires identifier")
			return nil
		}
		name = p.curTok.Literal
	} else {
		if p.curTok.Type != IDENT {
			p.addErr("assignment requires identifier")
			return nil
		}
		name = p.curTok.Literal
	}

	if p.peekTok.Type != ASSIGN {
		p.addErr("expected '=' after identifier")
		return nil
	}
	p.nextToken() // '='
	p.nextToken() // expr start

	expr := p.parseExpr(LOWEST)
	if expr == nil {
		return nil
	}
	return &LetStmt{Name: name, Expr: expr}
}

func (p *Parser) parsePrintStmt() Stmt {
	if p.peekTok.Type == EOF {
		return &PrintStmt{Exprs: nil}
	}
	p.nextToken() // move to first expr
	exprs := []Expr{}
	first := p.parseExpr(LOWEST)
	if first == nil {
		return nil
	}
	exprs = append(exprs, first)

	for p.peekTok.Type == COMMA {
		p.nextToken() // comma
		p.nextToken() // expr
		e := p.parseExpr(LOWEST)
		if e == nil {
			return nil
		}
		exprs = append(exprs, e)
	}
	return &PrintStmt{Exprs: exprs}
}

func (p *Parser) parseInputStmt() Stmt {
	p.nextToken()
	if p.curTok.Type != IDENT {
		p.addErr("INPUT requires identifier")
		return nil
	}
	return &InputStmt{Name: p.curTok.Literal}
}

func (p *Parser) parseIfStmt() Stmt {
	p.nextToken()
	cond := p.parseExpr(LOWEST)
	if cond == nil {
		return nil
	}

	if p.peekTok.Type != THEN {
		p.addErr("THEN requires statement or line number")
		return nil
	}
	p.nextToken() // token after THEN

	// THEN linenumber
	if p.curTok.Type == NUMBER && p.peekTok.Type == EOF {
		n, err := parseIntStrict(p.curTok.Literal)
		if err != nil {
			p.addErr("invalid line number after THEN: %v", err)
			return nil
		}
		return &IfStmt{Cond: cond, ThenLine: n, HasLine: true}
	}

	// THEN statement
	thenStmt := p.ParseStatement()
	if thenStmt == nil {
		return nil
	}
	return &IfStmt{Cond: cond, ThenStmt: thenStmt}
}

func (p *Parser) parseGotoStmt() Stmt {
	p.nextToken()
	if p.curTok.Type != NUMBER {
		p.addErr("GOTO requires line number")
		return nil
	}
	n, err := parseIntStrict(p.curTok.Literal)
	if err != nil {
		p.addErr("invalid GOTO line number: %v", err)
		return nil
	}
	return &GotoStmt{Line: n}
}

func (p *Parser) parseExpr(pr precedence) Expr {
	left := p.parsePrefix()
	if left == nil {
		return nil
	}

	for p.peekTok.Type != EOF && pr < p.peekPrecedence() {
		switch p.peekTok.Type {
		case PLUS, MINUS, ASTER, SLASH, ASSIGN, NEQ, LT, LTE, GT, GTE:
			p.nextToken()
			left = p.parseInfix(left)
			if left == nil {
				return nil
			}
		default:
			return left
		}
	}
	return left
}

func (p *Parser) parsePrefix() Expr {
	switch p.curTok.Type {
	case NUMBER:
		v, err := strconv.ParseFloat(p.curTok.Literal, 64)
		if err != nil {
			p.addErr("invalid number %q", p.curTok.Literal)
			return nil
		}
		return &NumberLit{Value: v}
	case STRING:
		return &StringLit{Value: p.curTok.Literal}
	case IDENT:
		return &VarRef{Name: p.curTok.Literal}
	case PLUS, MINUS:
		op := p.curTok.Literal
		p.nextToken()
		rhs := p.parseExpr(PREFIX)
		if rhs == nil {
			return nil
		}
		return &UnaryExpr{Op: op, Rhs: rhs}
	case LPAREN:
		p.nextToken()
		e := p.parseExpr(LOWEST)
		if e == nil {
			return nil
		}
		if p.peekTok.Type != RPAREN {
			p.addErr("expexted ')'")
			return nil
		}
		p.nextToken() // consume ')'
		return e
	default:
		p.addErr("unexpected token in expression: %s", p.curTok.Type)
		return nil
	}
}

func (p *Parser) parseInfix(left Expr) Expr {
	opTok := p.curTok
	prec := p.curPrecedence()
	p.nextToken()
	right := p.parseExpr(prec)
	if right == nil {
		return nil
	}
	return &BinaryExpr{Op: opTok.Literal, Lhs: left, Rhs: right}
}

func (p *Parser) peekPrecedence() precedence {
	if pr, ok := precedences[p.peekTok.Type]; ok {
		return pr
	}
	return LOWEST
}

func (p *Parser) curPrecedence() precedence {
	if pr, ok := precedences[p.curTok.Type]; ok {
		return pr
	}
	return LOWEST
}
