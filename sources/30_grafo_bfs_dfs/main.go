// ============================================================
// CodeNomad Academy  |  Sources Go  |  30. Grafos: BFS e DFS num labirinto
// ------------------------------------------------------------
// O que faz: lê um labirinto em texto (# parede, . livre, S início, E chegada),
//            trata cada célula livre como um vértice ligado aos vizinhos, acha o
//            caminho mais curto com BFS, explora com DFS e desenha o caminho.
// Como usar: go run ./sources/30_grafo_bfs_dfs --arquivo labirinto.txt
//            go run ./sources/30_grafo_bfs_dfs --arquivo labirinto.txt --dfs
//            go run ./sources/30_grafo_bfs_dfs --demo
// O que você aprende: struct como chave de map (Ponto), fila com slice para a
//            BFS, pilha com slice para a DFS, map de "pai" para reconstruir o
//            caminho, [][]rune para a grade, bufio.Scanner e strings.Split.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Ponto é uma célula da grade. Como é um struct de valores comparáveis, pode
// ser chave de map: é assim que marcamos visitados sem uma matriz paralela.
type Ponto struct{ linha, coluna int }

// Labirinto é a grade de runas mais as posições de início e fim.
type Labirinto struct {
	grade       [][]rune
	inicio, fim Ponto
}

