package main

import "fmt"

type TreeNodeType int

const (
	TNT_ILLEGAL TreeNodeType = iota

	TNT_ROOT

	TNT_FUNC_LIST
	TNT_FUNC

	TNT_FUNC_IDENT
	TNT_FUNC_SIG
	TNT_FUNC_PARAM_LIST
	TNT_FUNC_PARAM
	TNT_FUNC_PARAM_IDENT
	TNT_FUNC_PARAM_TYPE
	TNT_FUNC_RETURN_TYPE

	TNT_STMT_LIST

	TNT_STMT_DECL
	TNT_STMT_DECL_IDENT
	TNT_STMT_DECL_TYPE

	TNT_STMT_EXPR
	TNT_STMT_ASSIGN
	TNT_STMT_STORE_STRING
	TNT_STMT_STRING

	TNT_STMT_WHILE
	TNT_STMT_IF
	TNT_STMT_ELSE

	TNT_STMT_RETURN
	TNT_STMT_BREAK
	TNT_STMT_CONTINUE

	TNT_EXPR
	TNT_EXPR_INT
	TNT_EXPR_FUNC
	TNT_EXPR_FUNC_PARM_LIST
	TNT_EXPR_FUNC_PARM
	TNT_EXPR_INT_LIT
	TNT_EXPR_NEG_INT_LIT
	TNT_EXPR_CHAR
	TNT_EXPR_BINARY
)

var TreeNodeTypeNames = map[TreeNodeType]string{
	TNT_ILLEGAL:             "ILLEGAL",
	TNT_ROOT:                "ROOT",
	TNT_FUNC_LIST:           "FUNC_LIST",
	TNT_FUNC:                "FUNC",
	TNT_FUNC_IDENT:          "FUNC_IDENT",
	TNT_FUNC_SIG:            "FUNC_SIG",
	TNT_FUNC_PARAM_LIST:     "FUNC_PARAM_LIST",
	TNT_FUNC_PARAM:          "FUNC_PARAM",
	TNT_FUNC_PARAM_IDENT:    "FUNC_PARAM_IDENT",
	TNT_FUNC_PARAM_TYPE:     "FUNC_PARAM_TYPE",
	TNT_FUNC_RETURN_TYPE:    "FUNC_RETURN_TYPE",
	TNT_STMT_LIST:           "STMT_LIST",
	TNT_STMT_DECL:           "STMT_DECL",
	TNT_STMT_DECL_IDENT:     "STMT_DECL_IDENT",
	TNT_STMT_DECL_TYPE:      "STMT_DECL_TYPE",
	TNT_STMT_EXPR:           "STMT_EXPR",
	TNT_STMT_ASSIGN:         "STMT_ASSIGN",
	TNT_STMT_STORE_STRING:   "STMT_STORE_STRING",
	TNT_STMT_STRING:         "STMT_STRING",
	TNT_STMT_WHILE:          "STMT_WHILE",
	TNT_STMT_IF:             "STMT_IF",
	TNT_STMT_ELSE:           "STMT_ELSE",
	TNT_STMT_RETURN:         "STMT_RETURN",
	TNT_STMT_BREAK:          "STMT_BREAK",
	TNT_STMT_CONTINUE:       "STMT_CONTINUE",
	TNT_EXPR:                "EXPR",
	TNT_EXPR_INT:            "EXPR_INT",
	TNT_EXPR_FUNC:           "EXPR_FUNC",
	TNT_EXPR_FUNC_PARM_LIST: "EXPR_FUNC_PARM_LIST",
	TNT_EXPR_FUNC_PARM:      "EXPR_FUNC_PARM",
	TNT_EXPR_INT_LIT:        "EXPR_INT_LIT",
	TNT_EXPR_NEG_INT_LIT:    "EXPR_NEG_INT_LIT",
	TNT_EXPR_CHAR:           "EXPR_CHAR",
	TNT_EXPR_BINARY:         "EXPR_BINARY",
}

type TreeNode struct {
	Kype     TreeNodeType
	Children []TreeNode
	Tok      TokenData
}

var curToks []TokenData

func peekTok() TokenData {
	if len(curToks) == 0 {
		PrintErrorAndExit(0)
	}
	return curToks[0]
}

func advanceTok() TokenData {
	if len(curToks) == 0 {
		PrintErrorAndExit(0)
	}
	tok := curToks[0]
	curToks = curToks[1:]
	return tok
}

func consumeTok(tokType TokenType) TokenData {
	if len(curToks) == 0 {
		PrintErrorAndExit(0)
	}
	tok := curToks[0]
	if tok.Kype != tokType {
		PrintErrorAndExit(tok.LineNumber)
	}
	curToks = curToks[1:]
	return tok
}

func matchTok(tokTypes ...TokenType) bool {
	for _, curTokType := range tokTypes {
		if curTokType == peekTok().Kype {
			return true
		}
	}
	return false
}

