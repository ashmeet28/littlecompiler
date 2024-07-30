package main

import "fmt"

type TreeNodeType int

const (
	TNT_ILLEGAL TreeNodeType = iota

	TNT_ROOT

	TNT_FUNCS
	TNT_FUNC
	TNT_FUNC_IDENT
	TNT_FUNC_SIG
	TNT_FUNC_PARAMS
	TNT_FUNC_PARAM
	TNT_FUNC_PARAM_IDENT
	TNT_FUNC_PARAM_TYPE

	TNT_STMTS
	TNT_STMT
	TNT_STMT_DECL
	TNT_STMT_DECL_IDENT
	TNT_STMT_DECL_TYPE
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

func handleFuncs(ptn TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNCS

	for matchTok(TT_FUNC) {
		tn = handleFunc(tn)
	}
	consumeTok(TT_EOF)

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleFunc(ptn TreeNode) TreeNode {
	consumeTok(TT_FUNC)

	var tn TreeNode
	tn.Kype = TNT_FUNC

	tn = handleFuncIdent(tn)
	tn = handleFuncSig(tn)

	consumeTok(TT_NEW_LINE)

	tn = handleStmts(tn)

	consumeTok(TT_END)
	consumeTok(TT_NEW_LINE)

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleFuncIdent(ptn TreeNode) TreeNode {
	tok := consumeTok(TT_IDENT)

	var tn TreeNode
	tn.Kype = TNT_FUNC_IDENT
	tn.tok = tok

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleFuncSig(ptn TreeNode) TreeNode {
	consumeTok(TT_LPAREN)

	var tn TreeNode
	tn.Kype = TNT_FUNC_SIG

	if matchTok(TT_IDENT) {
		tn = handleFuncParams(tn)
	}

	consumeTok(TT_RPAREN)

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleFuncParams(ptn TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAMS

	tn = handleFuncParam(tn)
	for matchTok(TT_COMMA) {
		consumeTok(TT_COMMA)
		tn = handleFuncParam(tn)
	}

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleFuncParam(ptn TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM

	tn = handleFuncParamIdent(tn)
	tn = handleFuncParamType(tn)

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleFuncParamIdent(ptn TreeNode) TreeNode {
	tok := consumeTok(TT_IDENT)

	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_IDENT
	tn.tok = tok

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleFuncParamType(ptn TreeNode) TreeNode {
	if !matchTok(TT_U8, TT_U16, TT_U32, TT_U64, TT_I8, TT_I16, TT_I32, TT_I64) {
		PrintErrorAndExit(peekTok().LineNumber)
	}

	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAM_TYPE
	tn.tok = advanceTok()

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleStmts(ptn TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMTS

	for matchTok(TT_LET) {
		tn = handleStmt(tn)
	}

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleStmt(ptn TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_STMT

	if matchTok(TT_LET) {
		tn = handleDeclStmt(tn)
	}

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleDeclStmt(ptn TreeNode) TreeNode {
	consumeTok(TT_LET)

	var tn TreeNode
	tn.Kype = TNT_STMT_DECL

	tn = handleDeclStmtIdent(tn)
	tn = handleDeclStmtType(tn)
	consumeTok(TT_NEW_LINE)

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleDeclStmtIdent(ptn TreeNode) TreeNode {
	tok := consumeTok(TT_IDENT)

	var tn TreeNode
	tn.Kype = TNT_STMT_DECL_IDENT
	tn.tok = tok

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleDeclStmtType(ptn TreeNode) TreeNode {
	if !matchTok(TT_U8, TT_U16, TT_U32, TT_U64, TT_I8, TT_I16, TT_I32, TT_I64) {
		PrintErrorAndExit(peekTok().LineNumber)
	}

	var tn TreeNode
	tn.Kype = TNT_STMT_DECL_TYPE
	tn.tok = advanceTok()

	ptn.children = append(ptn.children, tn)
	return ptn
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

	tn = handleFuncs(tn)

	return tn
}
