// ============================================================
// CodeNomad Academy  |  Sources Go  |  04. Conversor de unidades
// ------------------------------------------------------------
// O que faz: converte comprimento, massa, temperatura, volume e velocidade
//            entre quaisquer duas unidades da mesma categoria, a partir de
//            uma tabela de fatores. Também lista as unidades disponíveis.
// Como usar: go run ./sources/04_conversor_unidades -valor 10 -de km -para mi
//            go run ./sources/04_conversor_unidades -valor 100 -de f -para c
//            go run ./sources/04_conversor_unidades -listar
//            go run ./sources/04_conversor_unidades --demo
// O que você aprende: maps de structs como tabela de dados, por que a ordem
//            de um map é aleatória e como ordenar as chaves com sort.Strings,
//            uma fórmula linear (valor*fator + desloc) que cobre até a
//            temperatura, funções que devolvem (valor, error) e strings.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// unidade descreve como levar um valor para a unidade base da categoria:
// base = valor*fator + desloc. Quase todas têm desloc zero; a temperatura
// precisa dele porque 0 °F não é 0 °C.
type unidade struct {
	nome   string
	fator  float64
	desloc float64
}

// categoria agrupa as unidades que podem ser convertidas entre si.
type categoria struct {
	nome     string
	base     string
	unidades map[string]unidade
}

// A tabela inteira fica em um slice de categorias. As chaves dos maps são
// os símbolos em minúsculas, que é como a entrada do usuário é procurada.
var categorias = []categoria{
	{"Comprimento", "m", map[string]unidade{
		"m": {"metro", 1, 0}, "km": {"quilômetro", 1000, 0}, "cm": {"centímetro", 0.01, 0},
		"mm": {"milímetro", 0.001, 0}, "mi": {"milha", 1609.344, 0}, "ft": {"pé", 0.3048, 0},
		"in": {"polegada", 0.0254, 0}, "yd": {"jarda", 0.9144, 0},
	}},
	{"Massa", "kg", map[string]unidade{
		"kg": {"quilograma", 1, 0}, "g": {"grama", 0.001, 0}, "mg": {"miligrama", 0.000001, 0},
		"t": {"tonelada", 1000, 0}, "lb": {"libra", 0.45359237, 0}, "oz": {"onça", 0.028349523125, 0},
	}},
	{"Temperatura", "c", map[string]unidade{
		"c": {"grau Celsius", 1, 0},
		"f": {"grau Fahrenheit", 5.0 / 9.0, -160.0 / 9.0}, // C = (F - 32) * 5/9
		"k": {"kelvin", 1, -273.15},                       // C = K - 273,15
	}},
	{"Volume", "l", map[string]unidade{
		"l": {"litro", 1, 0}, "ml": {"mililitro", 0.001, 0}, "cm3": {"centímetro cúbico", 0.001, 0},
		"m3": {"metro cúbico", 1000, 0}, "gal": {"galão americano", 3.785411784, 0}, "pt": {"pinta americana", 0.473176473, 0},
	}},
	{"Velocidade", "m/s", map[string]unidade{
		"m/s": {"metro por segundo", 1, 0}, "km/h": {"quilômetro por hora", 1 / 3.6, 0},
		"mph": {"milha por hora", 0.44704, 0}, "kn": {"nó", 0.514444, 0}, "ft/s": {"pé por segundo", 0.3048, 0},
	}},
}

// localizar procura o símbolo em todas as categorias e devolve a categoria
// e a unidade. O segundo valor de retorno segue o costume de Go: um error
// nil significa que deu certo.
func localizar(simbolo string) (*categoria, unidade, error) {
	chave := strings.ToLower(strings.TrimSpace(simbolo))
	for i := range categorias {
		if u, ok := categorias[i].unidades[chave]; ok {
			return &categorias[i], u, nil
		}
	}
	return nil, unidade{}, fmt.Errorf("unidade %q desconhecida; use -listar para ver as disponíveis", simbolo)
}

