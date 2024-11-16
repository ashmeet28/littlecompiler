package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

var (
	OP_HALT  byte = 0x01
	OP_ECALL byte = 0x02

	OP_CALL         byte = 0x04
	OP_RETURN       byte = 0x05
	OP_RETURN_EMPTY byte = 0x06

	OP_JUMP   byte = 0x08
	OP_BRANCH byte = 0x09

	OP_PUSH byte = 0x0c
	OP_POP  byte = 0x0d

	OP_ADD byte = 0x10
	OP_SUB byte = 0x11

	OP_AND byte = 0x14
	OP_OR  byte = 0x15
	OP_XOR byte = 0x16

	OP_SL byte = 0x18
	OP_SR byte = 0x19

	OP_MUL byte = 0x1c
	OP_DIV byte = 0x1d
	OP_REM byte = 0x1e
)

var bytecode []byte

type IntInfo struct {
	IsSigned   bool
	BytesCount int
}

type LocalIntInfo struct {
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

type EmptyInfo struct {
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

	ii.BytesCount = int(bitSize) / 8
	return ii, true
}

var ADDR_BYTES_COUNT int = 8

var blockLevel int
var returnIntInfo interface{}
var framePointer int

var callStackInfo []interface{}

var STARTING_BLOCK_LEVEL int = 1

func callStackInfoReset() {
	blockLevel = STARTING_BLOCK_LEVEL
	returnIntInfo = nil
	framePointer = 0
	callStackInfo = make([]interface{}, 0)
}

func callStackInfoGetBytesCount(i interface{}) (int, bool) {
	switch v := i.(type) {
	case IntInfo:
		return v.BytesCount, true
	case LocalIntInfo:
		return v.BytesCount, true
	case LocalIntAddrInfo:
		return v.BytesCount, true
	case PrevFrameAddrInfo:
		return v.BytesCount, true
	case ReturnAddrInfo:
		return v.BytesCount, true
	case EmptyInfo:
		return v.BytesCount, true
	default:
		return 0, false
	}
}

func callStackInfoGetTotalBytesCount() int {
	var totalBytesCount int
	for _, i := range callStackInfo {
		if currBytesCount, ok := callStackInfoGetBytesCount(i); ok {
			totalBytesCount += currBytesCount
		} else {
			PrintErrorAndExit(0)
		}
	}

	return totalBytesCount
}

func callStackInfoInitFrame() {
	callStackInfo = append(callStackInfo, PrevFrameAddrInfo{BytesCount: ADDR_BYTES_COUNT})
	callStackInfo = append(callStackInfo, ReturnAddrInfo{BytesCount: ADDR_BYTES_COUNT})
	framePointer = callStackInfoGetTotalBytesCount()
}

func callStackInfoFindLocalIntInfo(localIntInfoIdent string) (LocalIntInfo, bool) {
	for i := len(callStackInfo) - 1; i >= 0; i-- {
		lii, ok := callStackInfo[i].(LocalIntInfo)
		if ok && (lii.Ident == localIntInfoIdent) {
			return lii, true
		}
	}
	return LocalIntInfo{}, false
}

func callStackInfoGetLocalIntAddr(localIntInfoIdent string) (int, bool) {
	var hasFound bool = false

	var totalBytesCount int

	for i := len(callStackInfo) - 1; i >= 0; i-- {
		if hasFound {
			if currBytesCount, ok := callStackInfoGetBytesCount(callStackInfo[i]); ok {
				totalBytesCount += currBytesCount
			} else {
				PrintErrorAndExit(0)
			}
		} else {
			lii, ok := callStackInfo[i].(LocalIntInfo)
			if ok && (lii.Ident == localIntInfoIdent) {
				hasFound = true
			}
		}
	}

	if !hasFound {
		return 0, false
	}

	return (totalBytesCount - framePointer), true
}

func incBlockLevel() {
	blockLevel++
}

func decBlockLevel() {
	blockLevel--
	for i := len(callStackInfo) - 1; i >= 0; i-- {
		if lii, ok := callStackInfo[i].(LocalIntInfo); ok && (lii.BlockLevel > blockLevel) {
			emitPopOp(IntInfo{IsSigned: lii.IsSigned, BytesCount: lii.BytesCount})
			callStackInfo = callStackInfo[:len(callStackInfo)-1]
		}
	}
}

type FuncSigInfo struct {
	ParamListInt []IntInfo
	ReturnInt    interface{}
}

var funcListInfo map[string]FuncSigInfo

func funcListInfoInit(tn TreeNode) {
	funcListInfo = make(map[string]FuncSigInfo)

	for _, funcTreeNode := range tn.Children {

		funcSigTreeNode := funcTreeNode.Children[1]
		var newFuncSigInfo FuncSigInfo

		for _, c := range funcSigTreeNode.Children {

			if c.Kype == TNT_FUNC_PARAM_LIST {
				funcParamListTreeNode := c

				for _, funcParmTreeNode := range funcParamListTreeNode.Children {

					funcParamTypeTreeNode := funcParmTreeNode.Children[1]

					ii, ok := getIntInfoFromTypeString(string(funcParamTypeTreeNode.Tok.Buf))
					if !ok {
						PrintErrorAndExit(funcParamTypeTreeNode.Tok.LineNumber)
					}

					newFuncSigInfo.ParamListInt = append(newFuncSigInfo.ParamListInt, ii)

				}

			} else if c.Kype == TNT_FUNC_RETURN_TYPE {

				funcReturnTypeTreeNode := c
				ii, ok := getIntInfoFromTypeString(string(funcReturnTypeTreeNode.Tok.Buf))
				if !ok {
					PrintErrorAndExit(tn.Tok.LineNumber)
				}
				newFuncSigInfo.ReturnInt = ii

			}

		}

		funcIdentTreeNode := funcTreeNode.Children[0]
		funcIdent := string(funcIdentTreeNode.Tok.Buf)

		_, doesAlreadyExists := funcListInfo[funcIdent]
		if doesAlreadyExists {
			PrintErrorAndExit(funcIdentTreeNode.Tok.LineNumber)
		}

		funcListInfo[funcIdent] = newFuncSigInfo
	}
}

var funcListAddr map[string]int

func encodeIntInfo(ii IntInfo) byte {
	var b byte

	if ii.BytesCount == 1 {
		b = 0b00
	} else if ii.BytesCount == 2 {
		b = 0b01
	} else if ii.BytesCount == 4 {
		b = 0b10
	} else if ii.BytesCount == 8 {
		b = 0b11
	}

	if ii.IsSigned {
		b = b | 0b100
	}
	return b
}

func encodeLocalIntAddrInfo(liai LocalIntAddrInfo) byte {
	return (encodeIntInfo(IntInfo{IsSigned: liai.IsSigned, BytesCount: liai.RealSize}) | 0b1000)
}

func emitBlankPushOp() int {
	addr := len(bytecode)
	emitPushOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT}, 0)
	return addr
}

