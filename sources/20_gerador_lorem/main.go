// ============================================================
// CodeNomad Academy  |  Sources Go  |  20. Gerador de texto de preenchimento
// ------------------------------------------------------------
// O que faz: gera parágrafos, frases ou palavras de texto de preenchimento
//            em latim clássico ("lorem ipsum") ou em português, com a
//            quantidade à escolha e uma semente opcional para repetir o resultado.
// Como usar: go run ./sources/20_gerador_lorem --paragrafos 3
//            go run ./sources/20_gerador_lorem --frases 5 --idioma pt --semente 42
//            go run ./sources/20_gerador_lorem --palavras 20 --idioma latim
//            go run ./sources/20_gerador_lorem --demo
// O que você aprende: math/rand/v2 com semente fixa (rand.NewPCG),
//            strings.Builder, slices de palavras, funções que recebem o
//            gerador como parâmetro, e o verbo %q do fmt.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"
	"unicode"
)

// Cada idioma é só uma lista de palavras: o algoritmo é o mesmo para os dois.
var vocabulario = map[string][]string{
	"latim": strings.Fields(`lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod
		tempor incididunt ut labore et dolore magna aliqua enim ad minim veniam quis nostrud
		exercitation ullamco laboris nisi aliquip ex ea commodo consequat duis aute irure in
		reprehenderit voluptate velit esse cillum fugiat nulla pariatur excepteur sint occaecat
		cupidatat non proident sunt culpa qui officia deserunt mollit anim id est laborum`),
	"pt": strings.Fields(`a casa o rio uma janela aberta o vento leve passa pela praça e leva
		folhas até a esquina da rua antiga onde um café serve pão quente de manhã enquanto o
		sol sobe devagar sobre as montanhas e a cidade acorda com som de bicicleta mercado
		conversa risada papel caneta livro escola aluno professor projeto código função
		número texto página mesa cadeira luz sombra caminho ponte barco mar areia pedra
		jardim flor chuva nuvem estrela noite dia tempo silêncio música voz passo porta`),
}

// Gerador carrega o vocabulário e a fonte de aleatoriedade. Receber o
// gerador em um campo (e não usar o global) é o que permite repetir a saída.
type Gerador struct {
	Palavras []string
	Fonte    *rand.Rand
}

// NovoGerador cria o gerador. Semente 0 usa o relógio, qualquer outra
// semente produz sempre o mesmo texto.
func NovoGerador(idioma string, semente uint64) (*Gerador, error) {
	palavras, ok := vocabulario[idioma]
	if !ok {
		return nil, fmt.Errorf("idioma desconhecido %q (use latim ou pt)", idioma)
	}
	if semente == 0 {
		semente = uint64(time.Now().UnixNano())
	}
	// PCG precisa de duas sementes; usar a mesma nas duas é suficiente aqui.
	fonte := rand.New(rand.NewPCG(semente, semente))
	return &Gerador{Palavras: palavras, Fonte: fonte}, nil
}

// entre sorteia um inteiro no intervalo [menor, maior].
func (g *Gerador) entre(menor, maior int) int {
	return menor + g.Fonte.IntN(maior-menor+1)
}

// Palavra sorteia uma palavra do vocabulário.
func (g *Gerador) Palavra() string {
	return g.Palavras[g.Fonte.IntN(len(g.Palavras))]
}

// ListaPalavras devolve n palavras separadas por espaço.
func (g *Gerador) ListaPalavras(n int) string {
	lista := make([]string, n)
	for i := range lista {
		lista[i] = g.Palavra()
	}
	return strings.Join(lista, " ")
}

// capitalizar põe a primeira letra em maiúscula respeitando acentos:
// por isso o texto vira []rune antes de mexer na primeira posição.
func capitalizar(s string) string {
	runas := []rune(s)
	if len(runas) == 0 {
		return s
	}
	runas[0] = unicode.ToUpper(runas[0])
	return string(runas)
}

// Frase monta de 6 a 14 palavras, às vezes com uma vírgula no meio,
// começa com maiúscula e termina com ponto.
func (g *Gerador) Frase() string {
	n := g.entre(6, 14)
	palavras := make([]string, n)
	for i := range palavras {
		palavras[i] = g.Palavra()
	}
	if n >= 9 && g.Fonte.IntN(2) == 0 {
		meio := g.entre(3, n-4)
		palavras[meio] += ","
	}
	return capitalizar(strings.Join(palavras, " ")) + "."
}

