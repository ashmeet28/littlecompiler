package main

import (
	"fmt"
	"os"
	"strconv"
)

func PrintErrorAndExit(l int) {
	s := "Compilation error"
	if l != 0 {
		s = s + " " + "(" + "line" + " " + strconv.FormatInt(int64(l), 10) + ")"
	}
	fmt.Println(s)
	os.Exit(1)
}

func main() {
	buf, _ := os.ReadFile(os.Args[1])
	toks := LexicalAnalyzer(append(buf, 0x0a))

	tn := SyntaxAnalyzer(toks)
	PrintTreeNode(tn, 4)

	BytecodeGenerator(tn)
}
