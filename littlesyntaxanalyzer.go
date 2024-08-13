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

	TNT_STMT_LIST
	TNT_STMT

	TNT_STMT_DECL
	TNT_STMT_DECL_IDENT
	TNT_STMT_DECL_TYPE

	TNT_STMT_EXPR
	TNT_STMT_ASSIGN

	TNT_EXPR
	TNT_EXPR_VAR
	TNT_EXPR_FUNC
	TNT_EXPR_FUNC_PARM_LIST
	TNT_EXPR_FUNC_PARM
	TNT_EXPR_INT
	TNT_EXPR_CHAR
	TNT_EXPR_BINARY
)

type TreeNode struct {
	Kype     TreeNodeType
	children []TreeNode
	tok      TokenData
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
		tn.children = append(tn.children, handleFunc())
	}
	consumeTok(TT_EOF)

	return tn
}

func handleFunc() TreeNode {
	consumeTok(TT_FUNC)

	var tn TreeNode
	tn.Kype = TNT_FUNC

	tn.children = append(tn.children, handleFuncIdent())
	tn.children = append(tn.children, handleFuncSig())

	consumeTok(TT_NEW_LINE)

	tn.children = append(tn.children, handleStmtList())

	consumeTok(TT_END)
	consumeTok(TT_NEW_LINE)

	return tn
}

func handleFuncIdent() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_IDENT
	tn.tok = consumeTok(TT_IDENT)

	return tn
}

func handleFuncSig() TreeNode {
	consumeTok(TT_LPAREN)

	var tn TreeNode
	tn.Kype = TNT_FUNC_SIG

	if matchTok(TT_IDENT) {
		tn.children = append(tn.children, handleFuncParamList())
	}

	consumeTok(TT_RPAREN)

	return tn
}

func handleFuncParamList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_LIST

	tn.children = append(tn.children, handleFuncParam())
	for matchTok(TT_COMMA) {
		consumeTok(TT_COMMA)
		tn.children = append(tn.children, handleFuncParam())
	}

	return tn
}

func handleFuncParam() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM

	tn.children = append(tn.children, handleFuncParamIdent())
	tn.children = append(tn.children, handleFuncParamType())

	return tn
}

func handleFuncParamIdent() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_IDENT
	tn.tok = consumeTok(TT_IDENT)
	return tn
}

func handleFuncParamType() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_TYPE
	tn.tok = consumeTok(TT_IDENT)

	return tn
}

func handleStmtList() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_LIST

	for matchTok(TT_LET, TT_IDENT, TT_LPAREN) {
		tn.children = append(tn.children, handleStmt())
	}

	return tn
}

func handleStmt() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT

	if matchTok(TT_LET) {
		tn.children = append(tn.children, handleDeclStmt())
	} else {
		exprTreeNode := handleExpr()

		if matchTok(TT_NEW_LINE) {
			tn.children = append(tn.children, handleExprStmt(exprTreeNode))
		} else if matchTok(TT_ASSIGN) {
			tn.children = append(tn.children, handleAssignStmt(exprTreeNode))
		}
	}

	return tn
}

func handleDeclStmt() TreeNode {
	consumeTok(TT_LET)

	var tn TreeNode
	tn.Kype = TNT_STMT_DECL

	tn.children = append(tn.children, handleDeclStmtIdent())
	tn.children = append(tn.children, handleDeclStmtType())

	consumeTok(TT_NEW_LINE)

	return tn
}

func handleDeclStmtIdent() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_DECL_IDENT
	tn.tok = consumeTok(TT_IDENT)

	return tn
}

func handleDeclStmtType() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_DECL_TYPE
	tn.tok = consumeTok(TT_IDENT)

	return tn
}

func handleExprStmt(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_EXPR

	consumeTok(TT_NEW_LINE)

	tn.children = append(tn.children, exprTreeNode)
	return tn
}

func handleAssignStmt(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT_ASSIGN

	consumeTok(TT_ASSIGN)

	tn.children = append(tn.children, exprTreeNode)
	tn.children = append(tn.children, handleExpr())

	consumeTok(TT_NEW_LINE)
	return tn
}

func handleExpr() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR
	tn.children = append(tn.children, handleExprCont(true))

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
		tn.tok = consumeTok(TT_IDENT)
		if matchTok(TT_LPAREN) {
			tn.Kype = TNT_EXPR_FUNC
			tn.children = append(tn.children, handleExprUnaryFuncParmList())
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

	if matchTok(TT_LPAREN, TT_IDENT) {
		tn.children = append(tn.children, handleExprUnaryFuncParm())
	}

	for matchTok(TT_COMMA) {
		consumeTok(TT_COMMA)
		tn.children = append(tn.children, handleExprUnaryFuncParm())
	}

	consumeTok(TT_RPAREN)

	return tn
}

func handleExprUnaryFuncParm() TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_FUNC_PARM

	tn.children = append(tn.children, handleExpr())

	return tn
}

func handleExprBinary(exprTreeNode TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_EXPR_BINARY
	tn.tok = advanceTok()

	tn.children = append(tn.children, exprTreeNode)
	tn.children = append(tn.children, handleExprCont(false))
	return tn
}

func PrintTreeNode(tn TreeNode, level int) {
	var s string
	for i := 0; i < level; i++ {
		s = s + " "
	}

	fmt.Println(s, "->", tn.Kype, tn.tok)

	for _, child := range tn.children {
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

	tn.children = append(tn.children, handleFuncList())

	return tn
}
