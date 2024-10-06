package main

import (
	"os"
	"strings"
)

func main() {
	instruções, err := os.ReadFile(os.Args[0])
    if err != nil {
        panic(err)
    }
    processador := NewProcessador(strings.Split(string(instruções), "\n"))
}

type TipoDeInstrução int

const (
	Add = iota
	Sub
	Beq
	Lw
	Sw
	Noop
	Halt
	Inválida
)

type instrução struct {
	linhaDeOrigem string
	tipo          TipoDeInstrução
	parâmetros    []string
	valores       []int
}

type processador struct {
	clock                                    int
	pc                                       int
	registradores                            [32]int
	labelsRegistradores                      map[string]int
	memória                                  []int
	labelsMemória                            map[string]int
	instruções                               []string
	fetch                                    string
	decode, execute, memoryAccess, writeBack instrução
}

func NewProcessador(instruções []string) *processador {
    return &processador{
        instruções: instruções,
    }
}

func (p *processador) Processar() error {
    // fetch
    
    // decode
    // execute
    // memoryAccess
    // writeBack
    // rotacionar instruções
    return nil
}
