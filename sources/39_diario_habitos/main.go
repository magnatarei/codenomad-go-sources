// ============================================================
// CodeNomad Academy  |  Sources Go  |  39. Diário de hábitos com sequência
// ------------------------------------------------------------
// O que faz: registra hábitos cumpridos por dia num arquivo de texto simples
//            (uma linha por registro), calcula a sequência atual e a maior
//            sequência de cada hábito e desenha um resumo semanal em grade.
// Como usar: go run ./sources/39_diario_habitos -arquivo habitos.txt marcar ler
//            go run ./sources/39_diario_habitos -arquivo habitos.txt marcar correr 2026-10-20
//            go run ./sources/39_diario_habitos -arquivo habitos.txt -semanas 3 resumo
//            go run ./sources/39_diario_habitos --demo
// O que você aprende: bufio.Scanner, os.OpenFile com O_APPEND, map de map
//            como conjunto, time.AddDate e Weekday, sort e Printf com largura fixa.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const formatoData = "2006-01-02"

// Diario guarda, por hábito, o conjunto de dias cumpridos (map[string]bool é um conjunto).
type Diario struct {
	dias map[string]map[string]bool
}

// carregar lê o arquivo linha a linha ("AAAA-MM-DD hábito"); vazias e # são ignoradas.
func carregar(caminho string) (*Diario, error) {
	d := &Diario{dias: map[string]map[string]bool{}}
	f, err := os.Open(caminho)
	if errors.Is(err, os.ErrNotExist) {
		return d, nil // diário novo
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		linha := strings.TrimSpace(sc.Text())
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue
		}
		data, habito, ok := strings.Cut(linha, " ")
		if _, err := time.Parse(formatoData, data); !ok || err != nil {
			return nil, fmt.Errorf("linha %d do diário está fora do formato \"AAAA-MM-DD hábito\": %q", n, linha)
		}
		d.marcar(strings.TrimSpace(habito), data)
	}
	return d, sc.Err()
}

func (d *Diario) marcar(habito, data string) {
	if d.dias[habito] == nil {
		d.dias[habito] = map[string]bool{}
	}
	d.dias[habito][data] = true
}

