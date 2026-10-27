// ============================================================
// CodeNomad Academy  |  Sources Go  |  15. Cifra de César e de Vigenère
// ------------------------------------------------------------
// O que faz: codifica e decodifica texto com a cifra de César (deslocamento)
//            e a de Vigenère (palavra-chave), preservando maiúsculas, espaços
//            e pontuação, e quebra um texto de César sem saber a chave,
//            comparando a frequência das letras com a do português.
// Como usar: go run ./sources/15_cifra_cesar_vigenere --chave 3 --texto "Olá, mundo!"
//            go run ./sources/15_cifra_cesar_vigenere --metodo vigenere --acao decifrar --chave LIMAO --texto "Ztá, yubow!"
//            go run ./sources/15_cifra_cesar_vigenere --acao quebrar --texto "Vsá, tbukv!"
//            go run ./sources/15_cifra_cesar_vigenere --demo
// O que você aprende: aritmética com runas, módulo com negativos, strings.Map,
//            strings.Builder, maps de frequência, sort.Slice e qui-quadrado.
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
	"strconv"
	"strings"
)

// deslocar move uma letra A-Z ou a-z k posições no alfabeto, mantendo a caixa.
// O ((k % 26) + 26) % 26 garante resultado positivo mesmo com k negativo.
// Qualquer outro símbolo (espaço, vírgula, letra acentuada) volta igual.
func deslocar(r rune, k int) rune {
	k = ((k % 26) + 26) % 26
	switch {
	case r >= 'a' && r <= 'z':
		return 'a' + (r-'a'+rune(k))%26
	case r >= 'A' && r <= 'Z':
		return 'A' + (r-'A'+rune(k))%26
	}
	return r
}

func ehLetra(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }

// Cesar cifra com deslocamento k; para decifrar, basta usar -k.
func Cesar(texto string, k int) string {
	return strings.Map(func(r rune) rune { return deslocar(r, k) }, texto)
}

// Vigenere usa uma letra da chave por letra do texto, em ciclo. A chave só
// avança nas letras: espaços e pontuação não consomem posição.
func Vigenere(texto, chave string, decifrar bool) (string, error) {
	chave = strings.ToUpper(strings.TrimSpace(chave))
	if chave == "" {
		return "", errors.New("a chave de Vigenère não pode ser vazia")
	}
	for _, r := range chave {
		if !ehLetra(r) {
			return "", fmt.Errorf("a chave só pode ter letras de A a Z, achei %q", r)
		}
	}
	var sb strings.Builder
	i := 0
	for _, r := range texto {
		if ehLetra(r) {
			k := int(chave[i%len(chave)] - 'A')
			if decifrar {
				k = -k
			}
			r = deslocar(r, k)
			i++
		}
		sb.WriteRune(r)
	}
	return sb.String(), nil
}

// frequenciaPT é a porcentagem média de cada letra em textos em português.
var frequenciaPT = map[rune]float64{
	'a': 14.63, 'b': 1.04, 'c': 3.88, 'd': 4.99, 'e': 12.57, 'f': 1.02, 'g': 1.30, 'h': 1.28, 'i': 6.18,
	'j': 0.40, 'k': 0.02, 'l': 2.78, 'm': 4.74, 'n': 5.05, 'o': 10.73, 'p': 2.52, 'q': 1.20, 'r': 6.53,
	's': 7.81, 't': 4.34, 'u': 4.63, 'v': 1.67, 'w': 0.01, 'x': 0.21, 'y': 0.01, 'z': 0.47,
}

// Candidato é uma tentativa de quebra: a chave testada, o texto resultante e
// a distância (qui-quadrado) até a frequência do português. Menor é melhor.
type Candidato struct {
	Chave     int
	Distancia float64
	Texto     string
}

// quiQuadrado mede quanto a contagem de letras se afasta da esperada:
// soma de (observado - esperado)^2 / esperado, letra por letra.
func quiQuadrado(texto string) float64 {
	contagem := map[rune]int{}
	total := 0
	for _, r := range strings.ToLower(texto) {
		if ehLetra(r) {
			contagem[r]++
			total++
		}
	}
	if total == 0 {
		return 0
	}
	soma := 0.0
	for letra, pct := range frequenciaPT {
		esperado := float64(total) * pct / 100
		diferenca := float64(contagem[letra]) - esperado
		soma += diferenca * diferenca / esperado
	}
	return soma
}

// QuebrarCesar testa os 26 deslocamentos e ordena do mais provável ao menos.
func QuebrarCesar(cifrado string) []Candidato {
	candidatos := make([]Candidato, 0, 26)
	for k := 0; k < 26; k++ {
		texto := Cesar(cifrado, -k)
		candidatos = append(candidatos, Candidato{k, quiQuadrado(texto), texto})
	}
	sort.Slice(candidatos, func(i, j int) bool { return candidatos[i].Distancia < candidatos[j].Distancia })
	return candidatos
}

