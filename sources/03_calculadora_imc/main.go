// ============================================================
// CodeNomad Academy  |  Sources Go  |  03. Calculadora de IMC
// ------------------------------------------------------------
// O que faz: recebe peso e altura, valida os dois e devolve o IMC, a
//            classificação da tabela oficial e a faixa de peso considerada
//            saudável para aquela altura, com quanto falta para chegar nela.
// Como usar: go run ./sources/03_calculadora_imc -peso 82 -altura 1.75
//            go run ./sources/03_calculadora_imc -peso 82 -altura 175 -nome "Ana"
//            go run ./sources/03_calculadora_imc --demo
// O que você aprende: structs com métodos, uma tabela de faixas percorrida
//            com for e range, math.Inf para o "sem limite", funções que
//            devolvem dois valores, sort.Slice para ordenar por um campo e
//            text/tabwriter para alinhar colunas sem contar espaços.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

// pessoa guarda os dados de entrada. A altura fica sempre em metros.
type pessoa struct {
	nome   string
	peso   float64 // em quilogramas
	altura float64 // em metros
}

// imc é o peso dividido pela altura ao quadrado.
func (p pessoa) imc() float64 {
	return p.peso / (p.altura * p.altura)
}

// faixa é uma linha da tabela de classificação: vale até o limite (exclusivo).
type faixa struct {
	limite float64
	nome   string
}

// A tabela precisa estar em ordem crescente: a busca para na primeira
// faixa cujo limite ainda não foi alcançado. math.Inf(1) é "infinito",
// então a última faixa serve para qualquer valor.
var tabelaIMC = []faixa{
	{18.5, "Abaixo do peso"},
	{25.0, "Peso normal"},
	{30.0, "Sobrepeso"},
	{35.0, "Obesidade grau I"},
	{40.0, "Obesidade grau II"},
	{math.Inf(1), "Obesidade grau III"},
}

// Limites do IMC considerado saudável, usados para a faixa de peso.
const (
	imcMinimo = 18.5
	imcMaximo = 24.9
)

func classificar(imc float64) string {
	for _, f := range tabelaIMC {
		if imc < f.limite {
			return f.nome
		}
	}
	return tabelaIMC[len(tabelaIMC)-1].nome
}

// pesoSaudavel devolve o menor e o maior peso com IMC saudável para a
// altura. Em Go uma função pode devolver mais de um valor.
func pesoSaudavel(altura float64) (menor, maior float64) {
	return imcMinimo * altura * altura, imcMaximo * altura * altura
}

// ajuste diz quantos quilos faltam para entrar na faixa saudável.
func ajuste(p pessoa) string {
	menor, maior := pesoSaudavel(p.altura)
	switch {
	case p.peso < menor:
		return "ganhar " + dec(menor-p.peso, 1) + " kg"
	case p.peso > maior:
		return "perder " + dec(p.peso-maior, 1) + " kg"
	}
	return "dentro da faixa"
}

// normalizar aceita a altura em metros (1.75) ou em centímetros (175) e
// confere se os valores fazem sentido para um ser humano.
func normalizar(nome string, peso, altura float64) (pessoa, error) {
	if altura > 3 {
		altura /= 100 // veio em centímetros
	}
	switch {
	case peso <= 0:
		return pessoa{}, fmt.Errorf("o peso deve ser maior que zero (recebi %s kg)", dec(peso, 1))
	case peso > 500:
		return pessoa{}, fmt.Errorf("peso de %s kg não parece real; confira o valor", dec(peso, 1))
	case altura <= 0:
		return pessoa{}, fmt.Errorf("a altura deve ser maior que zero (recebi %s m)", dec(altura, 2))
	case altura < 0.5 || altura > 2.8:
		return pessoa{}, fmt.Errorf("altura de %s m não parece real; use metros (1.75) ou centímetros (175)", dec(altura, 2))
	}
	if strings.TrimSpace(nome) == "" {
		nome = "Pessoa"
	}
	return pessoa{nome: nome, peso: peso, altura: altura}, nil
}

// dec formata um número com vírgula e o número de casas pedido.
func dec(v float64, casas int) string {
	return strings.Replace(strconv.FormatFloat(v, 'f', casas, 64), ".", ",", 1)
}