// registrar acrescenta uma linha ao fim do arquivo: O_APPEND preserva o que já existe.
func registrar(caminho, habito, data string) error {
	f, err := os.OpenFile(caminho, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s %s\n", data, habito)
	return err
}

func (d *Diario) habitos() []string {
	nomes := make([]string, 0, len(d.dias))
	for h := range d.dias {
		nomes = append(nomes, h)
	}
	sort.Strings(nomes)
	return nomes
}

// sequencias devolve a sequência atual e a maior de um hábito. A atual conta
// para trás a partir de hoje; se hoje ainda não foi marcado, começa em ontem.
func (d *Diario) sequencias(habito string, hoje time.Time) (atual, maior int) {
	feito := d.dias[habito]
	dia := hoje
	if !feito[dia.Format(formatoData)] {
		dia = dia.AddDate(0, 0, -1)
	}
	for feito[dia.Format(formatoData)] {
		atual++
		dia = dia.AddDate(0, 0, -1)
	}
	// Para a maior sequência, percorre as datas em ordem e conta as consecutivas.
	datas := make([]string, 0, len(feito))
	for data := range feito {
		datas = append(datas, data)
	}
	sort.Strings(datas) // o formato AAAA-MM-DD ordena certo como texto
	corrida := 0
	for i, data := range datas {
		t, _ := time.Parse(formatoData, data)
		if i > 0 && t.AddDate(0, 0, -1).Format(formatoData) == datas[i-1] {
			corrida++
		} else {
			corrida = 1
		}
		maior = max(maior, corrida)
	}
	return atual, maior
}

// inicioDaSemana volta até a segunda. Weekday() dá 0 para domingo, daí o ajuste.
func inicioDaSemana(t time.Time) time.Time {
	return t.AddDate(0, 0, -(int(t.Weekday())+6)%7)
}

// resumo imprime a tabela de sequências e a grade das últimas semanas.
func (d *Diario) resumo(hoje time.Time, semanas int) {
	nomes := d.habitos()
	const largura = 12
	fmt.Printf("Hoje: %s\n\n  %-*s  %6s  %6s  %6s\n", hoje.Format("02/01/2006"), largura, "Hábito", "Atual", "Maior", "Total")
	for _, h := range nomes {
		atual, maior := d.sequencias(h, hoje)
		fmt.Printf("  %-*s  %6d  %6d  %6d\n", largura, h, atual, maior, len(d.dias[h]))
	}
	fmt.Printf("\nGrade das últimas %d semanas (x = feito, . = não, espaço = futuro):\n", semanas)
	primeira := inicioDaSemana(hoje).AddDate(0, 0, -7*(semanas-1))
	for s := 0; s < semanas; s++ {
		inicio := primeira.AddDate(0, 0, 7*s)
		fmt.Printf("\n  Semana de %s   S T Q Q S S D\n", inicio.Format("02/01"))
		for _, h := range nomes {
			fmt.Printf("  %-*s  ", largura, h)
			for dia := 0; dia < 7; dia++ {
				t := inicio.AddDate(0, 0, dia)
				switch {
				case t.After(hoje):
					fmt.Print("  ")
				case d.dias[h][t.Format(formatoData)]:
					fmt.Print(" x")
				default:
					fmt.Print(" .")
				}
			}
			fmt.Println()
		}
	}
}

func executar(caminho string, hoje time.Time, semanas int, args []string) error {
	d, err := carregar(caminho)
	if err != nil {
		return err
	}
	switch args[0] {
	case "marcar":
		if len(args) < 2 || strings.TrimSpace(args[1]) == "" {
			return errors.New("uso: marcar <hábito> [AAAA-MM-DD]")
		}
		habito, data := strings.ToLower(args[1]), hoje.Format(formatoData)
		if len(args) > 2 {
			t, err := time.Parse(formatoData, args[2])
			if err != nil {
				return fmt.Errorf("data %q inválida: use AAAA-MM-DD", args[2])
			}
			if t.After(hoje) {
				return fmt.Errorf("a data %s ainda não chegou", args[2])
			}
			data = args[2]
		}
		if d.dias[habito][data] {
			fmt.Printf("  %q já estava marcado em %s.\n", habito, data)
			return nil
		}
		if err := registrar(caminho, habito, data); err != nil {
			return err
		}
		d.marcar(habito, data)
		atual, _ := d.sequencias(habito, hoje)
		fmt.Printf("  Marcado %q em %s. Sequência atual: %d dia(s).\n", habito, data, atual)
	case "resumo":
		d.resumo(hoje, semanas)
	default:
		return fmt.Errorf("comando %q desconhecido (use marcar ou resumo)", args[0])
	}
	return nil
}

// demo cria três semanas de registros com padrões diferentes por hábito.
func demo() error {
	dir, err := os.MkdirTemp("", "codenomad-habitos-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	caminho := filepath.Join(dir, "habitos.txt")
	hoje := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC) // uma quarta-feira
	for i := 20; i >= 0; i-- {
		dia := hoje.AddDate(0, 0, -i)
		if i != 9 { // "ler" falhou um dia só, há 9 dias
			_ = registrar(caminho, "ler", dia.Format(formatoData))
		}
		if dia.Weekday() != time.Sunday && dia.Weekday() != time.Saturday {
			_ = registrar(caminho, "estudar go", dia.Format(formatoData))
		}
		if i%2 == 0 && i > 0 { // "correr" em dias alternados, mas não hoje
			_ = registrar(caminho, "correr", dia.Format(formatoData))
		}
	}
	fmt.Println("== Demonstração: diário em", caminho)
	for _, args := range [][]string{{"marcar", "correr"}, {"marcar", "correr"}, {"marcar", "meditar", "2026-10-30"}, {"resumo"}} {
		fmt.Printf("\n$ habitos %s\n", strings.Join(args, " "))
		if err := executar(caminho, hoje, 3, args); err != nil {
			fmt.Println("  Erro:", err)
		}
	}
	return nil
}

func main() {
	arquivo := flag.String("arquivo", "habitos.txt", "arquivo de texto com os registros")
	semanas := flag.Int("semanas", 3, "quantas semanas mostrar na grade do resumo")
	dataHoje := flag.String("hoje", time.Now().Format(formatoData), "data de referência AAAA-MM-DD")
	modoDemo := flag.Bool("demo", false, "roda uma demonstração completa num arquivo temporário")
	flag.Parse()
	hoje, err := time.Parse(formatoData, *dataHoje)
	switch {
	case *modoDemo:
		err = demo()
	case flag.NArg() == 0:
		fmt.Fprintln(os.Stderr, "Uso: habitos [-arquivo habitos.txt] [-semanas 3] [-hoje AAAA-MM-DD] marcar <hábito> [data] | resumo")
		flag.PrintDefaults()
		os.Exit(2)
	case err != nil:
		err = fmt.Errorf("-hoje %q inválido: use AAAA-MM-DD", *dataHoje)
	default:
		err = executar(*arquivo, hoje, *semanas, flag.Args())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Crie o comando "desmarcar <hábito> <data>" regravando o arquivo sem aquela linha.
// 2. Mostre a taxa de cumprimento dos últimos 30 dias em porcentagem por hábito.
// 3. Aceite hábitos com meta semanal (ex.: correr 3x) e diga se a meta da semana foi batida.
