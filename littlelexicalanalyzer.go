package main

import (
	"fmt"
	"log"
)

const (
	TT_ILLEGAL int = iota
	TT_EOF

	TT_SPACE

	TT_NEW_LINE

	TT_IDENT // main
	TT_INT   // 12345
	TT_CHAR  // 'a'
	TT_STR   // "abc"

	TT_ADD // +
	TT_SUB // -
	TT_MUL // *
	TT_QUO // /
	TT_REM // %

	TT_AND // &
	TT_OR  // |
	TT_XOR // ^

	TT_SHL // <<
	TT_SHR // >>

	TT_LAND // &&
	TT_LOR  // ||

	TT_ARROW // <-

	TT_EQL // ==
	TT_NEQ // !=
	TT_LSS // <
	TT_GTR // >
	TT_LEQ // <=
	TT_GEQ // >=

	TT_ASSIGN // =

	TT_LPAREN // (
	TT_RPAREN // )

	TT_COMMA // ,

	TT_FUNC
	TT_RETURN

	TT_IF
	TT_ELSE

	TT_WHILE
	TT_BREAK
	TT_CONTINUE

	TT_LET

	TT_END
)

type TokenData struct {
	Kype   int
	Offset int
	Buf    []byte
}

func checkTokenType(buf []byte) (int, int) {
	if len(buf) == 0 {
		return TT_EOF, 0
	} else if buf[0] == 0x0a {
		return TT_NEW_LINE, 1
	} else if buf[0] == 0x20 {
		return TT_SPACE, 1
	}

	var srcLine string

	for i, c := range buf {
		if c == 0x0a {
			srcLine = string(buf[:i])
			break
		}
	}

	tokType := TT_ILLEGAL
	bytesConsumed := 0

	TokTypeToStr := map[int]string{
		TT_ADD: "+",
		TT_SUB: "-",
		TT_MUL: "*",
		TT_QUO: "/",
		TT_REM: "%",

		TT_AND: "&",
		TT_OR:  "|",
		TT_XOR: "^",

		TT_SHL: "<<",
		TT_SHR: ">>",

		TT_LAND: "&&",
		TT_LOR:  "||",

		TT_ARROW: "<-",

		TT_EQL: "==",
		TT_NEQ: "!=",
		TT_LSS: "<",
		TT_GTR: ">",
		TT_LEQ: "<=",
		TT_GEQ: ">=",

		TT_ASSIGN: "=",

		TT_LPAREN: "(",
		TT_RPAREN: ")",

		TT_COMMA: ",",

		TT_FUNC:   "func",
		TT_RETURN: "return",

		TT_IF:   "if",
		TT_ELSE: "else",

		TT_WHILE:    "while",
		TT_BREAK:    "break",
		TT_CONTINUE: "continue",

		TT_LET: "let",

		TT_END: "end",
	}

	var prevTokStr string
	for curTokType, curTokStr := range TokTypeToStr {
		if (len(srcLine) >= len(curTokStr) && srcLine[:len(curTokStr)] == curTokStr) &&
			(tokType == TT_ILLEGAL || len(prevTokStr) < len(curTokStr)) {

			tokType = curTokType
			bytesConsumed = len(curTokStr)
			prevTokStr = curTokStr

		}
	}

	if tokType != TT_ILLEGAL {
		return tokType, bytesConsumed
	}

	isDigit := func(c byte) bool {
		return c >= 0x30 && c <= 0x39
	}

	isAplabet := func(c byte) bool {
		return (c >= 0x41 && c <= 0x5a) || (c >= 0x61 && c <= 0x7a) || (c == 0x5f)
	}

	i := 0

	if isAplabet(srcLine[i]) {

		tokType = TT_IDENT
		for (i < len(srcLine)) && (isAplabet(srcLine[i]) || isDigit(srcLine[i])) {
			i++
		}
		bytesConsumed = i

	} else if isDigit(srcLine[i]) {

		tokType = TT_INT
		for (i < len(srcLine)) && (isAplabet(srcLine[i]) || isDigit(srcLine[i])) {
			i++
		}
		bytesConsumed = i

	} else if srcLine[i] == 0x22 {

		for i++; i < len(srcLine); i++ {
			if srcLine[i] == 0x22 && srcLine[i-1] != 0x5c {
				tokType = TT_STR
				bytesConsumed = i + 1
				break
			}
		}

	} else if srcLine[i] == 0x27 {

		for i++; i < len(srcLine); i++ {
			if srcLine[i] == 0x27 && srcLine[i-1] != 0x5c {
				tokType = TT_CHAR
				bytesConsumed = i + 1
				break
			}
		}

	}

	return tokType, bytesConsumed
}

func LexicalAnalyzer(buf []byte) {
	for {
		tokType, bytesConsumed := checkTokenType(buf)
		fmt.Println(tokType, bytesConsumed)

		if tokType == TT_EOF {
			break
		}

		if tokType == TT_ILLEGAL {
			log.Fatalln(string(buf))
		}

		buf = buf[bytesConsumed:]
	}

}