func backpatchBlankPushOp(addr int, v uint64) {
	for i, b := range binary.LittleEndian.AppendUint64(make([]byte, 0), v) {
		bytecode[addr+i+2] = b
	}
}

func emitOp(op byte) {
	bytecode = append(bytecode, op)
}

func emitPushOp(ii IntInfo, v uint64) {
	bytecode = append(bytecode, OP_PUSH)
	bytecode = append(bytecode, encodeIntInfo(ii))

	bytecode = append(bytecode,
		binary.LittleEndian.AppendUint64(make([]byte, 0), v)[:ii.BytesCount]...)
}

func emitReturn(i interface{}) bool {
	switch v := i.(type) {
	case IntInfo:
		bytecode = append(bytecode, OP_RETURN)
		bytecode = append(bytecode, encodeIntInfo(v))
		return true
	case LocalIntAddrInfo:
		bytecode = append(bytecode, OP_RETURN)
		bytecode = append(bytecode, encodeLocalIntAddrInfo(v))
		return true
	default:
		return false
	}
}

func emitReturnEmpty() {
	bytecode = append(bytecode, OP_RETURN_EMPTY)
}

func emitPopOp(ii IntInfo) {
	bytecode = append(bytecode, OP_POP)
	bytecode = append(bytecode, encodeIntInfo(ii))
}

func emitBinaryOp(op byte, v1 interface{}, v2 interface{}) (bool, IntInfo) {
	var vb1 byte = 0
	var vb2 byte = 0

	var ii IntInfo

	switch v := v1.(type) {
	case IntInfo:
		vb1 = encodeIntInfo(v)
		ii = IntInfo{IsSigned: v.IsSigned, BytesCount: v.BytesCount}
	case LocalIntAddrInfo:
		vb1 = encodeLocalIntAddrInfo(v)
		ii = IntInfo{IsSigned: v.IsSigned, BytesCount: v.RealSize}
	default:
		return false, IntInfo{}
	}

	switch v := v2.(type) {
	case IntInfo:
		vb2 = encodeIntInfo(v)
	case LocalIntAddrInfo:
		vb2 = encodeLocalIntAddrInfo(v)
	default:
		return false, IntInfo{}
	}

	if ((vb1 & 0b111) == (vb2 & 0b111)) || (op == OP_SL) || (op == OP_SR) {
		bytecode = append(bytecode, op)
		bytecode = append(bytecode, (vb2<<4)|vb1)

		return true, ii
	}

	return false, IntInfo{}
}

var rootTreeNode TreeNode

