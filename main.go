package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func main() {
	instruções, err := os.ReadFile(os.Args[0])
    if err != nil {
        panic(err)
    }
    processador := newProcessador(strings.Split(string(instruções), "\n"))
    err = processador.Processar()
    if err != nil {
        panic(err)
    }
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

type name struct {
    
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

func newProcessador(instruções []string) *processador {
    return &processador{
        instruções: instruções,
    }
}

func (p *processador) obterPróximaInstrução() (string, error) {
    if p.pc < 0 || p.pc > len(p.instruções) - 1 {
        return "", errors.New("PC [" + fmt.Sprint(p.pc) + "] aponta para uma instrução inexistente.")
    }
    return p.instruções[p.pc], nil
}

func (p *processador) Processar() error {
    var err error
    // fetch
    p.fetch, err = p.obterPróximaInstrução()
    if err != nil {
        return err
    }
    // decode

    // execute
    // memoryAccess
    // writeBack
    // rotacionar instruções
    return nil
}
