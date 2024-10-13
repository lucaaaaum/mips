package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
    "github.com/inancgumus/screen"
)

func main() {
	linhasDeInstrução, err := obterLinhasDeInstruçãoDoArquivo()
	if err != nil {
		panic(err)
	}
	p := newProcessador(linhasDeInstrução)
	err = p.processar()
	if err != nil {
		panic(err)
	}
}

func obterLinhasDeInstruçãoDoArquivo() ([]string, error) {
	quantidadeDeArgumentos := len(os.Args)

	if quantidadeDeArgumentos < 2 {
		return nil, errors.New("Deve ser fornecido um arquivo de instruções.")
	}

	if quantidadeDeArgumentos > 2 {
		return nil, errors.New("Somente um arquivo pode ser fornecido por vez.")
	}

	conteúdoDoArquivo, err := os.ReadFile(os.Args[1])
	if err != nil {
		return nil, err
	}
	linhasDeInstrução := strings.Split(string(conteúdoDoArquivo), "\n")
	linhasDeInstruçãoFormatadas := make([]string, 0)
	for i := 0; i < len(linhasDeInstrução); i++ {
		linhasDeInstrução[i] = strings.TrimSpace(linhasDeInstrução[i])
		if linhasDeInstrução[i] == "" {
			continue
		}
		linhasDeInstruçãoFormatadas = append(linhasDeInstruçãoFormatadas, linhasDeInstrução[i])
	}
	return linhasDeInstruçãoFormatadas, nil
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
	halt                                     bool
}

func newProcessador(instruções []string) *processador {
	return &processador{
		instruções:       instruções,
		labelsMemória:    make(map[string]int),
		labelsInstruções: make(map[string]int),
	}
}

func (p *processador) processar() error {
	var err error

	p.identificarLabels()

	err = p.processarFills()
	if err != nil {
		return err
	}

	for true {
		// imprimir
        screen.Clear()
		fmt.Printf("clock: %v\n", p.clock)
		fmt.Printf("pc: %v\n", p.pc)
		fmt.Printf("registradores: %v\n", p.registradores)
		fmt.Printf("memória: %v\n", p.memória)
		fmt.Printf("labelsMemória: %v\n", p.labelsMemória)
		fmt.Printf("labelsInstruções: %v\n", p.labelsInstruções)
		for i := 0; i < len(p.instruções); i++ {
			identificadorDaLinha := "       "
			if i == p.posiçõesDasInstruções[0] {
				identificadorDaLinha = "IF  -> "
			} else if i == p.posiçõesDasInstruções[1] {
				identificadorDaLinha = "ID  -> "
			} else if i == p.posiçõesDasInstruções[2] {
				identificadorDaLinha = "EX  -> "
			} else if i == p.posiçõesDasInstruções[3] {
				identificadorDaLinha = "Mem -> "
			} else if i == p.posiçõesDasInstruções[4] {
				identificadorDaLinha = "WB  -> "
			}
			fmt.Println("[" + fmt.Sprint(i) + "]" + identificadorDaLinha + p.instruções[i])
		}

		// fetch
		if !p.halt {
			p.fetch, err = p.obterPróximaInstrução()
			if err != nil {
				return err
			}
			p.posiçõesDasInstruções[0] = p.pc

			// incrementar PC
			p.pc++
		}

		// decode
		p.decode, err = decodificarInstrução(p.linhaDeOrigemDecode)
		if err != nil {
			return err
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
		p.linhaDeOrigemDecode = p.fetch
		p.writeBack = p.memoryAccess
		p.memoryAccess = p.execute
		p.execute = p.decode
		for i := 4; i > 0; i-- {
			p.posiçõesDasInstruções[i] = p.posiçõesDasInstruções[i-1]
		}
		p.clock++

		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')

		if p.writeBack != nil && p.writeBack.tipo == Halt {
			break
		}
	}
	return nil
}

func (p *processador) identificarLabels() {
	for i := 0; i < len(p.instruções); i++ {
		partes := strings.Split(p.instruções[i], " ")
		_, err := obterTipoDeInstrução(partes[0])
		if err != nil {
			tipo, _ := obterTipoDeInstrução(partes[1])
			if tipo != Fill {
				p.labelsInstruções[partes[0]] = i
			}
		}
	}
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
	case ".FILL":
		return Fill, nil
	default:
		return Inválida, errors.New("A string [" + tipo + "] não representa uma instrução conhecida.")
	}
}

func (p *processador) processarFills() error {
	for i := len(p.instruções) - 1; i >= 0; i-- {
		instrução, err := decodificarInstrução(p.instruções[i])
		if err != nil {
			return err
		}
		if instrução.tipo == Fill {
			valorDoFill, err := strconv.Atoi(instrução.parâmetros[0])
			if err != nil {
				return err
			}

			label, err := obterLabelDaLinhaDeInstrução(p.instruções[i])
			if err != nil {
				return err
			}
			if label == "" {
				return nil
			}

			p.memória = append(p.memória, valorDoFill)
			p.labelsMemória[label] = len(p.memória) - 1
		}
	}
	return nil
}

func obterLabelDaLinhaDeInstrução(linhaDeInstrução string) (string, error) {
	partes := strings.Split(linhaDeInstrução, " ")
	if len(partes) < 2 {
		return "", errors.New("O tamanho da instrução [" + linhaDeInstrução + "] é inválido.")
	}

	_, err := obterTipoDeInstrução(linhaDeInstrução)
	if err != nil {
		return partes[0], nil
	}

	return "", nil
}