func compileFuncList(tn TreeNode) {
	funcListAddr = make(map[string]int)
	compileTreeNodeChildren(tn.Children)
}

func compileFunc(tn TreeNode) {
	callStackInfoReset()
	compileTreeNodeChildren(tn.Children)
}

func compileFuncIdent(tn TreeNode) {
	funcListAddr[string(tn.Tok.Buf)] = len(bytecode)
}

func compileFuncSig(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
	callStackInfoInitFrame()
}

func compileFuncParamList(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileFuncParam(tn TreeNode) {
	var lii LocalIntInfo

	funcParamIdentTreeNode := tn.Children[0]
	funcParamTypeTreeNode := tn.Children[1]

	lii.Ident = string(funcParamIdentTreeNode.Tok.Buf)
	ii, ok := getIntInfoFromTypeString(string(funcParamTypeTreeNode.Tok.Buf))
	if !ok {
		PrintErrorAndExit(funcParamTypeTreeNode.Tok.LineNumber)
	}
	lii.IsSigned = ii.IsSigned
	lii.BytesCount = ii.BytesCount
	lii.BlockLevel = blockLevel

	callStackInfo = append(callStackInfo, lii)
}

func compileFuncReturnType(tn TreeNode) {
	ii, ok := getIntInfoFromTypeString(string(tn.Tok.Buf))
	if !ok {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}
	returnIntInfo = ii
}

func compileStmtList(tn TreeNode) {
	incBlockLevel()
	compileTreeNodeChildren(tn.Children)
	decBlockLevel()
}

func compileStmtDecl(tn TreeNode) {
	stmtDeclIdentTreeNode := tn.Children[0]
	stmtDeclTypeTreeNode := tn.Children[1]

	if lii, ok := callStackInfoFindLocalIntInfo(string(stmtDeclIdentTreeNode.Tok.Buf)); ok {
		if (lii.BlockLevel == blockLevel) ||
			((lii.BlockLevel == STARTING_BLOCK_LEVEL) && (blockLevel == STARTING_BLOCK_LEVEL+1)) {
			PrintErrorAndExit(stmtDeclIdentTreeNode.Tok.LineNumber)
		}
	}

	var lii LocalIntInfo

	lii.Ident = string(stmtDeclIdentTreeNode.Tok.Buf)
	ii, ok := getIntInfoFromTypeString(string(stmtDeclTypeTreeNode.Tok.Buf))
	if !ok {
		PrintErrorAndExit(stmtDeclTypeTreeNode.Tok.LineNumber)
	}
	lii.IsSigned = ii.IsSigned
	lii.BytesCount = ii.BytesCount
	lii.BlockLevel = blockLevel

	callStackInfo = append(callStackInfo, lii)

	emitPushOp(ii, 0)
}

func compileStmtExpr(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)

	switch v := callStackInfo[len(callStackInfo)-1].(type) {
	case IntInfo:
		emitPopOp(v)
	case LocalIntAddrInfo:
		emitPopOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT})
	case EmptyInfo:
	default:
		PrintErrorAndExit(0)
	}

	callStackInfo = callStackInfo[:len(callStackInfo)-1]
}

func compileStmtReturn(tn TreeNode) {
	if len(tn.Children) == 0 {
		if ii, ok := returnIntInfo.(IntInfo); ok {
			emitPushOp(ii, 0)
			if ok := emitReturn(ii); !ok {
				PrintErrorAndExit(0)
			}
		} else {
			emitReturnEmpty()
		}
	}
}

func compileExpr(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileExprInt(tn TreeNode) {
	if lii, ok := callStackInfoFindLocalIntInfo(string(tn.Tok.Buf)); ok {
		var liai LocalIntAddrInfo
		liai.RealSize = lii.BytesCount
		liai.IsSigned = lii.IsSigned
		liai.BytesCount = ADDR_BYTES_COUNT

		callStackInfo = append(callStackInfo, liai)

		if addr, ok := callStackInfoGetLocalIntAddr(string(tn.Tok.Buf)); ok {
			emitPushOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT}, uint64(addr))
		} else {
			PrintErrorAndExit(0)
		}
	} else {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}
}

func unescapeExprChar(b []byte) (byte, bool) {
	if len(b) < 3 || (b[0] != 0x27) || (b[len(b)-1] != 0x27) {
		return 0, false
	}

	if b[1] == 0x5c {
		if len(b) == 4 && (b[2] == 0x5c || b[2] == 0x27) {
			return b[2], true
		} else {
			return 0, false
		}
	} else if len(b) == 3 {
		return b[1], true
	} else {
		return 0, false
	}
}

