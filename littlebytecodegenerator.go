package main

import (
	"encoding/binary"
	"strconv"
)

var (
	OP_HALT  byte = 0x01
	OP_ECALL byte = 0x02

	OP_CALL   byte = 0x04
	OP_RETURN byte = 0x05

	OP_JUMP   byte = 0x08
	OP_BRANCH byte = 0x09

	OP_PUSH   byte = 0x0c
	OP_POP    byte = 0x0d
	OP_ASSIGN byte = 0x0e

	OP_ADD byte = 0x40
	OP_SUB byte = 0x41

	OP_AND byte = 0x44
	OP_OR  byte = 0x45
	OP_XOR byte = 0x46

	OP_SHL byte = 0x48
	OP_SHR byte = 0x49

	OP_MUL byte = 0x4c
	OP_QUO byte = 0x4d
	OP_REM byte = 0x4e

	OP_EQL byte = 0x50
	OP_NEQ byte = 0x51
	OP_LSS byte = 0x52
	OP_GTR byte = 0x53
	OP_LEQ byte = 0x54
	OP_GEQ byte = 0x55

	OP_CONVERT byte = 0x58

	OP_LOAD  byte = 0x20
	OP_STORE byte = 0x21

	OP_STORE_STRING byte = 0x22
)

var bytecode []byte

type IntInfo struct {
	IsSigned   bool
	BytesCount int
}

type IntStorageInfo struct {
	Ident      string
	IsSigned   bool
	BlockLevel int
	BytesCount int
}

type IntAddressInfo struct {
	RealSize   int
	IsSigned   bool
	BytesCount int
}

type PreviousFrameAddressInfo struct {
	BytesCount int
}

type ReturnAddressInfo struct {
	BytesCount int
}

type VoidInfo struct {
	BytesCount int
}

func getIntInfoFromTypeString(s string) (IntInfo, bool) {
	if ii, ok := map[string]IntInfo{
		"i8":  {IsSigned: true, BytesCount: 1},
		"i16": {IsSigned: true, BytesCount: 2},
		"i32": {IsSigned: true, BytesCount: 4},
		"i64": {IsSigned: true, BytesCount: 8},

		"u8":  {IsSigned: false, BytesCount: 1},
		"u16": {IsSigned: false, BytesCount: 2},
		"u32": {IsSigned: false, BytesCount: 4},
		"u64": {IsSigned: false, BytesCount: 8},
	}[s]; ok {
		return ii, true
	} else {
		return IntInfo{}, false
	}
}

var ADDR_BYTES_COUNT int = 8

var blockLevel int
var returnValueInfo interface{}
var framePointer int

var whileBlockLevel int

var callStackInfo []interface{}

var STARTING_BLOCK_LEVEL int = 1

func callStackInfoReset() {
	blockLevel = STARTING_BLOCK_LEVEL
	whileBlockLevel = STARTING_BLOCK_LEVEL
	returnValueInfo = VoidInfo{BytesCount: 0}
	framePointer = 0
	callStackInfo = make([]interface{}, 0)
}

func callStackInfoGetBytesCount(i interface{}) int {
	switch v := i.(type) {
	case IntInfo:
		return v.BytesCount
	case IntStorageInfo:
		return v.BytesCount
	case IntAddressInfo:
		return v.BytesCount
	case PreviousFrameAddressInfo:
		return v.BytesCount
	case ReturnAddressInfo:
		return v.BytesCount
	case VoidInfo:
		return v.BytesCount
	default:
		PrintErrorAndExit(0)
		return 0
	}
}

func callStackInfoGetTotalBytesCount() int {
	var totalBytesCount int
	for _, i := range callStackInfo {
		totalBytesCount += callStackInfoGetBytesCount(i)
	}

	return totalBytesCount
}

func callStackInfoInitFrame() {
	callStackInfo = append(callStackInfo, PreviousFrameAddressInfo{BytesCount: ADDR_BYTES_COUNT})
	callStackInfo = append(callStackInfo, ReturnAddressInfo{BytesCount: ADDR_BYTES_COUNT})
	framePointer = callStackInfoGetTotalBytesCount()
}

