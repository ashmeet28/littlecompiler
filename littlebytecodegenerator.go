package main

import (
	"encoding/binary"
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

type CallStackValueType int

const (
	VT_ILLEGAL CallStackValueType = iota

	VT_I8
	VT_I16
	VT_I32
	VT_I64

	VT_U8
	VT_U16
	VT_U32
	VT_U64

	VT_FUNC_ADDR
	VT_RETURN_ADDR
	VT_PREV_FRAME_PTR
)

type CallStackValue struct {
	Kype       byte
	Ident      string
	IsLValue   bool
	BytesCount int
}

var callStack []CallStackValue

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
	compileTreeChildren(tn.Children)
}

func compileFuncIdent(tn TreeNode) {
	symTableAddFuncIdent(string(tn.Tok.Buf))
}

func compileFuncSig(tn TreeNode) {
}

func compileStmtList(tn TreeNode) {
	compileTreeChildren(tn.Children)
}

func compileStmtDecl(tn TreeNode) {
}

func compileTreeChildren(treeChildren []TreeNode) {
	treeNodeFuncs := map[TreeNodeType]func(TreeNode){
		TNT_FUNC_LIST:  compileFuncList,
		TNT_FUNC:       compileFunc,
		TNT_FUNC_IDENT: compileFuncIdent,
		TNT_FUNC_SIG:   compileFuncSig,
		TNT_STMT_LIST:  compileStmtList,
		TNT_STMT_DECL:  compileStmtDecl,
	}

	for _, c := range treeChildren {
		treeNodeFuncs[c.Kype](c)
	}
}

func BytecodeGenerator(tn TreeNode) []byte {
	bytecode = make([]byte, 0)
	symTable = make([]SymData, 0)
	blankPushOpAddrStack = make([]int, 0)

	emitBlankPushOp()

	emitOp(OP_CALL)
	emitOp(OP_HALT)

	compileTreeChildren(tn.Children)

	if ok, s := symTableFindFunc("main"); ok {
		fillBlankPushOp(uint64(s.Addr))
	} else {
		PrintErrorAndExit(0)
	}

	return bytecode
}
