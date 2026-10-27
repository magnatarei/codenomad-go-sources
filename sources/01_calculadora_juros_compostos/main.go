// ============================================================
// CodeNomad Academy  |  Sources Go  |  01. Calculadora de juros compostos
// ------------------------------------------------------------
// O que faz: recebe um capital inicial, uma taxa mensal, um prazo em meses
//            e um aporte mensal opcional. Mostra o montante final, os juros
//            ganhos e uma tabela mês a mês com saldo inicial, juros, aporte
//            e saldo final.
// Como usar: go run ./sources/01_calculadora_juros_compostos -capital 1000 -taxa 0.8 -meses 24 -aporte 200
//            go run ./sources/01_calculadora_juros_compostos --demo
// O que você aprende: o pacote flag para ler argumentos, structs e slices
//            para guardar a tabela, métodos em structs, math.Pow, formatação
//            de dinheiro no padrão brasileiro com strings.Builder e o retorno
//            de error para validar a entrada sem derrubar o programa.
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

// linha guarda o que aconteceu em um único mês da simulação.
type linha struct {
	mes          int
	saldoInicial float64
	juros        float64
	aporte       float64
	saldoFinal   float64
}

// simulacao reúne os parâmetros de entrada e a tabela calculada.
type simulacao struct {
	capital float64
	taxa    float64 // taxa mensal em fração: 0,8% vira 0.008
	meses   int
	aporte  float64
	tabela  []linha
}

// validar confere os parâmetros antes de qualquer cálculo. Devolver um
// error, em vez de encerrar o programa aqui, deixa a função reutilizável.
func validar(capital, taxaPct float64, meses int, aporte float64) error {
	switch {
	case capital < 0:
		return errors.New("o capital inicial não pode ser negativo")
	case taxaPct < 0:
		return errors.New("a taxa mensal não pode ser negativa")
	case meses < 1 || meses > 1200:
		return fmt.Errorf("o prazo deve ficar entre 1 e 1200 meses (recebi %d)", meses)
	case aporte < 0:
		return errors.New("o aporte mensal não pode ser negativo")
	case capital == 0 && aporte == 0:
		return errors.New("informe um capital inicial ou um aporte mensal maior que zero")
	}
	return nil
}

// simular monta a tabela mês a mês. Os juros de cada mês incidem sobre o
// saldo do início do mês; o aporte entra no fim, depois dos juros.
func simular(capital, taxaPct float64, meses int, aporte float64) simulacao {
	s := simulacao{capital: capital, taxa: taxaPct / 100, meses: meses, aporte: aporte}
	saldo := capital
	for m := 1; m <= meses; m++ {
		juros := saldo * s.taxa
		l := linha{mes: m, saldoInicial: saldo, juros: juros, aporte: aporte}
		saldo += juros + aporte
		l.saldoFinal = saldo
		s.tabela = append(s.tabela, l)
	}
	return s
}

// montante é o saldo depois do último mês.
func (s simulacao) montante() float64 {
	if len(s.tabela) == 0 {
		return s.capital
	}
	return s.tabela[len(s.tabela)-1].saldoFinal
}

// totalInvestido soma o capital inicial com todos os aportes.
func (s simulacao) totalInvestido() float64 {
	return s.capital + s.aporte*float64(s.meses)
}

// jurosGanhos é o que o dinheiro rendeu além do que foi depositado.
func (s simulacao) jurosGanhos() float64 {
	return s.montante() - s.totalInvestido()
}

// montanteFormula confere a tabela pela fórmula fechada
// M = C(1+i)^n + A[((1+i)^n - 1)/i]. O laço e a fórmula chegam ao mesmo
// número, o que é uma boa forma de testar o próprio código.
func (s simulacao) montanteFormula() float64 {
	if s.taxa == 0 {
		return s.totalInvestido()
	}
	fator := math.Pow(1+s.taxa, float64(s.meses))
	return s.capital*fator + s.aporte*(fator-1)/s.taxa
}

// moeda formata um valor como R$ 1.234,56. Trabalhar em centavos inteiros
// evita o erro de arredondamento dos números de ponto flutuante.
func moeda(v float64) string {
	centavos := int64(math.Round(math.Abs(v) * 100))
	inteiro := strconv.FormatInt(centavos/100, 10)
	var b strings.Builder
	for i, d := range inteiro {
		if i > 0 && (len(inteiro)-i)%3 == 0 {
			b.WriteByte('.') // ponto de milhar a cada três dígitos
		}
		b.WriteRune(d)
	}
	texto := fmt.Sprintf("R$ %s,%02d", b.String(), centavos%100)
	if v < 0 {
		return "-" + texto
	}
	return texto
}

