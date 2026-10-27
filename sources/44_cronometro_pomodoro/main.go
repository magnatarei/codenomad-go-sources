// ============================================================
// CodeNomad Academy  |  Sources Go  |  44. Cronômetro Pomodoro no terminal
// ------------------------------------------------------------
// O que faz: alterna ciclos de foco e pausa com contagem regressiva que se
//            redesenha na mesma linha, barra de progresso, campainha ao fim
//            de cada fase e um resumo da sessão. Ctrl+C encerra com calma e
//            ainda mostra o resumo. Tempos e ciclos são configuráveis.
// Como usar: go run ./sources/44_cronometro_pomodoro --foco 25m --pausa 5m --ciclos 4
//            go run ./sources/44_cronometro_pomodoro --foco 90s --pausa 30s --ciclos 2 --sem-som
//            go run ./sources/44_cronometro_pomodoro --demo
// O que você aprende: time.Duration e flag.Duration ("25m", "90s"), ticker
//            com select e contexto, Ctrl+C via signal.NotifyContext, o
//            retorno de carro \r para redesenhar a linha, strings.Repeat para
//            a barra e um método em struct que acumula o que foi feito.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"
)

// O tipo abaixo guarda a configuração e o que já foi cumprido na sessão.
type Sessao struct {
	Foco, Pausa, PausaLonga time.Duration
	Ciclos                  int
	SemSom                  bool
	Terminal                bool // se a saída é um terminal, \r redesenha a linha

	completos  int
	tempoFoco  time.Duration
	tempoPausa time.Duration
}

func main() {
	foco := flag.Duration("foco", 25*time.Minute, "duração de cada ciclo de foco (ex.: 25m, 90s)")
	pausa := flag.Duration("pausa", 5*time.Minute, "duração da pausa curta")
	pausaLonga := flag.Duration("pausa-longa", 15*time.Minute, "pausa longa a cada 4 ciclos (0 desliga)")
	ciclos := flag.Int("ciclos", 4, "quantos ciclos de foco")
	semSom := flag.Bool("sem-som", false, "não toca a campainha do terminal")
	demo := flag.Bool("demo", false, "um ciclo de foco e uma pausa de 2 segundos cada")
	flag.Parse()

	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "Uso: cronometro_pomodoro --foco 25m --pausa 5m --ciclos 4 [--pausa-longa 15m] [--sem-som]   ou   --demo")
		flag.PrintDefaults()
		os.Exit(2)
	}
	if *demo {
		fmt.Println("=== CodeNomad Academy | Pomodoro: demonstração com ciclos de 2 segundos ===")
		*foco, *pausa, *pausaLonga, *ciclos = 2*time.Second, 2*time.Second, 0, 1
	}
	if *foco < time.Second || *pausa < time.Second {
		falhar("Erro: --foco e --pausa precisam ter pelo menos 1s (exemplos: 25m, 90s, 1h).")
	}
	if *ciclos < 1 || *ciclos > 100 {
		falhar("Erro: --ciclos precisa estar entre 1 e 100.")
	}
	if *pausaLonga < 0 {
		falhar("Erro: --pausa-longa não pode ser negativa.")
	}
	// NotifyContext cancela o contexto no Ctrl+C; cada fase observa ctx.Done().
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer parar()
	s := &Sessao{Foco: *foco, Pausa: *pausa, PausaLonga: *pausaLonga, Ciclos: *ciclos, SemSom: *semSom, Terminal: ehTerminal()}
	s.rodar(ctx)
}