func compileExprFunc(tn TreeNode) {
	if ii, ok := getIntInfoFromTypeString(string(tn.Tok.Buf)); ok {
		exprFuncParmListTreeNode := tn.Children[0]

		if len(exprFuncParmListTreeNode.Children) != 1 {
			PrintErrorAndExit(tn.Tok.LineNumber)
		}

		exprFuncParmTreeNode := exprFuncParmListTreeNode.Children[0]

		switch exprFuncParmTreeNode.Children[0].Kype {
		case TNT_EXPR_CHAR:
			exprCharTreeNode := exprFuncParmTreeNode.Children[0]
			if b, ok := unescapeExprChar(exprCharTreeNode.Tok.Buf); ok {
				if ii.BytesCount == 1 && (!ii.IsSigned) {
					emitPushOp(ii, uint64(b))
					callStackInfo = append(callStackInfo, ii)
				} else {
					PrintErrorAndExit(exprCharTreeNode.Tok.LineNumber)
				}
			} else {
				PrintErrorAndExit(exprCharTreeNode.Tok.LineNumber)
			}
		case TNT_EXPR_INT_LIT:
			exprIntLitTreeNode := exprFuncParmTreeNode.Children[0]

			if v, err := strconv.ParseUint(string(exprIntLitTreeNode.Tok.Buf), 0, 64); err == nil {
				if v > (^uint64(0))>>((8-ii.BytesCount)*8) {
					PrintErrorAndExit(exprIntLitTreeNode.Tok.LineNumber)
				}
				emitPushOp(ii, v)
				callStackInfo = append(callStackInfo, ii)
			} else {
				PrintErrorAndExit(exprIntLitTreeNode.Tok.LineNumber)
			}
		default:
			PrintErrorAndExit(tn.Tok.LineNumber)
		}
	} else {
		compileTreeNodeChildren(tn.Children)
	}
}

func compileExprFuncParmList(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileExprFuncParm(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileExprBinary(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)

	op, ok := map[TokenType]byte{
		TT_ADD: OP_ADD,
		TT_SUB: OP_SUB,

		TT_AND: OP_AND,
		TT_OR:  OP_OR,
		TT_XOR: OP_XOR,

		TT_SHL: OP_SL,
		TT_SHR: OP_SR,

		TT_MUL: OP_MUL,
		TT_QUO: OP_DIV,
		TT_REM: OP_REM,
	}[tn.Tok.Kype]

	if !ok {
		PrintErrorAndExit(0)
	}

	if ok, ii := emitBinaryOp(op,
		callStackInfo[len(callStackInfo)-2], callStackInfo[len(callStackInfo)-1]); ok {

		callStackInfo = callStackInfo[:len(callStackInfo)-2]
		callStackInfo = append(callStackInfo, ii)
	} else {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}
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

			TNT_STMT_LIST: compileStmtList,

			TNT_STMT_DECL: compileStmtDecl,
			// TNT_STMT_DECL_IDENT
			// TNT_STMT_DECL_TYPE

			TNT_STMT_EXPR: compileStmtExpr,
			// TNT_STMT_ASSIGN
			// TNT_STMT_STORE_STRING
			// TNT_STMT_STRING

			// TNT_STMT_WHILE
			// TNT_STMT_IF
			// TNT_STMT_ELSE

			TNT_STMT_RETURN: compileStmtReturn,
			// TNT_STMT_BREAK
			// TNT_STMT_CONTINUE

			TNT_EXPR:                compileExpr,
			TNT_EXPR_INT:            compileExprInt,
			TNT_EXPR_FUNC:           compileExprFunc,
			TNT_EXPR_FUNC_PARM_LIST: compileExprFuncParmList,
			TNT_EXPR_FUNC_PARM:      compileExprFuncParm,
			// TNT_EXPR_INT_LIT
			// TNT_EXPR_NEG_INT_LIT
			// TNT_EXPR_CHAR
			TNT_EXPR_BINARY: compileExprBinary,
		}[c.Kype](c)
	}
}

func BytecodeGenerator(tn TreeNode) []byte {
	rootTreeNode = tn

	funcListTreeNode := rootTreeNode.Children[0]

	funcListInfoInit(funcListTreeNode)

	sigInfo, ok := funcListInfo["main"]

	if (!ok) || (len(sigInfo.ParamListInt) != 0) || (sigInfo.ReturnInt != nil) {
		PrintErrorAndExit(0)
	}

	blankMainFuncAddrAddr := emitBlankPushOp()

	emitOp(OP_CALL)
	emitOp(OP_HALT)

	fmt.Println(funcListInfo)

	compileTreeNodeChildren(tn.Children)

	mainFuncAddr, ok := funcListAddr["main"]

	if !ok {
		PrintErrorAndExit(0)
	}

	backpatchBlankPushOp(blankMainFuncAddrAddr, uint64(mainFuncAddr))

	fmt.Printf("%x\n", bytecode)

	return bytecode
}
