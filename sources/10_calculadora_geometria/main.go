// ============================================================
// CodeNomad Academy  |  Sources Go  |  10. Calculadora de geometria
// ------------------------------------------------------------
// O que faz: calcula área e perímetro de quadrado, retângulo, triângulo
//            (conferindo se os lados formam um triângulo), círculo e
//            trapézio, e volume e área da superfície de cubo, esfera e cilindro.
// Como usar: go run ./sources/10_calculadora_geometria -forma triangulo -medidas 3,4,5
//            go run ./sources/10_calculadora_geometria -forma cilindro -medidas 2,10
//            go run ./sources/10_calculadora_geometria -listar
//            go run ./sources/10_calculadora_geometria --demo
// O que você aprende: interfaces (uma para figuras planas, outra para
//            sólidos), structs com métodos, type switch, um map de
//            "receitas" com funções como valores, math.Sqrt e math.Pi,
//            fórmula de Heron, strings.Split com strconv.ParseFloat e
//            validação que devolve error em vez de derrubar o programa.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Duas interfaces pequenas: quem tem Area e Perimetro é uma figura plana,
// quem tem Volume e Superficie é um sólido. O resto do programa só fala
// com as interfaces, nunca com os tipos concretos.
type plana interface {
	Area() float64
	Perimetro() float64
}

type solido interface {
	Volume() float64
	Superficie() float64
}

type quadrado struct{ lado float64 }
type retangulo struct{ base, altura float64 }
type triangulo struct{ a, b, c float64 }
type circulo struct{ raio float64 }
type trapezio struct{ baseMaior, baseMenor, altura, lado1, lado2 float64 }
type cubo struct{ aresta float64 }
type esfera struct{ raio float64 }
type cilindro struct{ raio, altura float64 }

func (q quadrado) Area() float64      { return q.lado * q.lado }
func (q quadrado) Perimetro() float64 { return 4 * q.lado }

func (r retangulo) Area() float64      { return r.base * r.altura }
func (r retangulo) Perimetro() float64 { return 2 * (r.base + r.altura) }

// Fórmula de Heron: com o semiperímetro s, a área é sqrt(s(s-a)(s-b)(s-c)).
func (t triangulo) Area() float64 {
	s := t.Perimetro() / 2
	return math.Sqrt(s * (s - t.a) * (s - t.b) * (s - t.c))
}
func (t triangulo) Perimetro() float64 { return t.a + t.b + t.c }

func (c circulo) Area() float64      { return math.Pi * c.raio * c.raio }
func (c circulo) Perimetro() float64 { return 2 * math.Pi * c.raio }

func (t trapezio) Area() float64      { return (t.baseMaior + t.baseMenor) * t.altura / 2 }
func (t trapezio) Perimetro() float64 { return t.baseMaior + t.baseMenor + t.lado1 + t.lado2 }

func (c cubo) Volume() float64     { return math.Pow(c.aresta, 3) }
func (c cubo) Superficie() float64 { return 6 * c.aresta * c.aresta }

func (e esfera) Volume() float64     { return 4.0 / 3.0 * math.Pi * math.Pow(e.raio, 3) }
func (e esfera) Superficie() float64 { return 4 * math.Pi * e.raio * e.raio }

func (c cilindro) Volume() float64     { return math.Pi * c.raio * c.raio * c.altura }
func (c cilindro) Superficie() float64 { return 2 * math.Pi * c.raio * (c.raio + c.altura) }

// receita descreve como montar uma forma: quais medidas ela pede e a
// função que recebe essas medidas (já validadas como positivas) e devolve
// o valor concreto. O retorno é "any" porque pode ser plana ou solido.
type receita struct {
	medidas []string
	montar  func(m []float64) (any, error)
}

var receitas = map[string]receita{
	"quadrado":  {[]string{"lado"}, func(m []float64) (any, error) { return quadrado{m[0]}, nil }},
	"retangulo": {[]string{"base", "altura"}, func(m []float64) (any, error) { return retangulo{m[0], m[1]}, nil }},
	"triangulo": {[]string{"lado a", "lado b", "lado c"}, func(m []float64) (any, error) {
		// Desigualdade triangular: cada lado precisa ser menor que a soma
		// dos outros dois, senão as pontas não se encontram.
		a, b, c := m[0], m[1], m[2]
		if a >= b+c || b >= a+c || c >= a+b {
			return nil, fmt.Errorf("os lados %s, %s e %s não formam um triângulo (cada lado deve ser menor que a soma dos outros dois)",
				num(a), num(b), num(c))
		}
		return triangulo{a, b, c}, nil
	}},
	"circulo": {[]string{"raio"}, func(m []float64) (any, error) { return circulo{m[0]}, nil }},
	"trapezio": {[]string{"base maior", "base menor", "altura", "lado 1", "lado 2"}, func(m []float64) (any, error) {
		if m[3] < m[2] || m[4] < m[2] {
			return nil, fmt.Errorf("os lados inclinados do trapézio não podem ser menores que a altura %s", num(m[2]))
		}
		return trapezio{m[0], m[1], m[2], m[3], m[4]}, nil
	}},
	"cubo":     {[]string{"aresta"}, func(m []float64) (any, error) { return cubo{m[0]}, nil }},
	"esfera":   {[]string{"raio"}, func(m []float64) (any, error) { return esfera{m[0]}, nil }},
	"cilindro": {[]string{"raio", "altura"}, func(m []float64) (any, error) { return cilindro{m[0], m[1]}, nil }},
}

