package instrucao

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Instrução struct {
	LinhaDeOrigem      string
	Tipo               TipoDeInstrução
	parâmetros         []string
	valores            []int
	ResultadoAlgébrico int
	ResultadoBooleano  bool
	ResultadoMemória   int
}

func New(linhaDeOrigem string) (*Instrução, error) {
	return Decodificar(linhaDeOrigem)
}

func NewNoop() *Instrução {
	instrução, _ := New("noop")
	return instrução
}

func Decodificar(linhaDeOrigem string) (*Instrução, error) {
	if linhaDeOrigem == "" {
		return nil, nil
	}

	var tipo TipoDeInstrução
	var parâmetros []string

	partes := strings.Split(linhaDeOrigem, " ")
	tipo, err := ObterTipoDeInstrução(partes[0])
	if err != nil {
		// Pode ser que o primeiro argumento não seja uma instrução, mas um label
		tipo, err = ObterTipoDeInstrução(partes[1])
		parâmetros = partes[2:]
		if err != nil {
			return &Instrução{LinhaDeOrigem: linhaDeOrigem, Tipo: tipo, parâmetros: parâmetros}, err
		}
	} else {
		parâmetros = partes[1:]
	}

	return &Instrução{LinhaDeOrigem: linhaDeOrigem, Tipo: tipo, parâmetros: parâmetros}, nil
}

func (i *Instrução) ObterRegistradorDosParâmetros(posição int) (int, error) {
	parâmetro, err := i.obterParâmetro(posição)
	if err != nil {
		return 0, err
	}
	parâmetroFormatado := strings.Trim(parâmetro, "R")
	return strconv.Atoi(parâmetroFormatado)
}

func (i *Instrução) obterParâmetro(posição int) (string, error) {
	if posição < 0 || posição >= len(i.parâmetros) {
		return "", errors.New("Não existe parâmetro na posição [" + fmt.Sprint(posição) + "]")
	}
	return i.parâmetros[posição], nil
}

func (i *Instrução) ObterValorDoParâmetro(posição int) (int, error) {
	parâmetro, err := i.obterParâmetro(posição)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(parâmetro)
}

func (i *Instrução) ObterParâmetro(posição int) string {
	return i.parâmetros[posição]
}

func (i *Instrução) AdicionarValor(valor int) int {
	i.valores = append(i.valores, valor)
	return len(i.valores) - 1
}

func (i *Instrução) ObterValor(posição int) (int, error) {
	if posição < 0 || posição > len(i.valores) {
		return 0, errors.New("Não existe valor na posição [" + fmt.Sprint(posição) + "]")
	}
	return i.valores[posição], nil
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
	Fill
	Inválida
)

func ObterTipoDeInstrução(tipo string) (TipoDeInstrução, error) {
	switch strings.ToUpper(tipo) {
	case "ADD":
		return Add, nil
	case "SUB":
		return Sub, nil
	case "BEQ":
		return Beq, nil
	case "LW":
		return Lw, nil
	case "SW":
		return Sw, nil
	case "NOOP":
		return Noop, nil
	case "HALT":
		return Halt, nil
	case ".FILL":
		return Fill, nil
	default:
		return Inválida, errors.New("A string [" + tipo + "] não representa uma instrução conhecida.")
	}
}
