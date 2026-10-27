// ============================================================
// CodeNomad Academy  |  Sources Go  |  38. Gerador de relatório em Markdown
// ------------------------------------------------------------
// O que faz: monta um documento Markdown (títulos, parágrafos, listas, tabela
//            alinhada e blocos de código) a partir de dados estruturados, com
//            um construtor encadeável. Também converte qualquer CSV numa
//            tabela Markdown com as colunas alinhadas e os números à direita.
// Como usar: go run ./sources/38_gerador_markdown -csv notas.csv -titulo "Notas da turma" -destino relatorio.md
//            go run ./sources/38_gerador_markdown -csv notas.csv
//            go run ./sources/38_gerador_markdown --demo
// O que você aprende: strings.Builder, métodos que devolvem o próprio
//            receptor (encadeamento), funções variádicas, encoding/csv,
//            strconv.ParseFloat para decidir o alinhamento e utf8.RuneCountInString
//            para alinhar texto com acentos.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Doc acumula o Markdown num strings.Builder. Cada método devolve *Doc,
// o que permite encadear: doc.Titulo(...).Paragrafo(...).Lista(...).
type Doc struct {
	b strings.Builder
}

func (d *Doc) Titulo(nivel int, texto string) *Doc {
	nivel = max(1, min(nivel, 6)) // Markdown só tem seis níveis
	fmt.Fprintf(&d.b, "%s %s\n\n", strings.Repeat("#", nivel), texto)
	return d
}

func (d *Doc) Paragrafo(texto string) *Doc {
	d.b.WriteString(texto + "\n\n")
	return d
}

// Lista é variádica: aceita quantos itens quiser, ou um slice com "itens...".
func (d *Doc) Lista(itens ...string) *Doc {
	for _, item := range itens {
		d.b.WriteString("- " + item + "\n")
	}
	d.b.WriteString("\n")
	return d
}

func (d *Doc) Codigo(linguagem, codigo string) *Doc {
	fmt.Fprintf(&d.b, "```%s\n%s\n```\n\n", linguagem, strings.TrimRight(codigo, "\n"))
	return d
}

func (d *Doc) Tabela(cabecalho []string, linhas [][]string) *Doc {
	d.b.WriteString(tabelaMarkdown(cabecalho, linhas) + "\n")
	return d
}

func (d *Doc) String() string {
	return d.b.String()
}

// numerico diz se uma célula é um número (aceita vírgula como separador e "%").
func numerico(s string) bool {
	s = strings.TrimSuffix(strings.ReplaceAll(strings.TrimSpace(s), ",", "."), "%")
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// largura conta letras visíveis, não bytes: "média" tem 5 letras e 6 bytes.
// Usar len() aqui desalinharia qualquer coluna com acento.
func largura(s string) int {
	return utf8.RuneCountInString(s)
}

// preencher completa a célula com espaços à esquerda (números) ou à direita (texto).
func preencher(s string, n int, direita bool) string {
	falta := strings.Repeat(" ", max(0, n-largura(s)))
	if direita {
		return falta + s
	}
	return s + falta
}

// tabelaMarkdown gera a tabela com todas as colunas na mesma largura e usa
// "---:" na linha separadora para alinhar colunas numéricas à direita.
func tabelaMarkdown(cabecalho []string, linhas [][]string) string {
	n := len(cabecalho)
	larguras := make([]int, n)
	numerica := make([]bool, n)
	for i, c := range cabecalho {
		larguras[i] = max(3, largura(c)) // 3 é o mínimo para "---"
		numerica[i] = len(linhas) > 0
	}
	for _, l := range linhas {
		for i := 0; i < n && i < len(l); i++ {
			larguras[i] = max(larguras[i], largura(l[i]))
			numerica[i] = numerica[i] && numerico(l[i]) // uma célula de texto já torna a coluna textual
		}
	}
	var b strings.Builder
	escrever := func(celulas []string) {
		b.WriteString("|")
		for i := 0; i < n; i++ {
			cel := ""
			if i < len(celulas) {
				cel = celulas[i]
			}
			b.WriteString(" " + preencher(cel, larguras[i], numerica[i]) + " |")
		}
		b.WriteString("\n")
	}
	escrever(cabecalho)
	b.WriteString("|")
	for i := 0; i < n; i++ {
		if numerica[i] {
			b.WriteString(" " + strings.Repeat("-", larguras[i]-1) + ": |")
		} else {
			b.WriteString(" " + strings.Repeat("-", larguras[i]) + " |")
		}
	}
	b.WriteString("\n")
	for _, l := range linhas {
		escrever(l)
	}
	return b.String()
}

// lerCSV devolve o cabeçalho e as linhas de um arquivo CSV.
func lerCSV(caminho string) ([]string, [][]string, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1 // aceita linhas com quantidades diferentes de colunas
	tudo, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("CSV inválido: %w", err)
	}
	if len(tudo) < 2 {
		return nil, nil, errors.New("o CSV precisa ter um cabeçalho e pelo menos uma linha de dados")
	}
	return tudo[0], tudo[1:], nil
}