func callStackInfoFindIntStorageInfo(intStorageInfoIdent string) (IntStorageInfo, bool) {
	for i := len(callStackInfo) - 1; i >= 0; i-- {
		if isi, ok := callStackInfo[i].(IntStorageInfo); ok && (isi.Ident == intStorageInfoIdent) {
			return isi, true
		}
	}

	return IntStorageInfo{}, false
}

func callStackInfoGetIntAddress(intStorageInfoIdent string) (uint64, bool) {
	var hasFound bool = false

	var totalBytesCount int

	for i := len(callStackInfo) - 1; i >= 0; i-- {
		if hasFound {
			totalBytesCount += callStackInfoGetBytesCount(callStackInfo[i])
		} else if isi, ok := callStackInfo[i].(IntStorageInfo); ok &&
			(isi.Ident == intStorageInfoIdent) {
			hasFound = true
		}
	}

	if hasFound {
		return (uint64(totalBytesCount) - uint64(framePointer)), true
	} else {
		return 0, false
	}
}

type FuncSigInfo struct {
	ParamListInt    []IntInfo
	ReturnValueInfo interface{}
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
				newFuncSigInfo.ReturnValueInfo = ii

			}

		}

		funcIdentTreeNode := funcTreeNode.Children[0]
		funcIdent := string(funcIdentTreeNode.Tok.Buf)

		if _, doesAlreadyExists := funcListInfo[funcIdent]; doesAlreadyExists {
			PrintErrorAndExit(funcIdentTreeNode.Tok.LineNumber)
		}

		if _, ok := newFuncSigInfo.ReturnValueInfo.(IntInfo); !ok {
			newFuncSigInfo.ReturnValueInfo = VoidInfo{BytesCount: 0}
		}

		funcListInfo[funcIdent] = newFuncSigInfo
	}
}

var funcAddrList map[string]int

var blankFuncAddrList map[string]int

var blankContinueStmtAddrList [][]int
var blankBreakStmtAddrList [][]int

func encodeIntInfo(ii IntInfo) byte {
	var b byte = byte(ii.BytesCount)

	if ii.IsSigned {
		return b | 0b10000
	}

	return b
}

