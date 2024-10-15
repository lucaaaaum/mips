package predicao

import (
	"errors"
	"fmt"
)

type TabelaDePredição struct {
	endereços, valores [32]int
}

func New() *TabelaDePredição {
	return &TabelaDePredição{}
}

func (t *TabelaDePredição) Incrementar(endereço int) error {
	for i := 0; i < len(t.endereços); i++ {
		if t.endereços[i] == endereço {
			if t.valores[i] < 2 {
				t.valores[i]++
			}
			return nil
		}
	}

	return errors.New("Não há mais espaço na tabela de predição.")
}

func (t *TabelaDePredição) Decrementar(endereço int) error {
	for i := 0; i < len(t.endereços); i++ {
		if t.endereços[i] == endereço {
			if t.valores[i] > 0 {
				t.valores[i]--
			}
			return nil
		}
	}

	return errors.New("Não há mais espaço na tabela de predição.")
}

func (t *TabelaDePredição) TomarDesvio(endereço int) (bool, error) {
	for i := 0; i < len(t.endereços); i++ {
		if t.endereços[i] == endereço {
			if t.valores[i] > 0 {
				return true, nil
			} else {
				return false, nil
			}
		}
	}

	return false, errors.New("O endereço [" + fmt.Sprint(endereço) + "] não foi encontrado na tabela de predição.")
}

func (t *TabelaDePredição) Imprimir() {
	fmt.Printf("tabela de predição: %v\n", t)
}
