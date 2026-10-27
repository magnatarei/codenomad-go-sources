// ============================================================
// CodeNomad Academy  |  Sources Go  |  46. Jogo da forca
// ------------------------------------------------------------
// O que faz: sorteia uma palavra do banco embutido (todas sobre programação,
//            cada uma com dica), desenha a forca em texto a cada erro, mostra
//            as letras já tentadas e anuncia vitória ou derrota. Pode jogar
//            no teclado ou assistir a um jogador automático.
// Como usar: go run ./sources/46_jogo_forca
//            go run ./sources/46_jogo_forca --erros 5 --semente 7
//            go run ./sources/46_jogo_forca --demo
// O que você aprende: structs com métodos, slices e strings.Builder para
//            desenhar, map[rune]bool como conjunto, bufio.Scanner no teclado,
//            math/rand com semente, unicode e strings.ToLower para validar
//            a letra digitada, e uma função que separa a regra do jogo da
//            origem dos palpites (teclado ou automático).
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"unicode"
)

// Palavra é um item do banco: o segredo e a dica que ajuda o jogador.
type Palavra struct {
	Texto string
	Dica  string
}

// banco só tem palavras sem acento, para o palpite ser uma letra simples.
var banco = []Palavra{
	{"goroutine", "roda em paralelo e custa poucos kilobytes"},
	{"ponteiro", "guarda o endereço de outra coisa"},
	{"compilador", "transforma o seu texto em programa"},
	{"interface", "conjunto de métodos que um tipo promete ter"},
	{"canal", "goroutines conversam por ele"},
	{"pacote", "a unidade de organização do código em Go"},
	{"fatia", "uma janela sobre um array, cresce com append"},
	{"estrutura", "agrupa campos com nome e tipo"},
	{"terminal", "onde este programa está rodando agora"},
	{"recursividade", "quando a função chama a si mesma"},
}

// Jogo guarda o estado de uma partida. Letras é usado como conjunto: o Go
// não tem um tipo set, então um map[rune]bool faz o papel.
type Jogo struct {
	Segredo   string
	Dica      string
	MaxErros  int
	Letras    map[rune]bool
	Erros     int
	Tentadas  []rune // em ordem, para mostrar ao jogador
	Acertadas int
}

func novoJogo(p Palavra, maxErros int) *Jogo {
	return &Jogo{Segredo: p.Texto, Dica: p.Dica, MaxErros: maxErros, Letras: map[rune]bool{}}
}

// tentar aplica uma letra e devolve uma mensagem para o jogador.
func (j *Jogo) tentar(letra rune) string {
	if j.Letras[letra] {
		return fmt.Sprintf("Você já tentou %q. Escolha outra letra.", letra)
	}
	j.Letras[letra] = true
	j.Tentadas = append(j.Tentadas, letra)
	ocorrencias := strings.Count(j.Segredo, string(letra))
	if ocorrencias == 0 {
		j.Erros++
		return fmt.Sprintf("Não tem %q. Erros: %d de %d.", letra, j.Erros, j.MaxErros)
	}
	j.Acertadas += ocorrencias
	return fmt.Sprintf("Boa! %q aparece %d vez(es).", letra, ocorrencias)
}

func (j *Jogo) venceu() bool { return j.Acertadas == len(j.Segredo) }
func (j *Jogo) perdeu() bool { return j.Erros >= j.MaxErros }
func (j *Jogo) acabou() bool { return j.venceu() || j.perdeu() }

