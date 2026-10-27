// ============================================================
// CodeNomad Academy  |  Sources Go  |  22. Busca binária e busca linear
// ------------------------------------------------------------
// O que faz: implementa a busca linear e a busca binária do zero, contando os
//            passos de cada uma em vetores de tamanhos diferentes, e mostra
//            por que a binária exige vetor ordenado e como os passos crescem.
// Como usar: go run ./sources/22_busca_binaria --alvo 42 3 8 15 42 57 90
//            go run ./sources/22_busca_binaria --alvo 7777 --tamanho 100000
//            go run ./sources/22_busca_binaria --demo
// O que você aprende: struct como resultado de função, laço com dois índices,
//            switch sem condição, slices.IsSorted e slices.Sort, math.Log2,
//            tabela formatada com fmt e a diferença entre O(n) e O(log n).
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"slices"
	"sort"
	"strconv"
)

// Resultado descreve o fim de uma busca: onde achou (ou -1), quantos passos
// deu e quais índices visitou, para desenhar o caminho da busca.
type Resultado struct {
	posicao int
	passos  int
	trilha  []int
}

// buscaLinear olha uma posição por vez, do começo ao fim. Funciona em qualquer
// vetor, ordenado ou não, mas no pior caso visita todas as posições.
func buscaLinear(v []int, alvo int) Resultado {
	r := Resultado{posicao: -1}
	for i, x := range v {
		r.passos++
		r.trilha = append(r.trilha, i)
		if x == alvo {
			r.posicao = i
			return r
		}
	}
	return r
}

// buscaBinaria olha o meio do trecho e descarta a metade onde o alvo não pode
// estar. Só funciona em vetor ordenado: é a ordem que autoriza o descarte.
func buscaBinaria(v []int, alvo int) Resultado {
	r := Resultado{posicao: -1}
	ini, fim := 0, len(v)-1
	for ini <= fim {
		// ini + (fim-ini)/2 em vez de (ini+fim)/2: em linguagens com int de
		// 32 bits a soma estoura em vetores enormes. O hábito vale a pena.
		meio := ini + (fim-ini)/2
		r.passos++
		r.trilha = append(r.trilha, meio)
		switch {
		case v[meio] == alvo:
			r.posicao = meio
			return r
		case v[meio] < alvo:
			ini = meio + 1 // o alvo, se existir, está à direita
		default:
			fim = meio - 1 // ou à esquerda
		}
	}
	return r
}

// maximoTeorico é o teto de passos da binária em n posições: o trecho cai
// pela metade a cada passo, então são floor(log2(n)) + 1 passos no pior caso.
func maximoTeorico(n int) int {
	if n == 0 {
		return 0
	}
	return int(math.Log2(float64(n))) + 1
}

// milhar formata 1234567 como 1.234.567, no padrão brasileiro.
func milhar(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return s
}

// descrever formata o resultado de uma busca em uma linha legível.
func descrever(nome string, r Resultado, mostrarTrilha bool) string {
	onde := "não encontrado"
	if r.posicao >= 0 {
		onde = fmt.Sprintf("posição %d", r.posicao)
	}
	linha := fmt.Sprintf("  %-8s %-15s %7s passo(s)", nome, onde, milhar(r.passos))
	if mostrarTrilha {
		linha += fmt.Sprintf("   índices visitados: %v", r.trilha)
	}
	return linha
}

// comparar roda as duas buscas no mesmo vetor e imprime os resultados. A
// trilha só aparece em vetores pequenos, para não encher a tela.
func comparar(v []int, alvo int) {
	fmt.Printf("Procurando %d em %s posições (teto da binária: %d passos)\n",
		alvo, milhar(len(v)), maximoTeorico(len(v)))
	mostrar := len(v) <= 16
	fmt.Println(descrever("linear", buscaLinear(v, alvo), mostrar))
	fmt.Println(descrever("binária", buscaBinaria(v, alvo), mostrar))
}

// vetorOrdenado gera 1, 3, 5, ... com n posições. Só ímpares: assim qualquer
// número par serve de exemplo de valor ausente.
func vetorOrdenado(n int) []int {
	v := make([]int, n)
	for i := range v {
		v[i] = 2*i + 1
	}
	return v
}

// tabelaCrescimento mostra como o pior caso cresce com o tamanho do vetor.
func tabelaCrescimento() {
	fmt.Printf("  %12s %14s %14s\n", "posições", "linear (pior)", "binária (pior)")
	for n := 10; n <= 10000000; n *= 10 {
		fmt.Printf("  %12s %14s %14d\n", milhar(n), milhar(n), maximoTeorico(n))
	}
}

