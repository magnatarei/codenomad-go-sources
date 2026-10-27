// ============================================================
// CodeNomad Academy  |  Sources Go  |  24. Lista ligada do zero
// ------------------------------------------------------------
// O que faz: implementa uma lista simplesmente ligada com inserir no início,
//            no fim e em uma posição, remover, buscar, pegar, inverter, tamanho
//            e impressão, e compara com o slice da linguagem em uma operação.
// Como usar: go run ./sources/24_lista_ligada inicio:5 fim:9 em:1:7 buscar:9 remover:5 inverter
//            go run ./sources/24_lista_ligada fim:1 fim:2 fim:3 pegar:1 tamanho
//            go run ./sources/24_lista_ligada --demo
// O que você aprende: struct com ponteiro para o próprio tipo, nil como fim da lista,
//            percorrer com for e ponteiro, fmt.Stringer, fmt.Errorf, strings.Cut e time.Since.
// Requisitos: Go 1.22+, só biblioteca padrão
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

// No é uma caixa com um valor e o endereço da próxima caixa. O último nó
// aponta para nil: é assim que a lista sabe onde termina.
type No struct {
	valor   int
	proximo *No
}

// Lista guarda só a cabeça; o resto se alcança seguindo os ponteiros.
type Lista struct {
	cabeca  *No
	tamanho int
}

// InserirInicio é a operação barata: o nó novo aponta para a cabeça antiga e vira a cabeça.
func (l *Lista) InserirInicio(v int) {
	l.cabeca = &No{valor: v, proximo: l.cabeca}
	l.tamanho++
}

// InserirFim precisa andar até o último nó, por isso custa o tamanho da lista.
func (l *Lista) InserirFim(v int) {
	if l.cabeca == nil {
		l.InserirInicio(v)
		return
	}
	atual := l.cabeca
	for atual.proximo != nil {
		atual = atual.proximo
	}
	atual.proximo = &No{valor: v}
	l.tamanho++
}

// InserirEm coloca v na posição pos (0 = início, tamanho = fim).
func (l *Lista) InserirEm(pos, v int) error {
	if pos < 0 || pos > l.tamanho {
		return fmt.Errorf("posição %d fora do intervalo 0..%d", pos, l.tamanho)
	}
	if pos == 0 {
		l.InserirInicio(v)
		return nil
	}
	anterior := l.cabeca
	for i := 1; i < pos; i++ {
		anterior = anterior.proximo
	}
	anterior.proximo = &No{valor: v, proximo: anterior.proximo}
	l.tamanho++
	return nil
}

// Remover tira a primeira ocorrência de v: o nó anterior "pula" por cima dele
// e o coletor de lixo recolhe o que ficou sem referência.
func (l *Lista) Remover(v int) bool {
	if l.cabeca != nil && l.cabeca.valor == v {
		l.cabeca = l.cabeca.proximo
		l.tamanho--
		return true
	}
	for atual := l.cabeca; atual != nil && atual.proximo != nil; atual = atual.proximo {
		if atual.proximo.valor == v {
			atual.proximo = atual.proximo.proximo
			l.tamanho--
			return true
		}
	}
	return false
}

// Buscar devolve a posição da primeira ocorrência de v, ou -1.
func (l *Lista) Buscar(v int) int {
	i := 0
	for atual := l.cabeca; atual != nil; atual = atual.proximo {
		if atual.valor == v {
			return i
		}
		i++
	}
	return -1
}

// Pegar devolve o valor na posição pos. Sem atalho: caminha nó a nó desde a cabeça.
func (l *Lista) Pegar(pos int) (int, error) {
	if pos < 0 || pos >= l.tamanho {
		return 0, fmt.Errorf("posição %d não existe: a lista tem %d nó(s)", pos, l.tamanho)
	}
	atual := l.cabeca
	for i := 0; i < pos; i++ {
		atual = atual.proximo
	}
	return atual.valor, nil
}

// Inverter vira cada ponteiro para trás, com três variáveis: anterior, atual
// e seguinte. O seguinte é guardado antes de virar o ponteiro, senão se perde.
func (l *Lista) Inverter() {
	var anterior *No
	atual := l.cabeca
	for atual != nil {
		seguinte := atual.proximo
		atual.proximo = anterior
		anterior = atual
		atual = seguinte
	}
	l.cabeca = anterior
}