func encodeIntAddressInfo(iai IntAddressInfo) byte {
	return (encodeIntInfo(IntInfo{IsSigned: iai.IsSigned, BytesCount: iai.RealSize}) | 0b100000)
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

func emitPopOp(ii IntInfo) {
	bytecode = append(bytecode, OP_POP)
	bytecode = append(bytecode, encodeIntInfo(ii))
}

func emitReturnOp(i interface{}) bool {
	switch v := i.(type) {
	case IntInfo:
		bytecode = append(bytecode, OP_RETURN)
		bytecode = append(bytecode, encodeIntInfo(v))

		return true
	case IntAddressInfo:
		bytecode = append(bytecode, OP_RETURN)
		bytecode = append(bytecode, encodeIntAddressInfo(v))

		return true
	case VoidInfo:
		bytecode = append(bytecode, OP_RETURN)
		bytecode = append(bytecode, encodeIntInfo(IntInfo{IsSigned: false, BytesCount: 0}))

		return true
	default:
		return false
	}
}

func emitBinaryOp(op byte, v1 interface{}, v2 interface{}) (bool, IntInfo) {
	var vb1 byte = 0
	var vb2 byte = 0

	var ii IntInfo

	switch v := v1.(type) {
	case IntInfo:
		vb1 = encodeIntInfo(v)
		ii = IntInfo{IsSigned: v.IsSigned, BytesCount: v.BytesCount}
	case IntAddressInfo:
		vb1 = encodeIntAddressInfo(v)
		ii = IntInfo{IsSigned: v.IsSigned, BytesCount: v.RealSize}
	default:
		return false, IntInfo{}
	}

	switch v := v2.(type) {
	case IntInfo:
		vb2 = encodeIntInfo(v)
	case IntAddressInfo:
		vb2 = encodeIntAddressInfo(v)
	default:
		return false, IntInfo{}
	}

	if _, ok := map[byte]bool{
		OP_EQL: true, OP_NEQ: true,
		OP_LSS: true, OP_GTR: true,
		OP_LEQ: true, OP_GEQ: true}[op]; ok {

		ii = IntInfo{IsSigned: false, BytesCount: 1}
	}

	if ((vb1 & 0b11111) == (vb2 & 0b11111)) || (op == OP_SHL) || (op == OP_SHR) {
		bytecode = append(bytecode, op)
		bytecode = append(bytecode, vb1)
		bytecode = append(bytecode, vb2)

		return true, ii
	}

	return false, IntInfo{}
}

func emitAssignOp(v1 interface{}, v2 interface{}) bool {
	var vb1 byte = 0
	var vb2 byte = 0

	switch v := v1.(type) {
	case IntAddressInfo:
		vb1 = encodeIntAddressInfo(v)
	default:
		return false
	}

	switch v := v2.(type) {
	case IntInfo:
		vb2 = encodeIntInfo(v)
	case IntAddressInfo:
		vb2 = encodeIntAddressInfo(v)
	default:
		return false
	}

	if (vb1 & 0b11111) == (vb2 & 0b11111) {
		bytecode = append(bytecode, OP_ASSIGN)
		bytecode = append(bytecode, vb1)
		bytecode = append(bytecode, vb2)

		return true
	}

	return false
}

func emitBranchOp(i interface{}) bool {
	switch v := i.(type) {
	case IntInfo:
		bytecode = append(bytecode, OP_BRANCH)
		bytecode = append(bytecode, encodeIntInfo(v))

		return true
	case IntAddressInfo:
		bytecode = append(bytecode, OP_BRANCH)
		bytecode = append(bytecode, encodeIntAddressInfo(v))

		return true
	default:
		return false
	}
}

func emitStoreStringOp(a interface{}, b []byte) bool {
	switch v := a.(type) {
	case IntAddressInfo:
		if (v.RealSize != 8) || (v.IsSigned) {
			return false
		}
	default:
		return false
	}

	bytecode = append(bytecode, OP_STORE_STRING)
	bytecode = append(bytecode, b...)
	bytecode = append(bytecode, 0)

	return true
}

func emitConvertOp(a interface{}, b IntInfo) bool {
	var vb byte

	switch v := a.(type) {
	case IntInfo:
		vb = encodeIntInfo(v)
	case IntAddressInfo:
		vb = encodeIntAddressInfo(v)
	default:
		return false
	}

	bytecode = append(bytecode, OP_CONVERT)
	bytecode = append(bytecode, vb)
	bytecode = append(bytecode, encodeIntInfo(b))

	return true
}

func compileFuncList(tn TreeNode) {
	funcAddrList = make(map[string]int)
	compileTreeNodeChildren(tn.Children)
}

func compileFunc(tn TreeNode) {
	callStackInfoReset()
	compileTreeNodeChildren(tn.Children)
}

func compileFuncIdent(tn TreeNode) {
	funcAddrList[string(tn.Tok.Buf)] = len(bytecode)
}

func compileFuncSig(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
	callStackInfoInitFrame()
}

func compileFuncParamList(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileFuncParam(tn TreeNode) {
	var isi IntStorageInfo

	funcParamIdentTreeNode := tn.Children[0]
	funcParamTypeTreeNode := tn.Children[1]

	isi.Ident = string(funcParamIdentTreeNode.Tok.Buf)

	if ii, ok := getIntInfoFromTypeString(string(funcParamTypeTreeNode.Tok.Buf)); ok {
		isi.IsSigned = ii.IsSigned
		isi.BytesCount = ii.BytesCount
	} else {
		PrintErrorAndExit(funcParamTypeTreeNode.Tok.LineNumber)
	}

	isi.BlockLevel = blockLevel

	callStackInfo = append(callStackInfo, isi)
}

func compileFuncReturnType(tn TreeNode) {
	if ii, ok := getIntInfoFromTypeString(string(tn.Tok.Buf)); ok {
		returnValueInfo = ii
	} else {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}
}

func compileStmtList(tn TreeNode) {
	blockLevel++

	compileTreeNodeChildren(tn.Children)

	blockLevel--

	for i := len(callStackInfo) - 1; i >= 0; i-- {
		if isi, ok := callStackInfo[i].(IntStorageInfo); ok && (isi.BlockLevel > blockLevel) {
			emitPopOp(IntInfo{IsSigned: isi.IsSigned, BytesCount: isi.BytesCount})
			callStackInfo = callStackInfo[:len(callStackInfo)-1]
		} else {
			break
		}
	}
}

func compileStmtDecl(tn TreeNode) {
	stmtDeclIdentTreeNode := tn.Children[0]
	stmtDeclTypeTreeNode := tn.Children[1]

	if isi, ok := callStackInfoFindIntStorageInfo(string(stmtDeclIdentTreeNode.Tok.Buf)); ok {
		if (isi.BlockLevel == blockLevel) ||
			((isi.BlockLevel == STARTING_BLOCK_LEVEL) && (blockLevel == STARTING_BLOCK_LEVEL+1)) {
			PrintErrorAndExit(stmtDeclIdentTreeNode.Tok.LineNumber)
		}
	}

	var isi IntStorageInfo

	isi.Ident = string(stmtDeclIdentTreeNode.Tok.Buf)

	if ii, ok := getIntInfoFromTypeString(string(stmtDeclTypeTreeNode.Tok.Buf)); ok {
		isi.IsSigned = ii.IsSigned
		isi.BytesCount = ii.BytesCount

		emitPushOp(ii, 0)
	} else {
		PrintErrorAndExit(stmtDeclTypeTreeNode.Tok.LineNumber)
	}

	isi.BlockLevel = blockLevel

	callStackInfo = append(callStackInfo, isi)
}

func compileStmtExpr(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)

	switch v := callStackInfo[len(callStackInfo)-1].(type) {
	case IntInfo:
		emitPopOp(v)
	case IntAddressInfo:
		emitPopOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT})
	}

	callStackInfo = callStackInfo[:len(callStackInfo)-1]
}