// demo procura valores em vetores de 10, 1.000 e 100.000 posições, mostra o
// que acontece em vetor fora de ordem e a tabela de crescimento.
func demo() {
	fmt.Println("=== Busca linear e busca binária: CodeNomad Academy ===")
	pequeno := vetorOrdenado(10)
	fmt.Printf("\nVetor de 10 posições: %v\n", pequeno)
	comparar(pequeno, 15) // existe, no meio
	comparar(pequeno, 1)  // primeiro: melhor caso da linear
	comparar(pequeno, 19) // último: pior caso da linear
	comparar(pequeno, 8)  // não existe
	for _, n := range []int{1000, 100000} {
		v := vetorOrdenado(n)
		fmt.Println()
		comparar(v, v[len(v)-1]) // último elemento: pior caso da linear
		comparar(v, 2*n)         // valor ausente: as duas vão até o fim
	}
	fmt.Println("\nPor que a binária exige vetor ordenado:")
	bagunca := []int{42, 7, 19, 3, 88, 15, 61}
	fmt.Printf("  vetor fora de ordem: %v, procurando 7\n", bagunca)
	fmt.Println(descrever("linear", buscaLinear(bagunca, 7), true))
	fmt.Println(descrever("binária", buscaBinaria(bagunca, 7), true))
	meio := len(bagunca) / 2
	fmt.Printf("  A binária confiou na ordem: viu v[%d] = %d, menor que 7, e descartou a metade\n", meio, bagunca[meio])
	fmt.Println("  esquerda inteira, justamente onde o 7 estava. Em vetor desordenado ela mente.")
	fmt.Println("\nComo o pior caso cresce com o tamanho:")
	tabelaCrescimento()
	fmt.Println("\nLeitura: dobrar o vetor custa um passo a mais para a binária e o dobro para a linear.")
	fmt.Printf("A biblioteca padrão já traz a binária pronta: sort.SearchInts(v, 15) = %d\n",
		sort.SearchInts(pequeno, 15))
}

// lerNumeros converte os argumentos em inteiros, com erro claro para quem usa.
func lerNumeros(args []string) ([]int, error) {
	numeros := make([]int, 0, len(args))
	for _, arg := range args {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return nil, fmt.Errorf("'%s' não é um número inteiro", arg)
		}
		numeros = append(numeros, n)
	}
	return numeros, nil
}

func main() {
	alvoTexto := flag.String("alvo", "", "valor procurado (obrigatório fora da demonstração)")
	tamanho := flag.Int("tamanho", 0, "gera um vetor ordenado com N posições em vez de usar os números informados")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Uso: go run ./sources/22_busca_binaria --alvo VALOR [--tamanho N | n1 n2 n3 ...]\n",
			"As opções vêm antes dos números: o pacote flag para no primeiro argumento sem hífen.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	alvo, err := strconv.Atoi(*alvoTexto)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro: informe --alvo com um número inteiro (recebi '%s')\n", *alvoTexto)
		flag.Usage()
		os.Exit(2)
	}
	var v []int
	switch {
	case *tamanho < 0 || *tamanho > 10000000:
		fmt.Fprintln(os.Stderr, "erro: --tamanho deve ficar entre 1 e 10.000.000")
		os.Exit(2)
	case *tamanho > 0:
		v = vetorOrdenado(*tamanho)
	case flag.NArg() > 0:
		v, err = lerNumeros(flag.Args())
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro:", err)
			os.Exit(2)
		}
	default:
		flag.Usage()
		os.Exit(2)
	}
	if !slices.IsSorted(v) {
		fmt.Println("Aviso: os números vieram fora de ordem. A binária exige ordem, então o vetor foi ordenado antes.")
		slices.Sort(v)
	}
	if len(v) <= 16 {
		fmt.Printf("Vetor: %v\n", v)
	}
	comparar(v, alvo)
}

// ------------------------------------------------------------
// Para ir além
// 1. Faça a busca binária devolver a primeira ocorrência quando há repetidos.
// 2. Implemente a busca por interpolação, que chuta a posição pela proporção
//    do alvo entre o menor e o maior valor, e compare os passos.
// 3. Escreva a binária recursiva e meça se a diferença de tempo é perceptível.
// ------------------------------------------------------------
