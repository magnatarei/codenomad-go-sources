// ============================================================
// CodeNomad Academy  |  Sources Go  |  09. Calculadora de datas
// ------------------------------------------------------------
// O que faz: cinco operações com datas: dias entre duas datas, dia da
//            semana de uma data, idade exata em anos, meses e dias, somar
//            ou subtrair dias de uma data e quantos dias faltam para uma data.
// Como usar: go run ./sources/09_calculadora_datas -entre 2024-01-15,2024-12-25
//            go run ./sources/09_calculadora_datas -idade 1990-05-20 -hoje 2026-10-27
//            go run ./sources/09_calculadora_datas --demo
// O que você aprende: o pacote time (Parse, Date, AddDate, Sub, Weekday),
//            por que o layout de data em Go é "2006-01-02", mais de um
//            formato de entrada, strings.Cut, maps para traduzir nomes,
//            funções como valores em uma tabela e retorno (valor, error).
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Em Go o layout de uma data é a data de referência 2 de janeiro de 2006.
var layouts = []string{"2006-01-02", "02/01/2006", "2/1/2006"}

// diasSemana traduz o time.Weekday (que só fala inglês) para português.
var diasSemana = map[time.Weekday]string{
	time.Sunday: "domingo", time.Monday: "segunda-feira", time.Tuesday: "terça-feira", time.Wednesday: "quarta-feira",
	time.Thursday: "quinta-feira", time.Friday: "sexta-feira", time.Saturday: "sábado",
}

// parseData aceita "2024-12-25" ou "25/12/2024". Parse devolve UTC sem hora: nada de fuso.
func parseData(texto string) (time.Time, error) {
	texto = strings.TrimSpace(texto)
	for _, layout := range layouts {
		if d, err := time.Parse(layout, texto); err == nil {
			return d, nil
		}
	}
	return time.Time{}, fmt.Errorf("data inválida %q: use AAAA-MM-DD ou DD/MM/AAAA", texto)
}

func formatar(d time.Time) string {
	return fmt.Sprintf("%s (%s)", d.Format("02/01/2006"), diasSemana[d.Weekday()])
}

// diasEntre conta os dias de a até b: Sub devolve uma Duration, e 24 h são um dia.
func diasEntre(a, b time.Time) int {
	return int(b.Sub(a).Hours() / 24)
}

// idade calcula anos, meses e dias completos entre o nascimento e a
// referência. A conta é feita campo a campo, "pedindo emprestado" do mês
// anterior quando os dias ficam negativos, como numa subtração manual.
func idade(nasc, ref time.Time) (anos, meses, dias int, err error) {
	if ref.Before(nasc) {
		return 0, 0, 0, fmt.Errorf("a data de referência (%s) é anterior ao nascimento (%s)",
			ref.Format("02/01/2006"), nasc.Format("02/01/2006"))
	}
	anos = ref.Year() - nasc.Year()
	meses = int(ref.Month()) - int(nasc.Month())
	dias = ref.Day() - nasc.Day()
	if dias < 0 {
		meses--
		// Truque do pacote time: o dia 0 de um mês vira o último dia do mês anterior.
		dias += time.Date(ref.Year(), ref.Month(), 0, 0, 0, 0, 0, time.UTC).Day()
	}
	if meses < 0 {
		anos--
		meses += 12
	}
	return anos, meses, dias, nil
}

// plural devolve "1 dia" ou "5 dias" sem repetir if em todo lugar.
func plural(n int, singular, pluralTexto string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, pluralTexto)
}

// Todas as operações têm a mesma assinatura (argumento, data de referência)
// para caberem em uma tabela; quem não usa a referência recebe "_".
func opEntre(arg string, _ time.Time) error {
	t1, t2, ok := strings.Cut(arg, ",") // "x,y" vira "x" e "y"
	if !ok {
		return fmt.Errorf("-entre espera duas datas separadas por vírgula, recebi %q", arg)
	}
	a, err := parseData(t1)
	if err != nil {
		return err
	}
	b, err := parseData(t2)
	if err != nil {
		return err
	}
	n := diasEntre(a, b)
	fmt.Printf("De %s até %s: %s", formatar(a), formatar(b), plural(n, "dia", "dias"))
	if n > 0 {
		fmt.Printf(" (%d semanas e %s)", n/7, plural(n%7, "dia", "dias"))
	}
	fmt.Println()
	return nil
}

func opSemana(arg string, _ time.Time) error {
	d, err := parseData(arg)
	if err != nil {
		return err
	}
	fmt.Printf("%s foi %s\n", d.Format("02/01/2006"), diasSemana[d.Weekday()])
	return nil
}

func opIdade(arg string, hoje time.Time) error {
	nasc, err := parseData(arg)
	if err != nil {
		return err
	}
	a, m, d, err := idade(nasc, hoje)
	if err != nil {
		return err
	}
	fmt.Printf("Nascido em %s, em %s tem %s, %s e %s (%d dias vividos)\n",
		formatar(nasc), hoje.Format("02/01/2006"), plural(a, "ano", "anos"),
		plural(m, "mês", "meses"), plural(d, "dia", "dias"), diasEntre(nasc, hoje))
	return nil
}

