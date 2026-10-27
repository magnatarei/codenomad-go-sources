// ============================================================
// CodeNomad Academy  |  Sources Go  |  25. Fila de prioridade com heap
// ------------------------------------------------------------
// O que faz: implementa um heap binário mínimo do zero (inserir, extrair,
//            espiar) e usa esse heap numa simulação de pronto-socorro em que
//            os pacientes são atendidos pela gravidade, não pela ordem de chegada.
// Como usar: go run ./sources/25_fila_prioridade Ana:3 Bruno:1 Carla:4 Diego:2
//            go run ./sources/25_fila_prioridade --consulta 8 Ana:3 Bruno:1 Carla:3
//            go run ./sources/25_fila_prioridade --demo
// O que você aprende: heap guardado em um slice (filho em 2i+1 e 2i+2, pai em
//            (i-1)/2), métodos com ponteiro, o padrão "valor, ok", strings.Cut,
//            desempate estável e a diferença entre fila comum e fila de prioridade.
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
)

// Paciente é o que entra na fila. gravidade vai de 1 (risco de vida) a 5
// (caso leve); chegada é a ordem em que a pessoa deu entrada e serve de desempate.
type Paciente struct {
	nome      string
	gravidade int
	chegada   int
}

// nomesGravidade dá um rótulo legível para cada nível.
var nomesGravidade = map[int]string{1: "emergência", 2: "muito urgente", 3: "urgente", 4: "pouco urgente", 5: "não urgente"}

// antes diz se p deve ser atendido antes de q: gravidade menor vence e, em
// empate, quem chegou primeiro. Sem o desempate, dois pacientes de mesma
// gravidade sairiam em ordem imprevisível.
func (p Paciente) antes(q Paciente) bool {
	if p.gravidade != q.gravidade {
		return p.gravidade < q.gravidade
	}
	return p.chegada < q.chegada
}

// HeapMin guarda a árvore num slice. O menor elemento fica sempre em itens[0].
// Os filhos do índice i estão em 2i+1 e 2i+2; o pai de i está em (i-1)/2.
// Não há ponteiros: a posição no slice já diz quem é pai de quem.
type HeapMin struct {
	itens []Paciente
}

// Inserir coloca no fim e "sobe" o item enquanto ele for menor que o pai.
// Custa O(log n), porque a árvore tem altura log n.
func (h *HeapMin) Inserir(p Paciente) {
	h.itens = append(h.itens, p)
	h.subir(len(h.itens) - 1)
}

func (h *HeapMin) subir(i int) {
	for i > 0 {
		pai := (i - 1) / 2
		if !h.itens[i].antes(h.itens[pai]) {
			return // o pai já é menor ou igual: a propriedade do heap vale
		}
		h.itens[i], h.itens[pai] = h.itens[pai], h.itens[i]
		i = pai
	}
}

// Extrair devolve o menor (a raiz). O último item vai para a raiz e "desce"
// até o lugar certo: assim a árvore continua completa e o custo é O(log n).
func (h *HeapMin) Extrair() (Paciente, bool) {
	if len(h.itens) == 0 {
		return Paciente{}, false
	}
	menor := h.itens[0]
	ultimo := len(h.itens) - 1
	h.itens[0] = h.itens[ultimo]
	h.itens = h.itens[:ultimo]
	h.descer(0)
	return menor, true
}

func (h *HeapMin) descer(i int) {
	n := len(h.itens)
	for {
		esq, dir, menor := 2*i+1, 2*i+2, i
		if esq < n && h.itens[esq].antes(h.itens[menor]) {
			menor = esq
		}
		if dir < n && h.itens[dir].antes(h.itens[menor]) {
			menor = dir
		}
		if menor == i {
			return // nenhum filho é menor: chegou ao lugar certo
		}
		h.itens[i], h.itens[menor] = h.itens[menor], h.itens[i]
		i = menor
	}
}

// Espiar mostra quem é o próximo sem tirar da fila.
func (h *HeapMin) Espiar() (Paciente, bool) {
	if len(h.itens) == 0 {
		return Paciente{}, false
	}
	return h.itens[0], true
}

func (h *HeapMin) Tamanho() int { return len(h.itens) }

// String mostra o slice interno como "nome(gravidade)": o aluno vê que o
// heap não é um vetor ordenado, só garante o menor na primeira posição.
func (h *HeapMin) String() string {
	partes := make([]string, len(h.itens))
	for i, p := range h.itens {
		partes[i] = fmt.Sprintf("%s(%d)", p.nome, p.gravidade)
	}
	return "[" + strings.Join(partes, " ") + "]"
}