// falhar imprime a mensagem na saída de erro e encerra com código 1.
func falhar(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// ehTerminal descobre se stdout é um terminal de verdade ou um arquivo/cano.
// Num arquivo, o \r não apaga nada: aí imprimimos uma linha por segundo.
func ehTerminal() bool {
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// rodar executa os ciclos e para na primeira fase interrompida.
func (s *Sessao) rodar(ctx context.Context) {
	inicio := time.Now()
	fmt.Printf("Plano: %d ciclo(s) de foco de %v com pausa de %v", s.Ciclos, s.Foco, s.Pausa)
	if s.PausaLonga > 0 {
		fmt.Printf(" e pausa longa de %v a cada 4 ciclos", s.PausaLonga)
	}
	fmt.Println(". Ctrl+C encerra a sessão.")
	interrompida := false
	for i := 1; i <= s.Ciclos && !interrompida; i++ {
		fmt.Printf("\nCiclo %d de %d: foco. Guarde o celular e vá fundo.\n", i, s.Ciclos)
		feito, completo := s.contagem(ctx, fmt.Sprintf("Foco %d/%d", i, s.Ciclos), s.Foco)
		s.tempoFoco += feito
		if !completo {
			interrompida = true
			break
		}
		s.completos++
		s.avisar("Fim do foco! Levante, beba água.")
		duracao, nome := s.Pausa, "Pausa"
		if s.PausaLonga > 0 && i%4 == 0 {
			duracao, nome = s.PausaLonga, "Pausa longa"
		}
		feito, completo = s.contagem(ctx, fmt.Sprintf("%s %d/%d", nome, i, s.Ciclos), duracao)
		s.tempoPausa += feito
		interrompida = !completo
		if completo {
			s.avisar("Fim da pausa.")
		}
	}
	s.resumo(time.Since(inicio), interrompida)
}

// contagem desenha a contagem regressiva de uma fase. Devolve quanto tempo
// passou e se a fase chegou ao fim (false quando o contexto foi cancelado).
func (s *Sessao) contagem(ctx context.Context, rotulo string, total time.Duration) (time.Duration, bool) {
	inicio := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop() // sem Stop, o ticker seguiria disparando depois do retorno
	for {
		restante := (total - time.Since(inicio)).Round(time.Second)
		if restante < 0 {
			restante = 0
		}
		fracao := 1 - float64(restante)/float64(total)
		linha := fmt.Sprintf("%-14s [%s] %s restantes", rotulo, barra(fracao, 20), relogio(restante))
		if s.Terminal {
			fmt.Print("\r" + linha + "   ") // volta ao começo da linha e sobrescreve
		} else {
			fmt.Println(linha)
		}
		if restante == 0 {
			if s.Terminal {
				fmt.Println()
			}
			return total, true
		}
		select { // espera o próximo segundo ou o cancelamento, o que vier antes
		case <-ticker.C:
		case <-ctx.Done():
			fmt.Println()
			return time.Since(inicio).Round(time.Second), false
		}
	}
}

// avisar imprime a mensagem e toca a campainha do terminal (\a) se permitido.
func (s *Sessao) avisar(msg string) {
	if !s.SemSom {
		fmt.Print("\a")
	}
	fmt.Println(">>", msg)
}

func (s *Sessao) resumo(total time.Duration, interrompida bool) {
	fmt.Println("\nResumo da sessão")
	fmt.Printf("  Ciclos de foco completos: %d de %d\n", s.completos, s.Ciclos)
	fmt.Printf("  Tempo em foco:            %s\n", relogio(s.tempoFoco))
	fmt.Printf("  Tempo em pausa:           %s\n", relogio(s.tempoPausa))
	fmt.Printf("  Duração total:            %s\n", relogio(total.Round(time.Second)))
	if interrompida {
		fmt.Println("  Encerrada antes do fim por Ctrl+C. Da próxima vez vai mais longe.")
	} else {
		fmt.Println("  Sessão completa. Bom trabalho!")
	}
}

// barra monta algo como [########............] para a fração concluída.
func barra(fracao float64, largura int) string {
	cheios := int(fracao*float64(largura) + 0.5)
	if cheios > largura {
		cheios = largura
	}
	return strings.Repeat("#", cheios) + strings.Repeat(".", largura-cheios)
}

// relogio formata uma duração como mm:ss, ou h:mm:ss quando passa de uma hora.
func relogio(d time.Duration) string {
	segundos := int(d.Seconds())
	h, m, s := segundos/3600, segundos%3600/60, segundos%60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. Grave cada ciclo completo em um arquivo pomodoro.csv (data, hora de
//    início, duração) com encoding/csv, e mostre o total do dia ao começar.
// 2. Aceite --tarefa "estudar slices" e mostre o nome da tarefa na linha da
//    contagem e no resumo.
// 3. Permita pausar e retomar: leia o teclado em outra rotina (bufio.Scanner
//    em os.Stdin) e mande sinais por um canal para a contagem.
// ------------------------------------------------------------
