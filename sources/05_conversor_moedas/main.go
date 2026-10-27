// ============================================================
// CodeNomad Academy  |  Sources Go  |  05. Conversor de moedas com tabela de taxas
// ------------------------------------------------------------
// O que faz: converte um valor entre duas moedas usando uma tabela de taxas
//            embutida no código (USD, EUR, GBP, JPY, ARS, BRL e outras),
//            mostra a taxa cruzada usada e lista a tabela inteira.
//            Para atualizar: troque os valores do map "taxas" (quantas
//            unidades de cada moeda 1 USD compra) e a constante dataTabela.
//            Nada mais precisa mudar.
// Como usar: go run ./sources/05_conversor_moedas -valor 100 -de USD -para BRL
//            go run ./sources/05_conversor_moedas -listar
//            go run ./sources/05_conversor_moedas --demo
// O que você aprende: constantes, map de structs como tabela de dados,
//            busca com o idioma "valor, ok := mapa[chave]", ordenação de
//            chaves com sort.Strings, strings.ToUpper para normalizar a
//            entrada e formatação de números com separador de milhar.
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
)

// dataTabela diz de quando são as taxas. Os valores são aproximados e
// servem para estudo; atualize os dois quando precisar de números atuais.
const dataTabela = "2025-06-30"

// moeda guarda o nome, quantas unidades 1 USD compra e quantas casas
// a moeda costuma mostrar (o iene, por exemplo, não usa centavos).
type moeda struct {
	nome     string
	porDolar float64
	casas    int
}

var taxas = map[string]moeda{
	"USD": {"dólar americano", 1, 2},
	"BRL": {"real brasileiro", 5.45, 2},
	"EUR": {"euro", 0.92, 2},
	"GBP": {"libra esterlina", 0.79, 2},
	"JPY": {"iene japonês", 158, 0},
	"ARS": {"peso argentino", 1190, 0},
	"CAD": {"dólar canadense", 1.37, 2},
	"AUD": {"dólar australiano", 1.53, 2},
	"CHF": {"franco suíço", 0.89, 2},
	"CNY": {"yuan chinês", 7.25, 2},
	"MXN": {"peso mexicano", 18.9, 2},
	"CLP": {"peso chileno", 940, 0},
	"UYU": {"peso uruguaio", 40.5, 2},
	"INR": {"rupia indiana", 85.6, 2},
}

// buscar normaliza o código (brl vira BRL) e procura na tabela. O "ok"
// do map diz se a chave existe; sem ele, uma chave ausente devolveria
// um struct vazio em silêncio.
func buscar(codigo string) (string, moeda, error) {
	chave := strings.ToUpper(strings.TrimSpace(codigo))
	m, ok := taxas[chave]
	if !ok {
		return "", moeda{}, fmt.Errorf("moeda %q desconhecida; use -listar para ver as disponíveis", codigo)
	}
	return chave, m, nil
}

// converter passa pelo dólar: primeiro leva o valor a USD, depois a USD
// para a moeda de destino. A taxa cruzada é o quanto 1 unidade de
// origem vale no destino.
func converter(valor float64, de, para string) (resultado, cruzada float64, err error) {
	if valor < 0 {
		return 0, 0, fmt.Errorf("o valor não pode ser negativo (recebi %s)", formatar(valor, 2))
	}
	_, mDe, err := buscar(de)
	if err != nil {
		return 0, 0, err
	}
	_, mPara, err := buscar(para)
	if err != nil {
		return 0, 0, err
	}
	cruzada = mPara.porDolar / mDe.porDolar
	return valor * cruzada, cruzada, nil
}