func compileStmtAssign(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)

	if ok := emitAssignOp(
		callStackInfo[len(callStackInfo)-2], callStackInfo[len(callStackInfo)-1]); ok {

		callStackInfo = callStackInfo[:len(callStackInfo)-2]
	} else {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}
}

func unescapeStmtString(b []byte) (bool, []byte) {
	var nb []byte

	if (len(b) < 3) || (b[0] != 0x22) || (b[len(b)-1] != 0x22) {
		return false, nb
	}

	b = b[1 : len(b)-1]

	for len(b) != 0 {
		if b[0] == 0x5c {
			b = b[1:]

			if len(b) == 0 {
				return false, nb
			}

			if (b[0] == 0x22) || (b[0] == 0x5c) {
				nb = append(nb, b[0])
				b = b[1:]
			} else {
				return false, nb
			}
		} else {
			nb = append(nb, b[0])
			b = b[1:]
		}
	}

	return true, nb
}

func compileStmtStoreString(tn TreeNode) {
	exprTreeNode := tn.Children[0]
	stmtStringTreeNode := tn.Children[1]

	compileTreeNode(exprTreeNode)

	if ok, b := unescapeStmtString(stmtStringTreeNode.Tok.Buf); ok {
		if ok := emitStoreStringOp(callStackInfo[len(callStackInfo)-1], b); ok {
			callStackInfo = callStackInfo[:len(callStackInfo)-1]
		} else {
			PrintErrorAndExit(tn.Tok.LineNumber)
		}
	} else {
		PrintErrorAndExit(stmtStringTreeNode.Tok.LineNumber)
	}
}

func compileStmtWhile(tn TreeNode) {
	stmtWhileStartingAddr := len(bytecode)

	exprTreeNode := tn.Children[0]
	stmtListTreeNode := tn.Children[1]

	compileTreeNode(exprTreeNode)

	stmtWhileBlankPushOpAddr := emitBlankPushOp()

	if ok := emitBranchOp(callStackInfo[len(callStackInfo)-1]); !ok {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}

	callStackInfo = callStackInfo[:len(callStackInfo)-1]

	blankBreakStmtAddrList = append(blankBreakStmtAddrList, make([]int, 0))
	blankContinueStmtAddrList = append(blankContinueStmtAddrList, make([]int, 0))

	whileBlockLevel = blockLevel

	compileTreeNode(stmtListTreeNode)

	emitPushOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT},
		uint64(stmtWhileStartingAddr)-(uint64(len(bytecode))+10))

	emitOp(OP_JUMP)

	backpatchBlankPushOp(stmtWhileBlankPushOpAddr,
		uint64(len(bytecode))-(uint64(stmtWhileBlankPushOpAddr)+10))

	for _, blankBreakStmtAddr := range blankBreakStmtAddrList[len(blankBreakStmtAddrList)-1] {
		backpatchBlankPushOp(blankBreakStmtAddr,
			uint64(len(bytecode))-(uint64(blankBreakStmtAddr)+10))
	}

	blankBreakStmtAddrList = blankBreakStmtAddrList[:len(blankBreakStmtAddrList)-1]

	for _, blankContinueStmtAddr := range blankContinueStmtAddrList[len(
		blankContinueStmtAddrList)-1] {

		backpatchBlankPushOp(blankContinueStmtAddr,
			uint64(stmtWhileStartingAddr)-(uint64(blankContinueStmtAddr)+10))
	}

	blankContinueStmtAddrList = blankContinueStmtAddrList[:len(blankContinueStmtAddrList)-1]
}

