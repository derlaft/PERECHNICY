package asm

import ()

//
// INSTRUCTIONS
//

type InstHandler func(*State, Instruction)

// Instruction map
var Handlers = map[string]InstHandler{
	// No operation
	"NOP": Nop,
	// Ariphmetic instructions
	"ADD": Add,
	"SUB": Sub,
	"INC": Inc,
	"DEC": Dec,
	"MUL": Mul,
	"DIV": Div,
	"MOD": Mod,
	// Logic instructions
	"AND": And,
	"OR":  Or,
	"XOR": Xor,
	"NOT": Not,
	// Shift instructions
	"ROL": Rol,
	"ROR": Ror,
	"RCL": Rcl,
	"RCR": Rcr,
	// Data move instructions
	"MOV": Mov,
	"LD":  Ld,
	"ST":  St,
	// Programm stack instructions
	"PUT": Push,
	"POP": Pop,
	// Compare instruction
	"CMP": Cmp,
	// Jumps instructions
	"JZ":  Jz,
	"JNZ": Jnz,
	"JO":  Jo,
	"JNO": Jno,
	"JF":  Jf,
	"JNF": Jnf,
	"JI":  Ji,
	"JNI": Jni,
	"JL":  Jl,
	"JNL": Jnl,
	"JE":  Je,
	"JNE": Jne,
	"JG":  Jg,
	"JNG": Jng,
	"JMP": Jmp,
	// Procedure instructions
	"CAL": Call,
	"RET": Ret,
	// Fault instruction
	"FLT": Flt,
	// Software interrupt
	"INT": Int,
}

// Instruction that changes first operand (return values)
var RetInstructions = []string{
	"ADD",
	"SUB",
	"INC",
	"DEC",
	"MUL",
	"DIV",
	"MOD",
	"AND",
	"OR",
	"XOR",
	"NOT",
	"ROL",
	"ROR",
	"RCL",
	"RCR",
	"MOV",
	"LD",
	"POP",
}

// Instruction argument
type Argument struct {
	Type byte
	Val  int
}

// Argument types: register, constant, label
const (
	ARG_REG   = 0
	ARG_CONST = 1
	ARG_LABEL = 2
)

// Command structure (generated by parser?)
type Command struct {
	InstName string      // Instruction name
	Args     [2]Argument // Instruction arguments
	Used     byte        // FIXME: WUT?
}

// Instruction structure
type Instruction struct {
	handler  InstHandler // Instruction handler
	Args     [2]byte     // Arguments
	JumpAddr int         // Jump address
	RetPtr   *byte       // Return value pointer
}

// Generate instruction from command
func (cmd *Command) toInstruction(state *State) Instruction {

	inst := Instruction{}
	inst.handler = Handlers[cmd.InstName]

	for i, arg := range cmd.Args {
		switch arg.Type {
		case ARG_REG:
			reg := arg.Val & 0xF
			inst.Args[i] = state.Reg[reg]
			if i == 0 {
				inst.RetPtr = &state.Reg[reg]
			}
		case ARG_CONST:
			inst.Args[i] = byte(arg.Val)
		case ARG_LABEL:
			inst.JumpAddr = arg.Val
		}
	}

	for _, v := range RetInstructions {
		if v == cmd.InstName && inst.RetPtr == nil {
			inst.handler = Flt
		}
	}

	return inst
}

//
// MACHINE STATE
//

// Digital machine state structure
type State struct {
	Reg [16]byte  // Registers: R0 ... RD, Stack pointer: RE (RSP), Flags RF
	Mem [256]byte // Data memory
	IP  int       // Instruction pointer (untouchable)
	Ret [32]int   // Returning stack array (untouchable)
	Rpt byte      // Return from procedure pointer

	Interrupt *byte // Interrupt flag; zero is no Interrupt
}

// Register alternate names
const (
	RA  = 10
	RB  = 11
	RC  = 12
	RD  = 13
	RE  = 14 // Programm stack pointer register
	RSP = 14
	RF  = 15 // Flag register
)

// Digital machine cycle
func (state *State) Cycle(cmd Command) error {

	inst := cmd.toInstruction(state)
	inst.handler(state, inst)

	return nil
}

//
// HANDLERS REALISATION
//

// JUMPS
func Jmp(state *State, inst Instruction) {
	state.IP = inst.JumpAddr
}

func JmpIfFlag(state *State, inst Instruction, flag byte, rev bool) {
	if (state.Reg[RF]&(1<<flag) > 0) == rev {
		Jmp(state, inst)
	}
}

// Flag register structure
const (
	F_ZERO  = 0
	F_OVER  = 1
	F_FAULT = 2
	F_INT   = 3
	F_LESS  = 4
	F_EQUAL = 5
	F_GREAT = 6
)

func Cmp(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)
}

func Jz(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_ZERO, false)
}

func Jnz(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_ZERO, true)
}

func Jo(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_OVER, true)
}

func Jno(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_OVER, true)
}

func Jf(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_FAULT, true)
}

func Jnf(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_FAULT, false)
}

func Ji(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_INT, false)
}

func Jni(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_INT, true)
}

func Jl(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_LESS, false)
}

func Jnl(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_LESS, true)
}

func Je(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_EQUAL, true)
}

func Jne(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_EQUAL, false)
}

func Jg(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_GREAT, false)
}
func Jng(state *State, inst Instruction) {
	JmpIfFlag(state, inst, F_GREAT, true)
}

// PROCEDURES

func Call(state *State, inst Instruction) {

	if state.Rpt == 31 {
		state.Reg[RF] |= (1 << F_FAULT)
		return
	}

	state.Rpt += 1
	state.Ret[state.Rpt] = state.IP
	state.IP = inst.JumpAddr
}