// relatorioDeCSV monta um documento simples em volta da tabela do CSV.
func relatorioDeCSV(titulo, caminho string) (string, error) {
	cab, linhas, err := lerCSV(caminho)
	if err != nil {
		return "", err
	}
	doc := &Doc{}
	doc.Titulo(1, titulo).
		Paragrafo(fmt.Sprintf("Tabela montada a partir de `%s` com %d linha(s) e %d coluna(s).", caminho, len(linhas), len(cab))).
		Tabela(cab, linhas)
	return doc.String(), nil
}

func demo() error {
	fmt.Println("== Demonstração: relatório montado a partir de dados em memória")
	fmt.Println()
	alunos := [][]string{
		{"Ana Souza", "Go", "9,5", "92%"}, {"Bruno Lima", "Rust", "7,8", "81%"},
		{"Cátia Mendes", "Python", "8,9", "88%"}, {"Diego Rocha", "Go", "6,4", "70%"},
	}
	doc := &Doc{}
	doc.Titulo(1, "Relatório da turma 2026.2").
		Paragrafo("Resumo do desempenho dos alunos na trilha de autoestudo da CodeNomad Academy.").
		Titulo(2, "Alunos").
		Tabela([]string{"Nome", "Linguagem", "Média", "Presença"}, alunos).
		Titulo(2, "Destaques").
		Lista("Maior média: Ana Souza (9,5)", "Trilha mais procurada: Go (2 alunos)", "Presença média da turma: 82,75%").
		Titulo(2, "Como a média foi calculada").
		Codigo("go", "media := (p1 + p2 + trabalho) / 3\nfmt.Printf(\"%.1f\\n\", media)").
		Paragrafo("Documento montado pelo source 38 da CodeNomad Academy.")
	fmt.Print(doc.String())

	fmt.Println("== Outra tabela, vista como texto cru (repare no alinhamento):")
	fmt.Println(tabelaMarkdown([]string{"Produto", "Qtd", "Preço"}, [][]string{{"Caderno", "12", "18,90"}, {"Mochila", "3", "149,00"}, {"Caneta azul", "150", "3,50"}}))
	fmt.Println("== Uma coluna com um texto no meio deixa de ser numérica:")
	fmt.Print(tabelaMarkdown([]string{"Item", "Valor"}, [][]string{{"a", "10"}, {"b", "n/d"}, {"c", "7"}}))
	return nil
}

func main() {
	arquivoCSV := flag.String("csv", "", "arquivo CSV para converter em tabela Markdown")
	titulo := flag.String("titulo", "Relatório", "título do documento")
	destino := flag.String("destino", "", "grava o Markdown neste arquivo (sem a flag, imprime na tela)")
	modoDemo := flag.Bool("demo", false, "gera um relatório de exemplo e imprime o Markdown")
	flag.Parse()

	var err error
	switch {
	case *modoDemo:
		err = demo()
	case *arquivoCSV == "":
		fmt.Fprintln(os.Stderr, "Uso: markdown -csv dados.csv [-titulo \"Meu relatório\"] [-destino relatorio.md]")
		flag.PrintDefaults()
		os.Exit(2)
	default:
		var md string
		if md, err = relatorioDeCSV(*titulo, *arquivoCSV); err == nil {
			if *destino == "" {
				fmt.Print(md)
			} else if err = os.WriteFile(*destino, []byte(md), 0o644); err == nil {
				fmt.Println("Markdown gravado em", *destino)
			}
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Acrescente o método Link(texto, url) e Imagem(alt, caminho) ao Doc.
// 2. Crie -ordenar <coluna> para ordenar as linhas do CSV antes de montar a tabela.
// 3. Gere também um sumário automático com links para cada título de nível 2.
