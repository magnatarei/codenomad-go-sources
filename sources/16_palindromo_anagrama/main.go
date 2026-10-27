// ============================================================
// CodeNomad Academy  |  Sources Go  |  16. Palíndromos e anagramas
// ------------------------------------------------------------
// O que faz: verifica se uma frase é palíndromo ignorando acentos, espaços
//            e pontuação; verifica se duas frases são anagramas uma da outra;
//            e agrupa uma lista de palavras pelos anagramas que formam.
// Como usar: go run ./sources/16_palindromo_anagrama --palindromo "Roma é amor"
//            go run ./sources/16_palindromo_anagrama --anagrama "pedra" --com "padre"
//            go run ./sources/16_palindromo_anagrama --agrupar "amor,roma,ramo,gato,toga"
//            go run ./sources/16_palindromo_anagrama --demo
// O que você aprende: runas contra bytes, strings.Map, unicode, sort.Slice
//            em runas, maps que agrupam por chave e ordenação de resultados.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

// acentos mapeia cada letra acentuada para a letra simples equivalente.
// Assim "é" e "e" contam como a mesma letra na comparação.
var acentos = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i', 'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u', 'ç': 'c', 'ñ': 'n',
}

// normalizar devolve só letras e algarismos, em minúsculas e sem acento.
// É a base das três operações: tudo o que não é letra simplesmente some.
func normalizar(texto string) []rune {
	limpo := strings.Map(func(r rune) rune {
		r = unicode.ToLower(r)
		if simples, ok := acentos[r]; ok {
			r = simples
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return -1 // devolver -1 no strings.Map apaga a runa
	}, texto)
	// Converter para []rune é obrigatório: indexar a string daria bytes,
	// e uma letra como "ç" ocupa dois bytes em UTF-8.
	return []rune(limpo)
}

// EhPalindromo compara a frase de fora para dentro, com dois índices.
func EhPalindromo(frase string) (bool, error) {
	letras := normalizar(frase)
	if len(letras) == 0 {
		return false, errors.New("a frase não tem nenhuma letra ou algarismo")
	}
	for i, j := 0, len(letras)-1; i < j; i, j = i+1, j-1 {
		if letras[i] != letras[j] {
			return false, nil
		}
	}
	return true, nil
}

// assinatura ordena as letras normalizadas: anagramas têm a mesma assinatura.
// "amor" e "roma" viram as duas "amor".
func assinatura(texto string) string {
	letras := normalizar(texto)
	sort.Slice(letras, func(i, j int) bool { return letras[i] < letras[j] })
	return string(letras)
}

// SaoAnagramas diz se as duas frases usam exatamente as mesmas letras.
func SaoAnagramas(a, b string) (bool, error) {
	sa, sb := assinatura(a), assinatura(b)
	if sa == "" || sb == "" {
		return false, errors.New("as duas frases precisam ter pelo menos uma letra")
	}
	return sa == sb, nil
}

// Grupo reúne palavras que são anagramas entre si.
type Grupo struct {
	Assinatura string
	Palavras   []string
}

// Agrupar coloca cada palavra no balde da sua assinatura e devolve os
// grupos ordenados: os maiores primeiro e, no empate, por ordem alfabética.
func Agrupar(palavras []string) []Grupo {
	baldes := map[string][]string{}
	for _, p := range palavras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		chave := assinatura(p)
		baldes[chave] = append(baldes[chave], p)
	}
	grupos := make([]Grupo, 0, len(baldes))
	for chave, lista := range baldes {
		sort.Strings(lista)
		grupos = append(grupos, Grupo{chave, lista})
	}
	// A ordem de um map muda a cada execução; ordenar deixa a saída estável.
	sort.Slice(grupos, func(i, j int) bool {
		if len(grupos[i].Palavras) != len(grupos[j].Palavras) {
			return len(grupos[i].Palavras) > len(grupos[j].Palavras)
		}
		return grupos[i].Palavras[0] < grupos[j].Palavras[0]
	})
	return grupos
}