func Ret(state *State, inst Instruction) {

	if state.Rpt == 0 {
		state.Reg[RF] |= (1 << F_FAULT)
		return
	}

	state.IP = state.Ret[state.Rpt]
	state.Rpt -= 1
}

// OTHER OPERATIONS

// Dual args flags: LESS, EQUAL, GREAT
func (state *State) flagArgs(first byte, second byte) {

	// Reset args flags
	state.Reg[RF] &= ((1 << F_LESS) ^ 0xFF)
	state.Reg[RF] &= ((1 << F_EQUAL) ^ 0xFF)
	state.Reg[RF] &= ((1 << F_GREAT) ^ 0xFF)

	// Set args flags
	if first < second {
		state.Reg[RF] |= (1 << F_LESS)
	}

	if first == second {
		state.Reg[RF] |= (1 << F_EQUAL)
	}

	if first > second {
		state.Reg[RF] |= (1 << F_GREAT)
	}
}

// Result flags: OVER, FAULT, ZERO
func (state *State) flagResult(result int) byte {

	// Reset args flags
	state.Reg[RF] &= ((1 << F_FAULT) ^ 0xFF)
	state.Reg[RF] &= ((1 << F_OVER) ^ 0xFF)
	state.Reg[RF] &= ((1 << F_ZERO) ^ 0xFF)

	// Check byte bounds
	switch {
	case result < 0:
		result = 0xFF + (result % 0xFF) // WARNING: may be need zero value or `0xFF - result`
		state.Reg[RF] |= (1 << F_OVER)
	case result > 0xFF:
		result = result & 0xFF
		state.Reg[RF] |= (1 << F_OVER)
	case result == 0: // Check for zero
		state.Reg[RF] |= (1 << F_ZERO)
	}

	return byte(result)
}

func Add(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)
	*inst.RetPtr = state.flagResult(int(first) + int(second))
}

func Sub(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)
	*inst.RetPtr = state.flagResult(int(first) - int(second))
}

func Inc(state *State, inst Instruction) {
	*inst.RetPtr = state.flagResult(int(inst.Args[0]) + 1)
}

func Dec(state *State, inst Instruction) {
	*inst.RetPtr = state.flagResult(int(inst.Args[0]) - 1)
}

func Mul(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)
	*inst.RetPtr = state.flagResult(int(first) * int(second))
}

func Div(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)

	if inst.Args[1] == 0 {
		// Divide by zero fault
		state.Reg[RF] |= (1 << F_FAULT)
	} else {
		*inst.RetPtr = state.flagResult(int(inst.Args[0]) / int(inst.Args[1]))
	}
}

func Mod(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)

	if inst.Args[1] == 0 {
		// Divide by zero fault
		state.Reg[RF] |= (1 << F_FAULT)
	} else {
		*inst.RetPtr = state.flagResult((int(inst.Args[0]) % int(inst.Args[1])))
	}
}

func And(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)
	*inst.RetPtr = state.flagResult(int(first) & int(second))
}

func Or(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)
	*inst.RetPtr = state.flagResult(int(first) | int(second))
}

func Xor(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)
	*inst.RetPtr = state.flagResult(int(first) ^ int(second))
}

func Not(state *State, inst Instruction) {
	*inst.RetPtr = state.flagResult(int(inst.Args[0] ^ 0xFF))
}

func Rol(state *State, inst Instruction) {
	*inst.RetPtr = state.flagResult(int(inst.Args[0] << 1))
}

func Ror(state *State, inst Instruction) {
	*inst.RetPtr = state.flagResult(int(inst.Args[0] >> 1))
}

func Rcl(state *State, inst Instruction) {
	*inst.RetPtr = state.flagResult(int(inst.Args[0]<<1) | int(inst.Args[0]>>7))
}

func Rcr(state *State, inst Instruction) {
	*inst.RetPtr = state.flagResult(int(inst.Args[0]>>1) | int(inst.Args[0]<<7))
}

// Data move
func Mov(state *State, inst Instruction) {
	first := inst.Args[0]
	second := inst.Args[1]
	state.flagArgs(first, second)
	*inst.RetPtr = state.flagResult(int(second))
}

func Ld(state *State, inst Instruction) {
	*inst.RetPtr = state.flagResult(int(state.Mem[inst.Args[1]]))
}

func St(state *State, inst Instruction) {
	state.flagResult(int(inst.Args[0]))
	state.Mem[inst.Args[1]] = inst.Args[0]
}

// Stack
func Push(state *State, inst Instruction) {
	if state.Reg[RSP] < 128 || state.Reg[RSP] > 127+64 {
		state.Reg[RF] |= (1 << F_FAULT)
	} else {
		state.flagResult(int(inst.Args[0]))
		state.Mem[state.Reg[RSP]] = inst.Args[0]
		state.Reg[RSP] = byte(int(state.Reg[RSP]) + 1)
	}
}

func Pop(state *State, inst Instruction) {
	if state.Reg[RSP] < 128 || state.Reg[RSP] > 127+64 {
		state.Reg[RF] |= (1 << F_FAULT)
	} else {
		state.flagResult(int(inst.Args[0]))
		state.Reg[RSP] = byte(int(state.Reg[RSP]) - 1)
		*inst.RetPtr = state.flagResult(int(state.Mem[state.Reg[RSP]]))
	}
}

// Generate fault state
func Flt(state *State, inst Instruction) {
	state.Reg[RF] |= (1 << F_FAULT)
}

// Nop-nop-nop
func Nop(state *State, inst Instruction) {
	// NO OPERATION
}

func Int(state *State, inst Instruction) {
	state.Interrupt = &inst.Args[0]
}
