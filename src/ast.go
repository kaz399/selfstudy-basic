/**************************************************************/
/*
   ast.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Stmt interface {
	stmtNode()
	String() string
}

type Expr interface {
	exprNode()
	String() string
}

// statements

type RemStmt struct{}

func (s *RemStmt) stmtNode()      {}
func (s *RemStmt) String() string { return "REM" }

type LetStmt struct {
	Name string
	Expr Expr
}

func (s *LetStmt) stmtNode()      {}
func (s *LetStmt) String() string { return fmt.Sprintf("%s = %s", s.Name, s.Expr.String()) }

type PrintStmt struct {
	Exprs []Expr // empty => PRINT only (blank line)
}

func (s *PrintStmt) stmtNode() {}
func (s *PrintStmt) String() string {
	if len(s.Exprs) == 0 {
		return "PRINT"
	}
	parts := make([]string, 0, len(s.Exprs))
	for _, e := range s.Exprs {
		parts = append(parts, e.String())
	}
	return "PRINT " + strings.Join(parts, ", ")
}

type InputStmt struct {
	Name string
}

func (s *InputStmt) stmtNode()      {}
func (s *InputStmt) String() string { return "INPUT " + s.Name }

type IfStmt struct {
	Cond     Expr
	ThenStmt Stmt // either ThenStmt or ThenLine is used
	ThenLine int
	HasLine  bool
}

func (s *IfStmt) stmtNode() {}
func (s *IfStmt) String() string {
	if s.HasLine {
		return fmt.Sprintf("IF %s THEN %d", s.Cond.String(), s.ThenLine)
	}
	return fmt.Sprintf("IF %s THEN %s", s.Cond.String(), s.ThenStmt.String())
}

type GotoStmt struct {
	Line int
}

func (s *GotoStmt) stmtNode()      {}
func (s *GotoStmt) String() string { return fmt.Sprintf("GOTO %d", s.Line) }

type EndStmt struct{}

func (s *EndStmt) stmtNode()      {}
func (s *EndStmt) String() string { return "END" }

// expressions

type NumberLit struct {
	Value float64
}

func (e *NumberLit) exprNode() {}
func (e *NumberLit) String() string {
	return strconv.FormatFloat(e.Value, 'g', -1, 64)
}

type StringLit struct {
	Value string
}

func (e *StringLit) exprNode()      {}
func (e *StringLit) String() string { return strconv.Quote(e.Value) }

type VarRef struct {
	Name string
}

func (e *VarRef) exprNode()      {}
func (e *VarRef) String() string { return e.Name }

type UnaryExpr struct {
	Op  string
	Rhs Expr
}

func (e *UnaryExpr) exprNode()      {}
func (e *UnaryExpr) String() string { return "(" + e.Op + e.Rhs.String() + ")" }

type BinaryExpr struct {
	Op  string
	Lhs Expr
	Rhs Expr
}

func (e *BinaryExpr) exprNode() {}
func (e *BinaryExpr) String() string {
	return "(" + e.Lhs.String() + " " + e.Op + " " + e.Rhs.String() + ")"
}
