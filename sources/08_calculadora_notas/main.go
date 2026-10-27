// ============================================================
// CodeNomad Academy  |  Sources Go  |  08. Calculadora de notas e média ponderada
// ------------------------------------------------------------
// O que faz: recebe notas e pesos, calcula a média ponderada, diz a situação
//            (aprovado, recuperação ou reprovado) e quanto o aluno precisa
//            tirar na próxima prova para alcançar a média de aprovação.
// Como usar: go run ./sources/08_calculadora_notas -notas 7.5,8,6 -pesos 2,3,5
//            go run ./sources/08_calculadora_notas -notas 4,6 -aprova 7 -recupera 5 -pesoprova 2
//            go run ./sources/08_calculadora_notas --demo
// O que você aprende: strings.Split e strconv.ParseFloat para ler listas,
//            slices de float64, funções que devolvem (valor, error), métodos
//            em structs, strings.Join para montar a conta por extenso e um
//            cálculo algébrico simples virando código legível.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// aluno guarda as notas e os pesos na mesma ordem: notas[i] vale pesos[i].
type aluno struct {
	nome  string
	notas []float64
	pesos []float64
}

// criterio reúne as regras de aprovação usadas na avaliação.
type criterio struct {
	aprova    float64 // média mínima para aprovação direta
	recupera  float64 // média mínima para ter direito à recuperação
	pesoProva float64 // peso da próxima prova, para a projeção
}

// lerLista transforma "7.5, 8,6" em []float64{7.5, 8, 6}. Aceita vírgula
// no lugar do ponto se os itens forem separados por ponto e vírgula ("7,5;8;6").
func lerLista(texto string) ([]float64, error) {
	separador := ","
	if strings.Contains(texto, ";") {
		separador = ";"
		texto = strings.ReplaceAll(texto, ",", ".")
	}
	var valores []float64
	for _, parte := range strings.Split(texto, separador) {
		parte = strings.TrimSpace(parte)
		if parte == "" {
			continue
		}
		v, err := strconv.ParseFloat(parte, 64)
		if err != nil {
			return nil, fmt.Errorf("%q não é um número", parte)
		}
		valores = append(valores, v)
	}
	if len(valores) == 0 {
		return nil, errors.New("a lista está vazia")
	}
	return valores, nil
}

// novoAluno valida as listas: notas entre 0 e 10, pesos positivos e o
// mesmo tamanho. Sem pesos, todos valem 1 (média simples).
func novoAluno(nome string, notas, pesos []float64) (aluno, error) {
	if len(pesos) == 0 {
		pesos = make([]float64, len(notas))
		for i := range pesos {
			pesos[i] = 1
		}
	}
	if len(notas) != len(pesos) {
		return aluno{}, fmt.Errorf("são %d notas para %d pesos; as listas precisam ter o mesmo tamanho", len(notas), len(pesos))
	}
	for i, n := range notas {
		if n < 0 || n > 10 {
			return aluno{}, fmt.Errorf("a nota %s está fora da escala de 0 a 10", dec(n))
		}
		if pesos[i] <= 0 {
			return aluno{}, fmt.Errorf("o peso %s precisa ser maior que zero", dec(pesos[i]))
		}
	}
	if strings.TrimSpace(nome) == "" {
		nome = "Aluno"
	}
	return aluno{nome: nome, notas: notas, pesos: pesos}, nil
}

// somas devolve a soma de nota*peso e a soma dos pesos, que as outras
// contas reaproveitam.
func (a aluno) somas() (produtos, pesos float64) {
	for i, n := range a.notas {
		produtos += n * a.pesos[i]
		pesos += a.pesos[i]
	}
	return produtos, pesos
}

func (a aluno) media() float64 {
	produtos, pesos := a.somas()
	return produtos / pesos
}

// contaPorExtenso monta "(8×1 + 7,5×1 + 9×2) / 4" para o aluno conferir.
func (a aluno) contaPorExtenso() string {
	termos := make([]string, len(a.notas))
	for i, n := range a.notas {
		termos[i] = dec(n) + "×" + dec(a.pesos[i])
	}
	_, pesos := a.somas()
	return "(" + strings.Join(termos, " + ") + ") / " + dec(pesos)
}

func (a aluno) situacao(c criterio) string {
	switch m := a.media(); {
	case m >= c.aprova:
		return "Aprovado"
	case m >= c.recupera:
		return "Recuperação"
	}
	return "Reprovado"
}