// pct formata uma porcentagem com vírgula, como 0,80%.
func pct(v float64) string {
	return strings.Replace(strconv.FormatFloat(v, 'f', 2, 64), ".", ",", 1) + "%"
}

// imprimir mostra os parâmetros, a tabela e o resumo final.
func imprimir(s simulacao) {
	fmt.Printf("Capital: %s | Taxa: %s ao mês | Prazo: %d meses | Aporte mensal: %s\n",
		moeda(s.capital), pct(s.taxa*100), s.meses, moeda(s.aporte))
	fmt.Println()
	fmt.Printf("%4s  %16s  %14s  %14s  %16s\n", "Mês", "Saldo inicial", "Juros", "Aporte", "Saldo final")
	fmt.Println(strings.Repeat("-", 72))
	for _, l := range s.tabela {
		fmt.Printf("%4d  %16s  %14s  %14s  %16s\n",
			l.mes, moeda(l.saldoInicial), moeda(l.juros), moeda(l.aporte), moeda(l.saldoFinal))
	}
	fmt.Println(strings.Repeat("-", 72))
	fmt.Printf("Montante final:            %s\n", moeda(s.montante()))
	fmt.Printf("Total investido:           %s\n", moeda(s.totalInvestido()))
	rendimento := 0.0
	if s.totalInvestido() > 0 {
		rendimento = s.jurosGanhos() / s.totalInvestido() * 100
	}
	fmt.Printf("Juros ganhos:              %s (%s do investido)\n", moeda(s.jurosGanhos()), pct(rendimento))
	fmt.Printf("Conferência pela fórmula:  %s\n", moeda(s.montanteFormula()))
}

// rodarDemo simula R$ 1.000 a 0,8% ao mês por 24 meses com aporte de R$ 200.
func rodarDemo() {
	fmt.Println("=== Demonstração: juros compostos com aporte mensal ===")
	fmt.Println()
	imprimir(simular(1000, 0.8, 24, 200))
	fmt.Println()
	fmt.Println("=== Mesmo cenário, sem aporte: veja a diferença no final ===")
	sem := simular(1000, 0.8, 24, 0)
	fmt.Printf("Montante sem aporte: %s | Juros ganhos: %s\n", moeda(sem.montante()), moeda(sem.jurosGanhos()))
}

// uso escreve no mesmo destino que flag.PrintDefaults usa (stderr), para
// que o texto e a lista de opções saiam juntos e na ordem certa.
func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Calculadora de juros compostos (CodeNomad Academy)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Uso: go run ./sources/01_calculadora_juros_compostos -capital 1000 -taxa 0.8 -meses 24 [-aporte 200]")
	fmt.Fprintln(w, "     go run ./sources/01_calculadora_juros_compostos --demo")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Opções:")
	flag.PrintDefaults()
}

func main() {
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	capital := flag.Float64("capital", 0, "capital inicial em reais")
	taxa := flag.Float64("taxa", 0, "taxa de juros mensal em porcentagem (0.8 significa 0,8%)")
	meses := flag.Int("meses", 0, "prazo em meses (1 a 1200)")
	aporte := flag.Float64("aporte", 0, "aporte mensal em reais, opcional")
	flag.Usage = uso
	flag.Parse()

	if *demo {
		rodarDemo()
		return
	}
	if flag.NFlag() == 0 { // nenhum argumento: mostra como usar
		uso()
		return
	}
	if err := validar(*capital, *taxa, *meses, *aporte); err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		fmt.Fprintln(os.Stderr, "Use -h para ver as opções.")
		os.Exit(1)
	}
	imprimir(simular(*capital, *taxa, *meses, *aporte))
}

// ------------------------------------------------------------
// Para ir além
// 1. Aceite uma taxa anual (-anual) e converta para mensal com
//    (1+anual)^(1/12) - 1, mostrando as duas taxas na saída.
// 2. Grave a tabela em um arquivo CSV com encoding/csv e abra em uma
//    planilha para desenhar o gráfico do saldo.
// 3. Descubra em quantos meses o saldo dobra: pare o laço assim que
//    saldoFinal >= 2 * capital e imprima o mês em que isso aconteceu.
// ------------------------------------------------------------
