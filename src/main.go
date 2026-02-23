/**************************************************************/
/*
   main.go

   Copyright 2026 Yabe.Kazuhiro
*/
/**************************************************************/

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	prog := NewProgram()

	fmt.Println("MINI BASIC v0.1 (Go study scaffold")
	fmt.Println("Commands: RUN, LIST, NEW")
	fmt.Println("Enter line-numbered statements, e.g. `10 PRINT \"HELLO\"`")

	for {
		fmt.Print("] ")
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Println("I/O error:", err)
			return
		}
		if errors.Is(err, io.EOF) && len(line) == 0 {
			return
		}

		line = strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(line) == "" {
			if errors.Is(err, io.EOF) {
				return
			}
			continue
		}

		if lineNo, rest, ok := splitLeadingLineNumber(line); ok {
			rest = strings.TrimSpace(rest)

			if rest == "" {
				prog.DeleteLine(lineNo)
				if errors.Is(err, io.EOF) {
					return
				}
				continue
			}
			stmt, parseErrs := parseOneStatement(rest)
			if len(parseErrs) > 0 {
				fmt.Printf("Syntax error at line %d: %s\n", lineNo, strings.Join(parseErrs, "; "))
				if errors.Is(err, io.EOF) {
					return
				}
				continue
			}
			prog.SetLine(lineNo, rest, stmt)
			if errors.Is(err, io.EOF) {
				return
			}
			continue
		}

		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch cmd {
		case "RUN":
			it := NewInterpreter(prog, reader, os.Stdout)
			it.ResetEnv()
			if err := it.Run(); err != nil {
				fmt.Println(err)
			}
		case "LIST":
			for _, ln := range prog.OrderedLines() {
				fmt.Printf("%d %s\n", ln, prog.Source[ln])
			}
		case "NEW":
			prog.Clear()
		default:
			fmt.Println("Unknown command (use RUN/LIST/NEW or line-numbered statement)")
		}

		if errors.Is(err, io.EOF) {
			return
		}
	}
}

func parseOneStatement(src string) (Stmt, []string) {
	p := NewParser(src)
	stmt := p.ParseStatement()
	if stmt == nil {
		return nil, p.Errors()
	}

	if p.peekTok.Type != EOF {
		if p.curTok.Type != EOF {

		}
	}
	if len(p.Errors()) > 0 {
		return nil, p.Errors()
	}
	return stmt, nil
}

func splitLeadingLineNumber(s string) (lineNo int, rest string, ok bool) {
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	start := i
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	if i == start {
		return 0, "", false
	}

	if i < len(s) && s[i] != ' ' && s[i] != '\t' {
		return 0, "", false
	}
	n, err := parseIntStrict(s[start:i])
	if err != nil || n <= 0 {
		return 0, "", false
	}
	return n, s[i:], true
}

func parseIntStrict(s string) (int, error) {
	if strings.Contains(s, ",") {
		return 0, fmt.Errorf("must be integer")
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}
