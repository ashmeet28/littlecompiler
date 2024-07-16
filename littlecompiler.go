package main

import (
	"fmt"
	"os"
	"strconv"
)

func PrintFatalCompilationError(l int) {
	s := "Compilation error"
	if l != 0 {
		s = s + " " + "(" + "line" + " " + strconv.FormatInt(int64(l), 10) + ")"
	}
	fmt.Println(s)
	os.Exit(1)
}

func main() {
	buf, _ := os.ReadFile(os.Args[1])
	LexicalAnalyzer(buf)
}
