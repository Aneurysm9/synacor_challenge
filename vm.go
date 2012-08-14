package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
)

type Stack struct {
	top *Element
	size int
}

type Element struct {
	value interface{}
	next *Element
}

// Return the stack's length
func (s *Stack) Len() int {
	return s.size
}

// Push a new element onto the stack
func (s *Stack) Push(value interface{}) {
	s.top = &Element{value, s.top}
	s.size++
}

// Remove the top element from the stack and return it's value
// If the stack is empty, return nil
func (s *Stack) Pop() (value interface{}) {
	if s.size > 0 {
		value, s.top = s.top.value, s.top.next
		s.size--
		return
	}
	return nil
}

const NREG = 8
const ROLL = (1 << 15)
const HALT = (1 << 16) - 1
const MEMSIZE = ROLL

var stack = new(Stack)
var memory [MEMSIZE]uint16
var registers [NREG]uint16
var debug bool
var breakpoint uint16 = MEMSIZE

// Opcodes receive their memory address and return address of next opcode
var opcodes = [22]func(uint16) uint16{
	0: func(addr uint16) uint16 {
		if debug {
			fmt.Printf("%d\thlt\n", addr)
		}
		return HALT
	},
	1: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get(addr + 2)
		set(a, b)
		if debug {
			fmt.Printf("%d\tset %d %d (%d)\n", addr, a, get_raw(addr+2), b)
		}
		return addr + 3
	},
	2: func(addr uint16) uint16 {
		stack.Push(get(addr + 1))
		if debug {
			fmt.Printf("%d\tpush %d (%d)\n", addr, get_raw(addr+1), get(addr+1))
		}
		return addr + 2
	},
	3: func(addr uint16) uint16 {
		if stack.Len() > 0 {
			set(get_raw(addr+1), stack.Pop().(uint16))
			if debug {
				fmt.Printf("%d\tpop %d\n", addr, get_raw(addr+1))
			}
		} else {
			if debug {
				fmt.Printf("%d\tpop with empty stack\n", addr)
			}
		}
		return addr + 2
	},
	4: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		val := uint16(0)
		if b == c {
			val = 1
		}
		set(a, val)
		if debug {
			fmt.Printf("%d\teq %d %d(%d) %d(%d)\n", addr, a, get_raw(addr+2), b, get_raw(addr+3), c)
		}
		return addr + 4
	},
	5: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		val := uint16(0)
		if b > c {
			val = 1
		}
		set(a, val)
		if debug {
			fmt.Printf("%d\tgt %d %d(%d) %d(%d)\n", addr, a, get_raw(addr+2), b, get_raw(addr+3), c)
		}
		return addr + 4
	},
	6: func(addr uint16) uint16 {
		if debug {
			fmt.Printf("%d\tjmp %d(%d)\n", addr, get_raw(addr+1), get(addr+1))
		}
		return get(addr + 1)
	},
	7: func(addr uint16) uint16 {
		a := get(addr + 1)
		b := get(addr + 2)
		if debug {
			fmt.Printf("%d\tjt %d(%d) %d(%d)\n", addr, get_raw(addr+1), a, get_raw(addr+2), b)
		}
		if a != 0 {
			return b
		}
		return addr + 3
	},
	8: func(addr uint16) uint16 {
		a := get(addr + 1)
		b := get(addr + 2)
		if debug {
			fmt.Printf("%d\tjf %d(%d) %d(%d)\n", addr, get_raw(addr+1), a, get_raw(addr+2), b)
		}
		if a == 0 {
			return b
		}
		return addr + 3
	},
	9: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, (b+c)%ROLL)
		if debug {
			fmt.Printf("%d\tadd %d %d(%d) %d(%d)\n", addr, get_raw(addr+1), get_raw(addr+2), b, get_raw(addr+3), c)
		}
		return addr + 4
	},
	10: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, (b*c)%ROLL)
		if debug {
			fmt.Printf("%d\tmult %d %d(%d) %d(%d)\n", addr, get_raw(addr+1), get_raw(addr+2), b, get_raw(addr+3), c)
		}
		return addr + 4
	},
	11: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, b%c)
		if debug {
			fmt.Printf("%d\tmod %d %d(%d) %d(%d)\n", addr, get_raw(addr+1), get_raw(addr+2), b, get_raw(addr+3), c)
		}
		return addr + 4
	},
	12: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, b&c)
		if debug {
			fmt.Printf("%d\tand %d %d(%d) %d(%d)\n", addr, get_raw(addr+1), get_raw(addr+2), b, get_raw(addr+3), c)
		}
		return addr + 4
	},
	13: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get(addr + 2)
		c := get(addr + 3)
		set(a, b|c)
		if debug {
			fmt.Printf("%d\tor %d %d(%d) %d(%d)\n", addr, get_raw(addr+1), get_raw(addr+2), b, get_raw(addr+3), c)
		}
		return addr + 4
	},
	14: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get(addr + 2)
		b = b << 1
		b = ^b
		b = b >> 1
		set(a, b)
		if debug {
			fmt.Printf("%d\tnot %d %d(%d) %d(%d)\n", addr, get_raw(addr+1), get_raw(addr+2), b)
		}
		return addr + 3
	},
	15: func(addr uint16) uint16 {
		a := get_raw(addr + 1)
		b := get_raw(get(addr + 2))
		set(a, b)
		if debug {
			fmt.Printf("%d\trmem %d %d(%d)\n", addr, get_raw(addr+1), get(addr+2), b)
		}
		return addr + 3
	},
	16: func(addr uint16) uint16 {
		a := get(addr + 1)
		b := get(addr + 2)
		set(a, b)
		if debug {
			fmt.Printf("%d\twmem %d(%d) %d(%d)\n", addr, get_raw(addr+1), a, get_raw(addr+2), b)
		}
		return addr + 3
	},
	17: func(addr uint16) uint16 {
		a := get(addr + 1)
		stack.Push(addr + 2)
		if debug {
			fmt.Printf("%d\tcall %d (%d)\n", addr, get_raw(addr+1), get(addr+1))
		}
		return a
	},
	18: func(addr uint16) uint16 {
		if stack.Len() == 0 {
			if debug {
				fmt.Printf("%d\tret with empty stack, HALTing\n", addr)
			}
			return HALT
		} // halt when stack is empty
		val := stack.Pop().(uint16)
		if debug {
			fmt.Printf("%d\tret (%v)\n", addr, val)
		}
		return val
	},
	19: func(addr uint16) uint16 {
		a := get(addr + 1)
		if debug {
			fmt.Printf("%d\tout %d(%d)\n", addr, get_raw(addr+1), a)
		}
		fmt.Printf("%c", a)
		return addr + 2
	},
	20: func(addr uint16) uint16 {
		if debug {
			fmt.Printf("%d\tin %d\n", addr, get_raw(addr+1))
		}
		a := get_raw(addr + 1)
		buf := make([]byte, 1)
		os.Stdin.Read(buf)
		set(a, uint16(buf[0]))
		return addr + 2
	},
	21: func(addr uint16) uint16 {
		if debug {
			fmt.Printf("%d\tnoop\n", addr)
		}
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
	var interrupted bool
	var exit bool

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		for _ = range sig {
			if interrupted {
				exit = true
			}
			interrupted = true
		}
	}()

	load_bin()
	for addr := uint16(0); addr < ROLL; addr = opcodes[get(addr)](addr) {
		if exit {
			return
		}
		if interrupted || breakpoint == addr {
			interrupted = false
			addr = debugger(addr)
		}
	}
}

