package predicao

import (
	"errors"
	"fmt"
)

type item struct {
	endereço, valor int
	preenchido      bool
}

type TabelaDePredição struct {
	predições [32]item
}

func New() *TabelaDePredição {
	return &TabelaDePredição{}
}

func (t *TabelaDePredição) Incrementar(endereço int) error {
	for i := 0; i < len(t.predições); i++ {
		p := t.predições[i]

		if p.endereço == endereço {
			if p.valor < 2 {
				p.valor++
			}
			return nil
		}
	}

	for i := 0; i < len(t.predições); i++ {
		p := t.predições[i]

		if !p.preenchido {
			p.endereço = endereço
			p.valor = 1
			p.preenchido = true
			return nil
		}
	}

	return errors.New("Não há mais espaço na tabela de predição.")
}

func (t *TabelaDePredição) Decrementar(endereço int) error {
	for i := 0; i < len(t.predições); i++ {
		p := t.predições[i]

		if p.endereço == endereço {
			if p.valor < 2 {
				p.valor--
			}
			return nil
		}
	}

	for i := 0; i < len(t.predições); i++ {
		p := t.predições[i]

		if !p.preenchido {
			p.endereço = endereço
			p.valor = 0
			p.preenchido = true
			return nil
		}
	}

	return errors.New("Não há mais espaço na tabela de predição.")
}

func (t *TabelaDePredição) TomarDesvio(endereço int) (bool, error) {
	for i := 0; i < len(t.predições); i++ {
		if t.predições[i].endereço == endereço {
			if t.predições[i].valor > 0 {
				return true, nil
			} else {
				return false, nil
			}
		}
	}

	return false, errors.New("O endereço [" + fmt.Sprint(endereço) + "] não foi encontrado na tabela de predição.")
}

func (t *TabelaDePredição) Imprimir() {
	fmt.Printf("tabela de predição:\n")
	for i := 0; i < len(t.predições); i++ {
		p := t.predições[i]
		if p.preenchido {
			fmt.Printf("[%v] predição: %v\n", i, p)
		}
	}
}