func compileStmtIf(tn TreeNode) {
	exprTreeNode := tn.Children[0]
	stmtListTreeNode := tn.Children[1]

	compileTreeNode(exprTreeNode)

	stmtIfBlankPushOpAddr := emitBlankPushOp()

	if ok := emitBranchOp(callStackInfo[len(callStackInfo)-1]); !ok {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}

	callStackInfo = callStackInfo[:len(callStackInfo)-1]

	compileTreeNode(stmtListTreeNode)

	if len(tn.Children) == 3 {
		stmtIfStmtListEndBlankPushOpAddr := emitBlankPushOp()

		emitOp(OP_JUMP)

		backpatchBlankPushOp(stmtIfBlankPushOpAddr,
			uint64(len(bytecode))-(uint64(stmtIfBlankPushOpAddr)+10))

		stmtElseTreeNode := tn.Children[2]
		compileTreeNode(stmtElseTreeNode)

		backpatchBlankPushOp(stmtIfStmtListEndBlankPushOpAddr,
			uint64(len(bytecode))-(uint64(stmtIfStmtListEndBlankPushOpAddr)+10))
	} else {
		backpatchBlankPushOp(stmtIfBlankPushOpAddr,
			uint64(len(bytecode))-(uint64(stmtIfBlankPushOpAddr)+10))
	}
}

func compileStmtElse(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileStmtReturn(tn TreeNode) {
	if len(tn.Children) == 0 {
		switch v := returnValueInfo.(type) {
		case IntInfo:
			emitPushOp(v, 0)
			emitPushOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT},
				uint64(-framePointer))
			if ok := emitReturnOp(v); !ok {
				PrintErrorAndExit(0)
			}
		case VoidInfo:
			emitPushOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT},
				uint64(-framePointer))
			if ok := emitReturnOp(v); !ok {
				PrintErrorAndExit(0)
			}
		default:
			PrintErrorAndExit(0)
		}
	} else {
		compileTreeNodeChildren(tn.Children)

		switch returnII := returnValueInfo.(type) {
		case IntInfo:
			switch v := callStackInfo[len(callStackInfo)-1].(type) {
			case IntInfo:
				if (v.BytesCount != returnII.BytesCount) || (v.IsSigned != returnII.IsSigned) {
					PrintErrorAndExit(tn.Tok.LineNumber)
				}
			case IntAddressInfo:
				if (v.RealSize != returnII.BytesCount) || (v.IsSigned != returnII.IsSigned) {
					PrintErrorAndExit(tn.Tok.LineNumber)
				}
			default:
				PrintErrorAndExit(tn.Tok.LineNumber)
			}

			emitPushOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT},
				uint64(-framePointer))

			if ok := emitReturnOp(returnII); !ok {
				PrintErrorAndExit(0)
			}
		case VoidInfo:
			PrintErrorAndExit(tn.Tok.LineNumber)
		default:
			PrintErrorAndExit(0)
		}
	}
}

