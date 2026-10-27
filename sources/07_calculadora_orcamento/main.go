// ============================================================
// CodeNomad Academy  |  Sources Go  |  07. Calculadora de orçamento mensal
// ------------------------------------------------------------
// O que faz: soma receitas e despesas por categoria, lidas dos argumentos ou
//            de um arquivo simples, e mostra total, saldo, percentual de cada
//            categoria sobre a receita, barras de texto e a categoria que mais pesa.
// Como usar: go run ./sources/07_calculadora_orcamento -receita "Salário=3500" -despesa "Aluguel=1200" -despesa "Mercado=800"
//            go run ./sources/07_calculadora_orcamento -arquivo orcamento.txt
//            go run ./sources/07_calculadora_orcamento --demo
//            O arquivo tem uma linha "tipo;categoria;valor" por lançamento; # abre comentário.
// O que você aprende: a interface flag.Value para repetir uma opção, bufio.Scanner
//            lendo de qualquer io.Reader (arquivo ou string), strings.Cut, maps para
//            somar por categoria, sort.Slice com desempate, defer e erros com %w.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// lista implementa flag.Value: com ela a mesma opção pode aparecer várias
// vezes na linha de comando e cada valor é guardado no slice.
type lista []string

func (l *lista) String() string     { return strings.Join(*l, ", ") }
func (l *lista) Set(v string) error { *l = append(*l, v); return nil }

// orcamento soma os valores por categoria, separando receitas de despesas.
type orcamento struct {
	receitas map[string]float64
	despesas map[string]float64
}

// adicionar lê "Categoria=valor", valida e acumula. Categorias repetidas somam.
func (o *orcamento) adicionar(tipo, texto string) error {
	categoria, valorTexto, achou := strings.Cut(texto, "=")
	categoria = strings.TrimSpace(categoria)
	if !achou || categoria == "" {
		return fmt.Errorf("%q não está no formato Categoria=valor", texto)
	}
	valor, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(valorTexto), ",", "."), 64)
	if err != nil || math.IsNaN(valor) || math.IsInf(valor, 0) || valor <= 0 {
		return fmt.Errorf("o valor de %q precisa ser um número maior que zero", categoria)
	}
	switch strings.ToLower(strings.TrimSpace(tipo)) {
	case "receita":
		o.receitas[categoria] += valor
	case "despesa":
		o.despesas[categoria] += valor
	default:
		return fmt.Errorf("tipo %q desconhecido: use receita ou despesa", tipo)
	}
	return nil
}

// carregar lê linhas "tipo;categoria;valor" de qualquer io.Reader: serve
// tanto para um arquivo aberto com os.Open quanto para uma string.
func (o *orcamento) carregar(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		linha := strings.TrimSpace(sc.Text())
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue
		}
		partes := strings.Split(linha, ";")
		if len(partes) != 3 {
			return fmt.Errorf("linha %d: esperava tipo;categoria;valor, recebi %q", n, linha)
		}
		if err := o.adicionar(partes[0], partes[1]+"="+partes[2]); err != nil {
			return fmt.Errorf("linha %d: %w", n, err)
		}
	}
	return sc.Err()
}

// carregarArquivo abre o arquivo; o defer garante que ele fecha mesmo com erro.
func (o *orcamento) carregarArquivo(caminho string) error {
	f, err := os.Open(caminho)
	if err != nil {
		return fmt.Errorf("não consegui abrir o arquivo: %w", err)
	}
	defer f.Close()
	return o.carregar(f)
}

// ordenar devolve as categorias do maior para o menor valor, desempatando
// pelo nome (a ordem de um map é aleatória), e o total somado.
func ordenar(m map[string]float64) ([]string, float64) {
	chaves, total := make([]string, 0, len(m)), 0.0
	for c, v := range m {
		chaves = append(chaves, c)
		total += v
	}
	sort.Slice(chaves, func(i, j int) bool {
		if m[chaves[i]] != m[chaves[j]] {
			return m[chaves[i]] > m[chaves[j]]
		}
		return chaves[i] < chaves[j]
	})
	return chaves, total
}