// lerPacientes converte "Ana:3" em Paciente, com erro claro para cada formato ruim.
func lerPacientes(args []string) ([]Paciente, error) {
	var lista []Paciente
	for i, arg := range args {
		nome, g, achou := strings.Cut(arg, ":")
		gravidade, err := strconv.Atoi(g)
		if !achou || nome == "" || err != nil {
			return nil, fmt.Errorf("'%s' não está no formato nome:gravidade", arg)
		}
		if gravidade < 1 || gravidade > 5 {
			return nil, fmt.Errorf("gravidade de %s deve ficar entre 1 (emergência) e 5 (não urgente), recebi %d", nome, gravidade)
		}
		lista = append(lista, Paciente{nome, gravidade, i + 1})
	}
	return lista, nil
}

// simular dá entrada em todos, mostra o heap crescendo e atende na ordem do
// heap, com um único médico que leva "consulta" minutos por paciente.
func simular(pacientes []Paciente, consulta int, mostrarHeap bool) {
	var fila HeapMin
	fmt.Println("Entrada dos pacientes, na ordem de chegada:")
	for _, p := range pacientes {
		fila.Inserir(p)
		if mostrarHeap {
			fmt.Printf("  chega %-7s gravidade %d   heap: %v\n", p.nome, p.gravidade, &fila)
		} else {
			fmt.Printf("  chega %-7s gravidade %d (%s)\n", p.nome, p.gravidade, nomesGravidade[p.gravidade])
		}
	}
	if proximo, ok := fila.Espiar(); ok {
		fmt.Printf("\nEspiar: o próximo é %s, sem tirar ninguém da fila (%d na espera)\n", proximo.nome, fila.Tamanho())
	}
	fmt.Printf("\nAtendimento (cada consulta leva %d min):\n", consulta)
	fmt.Printf("  %-3s %-8s %-16s %-8s %s\n", "#", "nome", "gravidade", "chegou", "esperou")
	relogio, somaEspera := 0, 0
	for ordem := 1; fila.Tamanho() > 0; ordem++ {
		p, _ := fila.Extrair() // o laço garante que há alguém, por isso ignoramos o ok
		somaEspera += relogio
		fmt.Printf("  %-3d %-8s %-16s %-8s %d min\n", ordem, p.nome,
			fmt.Sprintf("%d %s", p.gravidade, nomesGravidade[p.gravidade]), fmt.Sprintf("%dº", p.chegada), relogio)
		relogio += consulta
	}
	fmt.Printf("  espera média: %.1f min\n", float64(somaEspera)/float64(len(pacientes)))
}

func demo() {
	fmt.Println("=== Fila de prioridade com heap: CodeNomad Academy ===")
	fmt.Println("Gravidade: 1 = emergência ... 5 = não urgente. Menor número sai primeiro.")
	fmt.Println()
	pacientes, _ := lerPacientes([]string{"Ana:4", "Bruno:2", "Carla:5", "Diego:1", "Elisa:3", "Fábio:2", "Gabi:4", "Hugo:1"})
	simular(pacientes, 10, true)
	fmt.Println("\nRepare: Hugo chegou por último e foi atendido em 2º, porque gravidade 1 vence.")
	fmt.Println("Bruno e Fábio têm a mesma gravidade; Bruno saiu antes porque chegou antes (desempate).")
	fmt.Println("\nSe fosse uma fila comum (ordem de chegada), Diego e Hugo esperariam:")
	for i, p := range pacientes {
		if p.gravidade == 1 {
			fmt.Printf("  %-7s gravidade 1, seria o %dº atendido, após %d min de espera\n", p.nome, i+1, i*10)
		}
	}
	fmt.Println("\nLeitura: o heap não ordena a fila inteira; ele só garante que o menor está\n" +
		"na raiz, e repara a árvore em O(log n) a cada entrada ou saída.")
}

func main() {
	consulta := flag.Int("consulta", 10, "minutos que cada consulta leva")
	heap := flag.Bool("heap", false, "mostra o slice interno do heap a cada entrada")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Uso: go run ./sources/25_fila_prioridade [--consulta MIN] [--heap] nome:gravidade ...")
		fmt.Fprintln(os.Stderr, "A gravidade vai de 1 (emergência) a 5 (não urgente). As opções vêm antes dos pacientes.")
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
	if *consulta < 1 || *consulta > 600 {
		fmt.Fprintln(os.Stderr, "erro: --consulta deve ficar entre 1 e 600 minutos")
		os.Exit(2)
	}
	pacientes, err := lerPacientes(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(2)
	}
	simular(pacientes, *consulta, *heap)
}

// ------------------------------------------------------------
// Para ir além
// 1. Troque o critério de antes() para um heap máximo e use-o num ranking de pontos.
// 2. Implemente o heap sort: insira todos os números e extraia um a um. Compare com slices.Sort.
// 3. Dê a cada paciente um horário de chegada real e faça a gravidade subir
//    de nível quando a espera passar de 30 minutos (envelhecimento da fila).
// ------------------------------------------------------------
