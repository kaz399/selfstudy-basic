/**************************************************************/
/*
   main.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

import (
	"fmt"
)

func main() {
	input := "10 + 20"
	lexer := NewLexer(input)
	parser := NewParser(lexer)
	ast := parser.ParseExpression()

	fmt.Printf("input: %q\n\n", input)

	if infix, ok := ast.(*InfixExpression); ok {
		fmt.Printf("AST structure\n")
		fmt.Printf("Parent Node: %s\n", infix.Operator)

		left := infix.Left.(*IntegerLiteral)
		right := infix.Right.(*IntegerLiteral)

		fmt.Printf(" Left  Node: %d\n", left.Value)
		fmt.Printf(" Right Node: %d\n", right.Value)
	}
}
