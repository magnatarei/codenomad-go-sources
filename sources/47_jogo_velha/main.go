// ============================================================
// CodeNomad Academy  |  Sources Go  |  47. Jogo da velha com computador
// ------------------------------------------------------------
// O que faz: jogo da velha no terminal para dois jogadores ou contra o
//            computador, que usa o algoritmo minimax e por isso nunca perde.
//            Mostra o tabuleiro a cada jogada e detecta vitória e empate.
// Como usar: go run ./sources/47_jogo_velha --modo cpu --comeca X
//            go run ./sources/47_jogo_velha --modo 2j
//            go run ./sources/47_jogo_velha --demo
// O que você aprende: array de tamanho fixo como tabuleiro, recursão com o
//            minimax (explorar todas as jogadas futuras e escolher a melhor),
//            funções que recebem funções para separar quem decide a jogada,
//            bufio.Scanner e strconv.Atoi com validação da entrada.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Tabuleiro tem 9 casas; cada uma guarda 'X', 'O' ou ' ' (vazia). Um array,
// e não um slice, porque o tamanho nunca muda e a cópia é barata: o minimax
// copia o tabuleiro a cada jogada simulada sem medo de alterar o original.
type Tabuleiro [9]byte

// linhas são as oito combinações que dão vitória: três horizontais, três
// verticais e duas diagonais.
var linhas = [8][3]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {0, 3, 6}, {1, 4, 7}, {2, 5, 8}, {0, 4, 8}, {2, 4, 6}}

func novoTabuleiro() Tabuleiro {
	var t Tabuleiro
	for i := range t {
		t[i] = ' '
	}
	return t
}

// vencedor devolve 'X' ou 'O' se alguém fechou uma linha, ou ' ' se não.
func (t Tabuleiro) vencedor() byte {
	for _, l := range linhas {
		if t[l[0]] != ' ' && t[l[0]] == t[l[1]] && t[l[1]] == t[l[2]] {
			return t[l[0]]
		}
	}
	return ' '
}

func (t Tabuleiro) cheio() bool {
	for _, c := range t {
		if c == ' ' {
			return false
		}
	}
	return true
}

// desenhar mostra as casas vazias com o seu número, para o jogador escolher.
func (t Tabuleiro) desenhar() string {
	var b strings.Builder
	for i, c := range t {
		if c == ' ' {
			b.WriteString(strconv.Itoa(i + 1))
		} else {
			b.WriteByte(c)
		}
		switch {
		case i%3 < 2:
			b.WriteString(" | ")
		case i < 8:
			b.WriteString("\n---+---+---\n")
		}
	}
	return b.String()
}

func adversario(j byte) byte {
	if j == 'X' {
		return 'O'
	}
	return 'X'
}

// minimax avalia o tabuleiro do ponto de vista de quem vai jogar: +10 se
// ele pode vencer, -10 se vai perder, 0 se o melhor resultado é empate.
// A profundidade entra na conta para preferir vencer logo e perder tarde.
func minimax(t Tabuleiro, vez byte, profundidade int) int {
	if v := t.vencedor(); v != ' ' {
		if v == vez {
			return 10 - profundidade
		}
		return profundidade - 10
	}
	if t.cheio() {
		return 0
	}
	melhor := -100
	for i := range t {
		if t[i] != ' ' {
			continue
		}
		t[i] = vez // t é uma cópia, então a marca some ao sair da função
		// O que é bom para o adversário é ruim para nós: por isso o sinal inverte.
		pontos := -minimax(t, adversario(vez), profundidade+1)
		t[i] = ' '
		if pontos > melhor {
			melhor = pontos
		}
	}
	return melhor
}

// melhorJogada testa cada casa livre com o minimax e devolve a de maior valor.
func melhorJogada(t Tabuleiro, vez byte) int {
	escolha, melhor := -1, -100
	for i := range t {
		if t[i] != ' ' {
			continue
		}
		t[i] = vez
		pontos := -minimax(t, adversario(vez), 1)
		t[i] = ' '
		if pontos > melhor {
			escolha, melhor = i, pontos
		}
	}
	return escolha
}

