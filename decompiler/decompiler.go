package main

import (
	"fmt"
	"io/ioutil"
	"flag"
	"stack"
)

const NREG = 8
const ROLL = (1 << 15)
const HALT = (1 << 16) - 1
const MEMSIZE = ROLL - 1

var stack = new(Stack)
var memory [MEMSIZE]uint16
var registers [NREG]uint16

// Opcodes receive their memory address and return address of next opcode
var opcodes = map[uint16]func(uint16) uint16{
	0: func(addr uint16) uint16 {
		fmt.Printf("%d\thlt\n", addr)
		return addr + 1
	},
	1: func(addr uint16) uint16 {
		fmt.Printf("%d\tset %d %d (%d)\n", addr, get_raw(addr + 1), get_raw(addr + 2), get(addr+2))
		a := get_raw(addr + 1)
		b := get(addr + 2)
		set(a, b)
		return addr + 3
	},
	2: func(addr uint16) uint16 {
		fmt.Printf("%d\tpush %d\n", addr, get_raw(addr+1))
		stack.Push(get(addr + 1))
		return addr + 2
	},
	3: func(addr uint16) uint16 {
		fmt.Printf("%d\tpop %d\n", addr, get_raw(addr+1))
		if stack.Len() > 0 {
			set(get_raw(addr+1), stack.Pop().(uint16))
		}
		return addr + 2
	},
	4: func(addr uint16) uint16 {
		fmt.Printf("%d\teq %d %d %d\n", addr, get_raw(addr+1), get_raw(addr+2), get_raw(addr+3))
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		val := uint16(0)
		if b == c {
			val = 1
		}
		set(a, val)
		return addr + 4
	},
	5: func(addr uint16) uint16 {
		fmt.Printf("%d\tgt %d %d %d\n", addr, get_raw(addr+1), get_raw(addr+2), get_raw(addr+3))
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		val := uint16(0)
		if b > c {
			val = 1
		}
		set(a, val)
		return addr + 4
	},
	6: func(addr uint16) uint16 {
		fmt.Printf("%d\tjmp %d\n", addr, get_raw(addr+1))
		return addr + 2
	},
	7: func(addr uint16) uint16 {
		fmt.Printf("%d\tjt %d(%d) %d(%d)\n", addr, get_raw(addr+1), get(addr+1), get_raw(addr+2), get(addr+2))
		return addr + 3
	},
	8: func(addr uint16) uint16 {
		fmt.Printf("%d\tjf %d(%d) %d(%d)\n", addr, get_raw(addr+1), get(addr+1), get_raw(addr+2), get(addr+2))
		return addr + 3
	},
	9: func(addr uint16) uint16 {
		fmt.Printf("%d\tadd %d %d %d\n", addr, get_raw(addr+1), get_raw(addr+2), get_raw(addr+3))
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, (b + c) % ROLL)
		return addr + 4
	},
	10: func(addr uint16) uint16 {
		fmt.Printf("%d\tmult %d %d %d\n", addr, get_raw(addr+1), get_raw(addr+2), get_raw(addr+3))
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, (b * c) % ROLL)
		return addr + 4
	},
	11: func(addr uint16) uint16 {
		fmt.Printf("%d\tmod %d %d %d\n", addr, get_raw(addr+1), get_raw(addr+2), get_raw(addr+3))
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, b%c)
		return addr + 4
	},
	12: func(addr uint16) uint16 {
		fmt.Printf("%d\tand %d %d %d\n", addr, get_raw(addr+1), get_raw(addr+2), get_raw(addr+3))
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, b&c)
		return addr + 4
	},
	13: func(addr uint16) uint16 {
		fmt.Printf("%d\tor %d %d %d\n", addr, get_raw(addr+1), get_raw(addr+2), get_raw(addr+3))
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, b|c)
		return addr + 4
	},
	14: func(addr uint16) uint16 {
		fmt.Printf("%d\tnot %d %d\n", addr, get_raw(addr+1), get_raw(addr+2))
		a := get_raw(addr + 1)
		b := get(addr + 2)
		b = b << 1
		b = ^b
		b = b >> 1
		set(a, b)
		return addr + 3
	},
	15: func(addr uint16) uint16 {
		fmt.Printf("%d\trmem %d %d\n", addr, get_raw(addr+1), get_raw(addr+2))
		a := get_raw(addr + 1)
		b := get_raw(addr + 2)
		set(a, b)
		return addr + 3
	},
	16: func(addr uint16) uint16 {
		fmt.Printf("%d\twmem %d %d\n", addr, get_raw(addr+1), get_raw(addr+2))
		a := get(addr + 1)
		b := get(addr + 2)
		set(a, b)
		return addr + 3
	},
	17: func(addr uint16) uint16 {
		fmt.Printf("%d\tcall %d (%d)\n", addr, get_raw(addr+1), get(addr+1))
		stack.Push(addr + 2)
		return addr + 2
	},
	18: func(addr uint16) uint16 {
		var val uint16
		if stack.Len() == 0 {
			val = 0
		} else {
		val = stack.Pop().(uint16)
		}
		fmt.Printf("%d\tret (%v)\n", addr, val)
		return addr + 1
	},
	19: func(addr uint16) uint16 {
		fmt.Printf("%d\tout %d\n", addr, get_raw(addr+1))
		return addr + 2
	},
	20: func(addr uint16) uint16 {
		fmt.Printf("%d\tin %d\n", addr, get_raw(addr+1))
		return addr + 2
	},
	21: func(addr uint16) uint16 {
		fmt.Printf("%d\tnoop\n", addr)
		return addr + 1
	},
}

func get_raw(addr uint16) uint16 {
	if is_mem(addr) {
		return memory[addr]
	} else if is_reg(addr) {
		return registers[addr%ROLL]
	}

	return 0
}

func get(addr uint16) uint16 {
	if is_mem(addr) {
		val := memory[addr]
		if is_reg(val) {
			val = get(val)
		}
		return val
	} else if is_reg(addr) {
		return registers[addr%ROLL]
	}

	return 0
}

func set(addr, val uint16) uint16 {
	if is_reg(val) {
		val = get(val)
	}

	if is_mem(addr) {
		memory[addr] = val
		return memory[addr]
	} else if is_reg(addr) {
		registers[addr%ROLL] = val
		return registers[addr%ROLL]
	}

	return 0
}

func main() {
	load_bin()
	for addr := uint16(0); addr < ROLL; {
		tmp := get(addr)
		if fnc, ok := opcodes[tmp]; ok {
			addr = fnc(addr)
		} else { addr++ }
	}
}

func load_bin() {
	flag.Parse()
	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Printf("Read error: %v\n", err.String())
		return
	}
	n, j := len(data), 0
	for i := 0; j < n && i < MEMSIZE; i++ {
		memory[i] = uint16(data[j]) | (uint16(data[j+1]) << 8)
		j = j + 2
	}
}

func is_mem(a uint16) bool { return a < ROLL }
func is_reg(a uint16) bool { return a >= ROLL && a <= MEMSIZE+NREG }