func matchBinaryTok() bool {
	return matchTok(TT_ADD, TT_SUB,
		TT_MUL, TT_QUO, TT_REM,
		TT_AND, TT_OR, TT_XOR,
		TT_SHL, TT_SHR,
		TT_LAND, TT_LOR,
		TT_EQL, TT_NEQ,
		TT_LSS, TT_GTR,
		TT_LEQ, TT_GEQ)
}

func parseFuncList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_LIST

	for matchTok(TT_FUNC) {
		tn.Children = append(tn.Children, parseFunc())
	}
	consumeTok(TT_EOF)

	return tn
}

func parseFunc() TreeNode {
	consumeTok(TT_FUNC)

	var tn TreeNode
	tn.Kype = TNT_FUNC

	tn.Children = append(tn.Children, parseFuncIdent())
	tn.Children = append(tn.Children, parseFuncSig())

	consumeTok(TT_NEW_LINE)

	tn.Children = append(tn.Children, parseStmtList())

	consumeTok(TT_END)
	consumeTok(TT_NEW_LINE)

	return tn
}

func parseFuncIdent() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_IDENT
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func parseFuncSig() TreeNode {
	consumeTok(TT_LPAREN)

	var tn TreeNode
	tn.Kype = TNT_FUNC_SIG

	if matchTok(TT_IDENT) {
		tn.Children = append(tn.Children, parseFuncParamList())
	}

	consumeTok(TT_RPAREN)

	if matchTok(TT_IDENT) {
		tn.Children = append(tn.Children, parseFuncReturnType())
	}

	return tn
}

func parseFuncParamList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_LIST

	tn.Children = append(tn.Children, parseFuncParam())
	for matchTok(TT_COMMA) {
		consumeTok(TT_COMMA)
		tn.Children = append(tn.Children, parseFuncParam())
	}

	return tn
}

func parseFuncParam() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM

	tn.Children = append(tn.Children, parseFuncParamIdent())
	tn.Children = append(tn.Children, parseFuncParamType())

	return tn
}

func parseFuncParamIdent() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_IDENT
	tn.Tok = consumeTok(TT_IDENT)
	return tn
}

func parseFuncParamType() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_TYPE
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func parseFuncReturnType() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_RETURN_TYPE
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func parseStmtList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_LIST

	for matchTok(TT_LET, TT_WHILE, TT_IF, TT_RETURN, TT_BREAK, TT_CONTINUE, TT_IDENT, TT_LPAREN) {
		tn.Children = append(tn.Children, parseStmt())
	}

	return tn
}

func parseStmt() TreeNode {
	if matchTok(TT_LET) {
		return parseStmtDecl()
	} else if matchTok(TT_WHILE) {
		return parseStmtWhile()
	} else if matchTok(TT_IF) {
		return parseStmtIf()
	} else if matchTok(TT_RETURN) {
		return parseStmtReturn()
	} else if matchTok(TT_BREAK) {
		return parseStmtBreak()
	} else if matchTok(TT_CONTINUE) {
		return parseStmtContinue()
	} else {
		exprTreeNode := parseExpr()

		if matchTok(TT_ASSIGN) {
			return parseStmtAssign(exprTreeNode)
		} else if matchTok(TT_ARROW) {
			return parseStmtStoreString(exprTreeNode)
		} else {
			return parseStmtExpr(exprTreeNode)
		}
	}
}

func parseStmtDecl() TreeNode {
	consumeTok(TT_LET)

	var tn TreeNode
	tn.Kype = TNT_STMT_DECL

	tn.Children = append(tn.Children, parseStmtDeclIdent())
	tn.Children = append(tn.Children, parseStmtDeclType())

	consumeTok(TT_NEW_LINE)

	return tn
}

func parseStmtDeclIdent() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_DECL_IDENT
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func parseStmtDeclType() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_DECL_TYPE
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func parseStmtExpr(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_EXPR

	consumeTok(TT_NEW_LINE)

	tn.Children = append(tn.Children, exprTreeNode)
	return tn
}

func parseStmtAssign(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_ASSIGN

	consumeTok(TT_ASSIGN)

	tn.Children = append(tn.Children, exprTreeNode)
	tn.Children = append(tn.Children, parseExpr())

	consumeTok(TT_NEW_LINE)
	return tn
}

func parseStmtStoreString(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_STORE_STRING

	consumeTok(TT_ARROW)

	tn.Children = append(tn.Children, exprTreeNode)
	tn.Children = append(tn.Children, parseStmtString())

	consumeTok(TT_NEW_LINE)
	return tn
}

func parseStmtString() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_STRING
	tn.Tok = consumeTok(TT_STR)

	return tn
}

func parseStmtWhile() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_WHILE

	consumeTok(TT_WHILE)

	tn.Children = append(tn.Children, parseExpr())

	consumeTok(TT_NEW_LINE)

	tn.Children = append(tn.Children, parseStmtList())

	consumeTok(TT_END)
	consumeTok(TT_NEW_LINE)

	return tn
}

