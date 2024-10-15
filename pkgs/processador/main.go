package processador

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/inancgumus/screen"
	inst "github.com/lucaaaaum/mips/pkgs/instrucao"
)

type Processador struct {
	clock                                    int
	pc                                       int
	registradores                            [32]int
	memória                                  []int
	labelsMemória                            map[string]int
	instruções                               []string
	labelsInstruções                         map[string]int
	fetch, linhaDeOrigemDecode               string
	decode, execute, memoryAccess, writeBack *inst.Instrução
	posiçõesDasInstruções                    [5]int
	halt                                     bool
}

func New(instruções []string) *Processador {
	return &Processador{
		instruções:       instruções,
		labelsMemória:    make(map[string]int),
		labelsInstruções: make(map[string]int),
	}
}

func (p *Processador) armazenarNaMemóriaComLabel(label string, valor int) error {
	posição := p.labelsInstruções[label]
	return p.armazenarNaMemóriaComPosição(posição, valor)
}

func (p *Processador) armazenarNaMemóriaComPosição(posição int, valor int) error {
	if posição < 0 || posição >= len(p.memória) {
		return errors.New("Não há memória para a posição [" + fmt.Sprint(posição) + "]")
	}
	p.memória[posição] = valor
	return nil
}

func (p *Processador) Processar() error {
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
		fmt.Printf("posiçõesDasInstruções: %v\n", p.posiçõesDasInstruções)
		for i := 0; i < len(p.instruções); i++ {
			identificadorDaLinha := "       "
			switch i {
			case p.posiçõesDasInstruções[0]:
				identificadorDaLinha = "IF  -> "
			case p.posiçõesDasInstruções[1]:
				identificadorDaLinha = "ID  -> "
			case p.posiçõesDasInstruções[2]:
				identificadorDaLinha = "EX  -> "
			case p.posiçõesDasInstruções[3]:
				identificadorDaLinha = "Mem -> "
			case p.posiçõesDasInstruções[4]:
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
		p.decode, err = inst.Decodificar(p.linhaDeOrigemDecode)
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
		p.posiçõesDasInstruções[0] = p.pc
		p.clock++

		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')

		if p.writeBack != nil && p.writeBack.Tipo == inst.Halt {
			break
		}
	}
	return nil
}

func (p *Processador) identificarLabels() {
	for i := 0; i < len(p.instruções); i++ {
		partes := strings.Split(p.instruções[i], " ")
		_, err := inst.ObterTipoDeInstrução(partes[0])
		if err != nil {
			tipo, _ := inst.ObterTipoDeInstrução(partes[1])
			if tipo != inst.Fill {
				p.labelsInstruções[partes[0]] = i
			}
		}
	}
}

func (p *Processador) processarFills() error {
	for i := len(p.instruções) - 1; i >= 0; i-- {
		instrução, err := inst.Decodificar(p.instruções[i])
		if err != nil {
			return err
		}
		if instrução.Tipo == inst.Fill {
			valorDoFill, err := instrução.ObterValorDoParâmetro(0)
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
		return "", errors.New("O tamanho da instrução [" + linhaDeInstrução + "] é inválido para obtenção de label.")
	}

	_, err := inst.ObterTipoDeInstrução(linhaDeInstrução)
	if err != nil {
		return partes[0], nil
	}

	return "", nil
}

func (p *Processador) obterPróximaInstrução() (string, error) {
	var instruçõesMips []string
	for i := 0; i < len(p.instruções); i++ {
		instrução, err := inst.Decodificar(p.instruções[i])
		if err != nil || instrução.Tipo == inst.Fill {
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

func (p *Processador) carregarValoresDosRegistradores() error {
	if p.decode == nil {
		return nil
	}

	switch p.decode.Tipo {
	case inst.Add, inst.Sub:
		reg1, err := p.decode.ObterRegistradorDosParâmetros(1)
		if err != nil {
			return err
		}
		reg2, err := p.decode.ObterRegistradorDosParâmetros(2)
		if err != nil {
			return err
		}
		p.decode.AdicionarValor(p.registradores[reg1])
		p.decode.AdicionarValor(p.registradores[reg2])
	case inst.Beq:
		reg0, err := p.decode.ObterRegistradorDosParâmetros(0)
		if err != nil {
			return err
		}
		reg1, err := p.decode.ObterRegistradorDosParâmetros(1)
		if err != nil {
			return err
		}
		p.decode.AdicionarValor(p.registradores[reg0])
		p.decode.AdicionarValor(p.registradores[reg1])
	case inst.Lw:
		regOffset, err := p.decode.ObterRegistradorDosParâmetros(0)
		if err != nil {
			return err
		}
		p.decode.AdicionarValor(regOffset)
		regDestino, err := p.decode.ObterRegistradorDosParâmetros(1)
		if err != nil {
			return err
		}
		p.decode.AdicionarValor(regDestino)

		parâmetroDePosiçãoDeMemória := p.decode.ObterParâmetro(2)
		if éNúmero(parâmetroDePosiçãoDeMemória) {
			posiçãoMemória, _ := strconv.Atoi(parâmetroDePosiçãoDeMemória)
			p.decode.AdicionarValor(posiçãoMemória)
		} else {
			posiçãoMemória := p.labelsMemória[parâmetroDePosiçãoDeMemória]
			p.decode.AdicionarValor(posiçãoMemória)
		}
	case inst.Sw:
		regOffset, err := p.decode.ObterRegistradorDosParâmetros(0)
		if err != nil {
			return err
		}
		p.decode.AdicionarValor(regOffset)

		parâmetroDePosiçãoDeMemória := p.decode.ObterParâmetro(2)
		if éNúmero(parâmetroDePosiçãoDeMemória) {
			posiçãoMemória, _ := strconv.Atoi(parâmetroDePosiçãoDeMemória)
			p.decode.AdicionarValor(posiçãoMemória)
		} else {
			posiçãoMemória := p.labelsMemória[parâmetroDePosiçãoDeMemória]
			p.decode.AdicionarValor(posiçãoMemória)
		}
	case inst.Noop, inst.Halt:
		return nil
	default:
		return errors.New("Instrução [" + p.decode.LinhaDeOrigem + "] é inválida e não pode ser decodificada.")
	}
	return nil
}

func éNúmero(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func (p *Processador) executarInstrução() error {
	if p.execute == nil {
		return nil
	}

	switch p.execute.Tipo {
	case inst.Lw, inst.Sw:
		valor0, err := p.execute.ObterValor(0)
		if err != nil {
			return err
		}
		valor2, err := p.execute.ObterValor(2)
		if err != nil {
			return err
		}
		p.execute.ResultadoAlgébrico = valor0 + valor2
	case inst.Add:
		valor0, err := p.execute.ObterValor(0)
		fmt.Printf("valor0: %v\n", valor0)
		if err != nil {
			return err
		}
		valor1, err := p.execute.ObterValor(1)
		fmt.Printf("valor1: %v\n", valor1)
		if err != nil {
			return err
		}
		p.execute.ResultadoAlgébrico = valor0 + valor1
	case inst.Sub:
		valor0, err := p.execute.ObterValor(0)
		if err != nil {
			return err
		}
		valor1, err := p.execute.ObterValor(1)
		if err != nil {
			return err
		}
		p.execute.ResultadoAlgébrico = valor0 + valor1
	case inst.Beq:
		valor0, err := p.execute.ObterValor(0)
		if err != nil {
			return err
		}
		valor1, err := p.execute.ObterValor(1)
		if err != nil {
			return err
		}
		p.execute.ResultadoBooleano = valor0 == valor1
		if p.execute.ResultadoBooleano {
			parâmetroDePosiçãoDeInstrução := p.execute.ObterParâmetro(2)
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
			p.decode = &inst.Instrução{LinhaDeOrigem: "noop", Tipo: inst.Noop}
			p.fetch = "noop"
			p.posiçõesDasInstruções[0] = -1
			p.posiçõesDasInstruções[1] = -1
		}
	case inst.Halt:
		p.halt = true
	case inst.Noop:
		return nil
	default:
		return errors.New("Instrução [" + p.execute.LinhaDeOrigem + "] é inválida e não pode ser executada.")
	}
	return nil
}

func (p *Processador) acessarMemória() error {
	if p.memoryAccess == nil {
		return nil
	}

	switch p.memoryAccess.Tipo {
	case inst.Lw:
		p.memoryAccess.ResultadoMemória = p.memória[p.memoryAccess.ResultadoAlgébrico]
	case inst.Sw:
		valor, err := p.memoryAccess.ObterValor(2)
		if err != nil {
			return err
		}
		p.armazenarNaMemóriaComPosição(p.memoryAccess.ResultadoAlgébrico, valor)
	case inst.Add, inst.Sub, inst.Beq, inst.Noop, inst.Halt:
		return nil
	default:
		return errors.New("Instrução [" + p.memoryAccess.LinhaDeOrigem + "] é inválida e não pode acessar a memória.")
	}
	return nil
}

func (p *Processador) escreverRegistradores() error {
	if p.writeBack == nil {
		return nil
	}

	switch p.writeBack.Tipo {
	case inst.Add, inst.Sub:
		regDestino, err := p.writeBack.ObterRegistradorDosParâmetros(2)
		if err != nil {
			return err
		}
		p.registradores[regDestino] = p.writeBack.ResultadoAlgébrico
	case inst.Lw:
		regDestino, err := p.writeBack.ObterRegistradorDosParâmetros(1)
		if err != nil {
			return err
		}
		p.registradores[regDestino] = p.writeBack.ResultadoMemória
	case inst.Beq, inst.Noop, inst.Halt:
		return nil
	default:
		return errors.New("Instrução [" + p.writeBack.LinhaDeOrigem + "] é inválida e não pode acessar os registradores para write back.")
	}
	return nil
}
