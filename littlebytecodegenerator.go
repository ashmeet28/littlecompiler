package main

import (
	"strconv"
)

var bytecode []byte

type IntInfo struct {
	RealSize   int
	IsSigned   bool
	BytesCount int
}

type LocalIntInfo struct {
	RealSize   int
	Ident      string
	IsSigned   bool
	BlockLevel int
	BytesCount int
}

type LocalIntAddrInfo struct {
	RealSize   int
	IsSigned   bool
	BytesCount int
}

type PrevFrameAddrInfo struct {
	BytesCount int
}

type ReturnAddrInfo struct {
	BytesCount int
}

func getIntInfoFromTypeString(s string) (IntInfo, bool) {
	_, ok := map[string]bool{"i8": true, "i16": true, "i32": true, "i64": true,
		"u8": true, "u16": true, "u32": true, "u64": true}[s]

	if !ok {
		return IntInfo{}, false
	}

	var ii IntInfo

	if s[0] == 0x75 {
		ii.IsSigned = false
	} else if s[0] == 0x69 {
		ii.IsSigned = true
	} else {
		return IntInfo{}, false
	}

	bitSize, err := strconv.ParseInt(s[1:], 10, 64)
	if err != nil {
		return IntInfo{}, false
	}

	ii.RealSize = int(bitSize) / 8
	ii.BytesCount = int(bitSize) / 8
	return ii, true
}

var blockLevel int
var returnIntInfo interface{}
var framePointer int

var callStackInfo []interface{}

func callStackInfoReset() {
	blockLevel = 1
	returnIntInfo = nil
	framePointer = 0
	callStackInfo = make([]interface{}, 0)
}

func callStackInfoGetTotalBytesCount() int {
	var c int
	for _, i := range callStackInfo {
		switch v := i.(type) {
		case IntInfo:
			c += v.BytesCount
		case LocalIntInfo:
			c += v.BytesCount
		case LocalIntAddrInfo:
			c += v.BytesCount
		case PrevFrameAddrInfo:
			c += v.BytesCount
		case ReturnAddrInfo:
			c += v.BytesCount
		default:
			PrintErrorAndExit(0)
		}
	}
	return c
}

func callStackInfoInitFrame() {
	callStackInfo = append(callStackInfo, PrevFrameAddrInfo{BytesCount: 8})
	callStackInfo = append(callStackInfo, ReturnAddrInfo{BytesCount: 8})
	framePointer = callStackInfoGetTotalBytesCount()
}

var rootTreeNode TreeNode

func compileFuncList(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileFunc(tn TreeNode) {
	callStackInfoReset()
	compileTreeNodeChildren(tn.Children)
}

func compileFuncIdent(tn TreeNode) {
}

func compileFuncSig(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileFuncParamList(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileFuncParam(tn TreeNode) {
	var lii LocalIntInfo
	lii.Ident = string(tn.Children[0].Tok.Buf)
	ii, isOk := getIntInfoFromTypeString(string(tn.Children[1].Tok.Buf))
	if !isOk {
		PrintErrorAndExit(tn.Children[1].Tok.LineNumber)
	}
	lii.RealSize = ii.RealSize
	lii.IsSigned = ii.IsSigned
	lii.BytesCount = ii.BytesCount
	lii.BlockLevel = blockLevel

	callStackInfo = append(callStackInfo, lii)
}

func compileFuncReturnType(tn TreeNode) {
	ii, isOk := getIntInfoFromTypeString(string(tn.Tok.Buf))
	if !isOk {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}
	returnIntInfo = ii
}

func compileTreeNodeChildren(treeNodeChildren []TreeNode) {
	for _, c := range treeNodeChildren {
		map[TreeNodeType]func(TreeNode){
			// TNT_ROOT

			TNT_FUNC_LIST: compileFuncList,
			TNT_FUNC:      compileFunc,

			TNT_FUNC_IDENT:      compileFuncIdent,
			TNT_FUNC_SIG:        compileFuncSig,
			TNT_FUNC_PARAM_LIST: compileFuncParamList,
			TNT_FUNC_PARAM:      compileFuncParam,
			// TNT_FUNC_PARAM_IDENT
			// TNT_FUNC_PARAM_TYPE
			TNT_FUNC_RETURN_TYPE: compileFuncReturnType,

			// TNT_STMT_LIST

			// TNT_STMT_DECL
			// TNT_STMT_DECL_IDENT
			// TNT_STMT_DECL_TYPE

			// TNT_STMT_EXPR
			// TNT_STMT_ASSIGN
			// TNT_STMT_STORE_STRING
			// TNT_STMT_STRING

			// TNT_STMT_WHILE
			// TNT_STMT_IF
			// TNT_STMT_ELSE

			// TNT_STMT_RETURN
			// TNT_STMT_BREAK
			// TNT_STMT_CONTINUE

			// TNT_EXPR
			// TNT_EXPR_INT
			// TNT_EXPR_FUNC
			// TNT_EXPR_FUNC_PARM_LIST
			// TNT_EXPR_FUNC_PARM
			// TNT_EXPR_INT_LIT
			// TNT_EXPR_NEG_INT_LIT
			// TNT_EXPR_CHAR
			// TNT_EXPR_BINARY
		}[c.Kype](c)
	}
}

func BytecodeGenerator(tn TreeNode) []byte {
	rootTreeNode = tn

	return bytecode
}