// jogadorComputador decide pelo minimax e explica a escolha.
func jogadorComputador(t Tabuleiro, vez byte) (int, bool) {
	casa := melhorJogada(t, vez)
	fmt.Printf("Computador (%c) joga na casa %d.\n", vez, casa+1)
	return casa, true
}

// jogadorHumano devolve uma função que lê a casa no teclado e valida.
func jogadorHumano() func(Tabuleiro, byte) (int, bool) {
	scanner := bufio.NewScanner(os.Stdin)
	return func(t Tabuleiro, vez byte) (int, bool) {
		for {
			fmt.Printf("Jogador %c, escolha uma casa (1 a 9): ", vez)
			if !scanner.Scan() {
				return -1, false // fim da entrada: encerra a partida
			}
			n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || n < 1 || n > 9 {
				fmt.Println("Digite um número de 1 a 9.")
				continue
			}
			if t[n-1] != ' ' {
				fmt.Println("Essa casa já está ocupada.")
				continue
			}
			return n - 1, true
		}
	}
}

// partida alterna os jogadores até alguém vencer ou o tabuleiro encher.
// Cada jogador é uma função: assim humano e computador se misturam à vontade.
func partida(jogadores map[byte]func(Tabuleiro, byte) (int, bool), comeca byte) {
	t := novoTabuleiro()
	vez := comeca
	fmt.Println(t.desenhar())
	for {
		casa, ok := jogadores[vez](t, vez)
		if !ok {
			fmt.Println("\nPartida interrompida.")
			return
		}
		t[casa] = vez
		fmt.Println()
		fmt.Println(t.desenhar())
		if v := t.vencedor(); v != ' ' {
			fmt.Printf("\nVitória de %c!\n", v)
			return
		}
		if t.cheio() {
			fmt.Println("\nEmpate: ninguém fechou uma linha.")
			return
		}
		vez = adversario(vez)
	}
}

func main() {
	modo := flag.String("modo", "cpu", "cpu (você contra o computador) ou 2j (dois jogadores)")
	comeca := flag.String("comeca", "X", "quem começa: X (você) ou O (computador), no modo cpu")
	demo := flag.Bool("demo", false, "computador contra computador, mostrando cada jogada")
	flag.Parse()

	if *demo {
		fmt.Println("=== CodeNomad Academy | Jogo da velha: computador contra computador ===")
		fmt.Println("Os dois lados usam o minimax, então nenhum consegue vencer: termina empatado.")
		fmt.Println()
		partida(map[byte]func(Tabuleiro, byte) (int, bool){'X': jogadorComputador, 'O': jogadorComputador}, 'X')
		return
	}
	primeiro := strings.ToUpper(strings.TrimSpace(*comeca))
	if primeiro != "X" && primeiro != "O" {
		fmt.Fprintln(os.Stderr, "Erro: --comeca precisa ser X ou O.")
		os.Exit(1)
	}
	humano := jogadorHumano()
	switch *modo {
	case "cpu":
		fmt.Println("=== Jogo da velha: você é X, o computador é O ===")
		partida(map[byte]func(Tabuleiro, byte) (int, bool){'X': humano, 'O': jogadorComputador}, primeiro[0])
	case "2j":
		fmt.Println("=== Jogo da velha: dois jogadores ===")
		partida(map[byte]func(Tabuleiro, byte) (int, bool){'X': humano, 'O': humano}, 'X')
	default:
		fmt.Fprintf(os.Stderr, "Erro: modo %q desconhecido. Use --modo cpu, --modo 2j ou --demo.\n", *modo)
		os.Exit(1)
	}
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. Conte quantas vezes o minimax foi chamado por jogada e mostre o número:
//    depois, acrescente a poda alfa-beta e compare.
// 2. Crie um nível fácil em que o computador joga numa casa livre ao acaso
//    (math/rand) em metade das vezes.
// 3. Generalize o Tabuleiro para 4x4 calculando as linhas vencedoras em vez
//    de escrevê-las à mão (o minimax vai ficar lento: por quê?).
// ------------------------------------------------------------
