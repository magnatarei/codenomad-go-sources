// ============================================================
// CodeNomad Academy  |  Sources Go  |  17. Conversor de formatos de texto
// ------------------------------------------------------------
// O que faz: converte uma frase entre snake_case, camelCase, PascalCase,
//            kebab-case e Title Case; gera slug para URL; remove acentos;
//            inverte o texto e capitaliza a primeira letra.
// Como usar: go run ./sources/17_conversor_texto --texto "Relatório de vendas do mês" --para snake
//            go run ./sources/17_conversor_texto --texto "numeroDoPedido" --para kebab
//            go run ./sources/17_conversor_texto --texto "Ação rápida" --para todos
//            go run ./sources/17_conversor_texto --demo
// O que você aprende: quebrar texto em palavras com runas e unicode,
//            strings.Builder, map de nome para função, inverter um []rune.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// acentos mapeia cada letra acentuada para a letra simples equivalente.
var acentos = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i', 'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u', 'ç': 'c', 'ñ': 'n',
	'Á': 'A', 'À': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'É': 'E', 'È': 'E', 'Ê': 'E', 'Ë': 'E',
	'Í': 'I', 'Ì': 'I', 'Î': 'I', 'Ï': 'I', 'Ó': 'O', 'Ò': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O',
	'Ú': 'U', 'Ù': 'U', 'Û': 'U', 'Ü': 'U', 'Ç': 'C', 'Ñ': 'N',
}

// RemoverAcentos troca cada letra acentuada pela simples e mantém o resto.
func RemoverAcentos(texto string) string {
	return strings.Map(func(r rune) rune {
		if simples, ok := acentos[r]; ok {
			return simples
		}
		return r
	}, texto)
}

// Palavras quebra o texto em palavras. Separa em espaços e pontuação, mas
// também na fronteira de um camelCase: "numeroDoPedido" vira 3 palavras.
func Palavras(texto string) []string {
	var palavras []string
	var atual strings.Builder
	runas := []rune(texto)
	fechar := func() {
		if atual.Len() > 0 {
			palavras = append(palavras, atual.String())
			atual.Reset()
		}
	}
	for i, r := range runas {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			fechar() // espaço, hífen, sublinhado ou pontuação: fim da palavra
			continue
		}
		// Maiúscula depois de minúscula ("oP" em "numeroDoPedido") começa palavra nova.
		if i > 0 && unicode.IsUpper(r) && unicode.IsLower(runas[i-1]) {
			fechar()
		}
		atual.WriteRune(r)
	}
	fechar()
	return palavras
}

// titulo põe a primeira letra em maiúscula e o resto em minúsculas.
func titulo(p string) string {
	runas := []rune(strings.ToLower(p))
	if len(runas) == 0 {
		return ""
	}
	runas[0] = unicode.ToUpper(runas[0])
	return string(runas)
}

// Nas formas usadas em código (snake, camel, Pascal, kebab) e no slug, os
// acentos saem: identificadores e URLs ficam só com letras simples.
func SnakeCase(texto string) string {
	return strings.ToLower(RemoverAcentos(strings.Join(Palavras(texto), "_")))
}

func KebabCase(texto string) string {
	return strings.ToLower(RemoverAcentos(strings.Join(Palavras(texto), "-")))
}

func PascalCase(texto string) string {
	var sb strings.Builder
	for _, p := range Palavras(RemoverAcentos(texto)) {
		sb.WriteString(titulo(p))
	}
	return sb.String()
}

func CamelCase(texto string) string {
	pascal := []rune(PascalCase(texto))
	if len(pascal) == 0 {
		return ""
	}
	pascal[0] = unicode.ToLower(pascal[0])
	return string(pascal)
}

// TitleCase mantém os acentos: é para gente ler, não para código.
func TitleCase(texto string) string {
	palavras := Palavras(texto)
	for i, p := range palavras {
		palavras[i] = titulo(p)
	}
	return strings.Join(palavras, " ")
}

