package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	proc "github.com/lucaaaaum/mips/pkgs/processador"
)

func main() {
	quantidadeDeArgumentos := len(os.Args)

	if quantidadeDeArgumentos < 2 {
		panic(errors.New("Deve ser fornecido um arquivo de instruções."))
	}

	if quantidadeDeArgumentos > 3 {
		panic(errors.New("Quantidade de argumentos inesperada."))
	}

	linhasDeInstrução, err := obterLinhasDeInstruçãoDoArquivo()
	if err != nil {
		panic(err)
	}

	utilizarPredição, err := utilizarPredição()
	if err != nil {
		panic(err)
	}

	p := proc.New(linhasDeInstrução, utilizarPredição)
	err = p.Processar()
	if err != nil {
		panic(err)
	}
}

func obterLinhasDeInstruçãoDoArquivo() ([]string, error) {
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

func utilizarPredição() (bool, error) {
	if len(os.Args) < 3 {
		return false, nil
	}

	parâmetroUtilizarPredição := os.Args[2]

	númeroUtilizarPredição, err := strconv.Atoi(parâmetroUtilizarPredição)
	if err != nil {
		return false, err
	}

	if númeroUtilizarPredição != 1 && númeroUtilizarPredição != 0 {
		return false, errors.New("O valor [" + fmt.Sprint(númeroUtilizarPredição) + "] é inválido. Por favor, passar 0 para false e 1 para true.")
	}

	return númeroUtilizarPredição == 1, nil
}
