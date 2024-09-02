package main

import (
	"encoding/binary"
	"strconv"
)

const (
	OP_ADD byte = iota
	OP_SUB
	OP_MUL
	OP_QUO
	OP_REM

	OP_AND
	OP_OR
	OP_XOR

	OP_SHL
	OP_SHR

	OP_EQL
	OP_NEQ
	OP_LSS
	OP_GTR
	OP_LEQ
	OP_GEQ
)

var (
	OP_HALT  byte = 0x01
	OP_ECALL byte = 0x02

	OP_CALL   byte = 0x04
	OP_JUMP   byte = 0x05
	OP_BRANCH byte = 0x06
	OP_RETURN byte = 0x07

	OP_PUSH byte = 0x08
	OP_POP  byte = 0x09
)

type IntInfo struct {
	IsSigned   bool
	RealSize   int
	IsLocal    bool
	Ident      string
	IsLValue   bool
	BlockLevel int
	BytesCount int
}

var blockLevel int
var returnIntInfo IntInfo
var framePointer int
var callStack []IntInfo

func isValidIntKype(kype string) bool {
	_, ok := map[string]bool{"i8": true, "i16": true, "i32": true, "i64": true,
		"u8": true, "u16": true, "u32": true, "u64": true}[kype]
	return ok
}

func callStackPushLocal(ident string, kype string) bool {
	if !isValidIntKype(kype) {
		return false
	}

	var v IntInfo

	v.IsSigned = true
	if kype[:1] == "u" {
		v.IsSigned = false
	}

	i, err := strconv.ParseInt(kype[1:], 10, 64)
	if err != nil {
		return false
	}

	v.RealSize = int(i / 8)

	v.IsLocal = true
	v.Ident = ident
	v.IsLValue = false
	v.BlockLevel = blockLevel
	v.BytesCount = v.RealSize

	callStack = append(callStack, v)

	return true
}

func callStackPushRetAddrAndFramePointer() {
	var v IntInfo

	v.IsSigned = false
	v.RealSize = 0
	v.IsLocal = false
	v.Ident = ""
	v.IsLValue = true
	v.BlockLevel = blockLevel
	v.BytesCount = v.RealSize

	callStack = append(callStack, v)
}

func setReturnIntInfo(kype string) bool {
	if !isValidIntKype(kype) {
		return false
	}

	var v IntInfo

	v.IsSigned = true
	if kype[:1] == "u" {
		v.IsSigned = false
	}

	i, err := strconv.ParseInt(kype[1:], 10, 64)
	if err != nil {
		return false
	}

	v.RealSize = int(i / 8)

	v.IsLocal = false
	v.Ident = ""
	v.IsLValue = false
	v.BlockLevel = blockLevel
	v.BytesCount = v.RealSize

	returnIntInfo = v

	return true
}

func incBlockLevel() {
	blockLevel++
}

func decBlockLevel() {
	blockLevel--
}

var funcAddrTable map[string]uint64

var bytecode []byte

var blankPushOpAddrStack []int

func emitBlankPushOp() {
	blankPushOpAddrStack = append(blankPushOpAddrStack, len(bytecode))
	bytecode = append(bytecode, OP_PUSH)
	bytecode = binary.LittleEndian.AppendUint64(bytecode, 0)
}

func fillBlankPushOp(v uint64) {
	addr := blankPushOpAddrStack[len(blankPushOpAddrStack)-1]
	for i, b := range binary.LittleEndian.AppendUint64(make([]byte, 0), v) {
		bytecode[addr+i+1] = b
	}
	blankPushOpAddrStack = blankPushOpAddrStack[:len(blankPushOpAddrStack)-1]
}

func emitOp(op byte) {
	bytecode = append(bytecode, op)
}

func compileFuncList(tn TreeNode) {
	compileTreeChildren(tn.Children)
}

func compileFunc(tn TreeNode) {
	blockLevel = 0
	returnIntInfo = IntInfo{}
	callStack = make([]IntInfo, 0)
	compileTreeChildren(tn.Children)
}

func compileFuncIdent(tn TreeNode) {
	funcAddrTable[string(tn.Tok.Buf)] = uint64(len(bytecode))
}

func compileFuncSig(tn TreeNode) {
	incBlockLevel()
	compileTreeChildren(tn.Children)
}

func compileFuncParamList(tn TreeNode) {
	compileTreeChildren(tn.Children)
}

func compileFuncParam(tn TreeNode) {
	callStackPushLocal(string(tn.Children[0].Tok.Buf), string(tn.Children[1].Tok.Buf))
}

func compileFuncReturnType(tn TreeNode) {
	setReturnIntInfo(string(tn.Tok.Buf))
}

func compileStmtList(tn TreeNode) {
	incBlockLevel()
	compileTreeChildren(tn.Children)
}

func compileTreeChildren(treeChildren []TreeNode) {
	treeNodeFuncs := map[TreeNodeType]func(TreeNode){

		TNT_FUNC_LIST: compileFuncList,
		TNT_FUNC:      compileFunc,

		TNT_FUNC_IDENT:       compileFuncIdent,
		TNT_FUNC_SIG:         compileFuncSig,
		TNT_FUNC_PARAM_LIST:  compileFuncParamList,
		TNT_FUNC_PARAM:       compileFuncParam,
		TNT_FUNC_RETURN_TYPE: compileFuncReturnType,

		TNT_STMT_LIST: compileStmtList,

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

	}

	for _, c := range treeChildren {
		treeNodeFuncs[c.Kype](c)
	}
}

func BytecodeGenerator(tn TreeNode) []byte {
	bytecode = make([]byte, 0)
	blankPushOpAddrStack = make([]int, 0)
	funcAddrTable = map[string]uint64{}

	emitBlankPushOp()

	emitOp(OP_CALL)
	emitOp(OP_HALT)

	compileTreeChildren(tn.Children)

	if funcAddr, ok := funcAddrTable["main"]; ok {
		fillBlankPushOp(uint64(funcAddr))
	} else {
		PrintErrorAndExit(0)
	}

	return bytecode
}
