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
	TNT_EXPR_VAR
	TNT_EXPR_FUNC
	TNT_EXPR_FUNC_PARM_LIST
	TNT_EXPR_FUNC_PARM
	TNT_EXPR_INT
	TNT_EXPR_NEG_INT
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
	TNT_EXPR_VAR:            "EXPR_VAR",
	TNT_EXPR_FUNC:           "EXPR_FUNC",
	TNT_EXPR_FUNC_PARM_LIST: "EXPR_FUNC_PARM_LIST",
	TNT_EXPR_FUNC_PARM:      "EXPR_FUNC_PARM",
	TNT_EXPR_INT:            "EXPR_INT",
	TNT_EXPR_NEG_INT:        "EXPR_NEG_INT",
	TNT_EXPR_CHAR:           "EXPR_CHAR",
	TNT_EXPR_BINARY:         "EXPR_BINARY",
}

type TreeNode struct {
	Kype     TreeNodeType
	Children []TreeNode
	Tok      TokenData
}

var peekTok func() TokenData

var advanceTok func() TokenData

var consumeTok func(TokenType) TokenData

var matchTok func(...TokenType) bool

var matchBinaryTok func() bool

func handleFuncList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_LIST

	for matchTok(TT_FUNC) {
		tn.Children = append(tn.Children, handleFunc())
	}
	consumeTok(TT_EOF)

	return tn
}

func handleFunc() TreeNode {
	consumeTok(TT_FUNC)

	var tn TreeNode
	tn.Kype = TNT_FUNC

	tn.Children = append(tn.Children, handleFuncIdent())
	tn.Children = append(tn.Children, handleFuncSig())

	consumeTok(TT_NEW_LINE)

	tn.Children = append(tn.Children, handleStmtList())

	consumeTok(TT_END)
	consumeTok(TT_NEW_LINE)

	return tn
}

func handleFuncIdent() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_IDENT
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func handleFuncSig() TreeNode {
	consumeTok(TT_LPAREN)

	var tn TreeNode
	tn.Kype = TNT_FUNC_SIG

	if matchTok(TT_IDENT) {
		tn.Children = append(tn.Children, handleFuncParamList())
	}

	consumeTok(TT_RPAREN)

	if matchTok(TT_IDENT) {
		tn.Children = append(tn.Children, handleFuncReturnType())
	}

	return tn
}

func handleFuncParamList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_LIST

	tn.Children = append(tn.Children, handleFuncParam())
	for matchTok(TT_COMMA) {
		consumeTok(TT_COMMA)
		tn.Children = append(tn.Children, handleFuncParam())
	}

	return tn
}

func handleFuncParam() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM

	tn.Children = append(tn.Children, handleFuncParamIdent())
	tn.Children = append(tn.Children, handleFuncParamType())

	return tn
}

func handleFuncParamIdent() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_IDENT
	tn.Tok = consumeTok(TT_IDENT)
	return tn
}

func handleFuncParamType() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_TYPE
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func handleFuncReturnType() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_RETURN_TYPE
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func handleStmtList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_LIST

	for matchTok(TT_LET, TT_IDENT, TT_LPAREN, TT_WHILE, TT_IF, TT_RETURN) {
		tn.Children = append(tn.Children, handleStmt())
	}

	return tn
}

func handleStmt() TreeNode {
	if matchTok(TT_LET) {
		return handleStmtDecl()
	} else if matchTok(TT_WHILE) {
		return handleStmtWhile()
	} else if matchTok(TT_IF) {
		return handleStmtIf()
	} else if matchTok(TT_RETURN) {
		return handleStmtReturn()
	} else if matchTok(TT_BREAK) {
		return handleStmtBreak()
	} else if matchTok(TT_CONTINUE) {
		return handleStmtContinue()
	} else {
		exprTreeNode := handleExpr()

		if matchTok(TT_ASSIGN) {
			return handleStmtAssign(exprTreeNode)
		} else if matchTok(TT_ARROW) {
			return handleStmtStoreString(exprTreeNode)
		} else {
			return handleStmtExpr(exprTreeNode)
		}
	}
}

func handleStmtDecl() TreeNode {
	consumeTok(TT_LET)

	var tn TreeNode
	tn.Kype = TNT_STMT_DECL

	tn.Children = append(tn.Children, handleStmtDeclIdent())
	tn.Children = append(tn.Children, handleStmtDeclType())

	consumeTok(TT_NEW_LINE)

	return tn
}

func handleStmtDeclIdent() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_DECL_IDENT
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func handleStmtDeclType() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_DECL_TYPE
	tn.Tok = consumeTok(TT_IDENT)

	return tn
}

func handleStmtExpr(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_EXPR

	consumeTok(TT_NEW_LINE)

	tn.Children = append(tn.Children, exprTreeNode)
	return tn
}

func handleStmtAssign(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_ASSIGN

	consumeTok(TT_ASSIGN)

	tn.Children = append(tn.Children, exprTreeNode)
	tn.Children = append(tn.Children, handleExpr())

	consumeTok(TT_NEW_LINE)
	return tn
}

func handleStmtStoreString(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_STORE_STRING

	consumeTok(TT_ARROW)

	tn.Children = append(tn.Children, exprTreeNode)
	tn.Children = append(tn.Children, handleStmtString())

	consumeTok(TT_NEW_LINE)
	return tn
}

func handleStmtString() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_STRING
	tn.Tok = consumeTok(TT_STR)

	return tn
}

