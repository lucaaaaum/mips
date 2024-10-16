# Simulador MIPS

## Acesso

O projeto está disponibilizado no github neste [link](https://github.com/lucaaaaum/mips).

## Como compilar

Para compilar o programa, é necessário ter o SDK de Go instalado. Em seguida, execute este comando no terminal:

```bash
go build -o mips
```

## Escrevendo programas

Para utilizar o programa, escreva um arquivo de texto contendo as instruções em assembly. Abaixo está um exemplo:

```assembly
lw R0 R1 neg1
lw R0 R2 ten
lw R0 R3 one
noop
noop
noop
loop add 2 1 2
noop
noop
beq 2 0 done
noop
noop
noop
beq 0 0 loop
noop
noop
noop
done halt
neg1 .fill -1
ten .fill 10
one .fill 1
```

## Como usar o simulador

Para executar o simulador, passe os seguintes argumentos:

```bash
./mips $1 $2
```

* $1: caminho para o arquivo assembly
* $2: booleano (0 ou 1) para habilitar/desabilitar a tabela de predição