func decodificarInstrução(linhaDeOrigem string) (*instrução, error) {
	if linhaDeOrigem == "" {
		return nil, nil
	}

	var tipo TipoDeInstrução
	var parâmetros []string

	partes := strings.Split(linhaDeOrigem, " ")
	tipo, err := obterTipoDeInstrução(partes[0])
	if err != nil {
		// Pode ser que o primeiro argumento não seja uma instrução, mas um label
		tipo, err = obterTipoDeInstrução(partes[1])
		parâmetros = partes[2:]
		if err != nil {
			return &instrução{linhaDeOrigem: linhaDeOrigem, tipo: tipo, parâmetros: parâmetros}, err
		}
	} else {
		parâmetros = partes[1:]
	}

	return &instrução{linhaDeOrigem: linhaDeOrigem, tipo: tipo, parâmetros: parâmetros}, nil
}

func (p *processador) obterPróximaInstrução() (string, error) {
	var instruçõesMips []string
	for i := 0; i < len(p.instruções); i++ {
		instrução, err := decodificarInstrução(p.instruções[i])
		if err != nil || instrução.tipo == Fill {
			continue
		}
		instruçõesMips = append(instruçõesMips, p.instruções[i])
	}

	if p.pc < 0 || p.pc > len(instruçõesMips)-1 {
		if p.pc == len(instruçõesMips) {
			return "noop", nil
		}
		return "", errors.New("PC [" + fmt.Sprint(p.pc) + "] aponta para uma instrução inexistente.")
	}
	return p.instruções[p.pc], nil
}

type instrução struct {
	linhaDeOrigem      string
	tipo               TipoDeInstrução
	parâmetros         []string
	valores            []int
	resultadoAlgébrico int
	resultadoBooleano  bool
	resultadoMemória   int
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

func (i *instrução) obterRegistradorDosParâmetros(posição int) (int, error) {
	if posição < 0 || posição >= len(i.parâmetros) {
		return 0, nil
	}
	parâmetro := i.parâmetros[posição]
	parâmetroFormatado := strings.Trim(parâmetro, "R")
	return strconv.Atoi(parâmetroFormatado)
}

func (p *processador) carregarValoresDosRegistradores() error {
	if p.decode == nil {
		return nil
	}

	switch p.decode.tipo {
	case Add, Sub:
		reg1, err := p.decode.obterRegistradorDosParâmetros(1)
		if err != nil {
			return err
		}
		reg2, err := p.decode.obterRegistradorDosParâmetros(2)
		if err != nil {
			return err
		}
		p.decode.valores = append(p.decode.valores, p.registradores[reg1])
		p.decode.valores = append(p.decode.valores, p.registradores[reg2])
		fmt.Printf("p.registradores[reg1]: %v\n", p.registradores[reg1])
		fmt.Printf("p.registradores[reg2]: %v\n", p.registradores[reg2])
		fmt.Printf("p.decode: %v\n", p.decode)
		fmt.Printf("reg1: %v\n", reg1)
		fmt.Printf("reg2: %v\n", reg2)
	case Beq:
		reg0, err := p.decode.obterRegistradorDosParâmetros(0)
		if err != nil {
			return err
		}
		reg1, err := p.decode.obterRegistradorDosParâmetros(1)
		if err != nil {
			return err
		}
		p.decode.valores = append(p.decode.valores, p.registradores[reg0])
		p.decode.valores = append(p.decode.valores, p.registradores[reg1])
	case Lw:
		regOffset, err := p.decode.obterRegistradorDosParâmetros(0)
		if err != nil {
			return err
		}
		p.decode.valores = append(p.decode.valores, regOffset)
		regDestino, err := p.decode.obterRegistradorDosParâmetros(1)
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
		regOffset, err := p.decode.obterRegistradorDosParâmetros(0)
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
	if p.execute == nil {
		return nil
	}

	switch p.execute.tipo {
	case Lw, Sw:
		p.execute.resultadoAlgébrico = p.execute.valores[0] + p.execute.valores[2]
	case Add:
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
			p.decode = &instrução{linhaDeOrigem: "noop", tipo: Noop}
			p.fetch = "noop"
			p.posiçõesDasInstruções[0] = -1
			p.posiçõesDasInstruções[1] = -1
		}
	case Halt:
		p.halt = true
	case Noop:
		return nil
	default:
		return errors.New("Instrução [" + p.execute.linhaDeOrigem + "] é inválida e não pode ser executada.")
	}
	return nil
}

func (p *processador) acessarMemória() error {
	if p.memoryAccess == nil {
		return nil
	}

	fmt.Printf("p.memoryAccess: %v\n", p.memoryAccess.resultadoAlgébrico)
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
	if p.writeBack == nil {
		return nil
	}

	switch p.writeBack.tipo {
	case Add, Sub:
		fmt.Printf("p.writeBack: %v\n", p.writeBack)
		regDestino, err := p.writeBack.obterRegistradorDosParâmetros(2)
		if err != nil {
			return err
		}
		p.registradores[regDestino] = p.writeBack.resultadoAlgébrico
	case Lw:
		regDestino, err := p.writeBack.obterRegistradorDosParâmetros(1)
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
