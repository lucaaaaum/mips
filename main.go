package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
    quantidadeDeArgumentos := len(os.Args)
    if quantidadeDeArgumentos < 2 {
        panic("Deve ser fornecido um arquivo de instruções.")
    }

    if quantidadeDeArgumentos > 2 {
        panic("Somente um arquivo pode ser fornecido por vez.")
    }

	instruções, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	processador := newProcessador(strings.Split(string(instruções), "\n"))
	err = processador.processar()
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
	linhaDeOrigem      string
	tipo               TipoDeInstrução
	parâmetros         []string
	valores            []int
	resultadoAlgébrico int
	resultadoBooleano  bool
	resultadoMemória   int
}

func decodificarInstrução(linhaDeOrigem string) (string, *instrução, error) {
	var label string
	var tipo TipoDeInstrução
	var parâmetros []string

	partes := strings.Split(linhaDeOrigem, " ")
	tipo, err := obterTipoDeInstrução(partes[0])
	if err != nil {
		// Pode ser que o primeiro argumento não seja uma instrução, mas um label
		label = partes[0]
		tipo, err = obterTipoDeInstrução(partes[1])
		parâmetros = partes[2:]
		if err != nil {
			return label, &instrução{linhaDeOrigem: linhaDeOrigem, tipo: tipo, parâmetros: parâmetros}, err
		}
	} else {
		parâmetros = partes[1:]
	}

	return label, &instrução{linhaDeOrigem: linhaDeOrigem, tipo: tipo, parâmetros: parâmetros}, nil
}

func obterTipoDeInstrução(tipo string) (TipoDeInstrução, error) {
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
	default:
		return Inválida, errors.New("A string [" + tipo + "] não representa uma instrução conhecida.")
	}
}

type processador struct {
	clock                                    int
	pc                                       int
	registradores                            [32]int
	memória                                  []int
	labelsMemória                            map[string]int
	instruções                               []string
	labelsInstruções                         map[string]int
	fetch, linhaDeOrigemDecode               string
	decode, execute, memoryAccess, writeBack *instrução
	posiçõesDasInstruções                    [5]int
}

func newProcessador(instruções []string) *processador {
	return &processador{
		instruções: instruções,
	}
}

func (p *processador) obterPróximaInstrução() (string, error) {
	if p.pc < 0 || p.pc > len(p.instruções)-1 {
		return "", errors.New("PC [" + fmt.Sprint(p.pc) + "] aponta para uma instrução inexistente.")
	}
	return p.instruções[p.pc], nil
}

func (p *processador) carregarValoresDosRegistradores() error {
	switch p.decode.tipo {
	case Add, Sub:
		reg1, err := strconv.Atoi(p.decode.parâmetros[1])
		if err != nil {
			return err
		}
		reg2, err := strconv.Atoi(p.decode.parâmetros[2])
		if err != nil {
			return err
		}
		p.decode.valores = append(p.decode.valores, p.registradores[reg1])
		p.decode.valores = append(p.decode.valores, p.registradores[reg2])
	case Beq:
		reg0, err := strconv.Atoi(p.decode.parâmetros[0])
		if err != nil {
			return err
		}
		reg1, err := strconv.Atoi(p.decode.parâmetros[1])
		if err != nil {
			return err
		}
		p.decode.valores = append(p.decode.valores, p.registradores[reg0])
		p.decode.valores = append(p.decode.valores, p.registradores[reg1])
	case Lw:
		regOffset, err := strconv.Atoi(p.decode.parâmetros[0])
		if err != nil {
			return err
		}
		p.decode.valores = append(p.decode.valores, regOffset)
		regDestino, err := strconv.Atoi(p.decode.parâmetros[1])
		if err != nil {
			return err
		}
		p.decode.valores = append(p.decode.valores, regDestino)
		parâmetroDePosiçãoDeMemória := p.decode.parâmetros[2]
		if éNúmero(parâmetroDePosiçãoDeMemória) {
			posiçãoMemória, err := strconv.Atoi(parâmetroDePosiçãoDeMemória)
			if err != nil {
				return err
			}
			p.decode.valores = append(p.decode.valores, posiçãoMemória)
		} else {
			posiçãoMemória := p.labelsMemória[parâmetroDePosiçãoDeMemória]
			p.decode.valores = append(p.decode.valores, posiçãoMemória)
		}
	case Sw:
		regOffset, err := strconv.Atoi(p.decode.parâmetros[0])
		if err != nil {
			return err
		}
		p.decode.valores = append(p.decode.valores, regOffset)
		parâmetroDePosiçãoDeMemória := p.decode.parâmetros[2]
		if éNúmero(parâmetroDePosiçãoDeMemória) {
			posiçãoMemória, err := strconv.Atoi(parâmetroDePosiçãoDeMemória)
			if err != nil {
				return err
			}
			p.decode.valores = append(p.decode.valores, posiçãoMemória)
		} else {
			posiçãoMemória := p.labelsMemória[parâmetroDePosiçãoDeMemória]
			p.decode.valores = append(p.decode.valores, posiçãoMemória)
		}
	case Noop, Halt:
		return nil
	default:
		return errors.New("Instrução [" + p.decode.linhaDeOrigem + "] é inválida e não pode ser decodificada.")
	}
	return nil
}

