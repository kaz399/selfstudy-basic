/**************************************************************/
/*
   parser.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ValueKind int

const (
	ValNumber ValueKind = iota
	ValString
)

type Value struct {
	Kind ValueKind
	Num  float64
	Str  string
}

func NumberValue(n float64) Value { return Value{Kind: ValNumber, Num: n} }
func StringValue(s string) Value  { return Value{Kind: ValString, Str: s} }

func (v Value) String() string {
	switch v.Kind {
	case ValNumber:
		return strconv.FormatFloat(v.Num, 'g', -1, 64)
	case ValString:
		return v.Str
	default:
		return "<invalid>"
	}
}

type Env struct {
	NumVars map[string]float64
	StrVars map[string]string
}

func NewEnv() *Env {
	return &Env{
		NumVars: map[string]float64{},
		StrVars: map[string]string{},
	}
}

func (e *Env) Get(name string) Value {
	name = strings.ToUpper(name)
	if strings.HasSuffix(name, "$") {
		return StringValue(e.StrVars[name]) // default value is ""
	}
	return NumberValue(e.NumVars[name]) // default value is 0
}

func (e *Env) Set(name string, v Value) error {
	name = strings.ToUpper(name)
	isStr := strings.HasSuffix(name, "$")
	if isStr && v.Kind != ValString {
		return fmt.Errorf("type mismatch: %s is string variable", name)
	}
	if !isStr && v.Kind != ValNumber {
		return fmt.Errorf("type mismatch: %s is numeric variable", name)
	}
	if isStr {
		e.StrVars[name] = v.Str
	} else {
		e.NumVars[name] = v.Num
	}
	return nil
}

type Interpreter struct {
	Prog   *Program
	Env    *Env
	In     *bufio.Reader
	Out    io.Writer
	MaxOps int // infinit loop limitation (0: unlimited)
}

func NewInterpreter(prog *Program, in *bufio.Reader, out io.Writer) *Interpreter {
	return &Interpreter{
		Prog:   prog,
		Env:    NewEnv(),
		In:     in,
		Out:    out,
		MaxOps: 1_000_000,
	}
}

func (it *Interpreter) ResetEnv() {
	it.Env = NewEnv()
}

func (it *Interpreter) Run() error {
	order := it.Prog.OrderedLines()
	if len(order) == 0 {
		return nil
	}
	lineIndex := make(map[int]int, len(order))
	for i, ln := range order {
		lineIndex[ln] = i
	}

	pc := 0
	ops := 0
	for pc >= 0 && pc < len(order) {
		if it.MaxOps > 0 {
			ops++
			if ops > it.MaxOps {
				return fmt.Errorf("runtime error: operation limit exceeded (possible infinite loop)")
			}
		}
		lineNo := order[pc]
		stmt := it.Prog.Stmts[lineNo]

		nextPC, end, err := it.execStmt(stmt, lineNo, lineIndex, pc)
		if err != nil {
			return fmt.Errorf("runtime error at line %d: %w", lineNo, err)
		}
		if end {
			return nil
		}
		pc = nextPC
	}
	return nil
}

func (it *Interpreter) execStmt(stmt Stmt, lineNo int, lineIndex map[int]int, pc int) (int, bool, error) {
	nextPC := pc + 1

	switch s := stmt.(type) {
	case *RemStmt:
		return nextPC, false, nil

	case *LetStmt:
		v, err := it.evalExpr(s.Expr)
		if err != nil {
			return 0, false, err
		}
		if err := it.Env.Set(s.Name, v); err != nil {
			return 0, false, err
		}
		return nextPC, false, nil

	case *PrintStmt:
		if len(s.Exprs) == 0 {
			fmt.Fprintln(it.Out)
			return nextPC, false, nil
		}
		parts := make([]string, 0, len(s.Exprs))
		for _, e := range s.Exprs {
			v, err := it.evalExpr(e)
			if err != nil {
				return 0, false, err
			}
			parts = append(parts, v.String())
		}
		fmt.Fprintln(it.Out, strings.Join(parts, " "))
		return nextPC, false, nil

	case *InputStmt:
		fmt.Fprint(it.Out, "? ")
		line, err := it.In.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, false, err
		}
		line = strings.TrimRight(line, "\r\n")
		if strings.HasSuffix(strings.ToUpper(s.Name), "$") {
			if err := it.Env.Set(s.Name, StringValue(line)); err != nil {
				return 0, false, err
			}
			return nextPC, false, nil
		}
		n, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
		if err != nil {
			return 0, false, fmt.Errorf("INPUT expects number")
		}
		if err := it.Env.Set(s.Name, NumberValue(n)); err != nil {
			return 0, false, err
		}
		return nextPC, false, nil

	case *IfStmt:
		cond, err := it.evalExpr(s.Cond)
		if err != nil {
			return 0, false, err
		}
		if cond.Kind != ValNumber {
			return 0, false, fmt.Errorf("IF condition must be numeric")
		}
		if cond.Num == 0 {
			return nextPC, false, nil
		}

		if s.HasLine {
			idx, ok := lineIndex[s.ThenLine]
			if !ok {
				return 0, false, fmt.Errorf("undefined line %d", s.ThenLine)
			}
			return idx, false, nil
		}

		return it.execStmt(s.ThenStmt, lineNo, lineIndex, pc)

	case *GotoStmt:
		idx, ok := lineIndex[s.Line]
		if !ok {
			return 0, false, fmt.Errorf("undefined line %d", s.Line)
		}
		return idx, false, nil

	case *EndStmt:
		return 0, true, nil

	default:
		return 0, false, fmt.Errorf("unknown statement type %T", stmt)
	}
}

func (it *Interpreter) evalExpr(e Expr) (Value, error) {
	switch x := e.(type) {
	case *NumberLit:
		return NumberValue(x.Value), nil
	case *StringLit:
		return StringValue(x.Value), nil
	case *VarRef:
		return it.Env.Get(x.Name), nil

	case *UnaryExpr:
		v, err := it.evalExpr(x.Rhs)
		if err != nil {
			return Value{}, err
		}
		if v.Kind != ValNumber {
			return Value{}, fmt.Errorf("unary %s requires number", x.Op)
		}
		switch x.Op {
		case "+":
			return v, nil
		case "-":
			return NumberValue(-v.Num), nil
		default:
			return Value{}, fmt.Errorf("unsupported unary op %s", x.Op)
		}

	case *BinaryExpr:
		lv, err := it.evalExpr(x.Lhs)
		if err != nil {
			return Value{}, err
		}
		rv, err := it.evalExpr(x.Rhs)
		if err != nil {
			return Value{}, err
		}
		return evalBinary(x.Op, lv, rv)

	default:
		return Value{}, fmt.Errorf("unknown expression type %l", e)
	}
}

func evalBinary(op string, l, r Value) (Value, error) {
	switch op {
	case "+", "-", "*", "/":
		if l.Kind != ValNumber || r.Kind != ValNumber {
			return Value{}, fmt.Errorf("arithmetic requires numbers")
		}
		switch op {
		case "+":
			return NumberValue(l.Num + r.Num), nil
		case "-":
			return NumberValue(l.Num - r.Num), nil
		case "*":
			return NumberValue(l.Num * r.Num), nil
		case "/":
			if r.Num == 0 {
				return Value{}, fmt.Errorf("division by zero")
			}
			return NumberValue(l.Num / r.Num), nil
		}
	case "=", "<>":
		// number-number or string-string
		if l.Kind != r.Kind {
			return Value{}, fmt.Errorf("type mismatch in comparison")
		}
		var ok bool
		if l.Kind == ValNumber {
			ok = (l.Num == r.Num)
		} else {
			ok = (l.Str == r.Str)
		}
		if op == "<>" {
			ok = !ok
		}
		if ok {
			return NumberValue(1), nil
		}
		return NumberValue(0), nil

	case "<", "<=", ">", ">=":
		if l.Kind != ValNumber || r.Kind != ValNumber {
			return Value{}, fmt.Errorf("ordered comparison requires numbers")
		}
		var ok bool
		switch op {
		case "<":
			ok = l.Num < r.Num
		case "<=":
			ok = l.Num <= r.Num
		case ">":
			ok = l.Num > r.Num
		case ">=":
			ok = l.Num > -r.Num
		}
		if ok {
			return NumberValue(1), nil
		}
		return NumberValue(0), nil
	}
	return Value{}, fmt.Errorf("unsupported operator %q", op)
}
