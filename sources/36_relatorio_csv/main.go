// ============================================================
// CodeNomad Academy  |  Sources Go  |  36. Relatório de vendas a partir de CSV
// ------------------------------------------------------------
// O que faz: lê um CSV de vendas (data, produto, categoria, quantidade, valor
//            unitário) e calcula total geral, totais por categoria e por mês,
//            ticket médio e produto mais vendido, num relatório alinhado.
// Como usar: go run ./sources/36_relatorio_csv -arquivo vendas.csv
//            go run ./sources/36_relatorio_csv -arquivo vendas.csv -top 5
//            go run ./sources/36_relatorio_csv --demo
// O que você aprende: encoding/csv linha a linha, strconv, time.Parse, maps
//            para somar por chave, funções como parâmetro, sort.Slice e Printf.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Venda é uma linha do CSV já convertida para os tipos certos.
type Venda struct {
	Data       time.Time
	Produto    string
	Categoria  string
	Quantidade int
	Valor      float64 // valor unitário
}

// Total é o que a venda rendeu: quantidade vezes valor unitário.
func (v Venda) Total() float64 { return float64(v.Quantidade) * v.Valor }

// lerVendas lê o CSV uma linha por vez e aponta o número da linha em qualquer erro.
func lerVendas(caminho string) ([]Venda, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = 5
	var vendas []Venda
	for linha := 1; ; linha++ {
		campos, err := r.Read()
		if err == io.EOF {
			break
		}
		if errors.Is(err, csv.ErrFieldCount) {
			return nil, fmt.Errorf("linha %d: esperava 5 colunas e achei %d", linha, len(campos))
		}
		if err != nil {
			return nil, fmt.Errorf("linha %d: %w", linha, err)
		}
		if linha == 1 && strings.EqualFold(campos[0], "data") {
			continue // cabeçalho
		}
		v, err := converter(campos)
		if err != nil {
			return nil, fmt.Errorf("linha %d: %w", linha, err)
		}
		vendas = append(vendas, v)
	}
	if len(vendas) == 0 {
		return nil, errors.New("o arquivo não tem nenhuma venda")
	}
	return vendas, nil
}

// converter valida os cinco campos de uma linha e monta a Venda.
func converter(c []string) (Venda, error) {
	data, err := time.Parse("2006-01-02", strings.TrimSpace(c[0]))
	if err != nil {
		return Venda{}, fmt.Errorf("data %q inválida (use AAAA-MM-DD)", c[0])
	}
	qtd, err := strconv.Atoi(strings.TrimSpace(c[3]))
	if err != nil || qtd <= 0 {
		return Venda{}, fmt.Errorf("quantidade %q inválida: precisa ser um inteiro positivo", c[3])
	}
	// Aceita "12,50" além de "12.50": troca a vírgula antes de converter.
	valor, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(c[4]), ",", "."), 64)
	if err != nil || valor < 0 {
		return Venda{}, fmt.Errorf("valor %q inválido", c[4])
	}
	produto, categoria := strings.TrimSpace(c[1]), strings.TrimSpace(c[2])
	if produto == "" || categoria == "" {
		return Venda{}, errors.New("produto e categoria não podem ficar vazios")
	}
	return Venda{data, produto, categoria, qtd, valor}, nil
}

// par liga um nome a um valor; usamos para ordenar o que veio de um map.
type par struct {
	Nome  string
	Valor float64
}

// somarPor agrupa as vendas por uma chave qualquer (categoria, mês, produto) e
// devolve a soma ordenada. Receber a função "chave" evita repetir o laço três vezes.
func somarPor(vendas []Venda, chave func(Venda) string, medida func(Venda) float64) []par {
	somas := map[string]float64{}
	for _, v := range vendas {
		somas[chave(v)] += medida(v)
	}
	pares := make([]par, 0, len(somas))
	for nome, valor := range somas {
		pares = append(pares, par{nome, valor})
	}
	sort.Slice(pares, func(i, j int) bool {
		if pares[i].Valor != pares[j].Valor {
			return pares[i].Valor > pares[j].Valor
		}
		return pares[i].Nome < pares[j].Nome // desempate estável
	})
	return pares
}