// String satisfaz fmt.Stringer: fmt.Println(l) chama este método sozinho (receptor ponteiro, use *Lista).
func (l *Lista) String() string {
	var sb strings.Builder
	for atual := l.cabeca; atual != nil; atual = atual.proximo {
		fmt.Fprintf(&sb, "%d -> ", atual.valor)
	}
	sb.WriteString("nil")
	return sb.String()
}

// precisa diz quantos números cada operação da linha de comando recebe.
var precisa = map[string]int{"inicio": 1, "fim": 1, "em": 2, "remover": 1, "buscar": 1, "pegar": 1, "inverter": 0, "tamanho": 0}

// executar interpreta uma operação em texto ("em:1:7") e a aplica na lista.
func executar(l *Lista, op string) error {
	nome, resto, temArgs := strings.Cut(op, ":")
	var args []int
	if temArgs {
		for _, p := range strings.Split(resto, ":") {
			n, err := strconv.Atoi(p)
			if err != nil {
				return fmt.Errorf("'%s' não é um número inteiro em '%s'", p, op)
			}
			args = append(args, n)
		}
	}
	if n, ok := precisa[nome]; !ok {
		return fmt.Errorf("operação desconhecida '%s'", nome)
	} else if len(args) != n {
		return fmt.Errorf("'%s' espera %d número(s), recebeu %d", nome, n, len(args))
	}
	switch nome {
	case "inicio":
		l.InserirInicio(args[0])
	case "fim":
		l.InserirFim(args[0])
	case "em":
		if err := l.InserirEm(args[0], args[1]); err != nil {
			return err
		}
	case "remover":
		if !l.Remover(args[0]) {
			fmt.Printf("  %-12s valor %d não está na lista\n", op, args[0])
		}
	case "buscar":
		fmt.Printf("  %-12s posição %d\n", op, l.Buscar(args[0]))
	case "pegar":
		v, err := l.Pegar(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("  %-12s valor %d\n", op, v)
	case "inverter":
		l.Inverter()
	case "tamanho":
		fmt.Printf("  %-12s %d nó(s)\n", op, l.tamanho)
	}
	fmt.Printf("  %-12s lista: %v\n", op, l)
	return nil
}

// compararComSlice mede onde a lista brilha: inserir no início, que no slice copia tudo uma casa.
func compararComSlice(n int) {
	inicio := time.Now()
	l := &Lista{}
	for i := 0; i < n; i++ {
		l.InserirInicio(i)
	}
	tLista := time.Since(inicio)
	inicio = time.Now()
	var s []int
	for i := 0; i < n; i++ {
		s = append([]int{i}, s...) // copia o slice inteiro a cada inserção
	}
	tSlice := time.Since(inicio)
	fmt.Printf("  lista ligada: %.3f ms   slice: %.3f ms   (os dois terminam com %d elementos)\n",
		float64(tLista.Microseconds())/1000, float64(tSlice.Microseconds())/1000, len(s))
}

func demo() {
	fmt.Println("=== Lista ligada do zero: CodeNomad Academy ===")
	fmt.Println("\nCada operação e a lista logo depois:")
	l := &Lista{}
	roteiro := []string{"inicio:5", "fim:9", "inicio:2", "em:1:7", "fim:4", "buscar:9", "buscar:8",
		"pegar:2", "remover:5", "remover:99", "inverter", "tamanho", "em:9:1"}
	for _, op := range roteiro {
		if err := executar(l, op); err != nil {
			fmt.Printf("  %-12s erro: %v\n", op, err)
		}
	}
	fmt.Println("\nInserir 10.000 vezes no início, lista ligada contra slice:")
	compararComSlice(10000)
	fmt.Println("\nLeitura: a lista insere no início sem mover ninguém, mas para ler uma posição\n" +
		"precisa caminhar nó a nó; o slice é o contrário. Escolha pela operação que mais repete.")
}

func main() {
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Uso: go run ./sources/24_lista_ligada OPERAÇÃO [OPERAÇÃO ...]")
		fmt.Fprintln(os.Stderr, "Operações: inicio:V  fim:V  em:POS:V  remover:V  buscar:V  pegar:POS  inverter  tamanho")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	l := &Lista{}
	for _, op := range flag.Args() {
		if err := executar(l, op); err != nil {
			fmt.Fprintln(os.Stderr, "erro:", err)
			os.Exit(2)
		}
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Guarde também um ponteiro para a cauda e faça InserirFim custar o mesmo que InserirInicio.
// 2. Transforme em lista duplamente ligada (cada nó conhece o anterior) e remova pelo fim.
// 3. Torne a Lista genérica com type Lista[T comparable] e teste com strings.
// ------------------------------------------------------------