func compileStmtBreak(tn TreeNode) {
	for i := len(callStackInfo) - 1; i >= 0; i-- {
		if isi, ok := callStackInfo[i].(IntStorageInfo); ok && (isi.BlockLevel > whileBlockLevel) {
			emitPopOp(IntInfo{IsSigned: isi.IsSigned, BytesCount: isi.BytesCount})
		} else {
			break
		}
	}

	if len(blankBreakStmtAddrList) != 0 {
		blankBreakStmtAddrList[len(blankBreakStmtAddrList)-1] = append(
			blankBreakStmtAddrList[len(blankBreakStmtAddrList)-1], emitBlankPushOp())

		emitOp(OP_JUMP)
	} else {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}
}

func compileStmtContinue(tn TreeNode) {
	for i := len(callStackInfo) - 1; i >= 0; i-- {
		if isi, ok := callStackInfo[i].(IntStorageInfo); ok && (isi.BlockLevel > whileBlockLevel) {
			emitPopOp(IntInfo{IsSigned: isi.IsSigned, BytesCount: isi.BytesCount})
		} else {
			break
		}
	}

	if len(blankContinueStmtAddrList) != 0 {
		blankContinueStmtAddrList[len(blankContinueStmtAddrList)-1] = append(
			blankContinueStmtAddrList[len(blankContinueStmtAddrList)-1], emitBlankPushOp())

		emitOp(OP_JUMP)
	} else {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}
}
func compileExpr(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileExprInt(tn TreeNode) {
	if isi, ok := callStackInfoFindIntStorageInfo(string(tn.Tok.Buf)); ok {
		var iai IntAddressInfo
		iai.RealSize = isi.BytesCount
		iai.IsSigned = isi.IsSigned
		iai.BytesCount = ADDR_BYTES_COUNT

		if a, ok := callStackInfoGetIntAddress(string(tn.Tok.Buf)); ok {
			emitPushOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT}, a)
			callStackInfo = append(callStackInfo, iai)
		} else {
			PrintErrorAndExit(0)
		}
	} else {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}
}

func unescapeExprChar(b []byte) (byte, bool) {
	if (len(b) < 3) || (b[0] != 0x27) || (b[len(b)-1] != 0x27) {
		return 0, false
	}

	if b[1] == 0x5c {
		if (len(b) == 4) && (b[2] == 0x5c || b[2] == 0x27) {
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

			if v, err :=
				strconv.ParseUint(string(exprIntLitTreeNode.Tok.Buf), 0, 64); err == nil {

				if v > ((^uint64(0)) >> ((8 - ii.BytesCount) * 8)) {
					PrintErrorAndExit(exprIntLitTreeNode.Tok.LineNumber)
				}
				emitPushOp(ii, v)
				callStackInfo = append(callStackInfo, ii)
			} else {
				PrintErrorAndExit(exprIntLitTreeNode.Tok.LineNumber)
			}

		case TNT_EXPR_NEG_INT_LIT:
			exprNegIntLitTreeNode := exprFuncParmTreeNode.Children[0]

			if v, err :=
				strconv.ParseUint(string(exprNegIntLitTreeNode.Tok.Buf), 0, 64); err == nil {

				if v > (uint64(1) << ((ii.BytesCount * 8) - 1)) {
					PrintErrorAndExit(exprNegIntLitTreeNode.Tok.LineNumber)
				}
				v = (^v) + 1
				emitPushOp(ii, v)
				callStackInfo = append(callStackInfo, ii)
			} else {
				PrintErrorAndExit(exprNegIntLitTreeNode.Tok.LineNumber)
			}
		case TNT_EXPR:
			compileTreeNode(exprFuncParmTreeNode.Children[0])

			if ok := emitConvertOp(callStackInfo[len(callStackInfo)-1], ii); ok {
				callStackInfo = callStackInfo[:len(callStackInfo)-1]
				callStackInfo = append(callStackInfo, ii)
			} else {
				PrintErrorAndExit(tn.Tok.LineNumber)
			}
		default:
			PrintErrorAndExit(0)
		}
	} else {
		callStackInfoLenBefore := len(callStackInfo)

		compileTreeNodeChildren(tn.Children)

		if fsi, ok := funcListInfo[string(tn.Tok.Buf)]; ok {
			if (len(callStackInfo) - callStackInfoLenBefore) != len(fsi.ParamListInt) {
				PrintErrorAndExit(tn.Tok.LineNumber)
			}

			for i, sigParam := range fsi.ParamListInt {
				if stackParam, ok :=
					callStackInfo[len(callStackInfo)-len(fsi.ParamListInt)+i].(IntInfo); ok {

					if (stackParam.BytesCount != sigParam.BytesCount) ||
						(stackParam.IsSigned != sigParam.IsSigned) {
						PrintErrorAndExit(tn.Tok.LineNumber)
					}

				} else {
					PrintErrorAndExit(tn.Tok.LineNumber)
				}
			}

			blankFuncAddrList[string(tn.Tok.Buf)] = emitBlankPushOp()

			emitOp(OP_CALL)

			callStackInfo = callStackInfo[:len(callStackInfo)-len(fsi.ParamListInt)]

			switch v := fsi.ReturnValueInfo.(type) {
			case IntInfo:
				callStackInfo = append(callStackInfo, v)
			case VoidInfo:
				callStackInfo = append(callStackInfo, v)
			default:
				PrintErrorAndExit(0)
			}
		} else {
			PrintErrorAndExit(tn.Tok.LineNumber)
		}
	}
}