// Paragrafo junta de 4 a 7 frases. O primeiro parágrafo em latim começa
// com a abertura clássica, como manda a tradição dos modelos de impressão.
func (g *Gerador) Paragrafo(classico bool) string {
	frases := make([]string, 0, 8)
	if classico {
		frases = append(frases, "Lorem ipsum dolor sit amet, consectetur adipiscing elit.")
	}
	for i := g.entre(4, 7); i > 0; i-- {
		frases = append(frases, g.Frase())
	}
	return strings.Join(frases, " ")
}

// Paragrafos devolve n parágrafos separados por uma linha em branco.
func (g *Gerador) Paragrafos(n int, idioma string) string {
	lista := make([]string, n)
	for i := range lista {
		lista[i] = g.Paragrafo(i == 0 && idioma == "latim")
	}
	return strings.Join(lista, "\n\n")
}

// Frases devolve n frases em uma linha só.
func (g *Gerador) Frases(n int) string {
	lista := make([]string, n)
	for i := range lista {
		lista[i] = g.Frase()
	}
	return strings.Join(lista, " ")
}

func contarPalavras(texto string) int { return len(strings.Fields(texto)) }

func demo() {
	fmt.Println("=== Gerador de texto de preenchimento: demonstração ===")
	for i, idioma := range []string{"latim", "pt"} {
		g, err := NovoGerador(idioma, 2024)
		if err != nil {
			fmt.Println("erro:", err)
			return
		}
		texto := g.Paragrafos(2, idioma)
		fmt.Printf("%d) Dois parágrafos em %s (semente 2024, %d palavras):\n\n", i+1, idioma, contarPalavras(texto))
		fmt.Println(texto)
		fmt.Println()
	}
	g, _ := NovoGerador("pt", 7) // o idioma "pt" existe, o erro é impossível aqui
	fmt.Println("3) Doze palavras soltas em português (semente 7):")
	fmt.Println("  ", g.ListaPalavras(12))
	fmt.Println("4) A mesma semente repete o resultado:")
	a, _ := NovoGerador("latim", 99)
	b, _ := NovoGerador("latim", 99)
	fa, fb := a.Frase(), b.Frase()
	fmt.Printf("   gerador A: %q\n   gerador B: %q\n   iguais: %v\n", fa, fb, fa == fb)
}

func uso() {
	fmt.Println("Gerador de texto de preenchimento (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/20_gerador_lorem --paragrafos 3")
	fmt.Println("  go run ./sources/20_gerador_lorem --frases 5 --idioma pt --semente 42")
	fmt.Println("  go run ./sources/20_gerador_lorem --palavras 20 --idioma latim")
	fmt.Println("  go run ./sources/20_gerador_lorem --demo")
	flag.PrintDefaults()
}

func main() {
	paragrafos := flag.Int("paragrafos", 0, "quantidade de parágrafos")
	frases := flag.Int("frases", 0, "quantidade de frases")
	palavras := flag.Int("palavras", 0, "quantidade de palavras soltas")
	idioma := flag.String("idioma", "latim", "latim ou pt")
	semente := flag.Uint64("semente", 0, "semente para repetir o resultado (0 = aleatório)")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	if *paragrafos == 0 && *frases == 0 && *palavras == 0 {
		uso()
		return
	}
	for nome, v := range map[string]int{"paragrafos": *paragrafos, "frases": *frases, "palavras": *palavras} {
		if v < 0 || v > 1000 {
			fmt.Fprintf(os.Stderr, "Erro: --%s precisa estar entre 0 e 1000, recebi %d\n", nome, v)
			os.Exit(1)
		}
	}
	g, err := NovoGerador(strings.ToLower(*idioma), *semente)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
	switch {
	case *paragrafos > 0:
		fmt.Println(g.Paragrafos(*paragrafos, strings.ToLower(*idioma)))
	case *frases > 0:
		fmt.Println(g.Frases(*frases))
	default:
		fmt.Println(g.ListaPalavras(*palavras))
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Acrescente a flag --html para envolver cada parágrafo em <p> e </p>.
// 2. Gere títulos (3 a 6 palavras em Title Case) antes de cada parágrafo.
// 3. Carregue um vocabulário de um arquivo de texto, uma palavra por
//    linha, e permita qualquer idioma.
// ------------------------------------------------------------
