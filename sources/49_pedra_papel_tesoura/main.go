// ============================================================
// CodeNomad Academy  |  Sources Go  |  49. Pedra, papel e tesoura com placar
// ------------------------------------------------------------
// O que faz: série de N rodadas contra o computador, com placar ao vivo e
//            encerramento assim que alguém garante a vitória. O computador
//            conta as jogadas que você mais usa e passa a jogar o que vence
//            a sua favorita, então repetir uma jogada sai caro.
// Como usar: go run ./sources/49_pedra_papel_tesoura --rodadas 5
//            go run ./sources/49_pedra_papel_tesoura --rodadas 7 --jogadas pedra,papel,pedra
//            go run ./sources/49_pedra_papel_tesoura --demo
// O que você aprende: map para o que vence o quê e para contar frequências,
//            struct com métodos para a estratégia, math/rand com semente,
//            bufio.Scanner com validação, strings.Split para jogadas prontas
//            e uma função que lê de duas origens diferentes.
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
)

// jogadas é a ordem fixa usada para sortear e para desempatar contagens.
var jogadas = []string{"pedra", "papel", "tesoura"}

// vence diz qual jogada ganha de qual: pedra quebra tesoura, e assim por diante.
var vence = map[string]string{"pedra": "tesoura", "papel": "pedra", "tesoura": "papel"}

// perdePara é o inverso de vence: o que ganha de cada jogada.
var perdePara = map[string]string{"tesoura": "pedra", "pedra": "papel", "papel": "tesoura"}

// Estrategia é a memória do computador: quantas vezes você usou cada jogada.
type Estrategia struct {
	contagem map[string]int
	gerador  *rand.Rand
}

func novaEstrategia(gerador *rand.Rand) *Estrategia {
	return &Estrategia{contagem: map[string]int{}, gerador: gerador}
}

// favorita devolve a jogada que você mais usou, ou "" se ainda não jogou.
// Em caso de empate, vence a que aparece primeiro na lista fixa de jogadas.
func (e *Estrategia) favorita() string {
	melhor, maior := "", 0
	for _, j := range jogadas {
		if e.contagem[j] > maior {
			melhor, maior = j, e.contagem[j]
		}
	}
	return melhor
}

// escolher joga ao acaso na primeira rodada e, depois, o que vence a sua
// jogada favorita. O método aprender é chamado depois de cada rodada.
func (e *Estrategia) escolher() (jogada, motivo string) {
	fav := e.favorita()
	if fav == "" {
		return jogadas[e.gerador.Intn(len(jogadas))], "primeira rodada, escolha ao acaso"
	}
	return perdePara[fav], fmt.Sprintf("você usou %s %d vez(es), e %s ganha de %s", fav, e.contagem[fav], perdePara[fav], fav)
}

func (e *Estrategia) aprender(jogadaDoJogador string) { e.contagem[jogadaDoJogador]++ }

// Placar acumula o resultado da série.
type Placar struct{ Jogador, Computador, Empates int }

// decidir compara as duas jogadas e atualiza o placar.
func (p *Placar) decidir(jogador, computador string) string {
	switch {
	case jogador == computador:
		p.Empates++
		return "empate"
	case vence[jogador] == computador:
		p.Jogador++
		return "ponto seu"
	}
	p.Computador++
	return "ponto do computador"
}

// serie roda até N rodadas, parando antes se alguém já não pode ser alcançado.
func serie(rodadas int, proximaJogada func(int) (string, bool), e *Estrategia) Placar {
	var placar Placar
	for r := 1; r <= rodadas; r++ {
		jogador, ok := proximaJogada(r)
		if !ok {
			fmt.Println("Sem mais jogadas. Série encerrada.")
			break
		}
		computador, motivo := e.escolher()
		resultado := placar.decidir(jogador, computador)
		e.aprender(jogador)
		fmt.Printf("Rodada %d: você %-7s x %-7s computador (%s) -> %s. Placar %d a %d\n",
			r, jogador, computador, motivo, resultado, placar.Jogador, placar.Computador)
		restantes := rodadas - r
		if restantes > 0 && (placar.Jogador > placar.Computador+restantes || placar.Computador > placar.Jogador+restantes) {
			fmt.Println("Série decidida: o outro lado não alcança mais.")
			break
		}
	}
	return placar
}

