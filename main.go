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

type Instrução struct {
	linhaDeOrigem string
	tipo          TipoDeInstrução
	parâmetros    []string
	valores       []int
}