// converter leva o valor para a base da categoria e da base para o destino.
func converter(valor float64, de, para string) (float64, error) {
	catDe, uDe, err := localizar(de)
	if err != nil {
		return 0, err
	}
	catPara, uPara, err := localizar(para)
	if err != nil {
		return 0, err
	}
	if catDe != catPara {
		return 0, fmt.Errorf("não dá para converter %s (%s) em %s (%s)",
			strings.ToLower(catDe.nome), de, strings.ToLower(catPara.nome), para)
	}
	base := valor*uDe.fator + uDe.desloc
	return (base - uPara.desloc) / uPara.fator, nil
}

// formatar mostra até 4 casas, sem zeros sobrando e com vírgula.
func formatar(v float64) string {
	s := strconv.FormatFloat(v, 'f', 4, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	if s == "" || s == "-0" {
		s = "0"
	}
	return strings.Replace(s, ".", ",", 1)
}

// listar imprime a tabela. As chaves de um map saem em ordem aleatória a
// cada execução, por isso elas são copiadas para um slice e ordenadas.
func listar() {
	for _, c := range categorias {
		fmt.Printf("%s (base: %s)\n", c.nome, c.base)
		simbolos := make([]string, 0, len(c.unidades))
		for s := range c.unidades {
			simbolos = append(simbolos, s)
		}
		sort.Strings(simbolos)
		for _, s := range simbolos {
			fmt.Printf("  %-5s %s\n", s, c.unidades[s].nome)
		}
	}
}

// mostrar executa uma conversão e imprime uma linha legível, ou o erro.
func mostrar(valor float64, de, para string) error {
	resultado, err := converter(valor, de, para)
	if err != nil {
		return err
	}
	fmt.Printf("%s %s = %s %s\n", formatar(valor), de, formatar(resultado), para)
	return nil
}

// rodarDemo converte dez exemplos variados e lista a tabela de unidades.
func rodarDemo() {
	fmt.Println("=== Demonstração: dez conversões ===")
	exemplos := []struct {
		valor    float64
		de, para string
	}{
		{10, "km", "mi"}, {100, "m", "ft"}, {12, "in", "cm"}, {5, "kg", "lb"},
		{16, "oz", "g"}, {100, "c", "f"}, {32, "f", "c"}, {300, "k", "c"},
		{2, "gal", "l"}, {100, "km/h", "mph"},
	}
	for _, e := range exemplos {
		if err := mostrar(e.valor, e.de, e.para); err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
		}
	}
	fmt.Println()
	fmt.Println("=== Unidades disponíveis ===")
	listar()
}

// uso escreve no mesmo destino que flag.PrintDefaults usa (stderr).
func uso() {
	w := flag.CommandLine.Output()
	fmt.Fprintln(w, "Conversor de unidades (CodeNomad Academy)")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Uso: go run ./sources/04_conversor_unidades -valor 10 -de km -para mi")
	fmt.Fprintln(w, "     go run ./sources/04_conversor_unidades -listar")
	fmt.Fprintln(w, "     go run ./sources/04_conversor_unidades --demo")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Opções:")
	flag.PrintDefaults()
}

func main() {
	demo := flag.Bool("demo", false, "roda uma demonstração completa, sem pedir nada")
	valor := flag.Float64("valor", 1, "quantidade a converter")
	de := flag.String("de", "", "unidade de origem (ex.: km)")
	para := flag.String("para", "", "unidade de destino (ex.: mi)")
	lista := flag.Bool("listar", false, "mostra todas as unidades e sai")
	flag.Usage = uso
	flag.Parse()

	switch {
	case *demo:
		rodarDemo()
	case *lista:
		listar()
	case flag.NFlag() == 0:
		uso()
	case *de == "" || *para == "":
		fmt.Fprintln(os.Stderr, "Erro: informe -de e -para (ex.: -de km -para mi)")
		os.Exit(1)
	default:
		if err := mostrar(*valor, *de, *para); err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			os.Exit(1)
		}
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Acrescente a categoria "Área" (m2, km2, hectare, acre) só mexendo na
//    tabela, sem tocar em nenhuma função.
// 2. Aceite nomes por extenso além dos símbolos ("quilômetro"), montando
//    um segundo map de nome para símbolo na inicialização.
// 3. Crie um modo -tabela que, dado -valor e -de, mostre o valor em todas
//    as unidades da mesma categoria de uma vez.
// ------------------------------------------------------------
