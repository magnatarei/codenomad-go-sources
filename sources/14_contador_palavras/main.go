// ============================================================
// CodeNomad Academy  |  Sources Go  |  14. Contador de palavras e frequência
// ------------------------------------------------------------
// O que faz: lê um texto (argumento ou arquivo) e conta letras, palavras,
//            linhas, frases e parágrafos; lista as N palavras mais frequentes
//            ignorando acentos e uma lista de palavras vazias do português.
// Como usar: go run ./sources/14_contador_palavras --arquivo redacao.txt --top 15
//            go run ./sources/14_contador_palavras --texto "Go é simples. Go é rápido."
//            go run ./sources/14_contador_palavras --demo
// O que você aprende: strings.Fields e Map, unicode, maps de contagem,
//            sort.Slice com desempate, os.ReadFile, structs com métodos.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Estatisticas guarda os totais de um texto.
type Estatisticas struct {
	Runas, SemEspacos, Palavras, Linhas, Frases, Paragrafos int
	MaiorPalavra                                            string
	Frequencia                                              map[string]int
	Forma                                                   map[string]string // primeira grafia vista de cada palavra normalizada
}

// Contagem é um par (palavra, vezes) usado no ranking.
type Contagem struct {
	Palavra string
	Vezes   int
}

// acentos mapeia cada letra acentuada para a letra simples equivalente.
var acentos = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i', 'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u', 'ç': 'c', 'ñ': 'n',
}

// vazias são palavras tão comuns que só atrapalham o ranking (artigos,
// preposições, pronomes). O map é preenchido no init, já normalizado.
var vazias = map[string]bool{}

func init() {
	lista := `de a o que e do da em um para é com não uma os no se na por mais as dos
	como mas foi ao ele das tem à seu sua ou ser quando muito há nos já está eu também
	só pelo pela até isso ela entre era depois sem mesmo aos ter seus quem nas me esse
	eles estão você tinha foram essa num nem suas meu às minha têm numa pelos elas havia
	seja qual será nós tenho lhe deles essas esses pelas este fosse dele são vai ter`
	for _, p := range strings.Fields(lista) {
		vazias[normalizar(p)] = true
	}
}

// normalizar põe a palavra em minúsculas e tira acentos e pontuação, para a
// mesma palavra com e sem acento (ou com vírgula colada) contar uma vez só.
func normalizar(p string) string {
	limpa := strings.Map(func(r rune) rune {
		r = unicode.ToLower(r)
		if simples, ok := acentos[r]; ok {
			return simples
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			return r
		}
		return -1 // -1 no strings.Map apaga a runa
	}, p)
	return strings.Trim(limpa, "-")
}

// Analisar percorre o texto uma vez por métrica. Cada laço é curto e
// independente: dá para ler um de cada vez.
func Analisar(texto string, ignorarVazias bool) Estatisticas {
	e := Estatisticas{Frequencia: map[string]int{}, Forma: map[string]string{}}
	e.Runas = utf8.RuneCountInString(texto)
	for _, r := range texto {
		if !unicode.IsSpace(r) {
			e.SemEspacos++
		}
	}
	emParagrafo := false
	for _, linha := range strings.Split(strings.TrimRight(texto, "\r\n"), "\n") {
		e.Linhas++
		vazia := strings.TrimSpace(linha) == ""
		if !vazia && !emParagrafo {
			e.Paragrafos++
		}
		emParagrafo = !vazia
	}
	// Uma frase termina em ponto, exclamação ou interrogação; "..." conta uma vez.
	anteriorFechou, temConteudo := false, false
	for _, r := range texto {
		fecha := strings.ContainsRune(".!?", r)
		if fecha && !anteriorFechou && temConteudo {
			e.Frases++
			temConteudo = false
		} else if !fecha && !unicode.IsSpace(r) {
			temConteudo = true
		}
		anteriorFechou = fecha
	}
	if temConteudo {
		e.Frases++ // texto que termina sem pontuação ainda é uma frase
	}
	for _, token := range strings.Fields(texto) {
		palavra := normalizar(token)
		if palavra == "" {
			continue // só pontuação, como um travessão solto
		}
		e.Palavras++
		if _, vista := e.Forma[palavra]; !vista {
			// guarda a grafia original (com acento) para o ranking ficar legível
			e.Forma[palavra] = strings.ToLower(strings.TrimFunc(token, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsDigit(r)
			}))
		}
		if utf8.RuneCountInString(palavra) > utf8.RuneCountInString(e.MaiorPalavra) {
			e.MaiorPalavra = e.Forma[palavra]
		}
		if !(ignorarVazias && vazias[palavra]) {
			e.Frequencia[palavra]++
		}
	}
	return e
}

