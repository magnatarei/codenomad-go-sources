// ============================================================
// CodeNomad Academy  |  Sources Go  |  02. Simulador de financiamento (Price e SAC)
// ------------------------------------------------------------
// O que faz: recebe o valor financiado, a taxa mensal e o prazo. Gera a
//            tabela de parcelas nos sistemas Price (parcela fixa) e SAC
//            (amortização fixa) e compara o total pago e os juros de cada um.
// Como usar: go run ./sources/02_simulador_financiamento -valor 50000 -taxa 1 -meses 12
//            go run ./sources/02_simulador_financiamento -valor 50000 -taxa 1 -meses 12 -sistema sac
//            go run ./sources/02_simulador_financiamento --demo
// O que você aprende: interfaces para tratar dois algoritmos do mesmo jeito,
//            tipos sem campos que só carregam métodos, slices com capacidade
//            reservada por make, math.Pow, formatação de tabelas com fmt e
//            validação de entrada devolvendo error.
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

// parcela é uma linha da tabela de amortização.
type parcela struct {
	numero      int
	valor       float64 // parcela = juros + amortização
	juros       float64
	amortizacao float64
	saldo       float64 // saldo devedor depois de pagar esta parcela
}

// sistema é o contrato que os dois métodos de cálculo cumprem. Como a
// impressão e o resumo só dependem desta interface, o mesmo código serve
// para Price, SAC ou qualquer sistema que você criar depois.
type sistema interface {
	nome() string
	gerar(valor, taxa float64, meses int) []parcela
}

// price e sac são tipos sem campos: existem só para carregar os métodos.
type price struct{}
type sac struct{}

func (price) nome() string { return "Price (parcela fixa)" }

// No Price a parcela é constante: P = V * i / (1 - (1+i)^-n). A cada mês
// os juros caem e a amortização sobe, mas a soma dos dois não muda.
func (price) gerar(valor, taxa float64, meses int) []parcela {
	pmt := valor / float64(meses) // sem juros, basta dividir
	if taxa > 0 {
		pmt = valor * taxa / (1 - math.Pow(1+taxa, -float64(meses)))
	}
	saldo := valor
	tabela := make([]parcela, 0, meses)
	for n := 1; n <= meses; n++ {
		juros := saldo * taxa
		amort := pmt - juros
		saldo -= amort
		if n == meses {
			saldo = 0 // elimina o resíduo de arredondamento da última parcela
		}
		tabela = append(tabela, parcela{numero: n, valor: pmt, juros: juros, amortizacao: amort, saldo: saldo})
	}
	return tabela
}

func (sac) nome() string { return "SAC (amortização fixa)" }

// No SAC a amortização é sempre V/n. Os juros incidem sobre um saldo que
// diminui a cada mês, então a parcela começa alta e vai caindo.
func (sac) gerar(valor, taxa float64, meses int) []parcela {
	amort := valor / float64(meses)
	saldo := valor
	tabela := make([]parcela, 0, meses)
	for n := 1; n <= meses; n++ {
		juros := saldo * taxa
		saldo -= amort
		if n == meses {
			saldo = 0
		}
		tabela = append(tabela, parcela{numero: n, valor: amort + juros, juros: juros, amortizacao: amort, saldo: saldo})
	}
	return tabela
}

// resumo totaliza uma tabela de parcelas.
type resumo struct {
	primeira, ultima, totalPago, totalJuros float64
}

func resumir(tabela []parcela) resumo {
	r := resumo{primeira: tabela[0].valor, ultima: tabela[len(tabela)-1].valor}
	for _, p := range tabela {
		r.totalPago += p.valor
		r.totalJuros += p.juros
	}
	return r
}

func validar(valor, taxaPct float64, meses int, nomeSistema string) error {
	switch {
	case valor <= 0:
		return errors.New("o valor financiado deve ser maior que zero")
	case taxaPct < 0:
		return errors.New("a taxa mensal não pode ser negativa")
	case meses < 1 || meses > 600:
		return fmt.Errorf("o prazo deve ficar entre 1 e 600 meses (recebi %d)", meses)
	}
	switch nomeSistema {
	case "price", "sac", "ambos":
		return nil
	}
	return fmt.Errorf("sistema %q desconhecido: use price, sac ou ambos", nomeSistema)
}