// num formata um número com duas casas e vírgula no lugar do ponto, como 12,50.
func num(v float64) string {
	return strings.Replace(strconv.FormatFloat(v, 'f', 2, 64), ".", ",", 1)
}

// nomesOrdenados devolve as chaves do map em ordem: a ordem de um map em
// Go é propositalmente aleatória, então ordenar é tarefa nossa.
func nomesOrdenados() []string {
	nomes := make([]string, 0, len(receitas))
	for nome := range receitas {
		nomes = append(nomes, nome)
	}
	sort.Strings(nomes)
	return nomes
}

// lerMedidas transforma "3,4,5" em []float64{3, 4, 5} e rejeita zero, negativo e lixo.
func lerMedidas(texto string, esperadas int) ([]float64, error) {
	partes := strings.Split(strings.TrimSpace(texto), ",")
	if esperadas == 1 && len(partes) == 2 { // "2,5" sozinho é um número com vírgula
		partes = []string{partes[0] + "." + partes[1]}
	}
	if len(partes) != esperadas {
		return nil, fmt.Errorf("esperava %d medida(s), recebi %d em %q", esperadas, len(partes), texto)
	}
	medidas := make([]float64, 0, esperadas)
	for _, p := range partes {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, fmt.Errorf("%q não é um número", p)
		}
		if v <= 0 {
			return nil, fmt.Errorf("toda medida precisa ser maior que zero (recebi %s)", num(v))
		}
		medidas = append(medidas, v)
	}
	return medidas, nil
}

// calcular monta a forma pelo nome e devolve uma linha pronta para imprimir.
func calcular(nome, medidas string) (string, error) {
	nome = strings.ToLower(strings.TrimSpace(nome))
	r, ok := receitas[nome]
	if !ok {
		return "", fmt.Errorf("forma %q desconhecida; use -listar para ver as disponíveis", nome)
	}
	m, err := lerMedidas(medidas, len(r.medidas))
	if err != nil {
		return "", fmt.Errorf("%s (%s pede: %s)", err, nome, strings.Join(r.medidas, ", "))
	}
	forma, err := r.montar(m)
	if err != nil {
		return "", err
	}
	// O type switch descobre qual interface o valor satisfaz. Cada tipo
	// concreto implementa só uma das duas, então não há ambiguidade.
	switch f := forma.(type) {
	case plana:
		return fmt.Sprintf("%-10s %-22s área = %10s   perímetro = %10s", nome, medidas, num(f.Area()), num(f.Perimetro())), nil
	case solido:
		return fmt.Sprintf("%-10s %-22s volume = %8s   superfície = %9s", nome, medidas, num(f.Volume()), num(f.Superficie())), nil
	}
	return "", fmt.Errorf("a forma %q não implementa nenhuma interface conhecida", nome)
}

func listar() {
	fmt.Println("Formas disponíveis e as medidas que cada uma pede (nessa ordem, separadas por vírgula):")
	for _, nome := range nomesOrdenados() {
		fmt.Printf("  %-10s %s\n", nome, strings.Join(receitas[nome].medidas, ", "))
	}
}

// rodarDemo calcula todas as formas com medidas de exemplo e mostra também
// o que acontece com um triângulo impossível.
func rodarDemo() {
	fmt.Println("=== Calculadora de geometria: demonstração ===")
	fmt.Println()
	fmt.Println("Figuras planas")
	exemplos := [][2]string{
		{"quadrado", "4"}, {"retangulo", "3,5"}, {"triangulo", "3,4,5"},
		{"circulo", "2"}, {"trapezio", "8,4,3,5,5"},
		{"cubo", "3"}, {"esfera", "2"}, {"cilindro", "2,10"},
	}
	for i, ex := range exemplos {
		if i == 5 {
			fmt.Println()
			fmt.Println("Sólidos")
		}
		linha, err := calcular(ex[0], ex[1])
		if err != nil {
			fmt.Println("  Erro:", err)
			continue
		}
		fmt.Println("  " + linha)
	}
	fmt.Println()
	fmt.Println("Validação: lados 1, 2 e 10")
	if _, err := calcular("triangulo", "1,2,10"); err != nil {
		fmt.Println("  Erro:", err)
	}
	fmt.Println()
	listar()
}

func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Calculadora de geometria (CodeNomad Academy)")
	fmt.Fprintln(w, "Uso: go run ./sources/10_calculadora_geometria -forma triangulo -medidas 3,4,5")
	fmt.Fprintln(w, "     go run ./sources/10_calculadora_geometria -listar | --demo")
	flag.PrintDefaults()
}

func main() {
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	forma := flag.String("forma", "", "nome da forma (veja -listar)")
	medidas := flag.String("medidas", "", "medidas separadas por vírgula, na ordem que -listar mostra")
	lista := flag.Bool("listar", false, "lista as formas e as medidas de cada uma")
	flag.Usage = uso
	flag.Parse()

	switch {
	case *demo:
		rodarDemo()
	case *lista:
		listar()
	case *forma == "" && *medidas == "":
		uso()
	default:
		linha, err := calcular(*forma, *medidas)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			os.Exit(1)
		}
		fmt.Println(linha)
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Acrescente o cone (volume = pi*r²*h/3) e o paralelepípedo: basta um
//    struct, dois métodos e uma entrada no map receitas.
// 2. Faça o triângulo também dizer se é equilátero, isósceles ou escaleno
//    e se é retângulo (a² + b² = c², cuidado com a comparação de floats).
// 3. Leia várias formas de um arquivo CSV (nome;medidas) com encoding/csv
//    e imprima a soma de todas as áreas.
// ------------------------------------------------------------