// reais formata um valor como R$ 1.234,56; pct formata 30.9 como 30,9%.
func reais(v float64) string {
	centavos := int64(math.Round(math.Abs(v) * 100))
	inteiro := strconv.FormatInt(centavos/100, 10)
	var b strings.Builder
	for i, d := range inteiro {
		if i > 0 && (len(inteiro)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(d)
	}
	sinal := ""
	if v < 0 {
		sinal = "-"
	}
	return fmt.Sprintf("%sR$ %s,%02d", sinal, b.String(), centavos%100)
}

func pct(v float64) string {
	return strings.Replace(strconv.FormatFloat(v, 'f', 1, 64), ".", ",", 1) + "%"
}

// relatorio imprime as duas listas, as barras e o alerta final.
func (o *orcamento) relatorio() error {
	receitas, receita := ordenar(o.receitas)
	if receita <= 0 {
		return fmt.Errorf("informe pelo menos uma receita para calcular os percentuais")
	}
	despesas, despesa := ordenar(o.despesas)
	fmt.Println("RECEITAS")
	for _, c := range receitas {
		fmt.Printf("  %-22s %14s\n", c, reais(o.receitas[c]))
	}
	fmt.Printf("  %-22s %14s\n\n", "Total", reais(receita))
	fmt.Println("DESPESAS (percentual sobre a receita)")
	for _, c := range despesas {
		parte := o.despesas[c] / receita * 100
		barra := strings.Repeat("#", int(math.Round(parte/2))) // cada # vale 2%
		fmt.Printf("  %-22s %14s  %6s  %s\n", c, reais(o.despesas[c]), pct(parte), barra)
	}
	fmt.Printf("  %-22s %14s  %6s\n\n", "Total", reais(despesa), pct(despesa/receita*100))
	saldo := receita - despesa
	fmt.Printf("SALDO DO MÊS: %s (%s da receita)\n", reais(saldo), pct(saldo/receita*100))
	switch {
	case saldo < 0:
		fmt.Println("  Alerta: as despesas passaram da receita. O mês fechou no vermelho.")
	case saldo < receita*0.1:
		fmt.Println("  Atenção: sobrou menos de 10% da receita. Pouca margem para imprevistos.")
	default:
		fmt.Println("  Bom sinal: sobrou mais de 10% da receita.")
	}
	if len(despesas) > 0 {
		c := despesas[0]
		fmt.Printf("  A categoria que mais pesa é %q: %s, ou %s da receita.\n", c, reais(o.despesas[c]), pct(o.despesas[c]/receita*100))
	}
	return nil
}

// exemplo é um orçamento embutido no mesmo formato do arquivo.
const exemplo = `# tipo;categoria;valor
receita;Salário;4200
receita;Trabalho extra;650
despesa;Aluguel;1500
despesa;Mercado;900
despesa;Transporte;380
despesa;Internet e celular;160
despesa;Lazer;420
despesa;Curso de programação;240
despesa;Mercado;150
`

// falhar imprime o erro e encerra o programa com código de saída 1.
func falhar(err error) {
	fmt.Fprintln(os.Stderr, "Erro:", err)
	os.Exit(1)
}

// rodarDemo carrega o exemplo com strings.NewReader e imprime o relatório.
func rodarDemo() {
	fmt.Println("=== Demonstração: orçamento de exemplo ===")
	o := &orcamento{receitas: map[string]float64{}, despesas: map[string]float64{}}
	if err := o.carregar(strings.NewReader(exemplo)); err != nil {
		falhar(err)
	}
	if err := o.relatorio(); err != nil {
		falhar(err)
	}
}

// uso escreve no mesmo destino que flag.PrintDefaults usa (stderr).
func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Calculadora de orçamento mensal (CodeNomad Academy)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Uso: go run ./sources/07_calculadora_orcamento -receita \"Salário=3500\" -despesa \"Aluguel=1200\" -despesa \"Mercado=800\"")
	fmt.Fprintln(w, "     go run ./sources/07_calculadora_orcamento -arquivo orcamento.txt   (linhas tipo;categoria;valor)")
	fmt.Fprintln(w, "     go run ./sources/07_calculadora_orcamento --demo")
	fmt.Fprintln(w)
	flag.PrintDefaults()
}

func main() {
	var receitas, despesas lista
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	arquivo := flag.String("arquivo", "", "arquivo com linhas tipo;categoria;valor")
	flag.Var(&receitas, "receita", "receita no formato Categoria=valor (pode repetir)")
	flag.Var(&despesas, "despesa", "despesa no formato Categoria=valor (pode repetir)")
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
	o := &orcamento{receitas: map[string]float64{}, despesas: map[string]float64{}}
	if *arquivo != "" {
		if err := o.carregarArquivo(*arquivo); err != nil {
			falhar(err)
		}
	}
	for _, r := range receitas {
		if err := o.adicionar("receita", r); err != nil {
			falhar(err)
		}
	}
	for _, d := range despesas {
		if err := o.adicionar("despesa", d); err != nil {
			falhar(err)
		}
	}
	if err := o.relatorio(); err != nil {
		falhar(err)
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Grave o relatório em CSV (categoria;valor;percentual) com encoding/csv.
// 2. Aceite um -limite por categoria ("Lazer=300") e avise quando passar dele.
// 3. Leia vários meses (um arquivo por mês) e mostre a evolução do saldo.
// ------------------------------------------------------------