// reais formata no padrão brasileiro: R$ 1.234,56.
func reais(v float64) string {
	inteiro, cents, _ := strings.Cut(strconv.FormatFloat(v, 'f', 2, 64), ".")
	var b strings.Builder
	for i, d := range inteiro {
		if i > 0 && (len(inteiro)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(d)
	}
	return "R$ " + b.String() + "," + cents
}

func imprimirTabela(titulo string, pares []par, total float64, top int) {
	fmt.Printf("\n%s\n%s\n", titulo, strings.Repeat("-", 48))
	if top > 0 && len(pares) > top {
		pares = pares[:top]
	}
	for _, p := range pares {
		barra := strings.Repeat("#", int(p.Valor/total*20+0.5))
		fmt.Printf("  %-14s %16s  %5.1f%% %s\n", p.Nome, reais(p.Valor), p.Valor/total*100, barra)
	}
}

func relatorio(vendas []Venda, top int) {
	var total float64
	for _, v := range vendas {
		total += v.Total()
	}
	fmt.Printf("RELATÓRIO DE VENDAS\n%s\n", strings.Repeat("=", 48))
	fmt.Printf("  Vendas:         %d\n", len(vendas))
	fmt.Printf("  Faturamento:    %s\n", reais(total))
	fmt.Printf("  Ticket médio:   %s por venda\n", reais(total/float64(len(vendas))))

	imprimirTabela("Por categoria", somarPor(vendas, func(v Venda) string { return v.Categoria }, Venda.Total), total, 0)
	// A chave "2006-01" é o ano-mês; ordenar por nome a põe em ordem cronológica.
	porMes := somarPor(vendas, func(v Venda) string { return v.Data.Format("2006-01") }, Venda.Total)
	sort.Slice(porMes, func(i, j int) bool { return porMes[i].Nome < porMes[j].Nome })
	imprimirTabela("Por mês", porMes, total, 0)
	porProduto := somarPor(vendas, func(v Venda) string { return v.Produto }, Venda.Total)
	imprimirTabela(fmt.Sprintf("Top %d produtos por faturamento", top), porProduto, total, top)
	unidades := somarPor(vendas, func(v Venda) string { return v.Produto }, func(v Venda) float64 { return float64(v.Quantidade) })
	fmt.Printf("\nProduto mais vendido em unidades: %s (%.0f unidades)\n", unidades[0].Nome, unidades[0].Valor)
}

// gerarExemplo grava um CSV com 60 vendas sorteadas, mas sempre as mesmas:
// a semente fixa deixa a demonstração reproduzível.
func gerarExemplo(caminho string) error {
	r := rand.New(rand.NewSource(42))
	produtos := []struct {
		nome, categoria string
		preco           float64
	}{
		{"Caderno", "Papelaria", 18.90}, {"Caneta", "Papelaria", 3.50}, {"Mochila", "Acessórios", 149.00},
		{"Fone", "Eletrônicos", 89.90}, {"Mouse", "Eletrônicos", 59.90}, {"Garrafa", "Acessórios", 42.00},
	}
	f, err := os.Create(caminho)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush() // sem o Flush, o fim do arquivo ficaria no buffer e não no disco
	if err := w.Write([]string{"data", "produto", "categoria", "quantidade", "valor"}); err != nil {
		return err
	}
	inicio := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 60; i++ {
		p := produtos[r.Intn(len(produtos))]
		dia := inicio.AddDate(0, 0, r.Intn(92)) // julho, agosto e setembro
		linha := []string{dia.Format("2006-01-02"), p.nome, p.categoria, strconv.Itoa(1 + r.Intn(5)), strconv.FormatFloat(p.preco, 'f', 2, 64)}
		if err := w.Write(linha); err != nil {
			return err
		}
	}
	return nil
}

func demo() error {
	dir, err := os.MkdirTemp("", "codenomad-vendas-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	caminho := filepath.Join(dir, "vendas.csv")
	if err := gerarExemplo(caminho); err != nil {
		return err
	}
	fmt.Println("== Demonstração: CSV com 60 vendas criado em", caminho)
	vendas, err := lerVendas(caminho)
	if err != nil {
		return err
	}
	relatorio(vendas, 3)
	fmt.Println("\n== Uma linha inválida é recusada com uma mensagem clara:")
	if _, err := converter([]string{"2026-07-02", "Fone", "Eletrônicos", "zero", "89.90"}); err != nil {
		fmt.Println("  Erro esperado:", err)
	}
	return nil
}

func main() {
	arquivo := flag.String("arquivo", "", "CSV com as colunas data,produto,categoria,quantidade,valor")
	top := flag.Int("top", 5, "quantos produtos mostrar no ranking")
	modoDemo := flag.Bool("demo", false, "gera um CSV de exemplo e imprime o relatório")
	flag.Parse()
	var err error
	switch {
	case *modoDemo:
		err = demo()
	case *arquivo == "":
		fmt.Fprintln(os.Stderr, "Uso: relatorio -arquivo vendas.csv [-top 5]")
		flag.PrintDefaults()
		os.Exit(2)
	case *top < 1:
		err = errors.New("-top precisa ser pelo menos 1")
	default:
		var vendas []Venda
		if vendas, err = lerVendas(*arquivo); err == nil {
			relatorio(vendas, *top)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Acrescente -inicio e -fim para filtrar o período antes de calcular.
// 2. Mostre o dia da semana com mais vendas (time.Weekday) e o pior.
// 3. Grave o relatório também em CSV, uma tabela por arquivo, com encoding/csv.