var direcoes = []Ponto{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

// lerLabirinto valida o texto: linhas do mesmo tamanho, um S, um E e só símbolos permitidos.
func lerLabirinto(texto string) (*Labirinto, error) {
	lab := &Labirinto{inicio: Ponto{-1, -1}, fim: Ponto{-1, -1}}
	for l, linha := range strings.Split(strings.TrimRight(texto, "\n"), "\n") {
		linha = strings.TrimRight(linha, "\r") // arquivos salvos no Windows terminam a linha com \r\n
		runas := []rune(linha)
		if l > 0 && len(runas) != len(lab.grade[0]) {
			return nil, fmt.Errorf("a linha %d tem %d colunas, mas a primeira tem %d", l+1, len(runas), len(lab.grade[0]))
		}
		for c, r := range runas {
			switch r {
			case 'S':
				lab.inicio = Ponto{l, c}
			case 'E':
				lab.fim = Ponto{l, c}
			case '#', '.', ' ':
			default:
				return nil, fmt.Errorf("caractere inválido %q na linha %d, coluna %d", r, l+1, c+1)
			}
		}
		lab.grade = append(lab.grade, runas)
	}
	if lab.inicio.linha < 0 || lab.fim.linha < 0 {
		return nil, fmt.Errorf("o labirinto precisa de um S (início) e um E (chegada)")
	}
	return lab, nil
}

// livre diz se o ponto está dentro da grade e não é parede.
func (lab *Labirinto) livre(p Ponto) bool {
	return p.linha >= 0 && p.linha < len(lab.grade) &&
		p.coluna >= 0 && p.coluna < len(lab.grade[0]) &&
		lab.grade[p.linha][p.coluna] != '#'
}

// vizinhos devolve as células livres ao redor de p: são as "arestas" do grafo.
func (lab *Labirinto) vizinhos(p Ponto) []Ponto {
	var vs []Ponto
	for _, d := range direcoes {
		q := Ponto{p.linha + d.linha, p.coluna + d.coluna}
		if lab.livre(q) {
			vs = append(vs, q)
		}
	}
	return vs
}

// BFS explora em camadas: primeiro todos a 1 passo, depois a 2, e assim por
// diante. Por isso o primeiro jeito de chegar ao fim é sempre o mais curto.
// Devolve o caminho (nil se não existe) e quantas células foram visitadas.
func (lab *Labirinto) BFS() ([]Ponto, int) {
	pai := map[Ponto]Ponto{lab.inicio: lab.inicio} // também serve de "visitado"
	fila := []Ponto{lab.inicio}
	for len(fila) > 0 {
		atual := fila[0]
		fila = fila[1:] // tira da frente: FIFO
		if atual == lab.fim {
			return reconstruir(pai, lab.fim), len(pai)
		}
		for _, v := range lab.vizinhos(atual) {
			if _, visto := pai[v]; !visto {
				pai[v] = atual
				fila = append(fila, v)
			}
		}
	}
	return nil, len(pai)
}

// DFS mergulha por um corredor até o fim antes de voltar e tentar outro: acha um caminho,
// não necessariamente o mais curto. Só muda tirar do FIM do slice (pilha), não da frente.
func (lab *Labirinto) DFS() ([]Ponto, int) {
	pai := map[Ponto]Ponto{lab.inicio: lab.inicio}
	pilha := []Ponto{lab.inicio}
	for len(pilha) > 0 {
		atual := pilha[len(pilha)-1]
		pilha = pilha[:len(pilha)-1] // tira do fim: LIFO
		if atual == lab.fim {
			return reconstruir(pai, lab.fim), len(pai)
		}
		for _, v := range lab.vizinhos(atual) {
			if _, visto := pai[v]; !visto {
				pai[v] = atual
				pilha = append(pilha, v)
			}
		}
	}
	return nil, len(pai)
}

// reconstruir volta do fim ao início seguindo os pais e inverte a lista.
func reconstruir(pai map[Ponto]Ponto, fim Ponto) []Ponto {
	var caminho []Ponto
	for p := fim; ; p = pai[p] {
		caminho = append(caminho, p)
		if pai[p] == p {
			break // o início é pai de si mesmo: chegamos
		}
	}
	for i, j := 0, len(caminho)-1; i < j; i, j = i+1, j-1 {
		caminho[i], caminho[j] = caminho[j], caminho[i]
	}
	return caminho
}

// desenhar devolve o labirinto com o caminho marcado com 'o', sem alterar a grade original.
func (lab *Labirinto) desenhar(caminho []Ponto) string {
	var sb strings.Builder
	noCaminho := map[Ponto]bool{}
	for _, p := range caminho {
		noCaminho[p] = true
	}
	for l, linha := range lab.grade {
		sb.WriteString("  ")
		for c, r := range linha {
			if noCaminho[Ponto{l, c}] && r != 'S' && r != 'E' {
				r = 'o'
			}
			sb.WriteRune(r)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// resolver roda a busca pedida, mostra o resultado e explica o que aconteceu.
func resolver(lab *Labirinto, usarDFS bool) {
	nome, busca := "BFS (largura)", lab.BFS
	if usarDFS {
		nome, busca = "DFS (profundidade)", lab.DFS
	}
	caminho, visitadas := busca()
	if caminho == nil {
		fmt.Printf("%s: não existe caminho de S até E (%d células exploradas)\n", nome, visitadas)
		return
	}
	fmt.Printf("%s: caminho com %d passos, %d células exploradas\n", nome, len(caminho)-1, visitadas)
	fmt.Print(lab.desenhar(caminho))
}

const labirintoDemo = `##############
#S...#.......#
#.##.#.#####.#
#.#..#.#...#.#
#.#.##.#.#.#.#
#.#....#.#...#
#.######.###.#
#........#..E#
##############`

const semSaida = `#########
#S..#...#
#.#.#.#.#
#.#...###
#.#####E#
#########`

func demo() {
	fmt.Println("=== Grafos: BFS e DFS num labirinto: CodeNomad Academy ===")
	fmt.Println("Cada célula livre é um vértice; vizinhos de cima, baixo, esquerda e direita são arestas.")
	lab, err := lerLabirinto(labirintoDemo)
	if err != nil {
		fmt.Println("erro no labirinto embutido:", err)
		return
	}
	fmt.Println("\nLabirinto original:")
	fmt.Print(lab.desenhar(nil))
	fmt.Println()
	resolver(lab, false)
	fmt.Println()
	resolver(lab, true)
	fmt.Println("\nLeitura: a BFS garante o caminho mais curto; a DFS acha um caminho qualquer, às vezes mais longo.")
	bloqueado, _ := lerLabirinto(semSaida)
	fmt.Println("\nLabirinto sem saída (o E está cercado de paredes):")
	fmt.Print(bloqueado.desenhar(nil))
	resolver(bloqueado, false)
	fmt.Println("\nErro de validação tratado:")
	if _, err := lerLabirinto("#####\n#S.X#\n#####"); err != nil {
		fmt.Println("  ", err)
	}
}

func main() {
	arquivo := flag.String("arquivo", "", "arquivo de texto com o labirinto (# parede, . livre, S início, E chegada)")
	dfs := flag.Bool("dfs", false, "usa busca em profundidade em vez da busca em largura")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	if *arquivo == "" {
		fmt.Fprintln(os.Stderr, "Uso: go run ./sources/30_grafo_bfs_dfs --arquivo labirinto.txt [--dfs]")
		flag.PrintDefaults()
		os.Exit(2)
	}
	f, err := os.Open(*arquivo)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	defer f.Close()
	var linhas []string
	leitor := bufio.NewScanner(f)
	for leitor.Scan() {
		linhas = append(linhas, leitor.Text())
	}
	lab, err := lerLabirinto(strings.Join(linhas, "\n"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	resolver(lab, *dfs)
}

// ------------------------------------------------------------
// Para ir além
// 1. Permita andar na diagonal (8 vizinhos) e veja como o caminho mais curto muda.
// 2. Dê pesos às células (areia custa 3, grama 1) e troque a BFS por Dijkstra com o heap do source 25.
// 3. Gere labirintos aleatórios com DFS recursiva (derrubando paredes) e resolva-os com a BFS.
// ------------------------------------------------------------