// Top devolve as n palavras mais frequentes. O empate é desfeito pela ordem
// alfabética, para o resultado ser sempre o mesmo.
func (e Estatisticas) Top(n int) []Contagem {
	lista := make([]Contagem, 0, len(e.Frequencia))
	for palavra, vezes := range e.Frequencia {
		lista = append(lista, Contagem{e.Forma[palavra], vezes})
	}
	sort.Slice(lista, func(i, j int) bool {
		if lista[i].Vezes != lista[j].Vezes {
			return lista[i].Vezes > lista[j].Vezes
		}
		return lista[i].Palavra < lista[j].Palavra
	})
	return lista[:min(n, len(lista))]
}

func imprimir(e Estatisticas, top int) {
	fmt.Printf("  Letras e símbolos: %d (sem espaços: %d)\n", e.Runas, e.SemEspacos)
	fmt.Printf("  Palavras: %d   Linhas: %d   Frases: %d   Parágrafos: %d\n", e.Palavras, e.Linhas, e.Frases, e.Paragrafos)
	fmt.Printf("  Maior palavra: %q   Palavras distintas: %d\n", e.MaiorPalavra, len(e.Frequencia))
	segundos := e.Palavras * 60 / 200 // leitura média de 200 palavras por minuto
	fmt.Printf("  Tempo de leitura: %d min %d s\n", segundos/60, segundos%60)
	ranking := e.Top(top)
	if len(ranking) == 0 {
		return
	}
	fmt.Printf("  Top %d palavras:\n", len(ranking))
	maior := ranking[0].Vezes
	for i, c := range ranking {
		fmt.Printf("  %2d. %-16s %3d  %s\n", i+1, c.Palavra, c.Vezes, strings.Repeat("#", c.Vezes*30/maior))
	}
}

const textoDemo = `Aprender a programar é parecido com aprender um instrumento. Ninguém toca bem
no primeiro dia: a mão erra, o ritmo escapa e a música sai torta. Mesmo assim, quem
pratica todo dia percebe que o erro de ontem já não aparece hoje.

A lógica de programação funciona do mesmo jeito. Cada laço, cada condição e cada
função que você escreve treina o olhar para padrões. Com o tempo, o problema que
parecia um muro vira uma escada: um degrau de cada vez, uma função de cada vez.

Os projetos pequenos são o melhor laboratório. Uma calculadora, um contador de
palavras, um jogo da forca: nenhum deles muda o mundo, mas cada um ensina uma peça
que você vai reutilizar depois. O segredo é terminar! Um projeto terminado vale mais
do que dez pela metade.

Por isso, escolha um problema pequeno, escreva o código, rode, quebre e conserte.
Depois leia o código de outra pessoa e compare. É assim que a prática vira fluência,
e a fluência vira confiança.`

func demo() {
	fmt.Println("=== Contador de palavras e frequência: demonstração ===")
	fmt.Println("Texto analisado: 4 parágrafos sobre aprender a programar (embutido no programa).")
	fmt.Println("1) Estatísticas e top 10 ignorando palavras vazias:")
	imprimir(Analisar(textoDemo, true), 10)
	fmt.Println("2) As 5 mais frequentes quando as palavras vazias entram na conta:")
	for i, c := range Analisar(textoDemo, false).Top(5) {
		fmt.Printf("  %d. %-8s %d\n", i+1, c.Palavra, c.Vezes)
	}
}

func uso() {
	fmt.Println("Contador de palavras e frequência (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/14_contador_palavras --arquivo redacao.txt --top 15")
	fmt.Println("  go run ./sources/14_contador_palavras --texto \"Go é simples. Go é rápido.\"")
	fmt.Println("  go run ./sources/14_contador_palavras --demo")
	flag.PrintDefaults()
}

func main() {
	texto := flag.String("texto", "", "texto passado direto na linha de comando")
	arquivo := flag.String("arquivo", "", "caminho de um arquivo de texto (UTF-8)")
	top := flag.Int("top", 10, "quantas palavras mostrar no ranking")
	comVazias := flag.Bool("com-vazias", false, "contar também artigos, preposições e pronomes")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	switch {
	case *demoFlag:
		demo()
	case *arquivo != "":
		conteudo, err := os.ReadFile(*arquivo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro: não consegui ler o arquivo %q: %v\n", *arquivo, err)
			os.Exit(1)
		}
		fmt.Printf("Arquivo %s:\n", *arquivo)
		imprimir(Analisar(string(conteudo), !*comVazias), *top)
	case *texto != "":
		imprimir(Analisar(*texto, !*comVazias), *top)
	default:
		uso()
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Leia o texto da entrada padrão (os.Stdin) quando vier de um pipe,
//    para encadear com outros comandos.
// 2. "Sr. Silva" conta como duas frases; trate abreviações comuns.
// 3. Mostre também os pares de palavras mais frequentes (bigramas).
// ------------------------------------------------------------
