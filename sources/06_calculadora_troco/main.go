// ============================================================
// CodeNomad Academy  |  Sources Go  |  06. Calculadora de troco
// ------------------------------------------------------------
// O que faz: recebe o valor da compra e o valor pago, calcula o troco e o
//            decompõe no menor número possível de cédulas e moedas do real
//            (algoritmo guloso). Avisa quando o pagamento não cobre a compra.
// Como usar: go run ./sources/06_calculadora_troco -compra 37,30 -pago 50
//            go run ./sources/06_calculadora_troco -compra 12.75 -pago 20
//            go run ./sources/06_calculadora_troco --demo
// O que você aprende: por que dinheiro se guarda em centavos inteiros e
//            não em float, divisão e resto inteiros, um algoritmo guloso,
//            slices de structs, strconv.ParseFloat, erros personalizados
//            com fmt.Errorf e uma função pequena para o plural.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// denominacao é uma cédula ou moeda, com o valor em centavos.
type denominacao struct {
	centavos int
	tipo     string // "cédula" ou "moeda"
}

// As denominações precisam estar em ordem decrescente: o algoritmo guloso
// sempre tenta a maior primeiro. Para o real isso dá o menor número de
// itens; em sistemas monetários exóticos o guloso pode falhar.
var denominacoes = []denominacao{
	{20000, "cédula"}, {10000, "cédula"}, {5000, "cédula"}, {2000, "cédula"},
	{1000, "cédula"}, {500, "cédula"}, {200, "cédula"},
	{100, "moeda"}, {50, "moeda"}, {25, "moeda"}, {10, "moeda"}, {5, "moeda"}, {1, "moeda"},
}

// item é uma linha do resultado: quantas unidades de uma denominação.
type item struct {
	denominacao
	quantidade int
}

// paraCentavos lê "37,30" ou "37.30" e devolve 3730. Em float, 0,1 + 0,2
// não dá exatamente 0,3; em centavos inteiros a conta é sempre exata.
func paraCentavos(texto string) (int, error) {
	limpo := strings.ReplaceAll(strings.TrimSpace(texto), ",", ".")
	if limpo == "" {
		return 0, errors.New("valor em branco")
	}
	v, err := strconv.ParseFloat(limpo, 64)
	if err != nil {
		return 0, fmt.Errorf("%q não é um valor em reais válido", texto)
	}
	if v < 0 {
		return 0, fmt.Errorf("valor negativo não faz sentido aqui: %s", texto)
	}
	return int(math.Round(v * 100)), nil
}

// decompor aplica o guloso: para cada denominação, quantas cabem no que
// sobrou (divisão inteira) e o que resta (operador %).
func decompor(troco int) []item {
	var itens []item
	resto := troco
	for _, d := range denominacoes {
		if quantidade := resto / d.centavos; quantidade > 0 {
			itens = append(itens, item{d, quantidade})
			resto %= d.centavos
		}
	}
	return itens
}

// calcular valida o pagamento e devolve o troco em centavos.
func calcular(compra, pago int) (int, error) {
	if compra == 0 {
		return 0, errors.New("o valor da compra deve ser maior que zero")
	}
	if pago < compra {
		return 0, fmt.Errorf("pagamento insuficiente: faltam %s", reais(compra-pago))
	}
	return pago - compra, nil
}

// reais formata centavos como R$ 1.234,56.
func reais(centavos int) string {
	inteiro := strconv.Itoa(centavos / 100)
	var b strings.Builder
	for i, d := range inteiro {
		if i > 0 && (len(inteiro)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(d)
	}
	return fmt.Sprintf("R$ %s,%02d", b.String(), centavos%100)
}

// plural escolhe a forma certa: "1 moeda", "2 moedas".
func plural(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// relatar imprime a conta e a decomposição, ou o aviso de insuficiência.
func relatar(textoCompra, textoPago string) error {
	compra, err := paraCentavos(textoCompra)
	if err != nil {
		return fmt.Errorf("compra: %w", err)
	}
	pago, err := paraCentavos(textoPago)
	if err != nil {
		return fmt.Errorf("pago: %w", err)
	}
	troco, err := calcular(compra, pago)
	if err != nil {
		return err
	}
	fmt.Printf("Compra: %s | Pago: %s | Troco: %s\n", reais(compra), reais(pago), reais(troco))
	if troco == 0 {
		fmt.Println("  Pagamento exato, nada a devolver.")
		return nil
	}
	cedulas, moedas := 0, 0
	for _, it := range decompor(troco) {
		fmt.Printf("  %2d %s de %s\n", it.quantidade, plural(it.quantidade, it.tipo, it.tipo+"s"), reais(it.centavos))
		if it.tipo == "moeda" {
			moedas += it.quantidade
		} else {
			cedulas += it.quantidade
		}
	}
	fmt.Printf("  Total: %d %s (%d %s e %d %s)\n", cedulas+moedas, plural(cedulas+moedas, "item", "itens"),
		cedulas, plural(cedulas, "cédula", "cédulas"), moedas, plural(moedas, "moeda", "moedas"))
	return nil
}

// rodarDemo resolve quatro compras e mostra um pagamento insuficiente.
func rodarDemo() {
	fmt.Println("=== Demonstração: troco com o menor número de itens ===")
	casos := [][2]string{{"37,30", "50"}, {"12.75", "20"}, {"199,99", "500"}, {"64,00", "64"}}
	for i, c := range casos {
		fmt.Printf("\nCaso %d\n", i+1)
		if err := relatar(c[0], c[1]); err != nil {
			fmt.Println("  Aviso:", err)
		}
	}
	fmt.Println("\nCaso extra: pagamento que não cobre a compra")
	if err := relatar("80", "50"); err != nil {
		fmt.Println("  Aviso:", err)
	}
}

// uso escreve no mesmo destino que flag.PrintDefaults usa (stderr).
func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Calculadora de troco (CodeNomad Academy)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Uso: go run ./sources/06_calculadora_troco -compra 37,30 -pago 50")
	fmt.Fprintln(w, "     go run ./sources/06_calculadora_troco --demo")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Opções:")
	flag.PrintDefaults()
}

func main() {
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	compra := flag.String("compra", "", "valor da compra em reais (aceita vírgula ou ponto)")
	pago := flag.String("pago", "", "valor entregue pelo cliente")
	flag.Usage = uso
	flag.Parse()

	switch {
	case *demo:
		rodarDemo()
	case flag.NFlag() == 0:
		uso()
	case *compra == "" || *pago == "":
		fmt.Fprintln(os.Stderr, "Erro: informe -compra e -pago (ex.: -compra 37,30 -pago 50)")
		os.Exit(1)
	default:
		if err := relatar(*compra, *pago); err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			os.Exit(1)
		}
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Aceite o pagamento em várias cédulas (-pago 50,20,5) somando tudo
//    antes de calcular, como um caixa de verdade recebe.
// 2. Receba a quantidade disponível de cada denominação no caixa e avise
//    quando não houver como montar o troco com o que existe.
// 3. Troque a tabela por uma moeda fictícia com valores 1, 3 e 4 e mostre
//    um caso em que o guloso não dá o menor número de itens (troco 6).
// ------------------------------------------------------------
