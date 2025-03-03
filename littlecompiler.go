package main

import (
	"fmt"
	"log"
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
	sourceCodeFilePath := os.Args[1]
	bytecodeFilePath := os.Args[2]

	data, err := os.ReadFile(sourceCodeFilePath)

	if err != nil {
		log.Fatal(err)
	}

	toks := LexicalAnalyzer(append(data, 0x0a))

	tn := SyntaxAnalyzer(toks)

	// PrintTreeNode(tn, 4)

	bytecode := BytecodeGenerator(tn)

	if err := os.WriteFile(bytecodeFilePath, bytecode, 0666); err != nil {
		log.Fatal(err)
	}
}
