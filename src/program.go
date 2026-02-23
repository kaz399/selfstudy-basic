/**************************************************************/
/*
   parser.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

import (
	"sort"
)

type Program struct {
	Source map[int]string // for LIST
	Stmts  map[int]Stmt   // for execution
}

func NewProgram() *Program {
	return &Program{
		Source: map[int]string{},
		Stmts:  map[int]Stmt{},
	}
}

func (p *Program) Clear() {
	p.Source = map[int]string{}
	p.Stmts = map[int]Stmt{}
}

func (p *Program) SetLine(lineNo int, src string, stmt Stmt) {
	p.Source[lineNo] = src
	p.Stmts[lineNo] = stmt
}

func (p *Program) DeleteLine(lineNo int) {
	delete(p.Source, lineNo)
	delete(p.Stmts, lineNo)
}

func (p *Program) OrderedLines() []int {
	keys := make([]int, 0, len(p.Stmts))
	for k := range p.Stmts {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