func relatarPalindromo(frase string) error {
	ok, err := EhPalindromo(frase)
	if err != nil {
		return err
	}
	veredito := "não é palíndromo"
	if ok {
		veredito = "é palíndromo"
	}
	fmt.Printf("  %-42q %s  (lido como %q)\n", frase, veredito, string(normalizar(frase)))
	return nil
}

func relatarAnagrama(a, b string) error {
	ok, err := SaoAnagramas(a, b)
	if err != nil {
		return err
	}
	veredito := "não são anagramas"
	if ok {
		veredito = "são anagramas"
	}
	fmt.Printf("  %-14q e %-14q %s\n", a, b, veredito)
	return nil
}

func relatarGrupos(palavras []string) {
	for _, g := range Agrupar(palavras) {
		if len(g.Palavras) == 1 {
			fmt.Printf("  [%s] %s  (sem par)\n", g.Assinatura, g.Palavras[0])
			continue
		}
		fmt.Printf("  [%s] %s\n", g.Assinatura, strings.Join(g.Palavras, ", "))
	}
}

func demo() {
	fmt.Println("=== Palíndromos e anagramas: demonstração ===")
	fmt.Println("1) Frases famosas (acentos, espaços e pontuação são ignorados):")
	frases := []string{
		"Roma é amor", "Socorram-me, subi no ônibus em Marrocos!", "A base do teto desaba",
		"Anotaram a data da maratona", "Luz azul", "Go é ótimo", "12321", "Aprender a programar",
	}
	for _, f := range frases {
		if err := relatarPalindromo(f); err != nil {
			fmt.Println("  erro:", err)
		}
	}
	fmt.Println("\n2) Pares de palavras e frases:")
	pares := [][2]string{
		{"pedra", "padre"}, {"amor", "Roma"}, {"alegria", "galeria"}, {"gato", "pato"},
		{"Dormitório", "Dormir ótio"}, {"escola", "sacole"},
	}
	for _, p := range pares {
		if err := relatarAnagrama(p[0], p[1]); err != nil {
			fmt.Println("  erro:", err)
		}
	}
	fmt.Println("\n3) Agrupando uma lista pelos anagramas:")
	lista := strings.Fields("amor roma mora ramo pedra padre ator rota tora gato toga pato gelo logo casa")
	fmt.Println("  lista:", strings.Join(lista, " "))
	relatarGrupos(lista)
}

func uso() {
	fmt.Println("Palíndromos e anagramas (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/16_palindromo_anagrama --palindromo \"Roma é amor\"")
	fmt.Println("  go run ./sources/16_palindromo_anagrama --anagrama \"pedra\" --com \"padre\"")
	fmt.Println("  go run ./sources/16_palindromo_anagrama --agrupar \"amor,roma,ramo,gato,toga\"")
	fmt.Println("  go run ./sources/16_palindromo_anagrama --demo")
	flag.PrintDefaults()
}

func main() {
	palindromo := flag.String("palindromo", "", "frase para testar se é palíndromo")
	anagrama := flag.String("anagrama", "", "primeira frase do par de anagramas (use junto com --com)")
	com := flag.String("com", "", "segunda frase do par de anagramas")
	agrupar := flag.String("agrupar", "", "palavras separadas por vírgula para agrupar")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	var err error
	switch {
	case *demoFlag:
		demo()
	case *palindromo != "":
		err = relatarPalindromo(*palindromo)
	case *anagrama != "" || *com != "":
		if *anagrama == "" || *com == "" {
			err = errors.New("informe as duas frases: --anagrama \"uma\" --com \"outra\"")
		} else {
			err = relatarAnagrama(*anagrama, *com)
		}
	case *agrupar != "":
		relatarGrupos(strings.Split(*agrupar, ","))
	default:
		uso()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Leia uma lista de palavras de um arquivo (uma por linha) e mostre
//    só os grupos com dois ou mais anagramas.
// 2. Encontre o maior palíndromo contido dentro de uma frase qualquer.
// 3. Troque o sort.Slice da assinatura por um map de contagem de letras
//    e compare qual dos dois é mais rápido com 100 mil palavras.
// ------------------------------------------------------------
