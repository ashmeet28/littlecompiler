package main

type TokenType int

const (
	TT_ILLEGAL TokenType = iota

	TT_EOF

	TT_SPACE

	TT_NEW_LINE

	TT_COMMENT

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
	Kype       TokenType
	LineNumber int
	Buf        []byte
}

func checkTokenType(buf []byte) (TokenType, int) {
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

	if srcLine[0] == 0x23 {
		return TT_COMMENT, len(srcLine)
	}

	tokType := TT_ILLEGAL
	bytesConsumed := 0

	TokTypeToStr := map[TokenType]string{
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

		i++
		for i < len(srcLine) {
			if srcLine[i] == 0x5c {
				i++
				if i < len(srcLine) {
					if srcLine[i] == 0x5c || srcLine[i] == 0x22 {
						i++
					} else {
						break
					}
				}
			} else if srcLine[i] == 0x22 {
				i++
				tokType = TT_STR
				bytesConsumed = i
				break
			} else {
				i++
			}
		}

	} else if srcLine[i] == 0x27 {

		i++
		for i < len(srcLine) {
			if srcLine[i] == 0x5c {
				i++
				if i < len(srcLine) {
					if srcLine[i] == 0x5c || srcLine[i] == 0x27 {
						i++
					} else {
						break
					}
				}
			} else if srcLine[i] == 0x27 {
				i++
				tokType = TT_CHAR
				bytesConsumed = i
				break
			} else {
				i++
			}
		}

	}

	return tokType, bytesConsumed
}

func checkForInvalidBytes(buf []byte) {
	curLineNum := 1

	for _, c := range buf {
		if c == 0x0a {
			curLineNum++
		} else if c < 0x20 || c > 0x7e {
			PrintErrorAndExit(curLineNum)
		}
	}
}

func filterNewLineTokens(toks []TokenData) []TokenData {
	var filteredToks []TokenData

	var prevTok TokenData

	prevTok.Kype = TT_ILLEGAL

	allowedPrevTokTypes := []TokenType{TT_IDENT,
		TT_STR, TT_RPAREN, TT_RETURN,
		TT_ELSE, TT_BREAK, TT_CONTINUE,
		TT_END}

	for _, tok := range toks {
		if tok.Kype == TT_NEW_LINE {
			for _, allowedPrevTokType := range allowedPrevTokTypes {
				if allowedPrevTokType == prevTok.Kype {
					filteredToks = append(filteredToks, tok)
				}
			}
		} else {
			filteredToks = append(filteredToks, tok)
		}

		prevTok = tok
	}

	return filteredToks
}

func generateTokens(buf []byte) []TokenData {
	curLineNum := 1

	var toks []TokenData

	for {
		tokType, bytesConsumed := checkTokenType(buf)

		if tokType == TT_ILLEGAL {
			PrintErrorAndExit(curLineNum)
		}

		var tok TokenData
		tok.Kype = tokType
		tok.LineNumber = curLineNum

		if tokType == TT_IDENT || tokType == TT_INT ||
			tokType == TT_CHAR || tokType == TT_STR {
			tok.Buf = buf[:bytesConsumed]
		}

		if tokType != TT_SPACE && tokType != TT_COMMENT {
			toks = append(toks, tok)
		}

		if tokType == TT_EOF {
			break
		} else if tokType == TT_NEW_LINE {
			curLineNum++
		}

		buf = buf[bytesConsumed:]
	}

	toks = filterNewLineTokens(toks)

	return toks
}

func LexicalAnalyzer(buf []byte) []TokenData {
	checkForInvalidBytes(buf)
	toks := generateTokens(buf)
	return toks
}
