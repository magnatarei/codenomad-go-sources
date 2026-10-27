// ============================================================
// CodeNomad Academy  |  Sources Go  |  28. Operações com matrizes
// ------------------------------------------------------------
// O que faz: soma, subtração, multiplicação, transposta, determinante (recursivo,
//            por Laplace) e identidade, com validação de dimensões e impressão alinhada.
// Como usar: go run ./sources/28_matriz_operacoes --op mult --a "1,2;3,4" --b "5,6;7,8"
//            go run ./sources/28_matriz_operacoes --op det --a "2,0,1;1,3,2;1,1,4"
//            go run ./sources/28_matriz_operacoes --demo
// O que você aprende: tipo nomeado sobre [][]float64 com métodos, slices de slices, laços
//            aninhados, recursão, método como valor num map, slices.Concat, fmt.Errorf com %w.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Matriz [][]float64

func Nova(linhas, colunas int) Matriz {
	m := make(Matriz, linhas)
	for i := range m {
		m[i] = make([]float64, colunas) // cada linha precisa do seu próprio make
	}
	return m
}

// Identidade tem 1 na diagonal e 0 no resto: é o "1" da multiplicação de matrizes.
func Identidade(n int) Matriz {
	m := Nova(n, n)
	for i := range m {
		m[i][i] = 1
	}
	return m
}

func (m Matriz) forma() string { return fmt.Sprintf("%dx%d", len(m), len(m[0])) }

// combinar aplica f célula a célula (soma e subtração só mudam a f); exige as mesmas dimensões.
func (m Matriz) combinar(o Matriz, f func(a, b float64) float64) (Matriz, error) {
	if len(m) != len(o) || len(m[0]) != len(o[0]) {
		return nil, fmt.Errorf("dimensões diferentes: %s e %s", m.forma(), o.forma())
	}
	r := Nova(len(m), len(m[0]))
	for i := range m {
		for j := range m[i] {
			r[i][j] = f(m[i][j], o[i][j])
		}
	}
	return r, nil
}

// Multiplicar exige colunas de m == linhas de o. Cada célula é o produto escalar de uma linha por uma coluna.
func (m Matriz) Multiplicar(o Matriz) (Matriz, error) {
	if len(m[0]) != len(o) {
		return nil, fmt.Errorf("não dá para multiplicar %s por %s: colunas da primeira devem ser iguais às linhas da segunda", m.forma(), o.forma())
	}
	r := Nova(len(m), len(o[0]))
	for i := range r {
		for j := range r[i] {
			for k := range o {
				r[i][j] += m[i][k] * o[k][j]
			}
		}
	}
	return r, nil
}

// Transposta troca linhas por colunas: a célula [i][j] vai para [j][i].
func (m Matriz) Transposta() Matriz {
	r := Nova(len(m[0]), len(m))
	for i := range m {
		for j := range m[i] {
			r[j][i] = m[i][j]
		}
	}
	return r
}

// Determinante usa a expansão de Laplace pela primeira linha: soma, para cada coluna j,
// (-1)^j * m[0][j] * det(matriz sem a linha 0 e a coluna j). Caso base: 1x1. Custa n!.
func (m Matriz) Determinante() (float64, error) {
	if len(m) != len(m[0]) {
		return 0, fmt.Errorf("determinante só existe em matriz quadrada (esta é %s)", m.forma())
	}
	if len(m) == 1 {
		return m[0][0], nil
	}
	total, sinal := 0.0, 1.0
	for j := range m[0] {
		d, _ := m.menor(0, j).Determinante() // o menor de uma quadrada é quadrado: o erro é sempre nil
		total += sinal * m[0][j] * d
		sinal = -sinal // os sinais alternam: + - + - ...
	}
	return total, nil
}

func (m Matriz) menor(l, c int) Matriz {
	var r Matriz
	for i, linha := range m {
		if i != l {
			r = append(r, slices.Concat(linha[:c], linha[c+1:])) // Concat copia: a linha original fica intacta
		}
	}
	return r
}