// lerJogada valida um texto digitado ou vindo de --jogadas; aceita a inicial.
func lerJogada(texto string) (string, bool) {
	texto = strings.ToLower(strings.TrimSpace(texto))
	for _, j := range jogadas {
		if texto == j || (len(texto) == 1 && texto[0] == j[0]) {
			return j, true
		}
	}
	return "", false
}

// doTeclado devolve uma função que pergunta a jogada até receber uma válida.
func doTeclado() func(int) (string, bool) {
	scanner := bufio.NewScanner(os.Stdin)
	return func(r int) (string, bool) {
		for {
			fmt.Printf("Rodada %d, sua jogada (pedra, papel ou tesoura): ", r)
			if !scanner.Scan() {
				return "", false
			}
			if j, ok := lerJogada(scanner.Text()); ok {
				return j, true
			}
			fmt.Println("Jogada inválida. Digite pedra, papel, tesoura ou só a inicial.")
		}
	}
}

// daLista devolve uma função que entrega as jogadas prontas, uma por rodada.
func daLista(lista []string) func(int) (string, bool) {
	return func(r int) (string, bool) {
		if r > len(lista) {
			return "", false
		}
		return lista[r-1], true
	}
}

func resumo(p Placar, e *Estrategia) {
	fmt.Printf("\nPlacar final: você %d, computador %d, empates %d.\n", p.Jogador, p.Computador, p.Empates)
	switch {
	case p.Jogador > p.Computador:
		fmt.Println("Você venceu a série!")
	case p.Computador > p.Jogador:
		fmt.Println("O computador venceu a série.")
	default:
		fmt.Println("Série empatada.")
	}
	fmt.Println("O que o computador aprendeu sobre você:")
	for _, j := range jogadas {
		fmt.Printf("  %-8s %d vez(es)\n", j, e.contagem[j])
	}
	if fav := e.favorita(); fav != "" {
		fmt.Printf("Sua favorita é %s, então ele jogaria %s na próxima.\n", fav, perdePara[fav])
	}
}

func main() {
	rodadas := flag.Int("rodadas", 5, "quantas rodadas tem a série (1 a 99)")
	prontas := flag.String("jogadas", "", "jogadas separadas por vírgula, em vez do teclado (ex.: pedra,papel,p)")
	semente := flag.Int64("semente", 0, "semente do sorteio da primeira rodada (0 = aleatória)")
	demo := flag.Bool("demo", false, "simula uma série de 7 rodadas com jogadas pré-definidas")
	flag.Parse()

	if *demo {
		*rodadas, *semente = 7, 42
		*prontas = "pedra,pedra,papel,pedra,tesoura,pedra,papel"
		fmt.Println("=== CodeNomad Academy | Pedra, papel e tesoura: série simulada de 7 rodadas ===")
		fmt.Println("Jogadas do jogador:", *prontas)
	}
	if *rodadas < 1 || *rodadas > 99 {
		fmt.Fprintln(os.Stderr, "Erro: --rodadas precisa estar entre 1 e 99.")
		os.Exit(1)
	}
	gerador := rand.New(rand.NewSource(*semente))
	if *semente == 0 {
		gerador = rand.New(rand.NewSource(rand.Int63()))
	}
	proxima := doTeclado()
	if *prontas != "" {
		var lista []string
		for _, texto := range strings.Split(*prontas, ",") {
			j, ok := lerJogada(texto)
			if !ok {
				fmt.Fprintf(os.Stderr, "Erro: %q não é uma jogada; use pedra, papel ou tesoura.\n", strings.TrimSpace(texto))
				os.Exit(1)
			}
			lista = append(lista, j)
		}
		proxima = daLista(lista)
	}
	e := novaEstrategia(gerador)
	placar := serie(*rodadas, proxima, e)
	resumo(placar, e)
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. Faça o computador olhar só as últimas 3 jogadas, em vez do total:
//    ele passa a reagir a mudanças de tática.
// 2. Guarde a contagem num arquivo JSON para o computador lembrar de você
//    entre uma execução e outra.
// 3. Acrescente "lagarto" e "spock" ampliando os mapas vence e perdePara
//    (cada jogada ganha de duas e perde para duas).
// ------------------------------------------------------------