func éNúmero(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func (p *processador) executarInstrução() error {
	switch p.execute.tipo {
	case Add, Lw, Sw:
		p.execute.resultadoAlgébrico = p.execute.valores[0] + p.execute.valores[1]
	case Sub:
		p.execute.resultadoAlgébrico = p.execute.valores[0] - p.execute.valores[1]
	case Beq:
		p.execute.resultadoBooleano = p.execute.valores[0] == p.execute.valores[1]
        if p.execute.resultadoBooleano {
            parâmetroDePosiçãoDeInstrução := p.execute.parâmetros[2]
            if éNúmero(parâmetroDePosiçãoDeInstrução) {
                novoPc, err := strconv.Atoi(parâmetroDePosiçãoDeInstrução)
                if err != nil {
                    return err
                }
                p.pc = novoPc
            } else {
                novoPc := p.labelsInstruções[parâmetroDePosiçãoDeInstrução]
                p.pc = novoPc
            }
        }
	case Noop, Halt:
		return nil
	default:
		return errors.New("Instrução [" + p.execute.linhaDeOrigem + "] é inválida e não pode ser executada.")
	}
	return nil
}

func (p *processador) acessarMemória() error {
	switch p.memoryAccess.tipo {
	case Lw:
		p.memoryAccess.resultadoMemória = p.memória[p.memoryAccess.resultadoAlgébrico]
	case Sw:
		p.memória[p.memoryAccess.resultadoAlgébrico] = p.memoryAccess.valores[2]
	case Add, Sub, Beq, Noop, Halt:
		return nil
	default:
		return errors.New("Instrução [" + p.memoryAccess.linhaDeOrigem + "] é inválida e não pode acessar a memória.")
	}
	return nil
}

func (p *processador) escreverRegistradores() error {
	switch p.writeBack.tipo {
	case Add, Sub:
		regDestino, err := strconv.Atoi(p.writeBack.parâmetros[2])
		if err != nil {
			return err
		}
		p.registradores[regDestino] = p.writeBack.resultadoAlgébrico
	case Lw:
		regDestino, err := strconv.Atoi(p.writeBack.parâmetros[1])
		if err != nil {
			return err
		}
		p.registradores[regDestino] = p.writeBack.resultadoMemória
	case Beq, Noop, Halt:
		return nil
	default:
		return errors.New("Instrução [" + p.writeBack.linhaDeOrigem + "] é inválida e não pode acessar os registradores para write back.")
	}
	return nil
}

func (p *processador) processar() error {
	var err error

	// fetch
	p.fetch, err = p.obterPróximaInstrução()
	if err != nil {
		return err
	}
	p.posiçõesDasInstruções[0] = p.pc

    // incrementar PC
    p.pc++

	// decode
	var label string
	label, p.decode, err = decodificarInstrução(p.linhaDeOrigemDecode)
	if err != nil {
		return err
	}
	if label != "" {
		p.labelsInstruções[label] = p.posiçõesDasInstruções[1]
	}
	err = p.carregarValoresDosRegistradores()
	if err != nil {
		return err
	}

	// execute
    // caso seja BEQ, pc será manualmente alterado
	err = p.executarInstrução()
	if err != nil {
		return err
	}

	// memoryAccess
	err = p.acessarMemória()
	if err != nil {
		return err
	}

	// writeBack
    err = p.escreverRegistradores()
    if err != nil {
        return err
    }

	// rotacionar instruções
	p.fetch = ""
	p.linhaDeOrigemDecode = p.fetch
	p.writeBack = p.memoryAccess
	p.memoryAccess = p.execute
	p.execute = nil
	for i := 4; i > 0; i-- {
		p.posiçõesDasInstruções[i] = p.posiçõesDasInstruções[i-1]
	}
	return nil
}