func texto(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// String imprime cada coluna com a largura do texto mais longo, calculada antes.
func (m Matriz) String() string {
	largura := 1
	for _, linha := range m {
		for _, v := range linha {
			largura = max(largura, len(texto(v)))
		}
	}
	var sb strings.Builder
	for _, linha := range m {
		sb.WriteString("  |")
		for _, v := range linha {
			fmt.Fprintf(&sb, " %*s", largura, texto(v)) // %*s: a largura vem como argumento
		}
		sb.WriteString(" |\n")
	}
	return sb.String()
}

// lerMatriz converte "1,2;3,4" em Matriz: ponto e vírgula separa linhas, vírgula separa colunas.
func lerMatriz(campo string) (Matriz, error) {
	var m Matriz
	for i, linha := range strings.Split(strings.TrimSpace(campo), ";") {
		var valores []float64
		for _, celula := range strings.Split(linha, ",") {
			v, err := strconv.ParseFloat(strings.TrimSpace(celula), 64)
			if err != nil {
				return nil, fmt.Errorf("'%s' não é um número (linha %d)", strings.TrimSpace(celula), i+1)
			}
			valores = append(valores, v)
		}
		if i > 0 && len(valores) != len(m[0]) {
			return nil, fmt.Errorf("linha %d tem %d valores, mas a primeira tem %d", i+1, len(valores), len(m[0]))
		}
		m = append(m, valores)
	}
	return m, nil
}

// binarias liga o nome à função. Matriz.Multiplicar sem receptor é uma "expressão de método".
var binarias = map[string]func(Matriz, Matriz) (Matriz, error){
	"soma": func(a, b Matriz) (Matriz, error) { return a.combinar(b, func(x, y float64) float64 { return x + y }) },
	"sub":  func(a, b Matriz) (Matriz, error) { return a.combinar(b, func(x, y float64) float64 { return x - y }) },
	"mult": Matriz.Multiplicar,
}

// executar lê as matrizes dos textos, roda a operação e devolve o texto do resultado.
func executar(op, textoA, textoB string, n int) (string, error) {
	if op == "identidade" {
		if n < 1 || n > 20 {
			return "", fmt.Errorf("--n deve ficar entre 1 e 20 (recebi %d)", n)
		}
		return Identidade(n).String(), nil
	}
	if textoA == "" {
		return "", fmt.Errorf("a operação %s precisa da primeira matriz em --a", op)
	}
	a, err := lerMatriz(textoA)
	if err != nil {
		return "", fmt.Errorf("em --a: %w", err) // %w embrulha o erro original, que continua inspecionável
	}
	if f, ok := binarias[op]; ok {
		if textoB == "" {
			return "", fmt.Errorf("a operação %s precisa da segunda matriz em --b", op)
		}
		b, err := lerMatriz(textoB)
		if err != nil {
			return "", fmt.Errorf("em --b: %w", err)
		}
		r, err := f(a, b)
		if err != nil {
			return "", err
		}
		return r.String(), nil
	}
	switch op {
	case "transposta":
		return a.Transposta().String(), nil
	case "det":
		d, err := a.Determinante()
		if err != nil {
			return "", err
		}
		return "  det = " + texto(d) + "\n", nil
	}
	return "", fmt.Errorf("operação desconhecida '%s' (use soma, sub, mult, transposta, det ou identidade)", op)
}

func demo() {
	fmt.Println("=== Operações com matrizes: CodeNomad Academy ===")
	tA, tB, tR := "1,2;3,4", "5,6;7,8", "1,2,3;4,5,6" // R é 2x3: serve para a transposta e para os erros
	a, _ := lerMatriz(tA)
	b, _ := lerMatriz(tB)
	c, _ := lerMatriz("2,0,1;1,3,2;1,1,4")
	r, _ := lerMatriz(tR)
	fmt.Printf("\nA =\n%sB =\n%s", a, b)
	for _, op := range []string{"soma", "sub", "mult"} {
		saida, _ := executar(op, tA, tB, 0)
		fmt.Printf("%s(A, B) =\n%s", op, saida)
	}
	dA, _ := a.Determinante()
	dC, _ := c.Determinante()
	fmt.Printf("det(A) = %s (1*4 - 2*3)\n", texto(dA))
	fmt.Printf("\nC (3x3) =\n%stransposta(C) =\n%sdet(C) = %s\n", c, c.Transposta(), texto(dC))
	fmt.Printf("\nR (2x3) =\n%s", r)
	fmt.Printf("\nIdentidade 3x3:\n%s", Identidade(3))
	fmt.Println("\nErros tratados sem derrubar o programa, com R (2x3) e A (2x2):")
	for _, op := range []string{"soma", "mult", "det"} {
		_, err := executar(op, tR, tA, 0)
		fmt.Printf("  %-6s -> %v\n", op, err)
	}
}

func main() {
	op := flag.String("op", "", "soma, sub, mult, transposta, det ou identidade")
	a := flag.String("a", "", "primeira matriz, como \"1,2;3,4\" (vírgula entre colunas, ponto e vírgula entre linhas)")
	b := flag.String("b", "", "segunda matriz, para soma, sub e mult")
	n := flag.Int("n", 3, "tamanho da identidade")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	if *op == "" {
		fmt.Fprintln(os.Stderr, "Uso: go run ./sources/28_matriz_operacoes --op OPERAÇÃO --a \"1,2;3,4\" [--b \"5,6;7,8\"] [--n 4]")
		flag.PrintDefaults()
		os.Exit(2)
	}
	saida, err := executar(*op, *a, *b, *n)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	fmt.Print(saida)
}

// ------------------------------------------------------------
// Para ir além
// 1. Implemente a inversa por Gauss-Jordan e confira que A vezes inversa(A) é a identidade.
// 2. Calcule o determinante por eliminação (O(n³)) e compare o tempo com Laplace numa 10x10.
// 3. Leia a matriz de um arquivo CSV com encoding/csv em vez do texto "1,2;3,4".
// ------------------------------------------------------------