func parseStmtIf() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_IF

	consumeTok(TT_IF)

	tn.Children = append(tn.Children, parseExpr())

	consumeTok(TT_NEW_LINE)

	tn.Children = append(tn.Children, parseStmtList())

	if matchTok(TT_ELSE) {
		tn.Children = append(tn.Children, parseStmtElse())
	} else {
		consumeTok(TT_END)
		consumeTok(TT_NEW_LINE)
	}

	return tn
}

func parseStmtElse() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_ELSE

	consumeTok(TT_ELSE)

	if matchTok(TT_IF) {
		tn.Children = append(tn.Children, parseStmtIf())
	} else {
		consumeTok(TT_NEW_LINE)

		tn.Children = append(tn.Children, parseStmtList())

		consumeTok(TT_END)
		consumeTok(TT_NEW_LINE)
	}

	return tn
}

func parseExpr() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR
	tn.Children = append(tn.Children, parseExprCont(true))

	return tn
}

func parseExprCont(doesFollowBinary bool) TreeNode {
	tn := parseExprUnary()

	for doesFollowBinary && matchBinaryTok() {
		tn = parseExprBinary(tn)
	}

	return tn
}

func parseExprUnary() TreeNode {
	var tn TreeNode

	if matchTok(TT_IDENT) {
		tn.Tok = consumeTok(TT_IDENT)
		if matchTok(TT_LPAREN) {
			tn.Kype = TNT_EXPR_FUNC
			tn.Children = append(tn.Children, parseExprUnaryFuncParmList())
		} else {
			tn.Kype = TNT_EXPR_INT
		}
	} else {
		consumeTok(TT_LPAREN)
		tn = parseExprCont(true)
		consumeTok(TT_RPAREN)
	}

	return tn
}

func parseExprUnaryFuncParmList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_FUNC_PARM_LIST

	consumeTok(TT_LPAREN)

	if matchTok(TT_IDENT, TT_LPAREN, TT_INT, TT_CHAR, TT_SUB) {
		tn.Children = append(tn.Children, parseExprUnaryFuncParm())
		for matchTok(TT_COMMA) {
			consumeTok(TT_COMMA)
			tn.Children = append(tn.Children, parseExprUnaryFuncParm())
		}
	}

	consumeTok(TT_RPAREN)

	return tn
}

func parseExprUnaryFuncParm() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_FUNC_PARM

	if matchTok(TT_INT) {
		tn.Children = append(tn.Children, parseExprUnaryFuncParmInt())
	} else if matchTok(TT_SUB) {
		tn.Children = append(tn.Children, parseExprUnaryFuncParmNegInt())
	} else if matchTok(TT_CHAR) {
		tn.Children = append(tn.Children, parseExprUnaryFuncParmChar())
	} else {
		tn.Children = append(tn.Children, parseExpr())
	}

	return tn
}

func parseExprUnaryFuncParmInt() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_INT_LIT
	tn.Tok = consumeTok(TT_INT)
	return tn
}

func parseExprUnaryFuncParmNegInt() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_NEG_INT_LIT
	consumeTok(TT_SUB)
	tn.Tok = consumeTok(TT_INT)
	return tn
}

func parseExprUnaryFuncParmChar() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_CHAR
	tn.Tok = consumeTok(TT_CHAR)
	return tn
}

func parseExprBinary(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_BINARY
	tn.Tok = advanceTok()

	tn.Children = append(tn.Children, exprTreeNode)
	tn.Children = append(tn.Children, parseExprCont(false))
	return tn
}

func parseStmtReturn() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_RETURN
	consumeTok(TT_RETURN)

	if matchTok(TT_IDENT, TT_LPAREN) {
		tn.Children = append(tn.Children, parseExpr())
	}

	consumeTok(TT_NEW_LINE)
	return tn

}

func parseStmtBreak() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_BREAK
	consumeTok(TT_BREAK)
	consumeTok(TT_NEW_LINE)
	return tn
}

func parseStmtContinue() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_CONTINUE
	consumeTok(TT_CONTINUE)
	consumeTok(TT_NEW_LINE)
	return tn
}

func PrintTreeNode(tn TreeNode, level int) {
	var s string
	for i := 0; i < level; i++ {
		s = s + " "
	}

	fmt.Println(s, "->", TreeNodeTypeNames[tn.Kype], tn.Tok)

	for _, child := range tn.Children {
		PrintTreeNode(child, level+4)
	}
}

func normalizeWholeTree(tn TreeNode) TreeNode {
	funcListTreeNode := tn.Children[0]

	for _, c := range funcListTreeNode.Children {
		var newTreeNode TreeNode
		newTreeNode.Kype = TNT_STMT_RETURN

		stmtListNode := c.Children[2]
		stmtListNode.Children = append(stmtListNode.Children, newTreeNode)
		c.Children[2] = stmtListNode
	}

	return tn
}

func SyntaxAnalyzer(toks []TokenData) TreeNode {
	curToks = toks

	var tn TreeNode

	tn.Kype = TNT_ROOT

	tn.Children = append(tn.Children, parseFuncList())

	tn = normalizeWholeTree(tn)

	return tn
}
