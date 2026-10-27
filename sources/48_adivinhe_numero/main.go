// ============================================================
// CodeNomad Academy  |  Sources Go  |  48. Adivinhe o número, nos dois sentidos
// ------------------------------------------------------------
// O que faz: modo 1, você adivinha o número que o computador sorteou, com
//            dicas de "maior" ou "menor" e contagem de tentativas. Modo 2, o
//            computador adivinha o seu número por busca binária, e você só
//            responde se o palpite está acima, abaixo ou certo.
// Como usar: go run ./sources/48_adivinhe_numero --modo 1 --max 100
//            go run ./sources/48_adivinhe_numero --modo 2 --max 1000
//            go run ./sources/48_adivinhe_numero --demo
// O que você aprende: busca binária (metade do intervalo a cada passo),
//            math.Log2 para prever o número de palpites, bufio.Scanner e
//            strconv.Atoi com validação, math/rand com semente e funções
//            passadas como valor para trocar o teclado por uma resposta fixa.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// palpitesNecessarios é o teto teórico da busca binária: log2 do tamanho do
// intervalo, arredondado para cima. Para 100 números são 7 palpites.
func palpitesNecessarios(maximo int) int {
	return int(math.Ceil(math.Log2(float64(maximo))))
}

// modoVoceAdivinha sorteia e lê os palpites até acertar ou a entrada acabar.
func modoVoceAdivinha(maximo int, gerador *rand.Rand, leitor *bufio.Scanner) {
	segredo := gerador.Intn(maximo) + 1
	fmt.Printf("Pensei num número entre 1 e %d. Com busca binária você acerta em até %d palpites.\n", maximo, palpitesNecessarios(maximo))
	for tentativa := 1; ; tentativa++ {
		fmt.Printf("Palpite %d: ", tentativa)
		if !leitor.Scan() {
			fmt.Printf("\nEntrada encerrada. O número era %d.\n", segredo)
			return
		}
		n, err := strconv.Atoi(strings.TrimSpace(leitor.Text()))
		if err != nil || n < 1 || n > maximo {
			fmt.Printf("Digite um número inteiro entre 1 e %d.\n", maximo)
			tentativa-- // uma entrada inválida não conta como tentativa
			continue
		}
		switch {
		case n < segredo:
			fmt.Println("É maior.")
		case n > segredo:
			fmt.Println("É menor.")
		default:
			fmt.Printf("Acertou em %d tentativa(s)! O número era %d.\n", tentativa, segredo)
			return
		}
	}
}

// Resposta é o que o jogador diz sobre um palpite do computador.
type Resposta int

const (
	Certo Resposta = iota
	Maior
	Menor
	Desistiu
)

// modoComputadorAdivinha faz a busca binária. A função responder diz se o
// palpite está certo, se o número é maior ou menor: pode vir do teclado ou
// de um alvo fixo, e o algoritmo não sabe a diferença.
func modoComputadorAdivinha(maximo int, responder func(palpite int) Resposta) {
	fmt.Printf("Pense num número entre 1 e %d. Vou achar em até %d palpites.\n", maximo, palpitesNecessarios(maximo))
	baixo, alto := 1, maximo
	for tentativa := 1; baixo <= alto; tentativa++ {
		palpite := (baixo + alto) / 2 // o meio do intervalo que ainda é possível
		fmt.Printf("Palpite %d: %d  (intervalo possível: %d a %d)\n", tentativa, palpite, baixo, alto)
		switch responder(palpite) {
		case Certo:
			fmt.Printf("Acertei em %d palpite(s)!\n", tentativa)
			return
		case Maior:
			baixo = palpite + 1 // tudo até o palpite está descartado
		case Menor:
			alto = palpite - 1
		case Desistiu:
			fmt.Println("Entrada encerrada.")
			return
		}
	}
	// Se o intervalo ficou vazio, as respostas se contradizem.
	fmt.Println("Não sobrou nenhum número possível: alguma resposta foi trocada no meio do caminho.")
}

// respostaDoTeclado pede ao jogador uma letra: m (maior), n (menor) ou c (certo).
func respostaDoTeclado(leitor *bufio.Scanner) func(int) Resposta {
	return func(palpite int) Resposta {
		for {
			fmt.Printf("O seu número é maior (m), menor (n) ou é %d mesmo (c)? ", palpite)
			if !leitor.Scan() {
				return Desistiu
			}
			switch strings.ToLower(strings.TrimSpace(leitor.Text())) {
			case "m", "maior":
				return Maior
			case "n", "menor":
				return Menor
			case "c", "certo":
				return Certo
			}
			fmt.Println("Responda com m, n ou c.")
		}
	}
}

// respostaAutomatica compara o palpite com um alvo conhecido e explica.
func respostaAutomatica(alvo int) func(int) Resposta {
	return func(palpite int) Resposta {
		switch {
		case palpite < alvo:
			fmt.Println("  resposta: maior")
			return Maior
		case palpite > alvo:
			fmt.Println("  resposta: menor")
			return Menor
		}
		fmt.Println("  resposta: certo")
		return Certo
	}
}

func main() {
	modo := flag.Int("modo", 1, "1 = você adivinha, 2 = o computador adivinha")
	maximo := flag.Int("max", 100, "maior número possível (o menor é sempre 1)")
	alvo := flag.Int("alvo", 0, "no modo 2, responde sozinho como se o seu número fosse este (0 = pergunta no teclado)")
	semente := flag.Int64("semente", 0, "semente do sorteio no modo 1 (0 = aleatória)")
	demo := flag.Bool("demo", false, "o computador adivinha 73 entre 1 e 100, mostrando cada palpite")
	flag.Parse()

	if *demo {
		fmt.Println("=== CodeNomad Academy | Adivinhe o número: o computador adivinha 73 ===")
		fmt.Println("A cada palpite o intervalo possível cai pela metade: é a busca binária.")
		modoComputadorAdivinha(100, respostaAutomatica(73))
		fmt.Println()
		fmt.Printf("Para comparar: 1000 números precisam de só %d palpites, e um milhão, de %d.\n", palpitesNecessarios(1000), palpitesNecessarios(1000000))
		return
	}
	if *maximo < 2 {
		fmt.Fprintln(os.Stderr, "Erro: --max precisa ser pelo menos 2 (com um número só não há jogo).")
		os.Exit(1)
	}
	if *alvo < 0 || *alvo > *maximo {
		fmt.Fprintf(os.Stderr, "Erro: --alvo precisa estar entre 1 e %d.\n", *maximo)
		os.Exit(1)
	}
	gerador := rand.New(rand.NewSource(*semente))
	if *semente == 0 {
		gerador = rand.New(rand.NewSource(rand.Int63()))
	}
	leitor := bufio.NewScanner(os.Stdin)
	switch *modo {
	case 1:
		modoVoceAdivinha(*maximo, gerador, leitor)
	case 2:
		if *alvo > 0 {
			modoComputadorAdivinha(*maximo, respostaAutomatica(*alvo))
		} else {
			modoComputadorAdivinha(*maximo, respostaDoTeclado(leitor))
		}
	default:
		fmt.Fprintln(os.Stderr, "Erro: --modo precisa ser 1 ou 2.")
		os.Exit(1)
	}
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. No modo 1, mostre "quente" ou "frio" conforme a distância do palpite
//    ao segredo, além de maior e menor.
// 2. Registre o número de tentativas de cada partida num arquivo e mostre
//    o seu recorde ao começar.
// 3. No modo 2, troque a busca binária por palpites ao acaso dentro do
//    intervalo e meça, em 1000 partidas simuladas, a diferença de tentativas.
// ------------------------------------------------------------
