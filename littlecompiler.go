package main

import "os"

func main() {
	buf, _ := os.ReadFile("temp_file")
	LexicalAnalyzer(buf)
}
