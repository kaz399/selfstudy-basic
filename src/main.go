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
	lxer := NewLexer(input)

	fmt.Printf("input: %q\n\n", input)

	for {
		tok := lxer.NextToken()
		fmt.Printf("{Type: %-7s Literal: %q}\n", tok.Type, tok.Literal)

		if tok.Type == EOF {
			break
		}
	}
}