// Slug é o kebab-case de URL: só letras simples, algarismos e hífens.
func Slug(texto string) string {
	return KebabCase(texto)
}

// Inverter precisa de []rune: inverter os bytes quebraria letras acentuadas.
func Inverter(texto string) string {
	runas := []rune(texto)
	for i, j := 0, len(runas)-1; i < j; i, j = i+1, j-1 {
		runas[i], runas[j] = runas[j], runas[i]
	}
	return string(runas)
}

// Capitalizar deixa só a primeira letra do texto em maiúscula.
func Capitalizar(texto string) string {
	runas := []rune(strings.ToLower(strings.TrimSpace(texto)))
	for i, r := range runas {
		if unicode.IsLetter(r) {
			runas[i] = unicode.ToUpper(r)
			break
		}
	}
	return string(runas)
}

// Formato liga o nome usado na linha de comando à função que converte.
type Formato struct {
	Nome      string
	Converter func(string) string
}

// formatos é um slice (e não um map) para a ordem de impressão ser fixa.
var formatos = []Formato{
	{"snake", SnakeCase}, {"camel", CamelCase}, {"pascal", PascalCase}, {"kebab", KebabCase},
	{"title", TitleCase}, {"slug", Slug}, {"sem-acentos", RemoverAcentos},
	{"inverter", Inverter}, {"capitalizar", Capitalizar},
}

func buscarFormato(nome string) (Formato, bool) {
	for _, f := range formatos {
		if f.Nome == nome {
			return f, true
		}
	}
	return Formato{}, false
}

func nomesFormatos() string {
	nomes := make([]string, 0, len(formatos)+1)
	for _, f := range formatos {
		nomes = append(nomes, f.Nome)
	}
	return strings.Join(append(nomes, "todos"), ", ")
}

func mostrarTodos(texto string) {
	fmt.Printf("Texto: %q\n", texto)
	for _, f := range formatos {
		fmt.Printf("  %-12s %s\n", f.Nome+":", f.Converter(texto))
	}
}

func demo() {
	fmt.Println("=== Conversor de formatos de texto: demonstração ===")
	frases := []string{"Relatório de vendas do mês", "Ação rápida é a solução", "númeroDoPedido já confirmado"}
	for i, frase := range frases {
		fmt.Printf("%d) ", i+1)
		mostrarTodos(frase)
	}
	fmt.Println("Repare: nas formas de código e no slug os acentos saem; em Title Case eles ficam.")
}

func uso() {
	fmt.Println("Conversor de formatos de texto (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/17_conversor_texto --texto \"Relatório de vendas do mês\" --para snake")
	fmt.Println("  go run ./sources/17_conversor_texto --texto \"numeroDoPedido\" --para kebab")
	fmt.Println("  go run ./sources/17_conversor_texto --texto \"Ação rápida\" --para todos")
	fmt.Println("  go run ./sources/17_conversor_texto --demo")
	fmt.Println("Formatos:", nomesFormatos())
	flag.PrintDefaults()
}

func main() {
	texto := flag.String("texto", "", "texto de entrada")
	para := flag.String("para", "todos", "formato de saída: "+nomesFormatos())
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	if strings.TrimSpace(*texto) == "" {
		uso()
		return
	}
	nome := strings.ToLower(*para)
	if nome == "todos" {
		mostrarTodos(*texto)
		return
	}
	f, ok := buscarFormato(nome)
	if !ok {
		fmt.Fprintf(os.Stderr, "Erro: formato desconhecido %q. Use um destes: %s\n", *para, nomesFormatos())
		os.Exit(1)
	}
	fmt.Println(f.Converter(*texto))
}

// ------------------------------------------------------------
// Para ir além
// 1. Trate siglas no camelCase: "clienteHTTPNovo" deve virar as palavras
//    cliente, HTTP e Novo (olhe a letra seguinte, não só a anterior).
// 2. Acrescente o formato CONSTANT_CASE e o "sentence case".
// 3. Leia várias linhas da entrada padrão e converta todas de uma vez.
// ------------------------------------------------------------