func debugger(addr uint16) uint16 {
	var exit bool = false
	for !exit {
		var cmd uint16
		var reg uint16
		fmt.Printf("Debugger> \n")
		cmd = readCode()
		readCode()
		switch cmd {
		case 10:
			exit = true
			break
		case 114:
			fmt.Printf("Register to read> ")
			reg = readInt()
			if reg < NREG {
				fmt.Printf("Register value: %d\n", registers[reg])
			}
			break
		case 119:
			fmt.Printf("Register to write> ")
			reg = readInt()
			if reg < NREG {
				fmt.Printf("Current register value: %d\n", registers[reg])
				fmt.Printf("New value> ")
				registers[reg] = readInt()
			}
			break
		case 97:
			fmt.Printf("Current address: %d\n", addr)
			fmt.Printf("New value> ")
			addr = readInt()
			break
		case 98:
			fmt.Printf("Current breakpoint: %d\n", breakpoint)
			fmt.Printf("New value> ")
			breakpoint = readInt()
			break
		case 100:
			debug = !debug
			break
		}
		fmt.Printf("Registers: %v\n", registers)
		fmt.Printf("Stack: %v\n", stack)
		fmt.Printf("Pointer: %v\n", addr)
		fmt.Printf("Breakpoint: %v\n", breakpoint)
	}

	return addr
}

func readCode() uint16 {
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	return uint16(buf[0])
}

func readInt() uint16 {
	var tmp uint16
	var val uint16
	tmp = readCode()
	for tmp != 10 {
		val = val * 10
		val = val + (tmp - 48)
		tmp = readCode()
	}

	return val
}

func load_bin() {
	flag.Parse()
	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Printf("Read error: %v\n", err)
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