func compileExprFuncParmList(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)
}

func compileExprFuncParm(tn TreeNode) {
	compileTreeNodeChildren(tn.Children)

	if iai, ok := callStackInfo[len(callStackInfo)-1].(IntAddressInfo); ok {
		ii := IntInfo{IsSigned: iai.IsSigned, BytesCount: iai.RealSize}

		emitPushOp(ii, 0)
		callStackInfo = append(callStackInfo, ii)

		if ok, ii := emitBinaryOp(OP_ADD,
			callStackInfo[len(callStackInfo)-2], callStackInfo[len(callStackInfo)-1]); ok {

			callStackInfo = callStackInfo[:len(callStackInfo)-2]
			callStackInfo = append(callStackInfo, ii)
		} else {
			PrintErrorAndExit(0)
		}
	}
}

func compileExprBinaryLAND(tn TreeNode) {
	leftTreeNode := tn.Children[0]
	rightTreeNode := tn.Children[1]

	compileTreeNode(leftTreeNode)

	blankPushOpAAddr := emitBlankPushOp()

	if ok := emitBranchOp(callStackInfo[len(callStackInfo)-1]); !ok {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}

	callStackInfo = callStackInfo[:len(callStackInfo)-1]

	compileTreeNode(rightTreeNode)

	blankPushOpBAddr := emitBlankPushOp()

	if ok := emitBranchOp(callStackInfo[len(callStackInfo)-1]); !ok {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}

	callStackInfo = callStackInfo[:len(callStackInfo)-1]

	ii := IntInfo{IsSigned: false, BytesCount: 1}

	emitPushOp(ii, 1)

	blankPushOpCAddr := emitBlankPushOp()

	emitOp(OP_JUMP)

	backpatchBlankPushOp(blankPushOpAAddr, uint64(len(bytecode))-(uint64(blankPushOpAAddr)+10))
	backpatchBlankPushOp(blankPushOpBAddr, uint64(len(bytecode))-(uint64(blankPushOpBAddr)+10))

	emitPushOp(ii, 0)

	backpatchBlankPushOp(blankPushOpCAddr, uint64(len(bytecode))-(uint64(blankPushOpCAddr)+10))

	callStackInfo = append(callStackInfo, ii)
}

func compileExprBinaryLOR(tn TreeNode) {
	leftTreeNode := tn.Children[0]
	rightTreeNode := tn.Children[1]

	compileTreeNode(leftTreeNode)

	blankPushOpAAddr := emitBlankPushOp()

	if ok := emitBranchOp(callStackInfo[len(callStackInfo)-1]); !ok {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}

	callStackInfo = callStackInfo[:len(callStackInfo)-1]

	ii := IntInfo{IsSigned: false, BytesCount: 1}

	onePushOpStartingAddr := len(bytecode)

	emitPushOp(ii, 1)

	blankPushOpBAddr := emitBlankPushOp()

	emitOp(OP_JUMP)

	backpatchBlankPushOp(blankPushOpAAddr, uint64(len(bytecode))-(uint64(blankPushOpAAddr)+10))

	compileTreeNode(rightTreeNode)

	blankPushOpCAddr := emitBlankPushOp()

	if ok := emitBranchOp(callStackInfo[len(callStackInfo)-1]); !ok {
		PrintErrorAndExit(tn.Tok.LineNumber)
	}

	callStackInfo = callStackInfo[:len(callStackInfo)-1]

	emitPushOp(IntInfo{IsSigned: false, BytesCount: ADDR_BYTES_COUNT},
		uint64(onePushOpStartingAddr)-(uint64(len(bytecode))+10))

	emitOp(OP_JUMP)

	backpatchBlankPushOp(blankPushOpCAddr, uint64(len(bytecode))-(uint64(blankPushOpCAddr)+10))

	emitPushOp(ii, 0)

	backpatchBlankPushOp(blankPushOpBAddr, uint64(len(bytecode))-(uint64(blankPushOpBAddr)+10))

	callStackInfo = append(callStackInfo, ii)
}

