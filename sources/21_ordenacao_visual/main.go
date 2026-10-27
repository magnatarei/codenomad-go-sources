// ============================================================
// CodeNomad Academy  |  Sources Go  |  21. Algoritmos de ordenação com passos
// ------------------------------------------------------------
// O que faz: implementa bubble sort, insertion sort, merge sort e quick sort
//            do zero, contando comparações e trocas, com a opção de mostrar o
//            vetor a cada passo e uma tabela final comparando o desempenho.
// Como usar: go run ./sources/21_ordenacao_visual --algoritmo bubble --passos 5 2 9 1 7
//            go run ./sources/21_ordenacao_visual 5 2 9 1 7 3 8
//            go run ./sources/21_ordenacao_visual --demo
// O que você aprende: slices e por que copiar antes de ordenar, struct com métodos
//            e ponteiro, funções em um map, recursão, time.Since e n² contra n log n.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"time"
)

// Contador guarda as estatísticas de uma execução; os algoritmos recebem um
// ponteiro para ele e escrevem todos no mesmo lugar.
type Contador struct {
	comparacoes, trocas, passos int
	mostrar                     bool
}

// registrar conta um passo e, se pedido, imprime o vetor naquele instante. O
// rótulo só é formatado quando vai aparecer: 100 milhões de Sprintf custam caro.
func (c *Contador) registrar(v []int, formato string, args ...any) {
	c.passos++
	if c.mostrar {
		fmt.Printf("  passo %3d: %v  (%s)\n", c.passos, v, fmt.Sprintf(formato, args...))
	}
}

// bubbleSort compara vizinhos e troca os fora de ordem: a cada passada o
// maior valor do trecho "borbulha" até o fim.
func bubbleSort(v []int, c *Contador) []int {
	// Copiar é obrigatório: um slice aponta para o vetor de quem chamou.
	v = slices.Clone(v)
	for fim := len(v) - 1; fim > 0; fim-- {
		trocou := false
		for i := 0; i < fim; i++ {
			c.comparacoes++
			if v[i] > v[i+1] {
				v[i], v[i+1] = v[i+1], v[i]
				c.trocas++
				trocou = true
				c.registrar(v, "troca %d com %d", v[i+1], v[i])
			}
		}
		if !trocou {
			break // nenhuma troca na passada: o vetor já está em ordem
		}
	}
	return v
}

// insertionSort leva cada elemento para o lugar certo na parte já ordenada.
func insertionSort(v []int, c *Contador) []int {
	v = slices.Clone(v)
	for i := 1; i < len(v); i++ {
		chave := v[i]
		j := i - 1
		for j >= 0 {
			c.comparacoes++
			if v[j] <= chave {
				break
			}
			v[j+1] = v[j]
			c.trocas++
			j--
		}
		v[j+1] = chave
		c.registrar(v, "insere %d na posição %d", chave, j+1)
	}
	return v
}

// mergeSort divide ao meio, ordena cada metade e intercala as duas.
func mergeSort(v []int, c *Contador) []int {
	if len(v) <= 1 {
		return slices.Clone(v) // caso base da recursão: 0 ou 1 elemento já está ordenado
	}
	meio := len(v) / 2
	return intercalar(mergeSort(v[:meio], c), mergeSort(v[meio:], c), c)
}

// intercalar junta dois slices ordenados pegando sempre o menor da frente.
func intercalar(esq, dir []int, c *Contador) []int {
	resultado := make([]int, 0, len(esq)+len(dir))
	i, j := 0, 0
	for i < len(esq) && j < len(dir) {
		c.comparacoes++
		if esq[i] <= dir[j] {
			resultado = append(resultado, esq[i])
			i++
		} else {
			resultado = append(resultado, dir[j])
			j++
		}
	}
	resultado = append(resultado, esq[i:]...) // o que sobrou já está em ordem
	resultado = append(resultado, dir[j:]...)
	c.trocas += len(resultado) // no merge, cada escrita conta como movimento
	c.registrar(resultado, "intercala %v e %v", esq, dir)
	return resultado
}

// quickSort fixa um pivô, separa menores e maiores e repete em cada lado.
func quickSort(v []int, c *Contador) []int {
	v = slices.Clone(v)
	quick(v, 0, len(v)-1, c)
	return v
}

// quick ordena v[ini..fim] no lugar e para quando o trecho tem 0 ou 1 elemento.
func quick(v []int, ini, fim int, c *Contador) {
	if ini >= fim {
		return
	}
	p := particionar(v, ini, fim, c)
	c.registrar(v, "pivô %d fixado na posição %d", v[p], p)
	quick(v, ini, p-1, c)
	quick(v, p+1, fim, c)
}

