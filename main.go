package main

import (
	"errors"
	"os"
	"strings"

	proc "github.com/lucaaaaum/mips/pkgs/processador"
)

func main() {
	linhasDeInstrução, err := obterLinhasDeInstruçãoDoArquivo()
	if err != nil {
		panic(err)
	}
	p := proc.New(linhasDeInstrução)
	err = p.Processar()
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