func opSomar(arg string, _ time.Time) error {
	t1, t2, ok := strings.Cut(arg, ",")
	if !ok {
		return fmt.Errorf("-somar espera data,dias separados por vírgula, recebi %q", arg)
	}
	d, err := parseData(t1)
	if err != nil {
		return err
	}
	n, err := strconv.Atoi(strings.TrimSpace(t2))
	if err != nil {
		return fmt.Errorf("quantidade de dias inválida %q: use um inteiro, negativo para subtrair", t2)
	}
	// AddDate cuida sozinho de virada de mês, de ano e de ano bissexto.
	r := d.AddDate(0, 0, n)
	verbo, qtd := "mais", n
	if n < 0 {
		verbo, qtd = "menos", -n
	}
	fmt.Printf("%s %s %s = %s\n", formatar(d), verbo, plural(qtd, "dia", "dias"), formatar(r))
	return nil
}

func opFaltam(arg string, hoje time.Time) error {
	alvo, err := parseData(arg)
	if err != nil {
		return err
	}
	n := diasEntre(hoje, alvo)
	if n >= 0 {
		fmt.Printf("Faltam %s para %s (contando de %s)\n", plural(n, "dia", "dias"), formatar(alvo), hoje.Format("02/01/2006"))
	} else {
		fmt.Printf("%s já passou há %s (contando de %s)\n", formatar(alvo), plural(-n, "dia", "dias"), hoje.Format("02/01/2006"))
	}
	return nil
}

// operacao liga um texto de entrada à função que o trata.
type operacao struct {
	arg string
	fn  func(string, time.Time) error
}

// rodarDemo executa as cinco operações com datas fixas; o "hoje" é 27/10/2026.
func rodarDemo() {
	hoje := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	fmt.Println("=== Calculadora de datas: demonstração (hoje = 27/10/2026) ===")
	passos := []struct {
		titulo string
		op     operacao
	}{
		{"1. Dias entre duas datas", operacao{"2024-01-15,2024-12-25", opEntre}},
		{"   Ano bissexto no meio", operacao{"01/01/2024,01/01/2025", opEntre}},
		{"2. Dia da semana", operacao{"15/11/1889", opSemana}},
		{"   Outra data", operacao{"2000-01-01", opSemana}},
		{"3. Idade exata", operacao{"1990-05-20", opIdade}},
		{"   Nascido no fim do mês", operacao{"31/01/2000", opIdade}},
		{"4. Somar dias", operacao{"2024-01-31,30", opSomar}},
		{"   Subtrair dias", operacao{"2024-03-01,-1", opSomar}},
		{"5. Quantos dias faltam", operacao{"2026-12-25", opFaltam}},
		{"   Data que já passou", operacao{"2026-01-01", opFaltam}},
	}
	for _, p := range passos {
		fmt.Printf("%-28s ", p.titulo)
		if err := p.op.fn(p.op.arg, hoje); err != nil {
			fmt.Println("Erro:", err)
		}
	}
}

func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Calculadora de datas (CodeNomad Academy). Datas em AAAA-MM-DD ou DD/MM/AAAA. Uma operação por vez:")
	fmt.Fprintln(w, "  -entre D1,D2   -semana D   -idade D   -somar D,N   -faltam D   --demo")
	flag.PrintDefaults()
}

func main() {
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	entre := flag.String("entre", "", "duas datas separadas por vírgula: conta os dias entre elas")
	semana := flag.String("semana", "", "uma data: mostra o dia da semana")
	idadeDe := flag.String("idade", "", "data de nascimento: idade exata em anos, meses e dias")
	somar := flag.String("somar", "", "data,dias: soma (ou subtrai, se negativo) dias à data")
	faltam := flag.String("faltam", "", "uma data: quantos dias faltam para ela")
	hojeTxt := flag.String("hoje", "", "data de referência para -idade e -faltam (padrão: a data de hoje)")
	flag.Usage = uso
	flag.Parse()

	if *demo {
		rodarDemo()
		return
	}
	var err error
	hoje := time.Now().UTC().Truncate(24 * time.Hour)
	if *hojeTxt != "" {
		if hoje, err = parseData(*hojeTxt); err != nil {
			fmt.Fprintln(os.Stderr, "Erro em -hoje:", err)
			os.Exit(1)
		}
	}
	// A primeira opção preenchida vence; sem nenhuma, mostra o uso.
	for _, op := range []operacao{{*entre, opEntre}, {*semana, opSemana},
		{*idadeDe, opIdade}, {*somar, opSomar}, {*faltam, opFaltam}} {
		if op.arg == "" {
			continue
		}
		if err := op.fn(op.arg, hoje); err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			os.Exit(1)
		}
		return
	}
	uso()
}

// ------------------------------------------------------------
// Para ir além
// 1. Acrescente -uteis D1,D2 que conta só os dias úteis entre as datas,
//    pulando sábados e domingos com um laço e d.Weekday().
// 2. Crie -aniversario D que diz quantos dias faltam para o aniversário
//    seguinte de quem nasceu em D, lembrando de quem nasceu em 29/02.
// 3. Leia uma lista de datas de um arquivo (uma por linha) com bufio.Scanner
//    e imprima todas ordenadas com sort.Slice e d.Before().
// ------------------------------------------------------------
