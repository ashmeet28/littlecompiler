package main

const (
	T_ILLEGAL int = iota
	T_EOF

	T_NEW_LINE

	T_IDENT // main
	T_INT   // 12345
	T_CHAR  // 'a'
	T_STR   // "abc"

	T_ADD // +
	T_SUB // -
	T_MUL // *
	T_QUO // /
	T_REM // %

	T_AND // &
	T_OR  // |
	T_XOR // ^

	T_SHL // <<
	T_SHR // >>

	T_LAND // &&
	T_LOR  // ||

	T_ARROW // <-

	T_EQL // ==
	T_NEQ // !=
	T_LSS // <
	T_GTR // >
	T_LEQ // <=
	T_GEQ // >=

	T_ASSIGN // =

	T_LPAREN // (
	T_RPAREN // )

	T_COMMA // ,

	T_FUNC
	T_RETURN

	T_IF
	T_ELSE

	T_WHILE
	T_BREAK
	T_CONTINUE

	T_LET

	T_END
)

func LexicalAnalyzer(buf []byte) {

}
