package main

import "fmt"

type TreeNodeType int

const (
	TNT_ILLEGAL TreeNodeType = iota

	TNT_ROOT

	TNT_FUNC

	TNT_FUNC_IDENT
	TNT_FUNC_SIG
	TNT_FUNC_PARAMS
	TNT_FUNC_PARAM
	TNT_FUNC_PARAM_IDENT
	TNT_FUNC_PARAM_TYPE
)

type TreeNode struct {
	Kype     TreeNodeType
	children []TreeNode
	tok      TokenData
}

var tokFunc map[int]func(TreeNode) TreeNode

var peekTok func() TokenData

var advanceTok func() TokenData

var consumeTok func(TokenType) TokenData

var matchTok func(...TokenType) bool

func handleFunc(ptn TreeNode) TreeNode {
	consumeTok(TT_FUNC)

	var tn TreeNode
	tn.Kype = TNT_FUNC

	tn = handleFuncIdent(tn)
	tn = handleFuncSig(tn)

	consumeTok(TT_NEW_LINE)
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

	tn = handleFuncParams(tn)

	consumeTok(TT_RPAREN)

	ptn.children = append(ptn.children, tn)
	return ptn
}

func handleFuncParams(ptn TreeNode) TreeNode {
	var tn TreeNode
	tn.Kype = TNT_FUNC_PARAMS

	if !matchTok(TT_RPAREN) {
		tn = handleFuncParam(tn)
		for !matchTok(TT_RPAREN) {
			consumeTok(TT_COMMA)
			tn = handleFuncParam(tn)
		}
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

func PrintTreeNode(tn TreeNode, level int) {
	var s string
	for i := 0; i < level; i++ {
		s = s + " "
	}
	fmt.Println(s+"|", tn.Kype, "|", tn.tok, "|")
	for _, child := range tn.children {
		PrintTreeNode(child, level+4)
	}
}

func SyntaxAnalyzer(toks []TokenData) TreeNode {
	var rootTreeNode TreeNode

	rootTreeNode.Kype = TNT_ROOT

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

	for len(toks) > 1 {
		rootTreeNode = handleFunc(rootTreeNode)
	}

	consumeTok(TT_EOF)

	return rootTreeNode
}
