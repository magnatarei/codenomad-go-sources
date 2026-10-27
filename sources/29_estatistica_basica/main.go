// ============================================================
// CodeNomad Academy  |  Sources Go  |  29. Estatística descritiva
// ------------------------------------------------------------
// O que faz: calcula média, mediana, moda, variância, desvio padrão, mínimo,
//            máximo, amplitude e quartis de uma lista de números vinda dos
//            argumentos ou de um arquivo, e desenha um histograma em texto.
// Como usar: go run ./sources/29_estatistica_basica 7.5 8 6.5 9 7 8.5 10 6
//            go run ./sources/29_estatistica_basica --arquivo notas.txt --classes 5
//            go run ./sources/29_estatistica_basica --demo
// O que você aprende: slices.Sort e slices.Clone, map para contar frequências,
//            math.Sqrt e math.Pow, bufio.Scanner para ler arquivo, os.TempDir,
//            strings.Fields, strings.Repeat para desenhar barras e fmt com largura.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Resumo guarda todas as medidas de uma amostra. Calcular tudo de uma vez e
// devolver um struct é mais simples do que dez funções que ordenam de novo.
type Resumo struct {
	n                           int
	media, mediana              float64
	modas                       []float64
	varPop, varAmostral, desvio float64
	minimo, maximo, amplitude   float64
	q1, q3                      float64
}

// mediana de um slice JÁ ORDENADO: o do meio, ou a média dos dois do meio.
func mediana(ordenado []float64) float64 {
	n := len(ordenado)
	if n%2 == 1 {
		return ordenado[n/2]
	}
	return (ordenado[n/2-1] + ordenado[n/2]) / 2
}

// calcular ordena uma cópia (para não bagunçar o slice de quem chamou) e
// extrai cada medida. Os quartis são as medianas das duas metades.
func calcular(valores []float64) Resumo {
	v := slices.Clone(valores)
	slices.Sort(v)
	n := len(v)
	r := Resumo{n: n, minimo: v[0], maximo: v[n-1], mediana: mediana(v)}
	r.amplitude = r.maximo - r.minimo
	soma := 0.0
	for _, x := range v {
		soma += x
	}
	r.media = soma / float64(n)
	somaQuad := 0.0
	for _, x := range v {
		somaQuad += math.Pow(x-r.media, 2)
	}
	r.varPop = somaQuad / float64(n)
	if n > 1 {
		r.varAmostral = somaQuad / float64(n-1) // n-1: correção de Bessel para amostras
	}
	r.desvio = math.Sqrt(r.varAmostral)
	// Quartis: a metade de baixo vai até antes do meio; a de cima começa depois dele.
	r.q1, r.q3 = mediana(v[:n/2]), mediana(v[(n+1)/2:])
	if n < 4 {
		r.q1, r.q3 = r.minimo, r.maximo // com menos de 4 valores as metades ficam vazias
	}
	// Moda: o valor (ou valores) que mais se repete. Se nenhum repete, não há moda.
	contagem := map[float64]int{}
	maior := 0
	for _, x := range v {
		contagem[x]++
		maior = max(maior, contagem[x])
	}
	if maior > 1 {
		for _, x := range v {
			if contagem[x] == maior && !slices.Contains(r.modas, x) {
				r.modas = append(r.modas, x) // v está ordenado, então as modas saem em ordem
			}
		}
	}
	return r
}

// histograma divide o intervalo [mínimo, máximo] em classes de mesma largura
// e desenha uma barra proporcional à contagem de cada classe.
func histograma(valores []float64, classes int) string {
	r := calcular(valores)
	largura := r.amplitude / float64(classes)
	if largura == 0 {
		return fmt.Sprintf("  todos os %d valores são iguais a %g: não há o que distribuir\n", r.n, r.minimo)
	}
	contagem := make([]int, classes)
	for _, x := range valores {
		i := int((x - r.minimo) / largura)
		if i == classes {
			i-- // o máximo cairia na classe seguinte, que não existe: entra na última
		}
		contagem[i]++
	}
	maior := slices.Max(contagem)
	var sb strings.Builder
	for i, c := range contagem {
		ini := r.minimo + float64(i)*largura
		barra := strings.Repeat("#", c*40/maior) // a maior barra tem 40 caracteres
		fmt.Fprintf(&sb, "  %8.2f a %8.2f  %-40s %d\n", ini, ini+largura, barra, c)
	}
	return sb.String()
}