// executar traduz as flags em uma chamada e devolve o texto pronto para imprimir.
func executar(metodo, acao, chave, texto string) (string, error) {
	switch {
	case acao == "quebrar":
		if metodo != "cesar" {
			return "", errors.New("o quebrador por frequência só funciona para César")
		}
		var sb strings.Builder
		for i, c := range QuebrarCesar(texto)[:3] {
			fmt.Fprintf(&sb, "candidato %d: chave %2d  distância %7.1f  %s\n", i+1, c.Chave, c.Distancia, c.Texto)
		}
		return strings.TrimRight(sb.String(), "\n"), nil
	case acao != "cifrar" && acao != "decifrar":
		return "", fmt.Errorf("ação desconhecida %q (use cifrar, decifrar ou quebrar)", acao)
	case metodo == "cesar":
		k, err := strconv.Atoi(chave)
		if err != nil {
			return "", fmt.Errorf("a chave de César precisa ser um inteiro, recebi %q", chave)
		}
		if acao == "decifrar" {
			k = -k
		}
		return Cesar(texto, k), nil
	case metodo == "vigenere":
		return Vigenere(texto, chave, acao == "decifrar")
	}
	return "", fmt.Errorf("método desconhecido %q (use cesar ou vigenere)", metodo)
}

func demo() {
	fmt.Println("=== Cifra de César e de Vigenère: demonstração ===")
	frase := "Quem estuda todo dia vence o jogo da lógica!"
	fmt.Println("Frase original:          ", frase)
	cifrado := Cesar(frase, 3)
	fmt.Println("1) César, chave 3:       ", cifrado)
	fmt.Println("   decifrado com -3:     ", Cesar(cifrado, -3))
	v, err := Vigenere(frase, "LIMAO", false)
	if err != nil {
		fmt.Println("erro:", err)
		return
	}
	d, _ := Vigenere(v, "LIMAO", true) // a chave já foi validada acima
	fmt.Println("2) Vigenère, chave LIMAO:", v)
	fmt.Println("   decifrado:            ", d)
	fmt.Println("   Repare: acento, espaço e pontuação passam intactos; só A-Z muda.")
	fmt.Println("3) Quebrando um César sem saber a chave (chave real: 7):")
	segredo := Cesar("A prática diária de programação ensina mais do que qualquer teoria: "+
		"cada erro corrigido deixa o aluno mais perto da fluência.", 7)
	fmt.Println("   cifrado:", segredo)
	for i, c := range QuebrarCesar(segredo)[:3] {
		fmt.Printf("   candidato %d: chave %2d  distância %7.1f  %s\n", i+1, c.Chave, c.Distancia, c.Texto)
	}
	fmt.Println("   O qui-quadrado menor aponta a chave certa: as letras voltam à frequência do português.")
}

func uso() {
	fmt.Println("Cifra de César e de Vigenère (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/15_cifra_cesar_vigenere --chave 3 --texto \"Olá, mundo!\"")
	fmt.Println("  go run ./sources/15_cifra_cesar_vigenere --metodo vigenere --acao decifrar --chave LIMAO --texto \"Ztá, yubow!\"")
	fmt.Println("  go run ./sources/15_cifra_cesar_vigenere --acao quebrar --texto \"Vsá, tbukv!\"")
	fmt.Println("  go run ./sources/15_cifra_cesar_vigenere --demo")
	flag.PrintDefaults()
}

func main() {
	metodo := flag.String("metodo", "cesar", "cesar ou vigenere")
	acao := flag.String("acao", "cifrar", "cifrar, decifrar ou quebrar (quebrar só para César)")
	chave := flag.String("chave", "", "deslocamento (César, ex.: 3) ou palavra (Vigenère, ex.: LIMAO)")
	texto := flag.String("texto", "", "texto de entrada")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	if *texto == "" {
		uso()
		return
	}
	resultado, err := executar(strings.ToLower(*metodo), strings.ToLower(*acao), *chave, *texto)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
	fmt.Println(resultado)
}

// ------------------------------------------------------------
// Para ir além
// 1. Troque as letras acentuadas pelas simples antes de cifrar, para o
//    texto cifrado não entregar pistas.
// 2. Quebre Vigenère: descubra o tamanho da chave pelo índice de
//    coincidência e aplique o quebrador de César em cada coluna.
// 3. Leia o texto de um arquivo e grave o resultado em outro.
// ------------------------------------------------------------