// imprimirTabela usa o tabwriter: as colunas são separadas por \t e o
// pacote calcula a largura de cada uma sozinho.
func imprimirTabela(pessoas []pessoa) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Nome\tPeso\tAltura\tIMC\tClassificação\tPeso saudável\tAjuste")
	for _, p := range pessoas {
		menor, maior := pesoSaudavel(p.altura)
		fmt.Fprintf(w, "%s\t%s kg\t%s m\t%s\t%s\t%s a %s kg\t%s\n",
			p.nome, dec(p.peso, 1), dec(p.altura, 2), dec(p.imc(), 1), classificar(p.imc()),
			dec(menor, 1), dec(maior, 1), ajuste(p))
	}
	if err := w.Flush(); err != nil { // o tabwriter só escreve no Flush
		fmt.Fprintln(os.Stderr, "Erro ao imprimir a tabela:", err)
	}
}

// rodarDemo calcula para três pessoas e mostra a tabela ordenada por IMC.
func rodarDemo() {
	fmt.Println("=== Demonstração: IMC de três pessoas ===")
	fmt.Println()
	entradas := []struct {
		nome         string
		peso, altura float64
	}{
		{"Ana", 54, 1.68},
		{"Bruno", 98, 1.80},
		{"Carla", 47, 1.70},
	}
	var pessoas []pessoa
	for _, e := range entradas {
		p, err := normalizar(e.nome, e.peso, e.altura)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			continue
		}
		pessoas = append(pessoas, p)
	}
	// sort.Slice recebe a função "i vem antes de j?" e cuida do resto.
	sort.Slice(pessoas, func(i, j int) bool { return pessoas[i].imc() < pessoas[j].imc() })
	imprimirTabela(pessoas)

	fmt.Println()
	fmt.Println("Tabela de classificação usada:")
	inicio := 0.0
	for _, f := range tabelaIMC {
		if math.IsInf(f.limite, 1) {
			fmt.Printf("  %s ou mais: %s\n", dec(inicio, 1), f.nome)
			break
		}
		fmt.Printf("  de %s a %s: %s\n", dec(inicio, 1), dec(f.limite, 1), f.nome)
		inicio = f.limite
	}
}

// uso escreve no mesmo destino que flag.PrintDefaults usa (stderr).
func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Calculadora de IMC (CodeNomad Academy)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Uso: go run ./sources/03_calculadora_imc -peso 82 -altura 1.75 [-nome \"Ana\"]")
	fmt.Fprintln(w, "     go run ./sources/03_calculadora_imc --demo")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Opções:")
	flag.PrintDefaults()
}

func main() {
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	peso := flag.Float64("peso", 0, "peso em quilogramas")
	altura := flag.Float64("altura", 0, "altura em metros (1.75) ou centímetros (175)")
	nome := flag.String("nome", "", "nome da pessoa, opcional")
	flag.Usage = uso
	flag.Parse()

	if *demo {
		rodarDemo()
		return
	}
	if flag.NFlag() == 0 {
		uso()
		return
	}
	p, err := normalizar(*nome, *peso, *altura)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
	imprimirTabela([]pessoa{p})
	fmt.Printf("\n%s tem IMC %s: %s. Para a altura de %s m, o peso saudável vai de %s a %s kg (%s).\n",
		p.nome, dec(p.imc(), 1), classificar(p.imc()), dec(p.altura, 2),
		dec(imcMinimo*p.altura*p.altura, 1), dec(imcMaximo*p.altura*p.altura, 1), ajuste(p))
}

// ------------------------------------------------------------
// Para ir além
// 1. Leia várias pessoas de um arquivo CSV (nome;peso;altura) com
//    encoding/csv e imprima a tabela ordenada por nome.
// 2. Acrescente a idade e use a tabela de IMC por faixa etária, que tem
//    limites diferentes para crianças e para pessoas acima de 60 anos.
// 3. Transforme a classificação em um tipo próprio (type classe int) com
//    um método String(), e veja como o fmt passa a usá-lo sozinho.
// ------------------------------------------------------------