func handleStmtWhile() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_WHILE

	consumeTok(TT_WHILE)

	tn.Children = append(tn.Children, handleExpr())

	consumeTok(TT_NEW_LINE)

	tn.Children = append(tn.Children, handleStmtList())

	consumeTok(TT_END)
	consumeTok(TT_NEW_LINE)

	return tn
}

func handleStmtIf() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_IF

	consumeTok(TT_IF)

	tn.Children = append(tn.Children, handleExpr())

	consumeTok(TT_NEW_LINE)

	tn.Children = append(tn.Children, handleStmtList())

	if matchTok(TT_ELSE) {
		tn.Children = append(tn.Children, handleStmtElse())
	} else {
		consumeTok(TT_END)
		consumeTok(TT_NEW_LINE)
	}

	return tn
}

func handleStmtElse() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_ELSE

	consumeTok(TT_ELSE)

	if matchTok(TT_IF) {
		tn.Children = append(tn.Children, handleStmtIf())
	} else {
		consumeTok(TT_NEW_LINE)

		tn.Children = append(tn.Children, handleStmtList())

		consumeTok(TT_END)
		consumeTok(TT_NEW_LINE)
	}

	return tn
}

func handleExpr() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR
	tn.Children = append(tn.Children, handleExprCont(true))

	return tn
}

func handleExprCont(doesFollowBinary bool) TreeNode {
	tn := handleExprUnary()

	for doesFollowBinary && matchBinaryTok() {
		tn = handleExprBinary(tn)
	}

	return tn
}

func handleExprUnary() TreeNode {
	var tn TreeNode

	if matchTok(TT_IDENT) {
		tn.Tok = consumeTok(TT_IDENT)
		if matchTok(TT_LPAREN) {
			tn.Kype = TNT_EXPR_FUNC
			tn.Children = append(tn.Children, handleExprUnaryFuncParmList())
		} else {
			tn.Kype = TNT_EXPR_VAR
		}
	} else {
		consumeTok(TT_LPAREN)
		tn = handleExprCont(true)
		consumeTok(TT_RPAREN)
	}

	return tn
}

func handleExprUnaryFuncParmList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_FUNC_PARM_LIST

	consumeTok(TT_LPAREN)

	if matchTok(TT_IDENT, TT_LPAREN, TT_INT, TT_CHAR, TT_SUB) {
		tn.Children = append(tn.Children, handleExprUnaryFuncParm())
		for matchTok(TT_COMMA) {
			consumeTok(TT_COMMA)
			tn.Children = append(tn.Children, handleExprUnaryFuncParm())
		}
	}

	consumeTok(TT_RPAREN)

	return tn
}

func handleExprUnaryFuncParm() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_FUNC_PARM

	if matchTok(TT_INT) {
		tn.Children = append(tn.Children, handleExprUnaryFuncParmInt())
	} else if matchTok(TT_SUB) {
		tn.Children = append(tn.Children, handleExprUnaryFuncParmNegInt())
	} else if matchTok(TT_CHAR) {
		tn.Children = append(tn.Children, handleExprUnaryFuncParmChar())
	} else {
		tn.Children = append(tn.Children, handleExpr())
	}

	return tn
}

func handleExprUnaryFuncParmInt() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_INT
	tn.Tok = consumeTok(TT_INT)
	return tn
}

func handleExprUnaryFuncParmNegInt() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_NEG_INT
	consumeTok(TT_SUB)
	tn.Tok = consumeTok(TT_INT)
	return tn
}

func handleExprUnaryFuncParmChar() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_CHAR
	tn.Tok = consumeTok(TT_CHAR)
	return tn
}

func handleExprBinary(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_BINARY
	tn.Tok = advanceTok()

	tn.Children = append(tn.Children, exprTreeNode)
	tn.Children = append(tn.Children, handleExprCont(false))
	return tn
}

func handleStmtReturn() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_RETURN
	consumeTok(TT_RETURN)

	if matchTok(TT_IDENT, TT_LPAREN) {
		tn.Children = append(tn.Children, handleExpr())
	}

	consumeTok(TT_NEW_LINE)
	return tn

}

func handleStmtBreak() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_BREAK
	consumeTok(TT_BREAK)
	consumeTok(TT_NEW_LINE)
	return tn
}

func handleStmtContinue() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_CONTINUE
	consumeTok(TT_BREAK)
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

func SyntaxAnalyzer(toks []TokenData) TreeNode {
	var tn TreeNode

	tn.Kype = TNT_ROOT

	peekTok = func() TokenData {
		if len(toks) == 0 {
			PrintErrorAndExit(0)
		}
		return toks[0]
	}

	advanceTok = func() TokenData {
		if len(toks) == 0 {
			PrintErrorAndExit(0)
		}
		tok := toks[0]
		toks = toks[1:]
		return tok
	}

	consumeTok = func(tokType TokenType) TokenData {
		if len(toks) == 0 {
			PrintErrorAndExit(0)
		}
		tok := toks[0]
		if tok.Kype != tokType {
			PrintErrorAndExit(tok.LineNumber)
		}
		toks = toks[1:]
		return tok
	}

	matchTok = func(tokTypes ...TokenType) bool {
		for _, curTokType := range tokTypes {
			if curTokType == peekTok().Kype {
				return true
			}
		}
		return false
	}

	matchBinaryTok = func() bool {
		return matchTok(TT_ADD, TT_SUB, TT_MUL, TT_QUO, TT_REM,
			TT_AND, TT_OR, TT_XOR,
			TT_SHL, TT_SHR,
			TT_LAND, TT_LOR,
			TT_EQL, TT_NEQ, TT_LSS, TT_GTR, TT_LEQ, TT_GEQ)
	}

	tn.Children = append(tn.Children, handleFuncList())

	return tn
}