// mascara mostra as letras descobertas e um traço baixo para as demais.
func (j *Jogo) mascara() string {
	var b strings.Builder
	for _, r := range j.Segredo {
		if j.Letras[r] {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
		b.WriteByte(' ')
	}
	return strings.TrimSpace(b.String())
}

// desenho devolve a forca com as partes do corpo proporcionais aos erros.
// Com --erros diferente de 6, a figura avança em passos proporcionais.
func (j *Jogo) desenho() string {
	partes := []string{"", "O", "O\n |", "O\n/|", "O\n/|\\", "O\n/|\\\n/", "O\n/|\\\n/ \\"}
	nivel := j.Erros * (len(partes) - 1) / j.MaxErros
	corpo := strings.Split(partes[nivel], "\n")
	for len(corpo) < 3 {
		corpo = append(corpo, "")
	}
	return fmt.Sprintf("  +---+\n  |   %s\n  |  %-3s\n  |  %-3s\n  |\n=====", corpo[0], corpo[1], corpo[2])
}

func (j *Jogo) mostrar() {
	fmt.Println(j.desenho())
	fmt.Printf("Palavra: %s   (dica: %s)\n", j.mascara(), j.Dica)
	fmt.Printf("Tentadas: %s\n", string(j.Tentadas))
}

// jogar roda a partida inteira. A origem dos palpites vem de fora: assim a
// mesma função serve para o teclado e para o jogador automático.
func jogar(j *Jogo, proximoPalpite func(*Jogo) (rune, bool)) {
	fmt.Printf("A palavra tem %d letras. Você pode errar %d vezes.\n", len(j.Segredo), j.MaxErros)
	for !j.acabou() {
		fmt.Println()
		j.mostrar()
		letra, ok := proximoPalpite(j)
		if !ok {
			fmt.Println("\nSem mais palpites. Partida encerrada.")
			return
		}
		fmt.Printf("Palpite: %q -> %s\n", letra, j.tentar(letra))
	}
	fmt.Println()
	j.mostrar()
	if j.venceu() {
		fmt.Printf("\nVocê venceu com %d erro(s)! A palavra era %q.\n", j.Erros, j.Segredo)
	} else {
		fmt.Printf("\nEnforcado! A palavra era %q.\n", j.Segredo)
	}
}

// lerDoTeclado devolve uma função que pede uma letra ao jogador até receber
// uma válida. Fim da entrada (Ctrl+D ou Ctrl+Z) encerra o jogo.
func lerDoTeclado() func(*Jogo) (rune, bool) {
	scanner := bufio.NewScanner(os.Stdin)
	return func(*Jogo) (rune, bool) {
		for {
			fmt.Print("Digite uma letra: ")
			if !scanner.Scan() {
				return 0, false
			}
			texto := strings.ToLower(strings.TrimSpace(scanner.Text()))
			letras := []rune(texto)
			if len(letras) != 1 || !unicode.IsLetter(letras[0]) || letras[0] > 'z' {
				fmt.Println("Digite exatamente uma letra de a a z, sem acento.")
				continue
			}
			return letras[0], true
		}
	}
}

// jogadorAutomatico tenta as letras na ordem em que aparecem com mais
// frequência em português, pulando as que já foram usadas.
func jogadorAutomatico() func(*Jogo) (rune, bool) {
	ordem := []rune("aeosrindmutclpvgqbfhjzxkwy")
	return func(j *Jogo) (rune, bool) {
		for _, r := range ordem {
			if !j.Letras[r] {
				return r, true
			}
		}
		return 0, false
	}
}

func main() {
	maxErros := flag.Int("erros", 6, "quantos erros o jogador pode cometer (1 a 20)")
	semente := flag.Int64("semente", 0, "semente do sorteio (0 = aleatória), para repetir uma partida")
	demo := flag.Bool("demo", false, "um jogador automático joga uma partida inteira")
	flag.Parse()

	if *maxErros < 1 || *maxErros > 20 {
		fmt.Fprintln(os.Stderr, "Erro: --erros precisa estar entre 1 e 20.")
		os.Exit(1)
	}
	gerador := rand.New(rand.NewSource(*semente))
	if *semente == 0 {
		gerador = rand.New(rand.NewSource(rand.Int63()))
	}
	if *demo {
		fmt.Println("=== CodeNomad Academy | Jogo da forca: partida automática ===")
		fmt.Println("O jogador automático tenta as letras mais comuns do português primeiro.")
		// A palavra da demonstração é fixa, para a saída ser sempre a mesma.
		jogar(novoJogo(banco[7], *maxErros), jogadorAutomatico())
		return
	}
	fmt.Println("=== Jogo da forca ===")
	jogar(novoJogo(banco[gerador.Intn(len(banco))], *maxErros), lerDoTeclado())
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. Leia o banco de um arquivo texto (palavra;dica por linha) com bufio.Scanner
//    e aceite --arquivo palavras.txt.
// 2. Permita palpitar a palavra inteira: acertou, venceu; errou, conta dois erros.
// 3. Guarde um placar em JSON (vitórias, derrotas, melhor partida) e mostre
//    ao começar.
// ------------------------------------------------------------