// proximaProva resolve a equação (produtos + x*peso) / (pesos + peso) = alvo
// para x, que é a nota necessária na próxima prova.
func (a aluno) proximaProva(c criterio) string {
	produtos, pesos := a.somas()
	necessaria := (c.aprova*(pesos+c.pesoProva) - produtos) / c.pesoProva
	switch {
	case necessaria <= 0:
		return "já alcança a média mesmo tirando zero na próxima prova"
	case necessaria > 10:
		maxima := (produtos + 10*c.pesoProva) / (pesos + c.pesoProva)
		return "não alcança a média nem com 10 na próxima prova (chegaria a " + dec(maxima) + ")"
	}
	return "precisa tirar pelo menos " + dec(necessaria) + " na próxima prova (peso " + dec(c.pesoProva) + ")"
}

// dec formata com até duas casas e vírgula, sem zeros sobrando.
func dec(v float64) string {
	s := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(v, 'f', 2, 64), "0"), ".")
	return strings.Replace(s, ".", ",", 1)
}

func imprimir(a aluno, c criterio) {
	fmt.Printf("%s\n", a.nome)
	fmt.Printf("  Média ponderada: %s = %s\n", a.contaPorExtenso(), dec(a.media()))
	fmt.Printf("  Situação: %s (aprova com %s, recuperação a partir de %s)\n", a.situacao(c), dec(c.aprova), dec(c.recupera))
	fmt.Printf("  Projeção: %s\n", a.proximaProva(c))
}

// rodarDemo avalia três alunos com pesos diferentes.
func rodarDemo() {
	fmt.Println("=== Demonstração: três alunos, três situações ===")
	fmt.Println()
	c := criterio{aprova: 7, recupera: 5, pesoProva: 1}
	exemplos := []struct {
		nome         string
		notas, pesos []float64
	}{
		{"Ana", []float64{10, 9, 9.5}, []float64{1, 1, 2}},
		{"Bruno", []float64{6.5, 7, 7}, []float64{2, 3, 5}},
		{"Carla", []float64{3, 4.5}, nil}, // sem pesos: média simples
	}
	for _, e := range exemplos {
		a, err := novoAluno(e.nome, e.notas, e.pesos)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			continue
		}
		imprimir(a, c)
		fmt.Println()
	}
}

// uso escreve no mesmo destino que flag.PrintDefaults usa (stderr).
func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Calculadora de notas e média ponderada (CodeNomad Academy)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Uso: go run ./sources/08_calculadora_notas -notas 7.5,8,6 -pesos 2,3,5 [-aprova 7] [-recupera 5] [-pesoprova 1]")
	fmt.Fprintln(w, "     go run ./sources/08_calculadora_notas --demo")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Opções:")
	flag.PrintDefaults()
}

func main() {
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	notasTexto := flag.String("notas", "", "notas separadas por vírgula (ex.: 7.5,8,6)")
	pesosTexto := flag.String("pesos", "", "pesos na mesma ordem (ex.: 2,3,5); sem pesos, todos valem 1")
	nome := flag.String("nome", "", "nome do aluno, opcional")
	c := criterio{}
	flag.Float64Var(&c.aprova, "aprova", 7, "média para aprovação direta")
	flag.Float64Var(&c.recupera, "recupera", 5, "média a partir da qual há recuperação")
	flag.Float64Var(&c.pesoProva, "pesoprova", 1, "peso da próxima prova na projeção")
	flag.Usage = uso
	flag.Parse()

	if *demo {
		rodarDemo()
		return
	}
	if flag.NFlag() == 0 {
		uso()
		return
	}
	if c.pesoProva <= 0 || c.recupera > c.aprova || c.aprova > 10 {
		fmt.Fprintln(os.Stderr, "Erro: -pesoprova deve ser positivo e -recupera não pode passar de -aprova (máximo 10)")
		os.Exit(1)
	}
	notas, err := lerLista(*notasTexto)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro em -notas:", err)
		os.Exit(1)
	}
	var pesos []float64
	if *pesosTexto != "" {
		if pesos, err = lerLista(*pesosTexto); err != nil {
			fmt.Fprintln(os.Stderr, "Erro em -pesos:", err)
			os.Exit(1)
		}
	}
	a, err := novoAluno(*nome, notas, pesos)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
	imprimir(a, c)
}

// ------------------------------------------------------------
// Para ir além
// 1. Leia uma turma inteira de um CSV (nome;notas;pesos) e imprima um
//    ranking por média com sort.Slice.
// 2. Acrescente a frequência: abaixo de 75% o aluno reprova mesmo com
//    média alta, como em muitas escolas.
// 3. Calcule a nota da recuperação necessária quando a regra for
//    "média final = (média + recuperação) / 2".
// ------------------------------------------------------------