// relatorio imprime as medidas e o histograma de forma alinhada.
func relatorio(valores []float64, classes int) {
	r := calcular(valores)
	linha := func(nome string, v float64) { fmt.Printf("  %-22s %10.3f\n", nome, v) }
	fmt.Printf("  %-22s %10d\n", "quantidade", r.n)
	linha("média", r.media)
	linha("mediana", r.mediana)
	if len(r.modas) == 0 {
		fmt.Printf("  %-22s %10s\n", "moda", "nenhuma")
	} else {
		fmt.Printf("  %-22s %10s\n", "moda", strings.Trim(fmt.Sprint(r.modas), "[]"))
	}
	linha("variância (população)", r.varPop)
	linha("variância (amostra)", r.varAmostral)
	linha("desvio padrão", r.desvio)
	linha("mínimo", r.minimo)
	linha("1º quartil (Q1)", r.q1)
	linha("3º quartil (Q3)", r.q3)
	linha("máximo", r.maximo)
	linha("amplitude", r.amplitude)
	fmt.Printf("\n  Histograma em %d classes:\n", classes)
	fmt.Print(histograma(valores, classes))
}

// lerNumeros converte textos em float64, aceitando vírgula no lugar do ponto (7,5).
func lerNumeros(campos []string) ([]float64, error) {
	var valores []float64
	for _, c := range campos {
		x, err := strconv.ParseFloat(strings.ReplaceAll(c, ",", "."), 64)
		if err != nil {
			return nil, fmt.Errorf("'%s' não é um número", c)
		}
		valores = append(valores, x)
	}
	return valores, nil
}

// lerArquivo lê números separados por espaço, vírgula ou quebra de linha.
func lerArquivo(caminho string) ([]float64, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, err // o erro do sistema já diz "não existe" ou "sem permissão"
	}
	defer f.Close()
	var campos []string
	leitor := bufio.NewScanner(f)
	for leitor.Scan() {
		linha := strings.TrimSpace(leitor.Text())
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue // linhas vazias e comentários são ignorados
		}
		campos = append(campos, strings.Fields(strings.ReplaceAll(linha, ";", " "))...)
	}
	if err := leitor.Err(); err != nil {
		return nil, err
	}
	return lerNumeros(campos)
}

func demo() {
	fmt.Println("=== Estatística descritiva: CodeNomad Academy ===")
	// Tempos (em minutos) que 40 alunos levaram para terminar um exercício.
	amostra := []float64{12, 15, 9, 22, 18, 15, 11, 30, 14, 16, 19, 15, 8, 21, 17, 13, 25, 15, 10, 20,
		16, 12, 28, 14, 19, 15, 11, 23, 17, 13, 9, 26, 18, 14, 16, 12, 20, 15, 35, 17}
	fmt.Printf("\nAmostra de %d tempos (min): %v\n\n", len(amostra), amostra)
	relatorio(amostra, 6)
	fmt.Println("\nLeitura: média 16,9 acima da mediana 15,5 e cauda à direita no histograma:")
	fmt.Println("poucos alunos demoraram muito (30 e 35 min) e puxaram a média para cima.")
	// A mesma amostra lida de um arquivo na pasta temporária, para mostrar o --arquivo.
	caminho := filepath.Join(os.TempDir(), "codenomad_29_amostra.txt")
	conteudo := "# tempos em minutos\n12 15 9 22 18\n15, 11, 30\n"
	if err := os.WriteFile(caminho, []byte(conteudo), 0o644); err != nil {
		fmt.Println("não consegui escrever o arquivo de exemplo:", err)
		return
	}
	defer os.Remove(caminho)
	valores, err := lerArquivo(caminho)
	if err != nil {
		fmt.Println("erro ao ler:", err)
		return
	}
	fmt.Printf("\nLidos %d valores do arquivo %s: %v\n", len(valores), filepath.Base(caminho), valores)
	r := calcular(valores)
	fmt.Printf("  média %.2f, mediana %.2f, desvio %.2f\n", r.media, r.mediana, r.desvio)
}

func main() {
	arquivo := flag.String("arquivo", "", "arquivo com números separados por espaço, vírgula ou linha")
	classes := flag.Int("classes", 6, "número de classes do histograma (2 a 30)")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Uso: go run ./sources/29_estatistica_basica [--classes N] n1 n2 n3 ...")
		fmt.Fprintln(os.Stderr, "     go run ./sources/29_estatistica_basica [--classes N] --arquivo dados.txt")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	if *classes < 2 || *classes > 30 {
		fmt.Fprintln(os.Stderr, "erro: --classes deve ficar entre 2 e 30")
		os.Exit(2)
	}
	var valores []float64
	var err error
	switch {
	case *arquivo != "":
		valores, err = lerArquivo(*arquivo)
	case flag.NArg() > 0:
		valores, err = lerNumeros(flag.Args())
	default:
		flag.Usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	if len(valores) < 2 {
		fmt.Fprintln(os.Stderr, "erro: informe pelo menos 2 números")
		os.Exit(1)
	}
	relatorio(valores, *classes)
}

// ------------------------------------------------------------
// Para ir além
// 1. Calcule a assimetria (skewness) e diga se a cauda está à esquerda ou à direita.
// 2. Detecte outliers pela regra 1,5 x IQR (IQR = Q3 - Q1) e liste-os separados.
// 3. Leia um CSV com encoding/csv e escolha a coluna pelo nome do cabeçalho.
// ------------------------------------------------------------