// moeda formata um valor como R$ 1.234,56 usando centavos inteiros.
func moeda(v float64) string {
	centavos := int64(math.Round(math.Abs(v) * 100))
	inteiro := strconv.FormatInt(centavos/100, 10)
	var b strings.Builder
	for i, d := range inteiro {
		if i > 0 && (len(inteiro)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(d)
	}
	texto := fmt.Sprintf("R$ %s,%02d", b.String(), centavos%100)
	if v < 0 {
		return "-" + texto
	}
	return texto
}

func imprimirTabela(s sistema, tabela []parcela) {
	fmt.Printf("--- %s ---\n", s.nome())
	fmt.Printf("%4s  %14s  %14s  %14s  %16s\n", "Nº", "Parcela", "Juros", "Amortização", "Saldo devedor")
	for _, p := range tabela {
		fmt.Printf("%4d  %14s  %14s  %14s  %16s\n",
			p.numero, moeda(p.valor), moeda(p.juros), moeda(p.amortizacao), moeda(p.saldo))
	}
	fmt.Println()
}

// simular roda os sistemas escolhidos e imprime as tabelas e a comparação.
func simular(valor, taxaPct float64, meses int, escolha string) {
	taxa := taxaPct / 100
	fmt.Printf("Valor financiado: %s | Taxa: %s%% ao mês | Prazo: %d meses\n\n",
		moeda(valor), strings.Replace(strconv.FormatFloat(taxaPct, 'f', 2, 64), ".", ",", 1), meses)

	var sistemas []sistema
	if escolha == "price" || escolha == "ambos" {
		sistemas = append(sistemas, price{})
	}
	if escolha == "sac" || escolha == "ambos" {
		sistemas = append(sistemas, sac{})
	}

	resumos := make([]resumo, 0, len(sistemas))
	for _, s := range sistemas {
		tabela := s.gerar(valor, taxa, meses)
		imprimirTabela(s, tabela)
		resumos = append(resumos, resumir(tabela))
	}

	fmt.Printf("%-24s  %14s  %14s  %16s  %14s\n", "Resumo", "1ª parcela", "Última", "Total pago", "Total juros")
	for i, s := range sistemas {
		r := resumos[i]
		fmt.Printf("%-24s  %14s  %14s  %16s  %14s\n",
			s.nome(), moeda(r.primeira), moeda(r.ultima), moeda(r.totalPago), moeda(r.totalJuros))
	}
	if len(resumos) == 2 {
		diferenca := resumos[0].totalJuros - resumos[1].totalJuros
		fmt.Printf("\nNo SAC você paga %s a menos de juros, em troca de parcelas maiores no começo.\n", moeda(diferenca))
	}
}

// rodarDemo financia R$ 50.000 a 1% ao mês em 12 meses nos dois sistemas.
func rodarDemo() {
	fmt.Println("=== Demonstração: Price e SAC lado a lado ===")
	fmt.Println()
	simular(50000, 1, 12, "ambos")
}

// uso escreve no mesmo destino que flag.PrintDefaults usa (stderr).
func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Simulador de financiamento Price e SAC (CodeNomad Academy)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Uso: go run ./sources/02_simulador_financiamento -valor 50000 -taxa 1 -meses 12 [-sistema price|sac|ambos]")
	fmt.Fprintln(w, "     go run ./sources/02_simulador_financiamento --demo")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Opções:")
	flag.PrintDefaults()
}

func main() {
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	valor := flag.Float64("valor", 0, "valor financiado em reais")
	taxa := flag.Float64("taxa", 0, "taxa de juros mensal em porcentagem (1 significa 1%)")
	meses := flag.Int("meses", 0, "prazo em meses (1 a 600)")
	escolha := flag.String("sistema", "ambos", "price, sac ou ambos")
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
	if err := validar(*valor, *taxa, *meses, strings.ToLower(*escolha)); err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
	simular(*valor, *taxa, *meses, strings.ToLower(*escolha))
}

// ------------------------------------------------------------
// Para ir além
// 1. Crie um terceiro tipo que implemente a interface sistema, por exemplo
//    o sistema americano (só juros durante o prazo e o principal no fim),
//    e veja como nada mais precisa mudar.
// 2. Some uma taxa fixa mensal de seguro ou tarifa a cada parcela e mostre
//    o custo efetivo total do financiamento.
// 3. Receba a renda mensal do comprador e avise quando a primeira parcela
//    passar de 30% dela, um limite comum de comprometimento de renda.
// ------------------------------------------------------------