// particionar (esquema de Lomuto) leva o pivô para o fim do trecho e devolve
// sua posição final. Pivô do meio evita o pior caso em vetor já ordenado.
func particionar(v []int, ini, fim int, c *Contador) int {
	meio := (ini + fim) / 2
	v[meio], v[fim] = v[fim], v[meio]
	pivo := v[fim]
	i := ini // fronteira: tudo antes de i é menor que o pivô
	for j := ini; j < fim; j++ {
		c.comparacoes++
		if v[j] < pivo {
			if i != j {
				v[i], v[j] = v[j], v[i]
				c.trocas++
			}
			i++
		}
	}
	v[i], v[fim] = v[fim], v[i]
	c.trocas++
	return i
}

// nomes fixa a ordem de exibição: percorrer um map em Go tem ordem aleatória.
var nomes = []string{"bubble", "insertion", "merge", "quick"}
var algoritmos = map[string]func([]int, *Contador) []int{
	"bubble":    bubbleSort,
	"insertion": insertionSort,
	"merge":     mergeSort,
	"quick":     quickSort,
}

// ordenar roda um algoritmo pelo nome, mede o tempo e confere o resultado.
func ordenar(nome string, v []int, mostrar bool) ([]int, Contador, float64) {
	c := Contador{mostrar: mostrar}
	inicio := time.Now()
	resultado := algoritmos[nome](v, &c)
	ms := float64(time.Since(inicio).Nanoseconds()) / 1e6
	if !slices.IsSorted(resultado) {
		fmt.Fprintf(os.Stderr, "erro interno: %s devolveu um vetor fora de ordem\n", nome)
		os.Exit(1)
	}
	return resultado, c, ms
}

// milhar formata 1234567 como 1.234.567, no padrão brasileiro.
func milhar(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return s
}

// tabela compara os quatro algoritmos sobre o mesmo vetor.
func tabela(v []int) {
	fmt.Printf("  %-10s %12s %12s %11s\n", "algoritmo", "comparações", "trocas/mov.", "tempo (ms)")
	for _, nome := range nomes {
		_, c, ms := ordenar(nome, v, false)
		fmt.Printf("  %-10s %12s %12s %11.2f\n", nome, milhar(c.comparacoes), milhar(c.trocas), ms)
	}
}

// demo mostra os passos do bubble, a tabela num vetor pequeno e num grande.
func demo() {
	vetor := []int{5, 2, 9, 1, 7, 3, 8}
	fmt.Println("=== Ordenação passo a passo: CodeNomad Academy ===")
	fmt.Printf("Vetor original: %v\n\nBubble sort, mostrando cada troca:\n", vetor)
	resultado, c, _ := ordenar("bubble", vetor, true)
	fmt.Printf("Resultado: %v  (%d comparações, %d trocas)\n\n", resultado, c.comparacoes, c.trocas)
	fmt.Println("Os quatro algoritmos no mesmo vetor pequeno:")
	tabela(vetor)
	r := rand.New(rand.NewSource(42)) // semente fixa: a demonstração é igual em qualquer máquina
	grande := r.Perm(20000)           // os números de 0 a 19.999 embaralhados
	fmt.Printf("\nAgora com %s números embaralhados:\n", milhar(len(grande)))
	tabela(grande)
	fmt.Println("\nLeitura: bubble e insertion crescem com n², merge e quick com n log n.\n" +
		"Com 10 vezes mais dados, os dois primeiros ficam 100 vezes mais lentos.")
}

func main() {
	algoritmo := flag.String("algoritmo", "todos", "bubble, insertion, merge, quick ou todos (tabela comparativa)")
	passos := flag.Bool("passos", false, "imprime o vetor a cada passo")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Uso: go run ./sources/21_ordenacao_visual [opções] n1 n2 n3 ...\n",
			"As opções vêm antes dos números: o pacote flag para no primeiro argumento sem hífen.\n")
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
	numeros := make([]int, 0, flag.NArg())
	for _, arg := range flag.Args() {
		n, err := strconv.Atoi(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "erro: '%s' não é um número inteiro\n", arg)
			os.Exit(2)
		}
		numeros = append(numeros, n)
	}
	if *algoritmo == "todos" {
		tabela(numeros)
		return
	}
	if _, ok := algoritmos[*algoritmo]; !ok {
		fmt.Fprintf(os.Stderr, "erro: algoritmo '%s' não existe; use bubble, insertion, merge, quick ou todos\n", *algoritmo)
		os.Exit(2)
	}
	resultado, c, ms := ordenar(*algoritmo, numeros, *passos)
	fmt.Printf("%s: %v\n", *algoritmo, resultado)
	fmt.Printf("%d comparações, %d trocas, %.2f ms\n", c.comparacoes, c.trocas, ms)
}

// ------------------------------------------------------------
// Para ir além
// 1. Implemente o selection sort e inclua na tabela comparativa.
// 2. Rode a tabela com um vetor já ordenado e com um invertido: qual algoritmo sofre mais? Por quê?
// 3. Adicione a opção --reverso, que ordena do maior para o menor mudando só as comparações.
// ------------------------------------------------------------