func compileExprBinary(tn TreeNode) {
	if tn.Tok.Kype == TT_LAND {

		compileExprBinaryLAND(tn)

	} else if tn.Tok.Kype == TT_LOR {

		compileExprBinaryLOR(tn)

	} else {

		compileTreeNodeChildren(tn.Children)

		op, ok := map[TokenType]byte{
			TT_ADD: OP_ADD,
			TT_SUB: OP_SUB,

			TT_AND: OP_AND,
			TT_OR:  OP_OR,
			TT_XOR: OP_XOR,

			TT_SHL: OP_SHL,
			TT_SHR: OP_SHR,

			TT_MUL: OP_MUL,
			TT_QUO: OP_QUO,
			TT_REM: OP_REM,

			TT_EQL: OP_EQL,
			TT_NEQ: OP_NEQ,
			TT_LSS: OP_LSS,
			TT_GTR: OP_GTR,
			TT_LEQ: OP_LEQ,
			TT_GEQ: OP_GEQ,
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
}

func compileTreeNode(tn TreeNode) {
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

		TNT_STMT_EXPR:         compileStmtExpr,
		TNT_STMT_ASSIGN:       compileStmtAssign,
		TNT_STMT_STORE_STRING: compileStmtStoreString,
		// TNT_STMT_STRING

		TNT_STMT_WHILE: compileStmtWhile,
		TNT_STMT_IF:    compileStmtIf,
		TNT_STMT_ELSE:  compileStmtElse,

		TNT_STMT_RETURN:   compileStmtReturn,
		TNT_STMT_BREAK:    compileStmtBreak,
		TNT_STMT_CONTINUE: compileStmtContinue,

		TNT_EXPR:                compileExpr,
		TNT_EXPR_INT:            compileExprInt,
		TNT_EXPR_FUNC:           compileExprFunc,
		TNT_EXPR_FUNC_PARM_LIST: compileExprFuncParmList,
		TNT_EXPR_FUNC_PARM:      compileExprFuncParm,
		// TNT_EXPR_INT_LIT
		// TNT_EXPR_NEG_INT_LIT
		// TNT_EXPR_CHAR
		TNT_EXPR_BINARY: compileExprBinary,
	}[tn.Kype](tn)
}

func compileTreeNodeChildren(treeNodeChildren []TreeNode) {
	for _, tn := range treeNodeChildren {
		compileTreeNode(tn)
	}
}

func BytecodeGenerator(tn TreeNode) []byte {
	funcListTreeNode := tn.Children[0]

	funcListInfoInit(funcListTreeNode)

	sigInfo, ok := funcListInfo["main"]

	if (!ok) || (len(sigInfo.ParamListInt) != 0) {
		PrintErrorAndExit(0)
	}

	if _, ok := sigInfo.ReturnValueInfo.(VoidInfo); !ok {
		PrintErrorAndExit(0)
	}

	blankFuncAddrList = map[string]int{}

	blankFuncAddrList["main"] = emitBlankPushOp()

	emitOp(OP_CALL)
	emitOp(OP_HALT)

	compileTreeNodeChildren(tn.Children)

	for funcIdent, blankPushOpAddr := range blankFuncAddrList {
		if funcAddr, ok := funcAddrList[funcIdent]; ok {
			backpatchBlankPushOp(blankPushOpAddr, uint64(funcAddr)-(uint64(blankPushOpAddr)+10))
		} else {
			PrintErrorAndExit(0)
		}
	}

	return bytecode
}
