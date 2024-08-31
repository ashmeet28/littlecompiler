package main

import "encoding/binary"

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

type CSIntData struct {
	BytesCount int
	IsSigned   bool
	IsLocal    bool
	Ident      string
	IsLValue   bool
	BlockLevel int
}

var currentBlockLevel int
var callStack []CSIntData

func pushLocalToCallStack(ident string) {
	var v CSIntData
	v.IsLocal = true
	v.Ident = ident
	v.BlockLevel = currentBlockLevel
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
	callStack = make([]CSIntData, 0)
	compileTreeChildren(tn.Children)
}

func compileFuncIdent(tn TreeNode) {
	funcAddrTable[string(tn.Tok.Buf)] = uint64(len(bytecode))
}

func compileFuncSig(tn TreeNode) {
	compileTreeChildren(tn.Children)
}

func compileFuncParamList(tn TreeNode) {
	compileTreeChildren(tn.Children)
}

func compileFuncParam(tn TreeNode) {
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
	blankPushOpAddrStack = make([]int, 0)

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