// formatar escreve 1234567.891 como 1.234.567,89 (com as casas pedidas).
func formatar(v float64, casas int) string {
	escala := math.Pow(10, float64(casas))
	total := int64(math.Round(math.Abs(v) * escala))
	inteiro := strconv.FormatInt(total/int64(escala), 10)
	var b strings.Builder
	for i, d := range inteiro {
		if i > 0 && (len(inteiro)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(d)
	}
	if casas > 0 {
		fmt.Fprintf(&b, ",%0*d", casas, total%int64(escala)) // %0*d: largura vem do argumento
	}
	if v < 0 {
		return "-" + b.String()
	}
	return b.String()
}

// mostrar converte e imprime uma linha legível, ou devolve o erro.
func mostrar(valor float64, de, para string) error {
	resultado, cruzada, err := converter(valor, de, para)
	if err != nil {
		return err
	}
	codDe, mDe, _ := buscar(de) // já validados em converter
	codPara, mPara, _ := buscar(para)
	fmt.Printf("%s %s = %s %s   (1 %s = %s %s)\n",
		formatar(valor, mDe.casas), codDe, formatar(resultado, mPara.casas), codPara,
		codDe, formatar(cruzada, 4), codPara)
	return nil
}

// listar imprime a tabela em ordem alfabética de código, com o valor de
// cada moeda em dólar e em real.
func listar() {
	codigos := make([]string, 0, len(taxas))
	for c := range taxas {
		codigos = append(codigos, c)
	}
	sort.Strings(codigos)
	fmt.Printf("Tabela de taxas de %s (valores aproximados, para estudo)\n", dataTabela)
	fmt.Printf("%-5s %-20s %16s %16s\n", "Cód.", "Moeda", "Por 1 USD", "1 unidade em BRL")
	for _, c := range codigos {
		m := taxas[c]
		emReais := taxas["BRL"].porDolar / m.porDolar
		fmt.Printf("%-5s %-20s %16s %16s\n", c, m.nome, formatar(m.porDolar, 4), formatar(emReais, 4))
	}
}

// rodarDemo converte 100 unidades em seis pares e imprime a tabela.
func rodarDemo() {
	fmt.Println("=== Demonstração: 100 unidades em seis pares de moedas ===")
	pares := [][2]string{{"USD", "BRL"}, {"EUR", "BRL"}, {"BRL", "USD"}, {"GBP", "EUR"}, {"JPY", "BRL"}, {"ARS", "BRL"}}
	for _, p := range pares {
		if err := mostrar(100, p[0], p[1]); err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
		}
	}
	fmt.Println()
	listar()
}

// uso escreve no mesmo destino que flag.PrintDefaults usa (stderr).
func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Conversor de moedas com tabela embutida (CodeNomad Academy)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Uso: go run ./sources/05_conversor_moedas -valor 100 -de USD -para BRL")
	fmt.Fprintln(w, "     go run ./sources/05_conversor_moedas -listar")
	fmt.Fprintln(w, "     go run ./sources/05_conversor_moedas --demo")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Opções:")
	flag.PrintDefaults()
}

func main() {
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	valor := flag.Float64("valor", 1, "quantia a converter")
	de := flag.String("de", "", "código da moeda de origem (ex.: USD)")
	para := flag.String("para", "", "código da moeda de destino (ex.: BRL)")
	lista := flag.Bool("listar", false, "mostra a tabela de taxas e sai")
	flag.Usage = uso
	flag.Parse()

	switch {
	case *demo:
		rodarDemo()
	case *lista:
		listar()
	case flag.NFlag() == 0:
		uso()
	case *de == "" || *para == "":
		fmt.Fprintln(os.Stderr, "Erro: informe -de e -para (ex.: -de USD -para BRL)")
		os.Exit(1)
	default:
		if err := mostrar(*valor, *de, *para); err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			os.Exit(1)
		}
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Leia a tabela de um arquivo taxas.csv (código;nome;porDolar;casas)
//    com encoding/csv, mantendo a embutida como reserva se o arquivo faltar.
// 2. Acrescente um -spread em porcentagem para simular a diferença entre
//    a taxa de compra e a taxa de venda, como acontece no câmbio real.
// 3. Crie um modo -tabela que mostre quanto vale -valor da moeda -de em
//    todas as outras, ordenado da mais valorizada para a menos.
// ------------------------------------------------------------
