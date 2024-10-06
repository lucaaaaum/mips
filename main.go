package main

func main() {

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


