// ============================================================
// CodeNomad Academy  |  Sources Go  |  18. Conversor de números romanos
// ------------------------------------------------------------
// O que faz: converte inteiro para romano e romano para inteiro, de 1 a
//            3999, rejeita romanos mal formados (IIII, VX, IC) explicando
//            o motivo, e imprime uma tabela de referência dos símbolos.
// Como usar: go run ./sources/18_numeros_romanos --inteiro 1994
//            go run ./sources/18_numeros_romanos --romano MCMXCIV
//            go run ./sources/18_numeros_romanos --tabela
//            go run ./sources/18_numeros_romanos --demo
// O que você aprende: slice de structs como tabela de conversão, o
//            algoritmo guloso, validação por ida e volta, strings.Builder
//            e erros com fmt.Errorf.
// Requisitos: Go 1.22+, só biblioteca padrão
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

// Simbolo liga um valor ao seu símbolo. A tabela inclui os pares de
// subtração (CM, XC, IV) para o algoritmo guloso não precisar de casos especiais.
type Simbolo struct {
	Valor   int
	Romano  string
	Subtrai bool // true nos pares como IV e XL, só para a tabela de referência
}

// tabela vai do maior para o menor: o guloso sempre tira o maior que cabe.
var tabela = []Simbolo{
	{1000, "M", false}, {900, "CM", true}, {500, "D", false}, {400, "CD", true},
	{100, "C", false}, {90, "XC", true}, {50, "L", false}, {40, "XL", true},
	{10, "X", false}, {9, "IX", true}, {5, "V", false}, {4, "IV", true}, {1, "I", false},
}

// valorLetra devolve o valor de uma única letra romana, ou 0 se não existir.
func valorLetra(r rune) int {
	for _, s := range tabela {
		if !s.Subtrai && s.Romano == string(r) {
			return s.Valor
		}
	}
	return 0
}

// ParaRomano usa o algoritmo guloso: enquanto o número couber no símbolo
// atual, escreve o símbolo e subtrai; depois passa ao próximo menor.
func ParaRomano(n int) (string, error) {
	if n < 1 || n > 3999 {
		return "", fmt.Errorf("%d está fora do intervalo clássico (1 a 3999)", n)
	}
	var sb strings.Builder
	for _, s := range tabela {
		for n >= s.Valor {
			sb.WriteString(s.Romano)
			n -= s.Valor
		}
	}
	return sb.String(), nil
}

// ParaInteiro lê o romano da esquerda para a direita: uma letra menor
// seguida de uma maior (IV, XC) é subtraída, senão é somada. Depois
// converte o resultado de volta e compara: se não bater, o texto estava
// mal formado (IIII, VX, IC). Essa ida e volta é uma validação simples e completa.
func ParaInteiro(romano string) (int, error) {
	limpo := strings.ToUpper(strings.TrimSpace(romano))
	if limpo == "" {
		return 0, errors.New("o número romano está vazio")
	}
	runas := []rune(limpo)
	total := 0
	for i, r := range runas {
		v := valorLetra(r)
		if v == 0 {
			return 0, fmt.Errorf("%q não é um algarismo romano (use I, V, X, L, C, D ou M)", r)
		}
		if i+1 < len(runas) && v < valorLetra(runas[i+1]) {
			total -= v
		} else {
			total += v
		}
	}
	canonico, err := ParaRomano(total)
	if err != nil {
		return 0, fmt.Errorf("%s vale %d, fora do intervalo de 1 a 3999", limpo, total)
	}
	if canonico != limpo {
		return 0, fmt.Errorf("%s está mal formado: %d escreve-se %s", limpo, total, canonico)
	}
	return total, nil
}

// detectar aceita um argumento solto e descobre se é inteiro ou romano.
func detectar(entrada string) (string, error) {
	if n, err := strconv.Atoi(strings.TrimSpace(entrada)); err == nil {
		r, err := ParaRomano(n)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d = %s", n, r), nil
	}
	n, err := ParaInteiro(entrada)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s = %d", strings.ToUpper(strings.TrimSpace(entrada)), n), nil
}

func imprimirTabela() {
	fmt.Println("Tabela de referência:")
	fmt.Println("  Símbolo  Valor  Observação")
	for _, s := range tabela {
		obs := "símbolo básico"
		if s.Subtrai {
			obs = "par de subtração (menor antes do maior)"
		}
		fmt.Printf("  %-7s  %5d  %s\n", s.Romano, s.Valor, obs)
	}
	fmt.Println("  Regras: no máximo três repetições seguidas (III, XXX); V, L e D nunca repetem;")
	fmt.Println("  só I, X e C podem subtrair, e só do par de símbolos logo acima deles.")
}

func demo() {
	fmt.Println("=== Conversor de números romanos: demonstração ===")
	fmt.Println("1) Inteiro para romano:")
	for _, n := range []int{1, 4, 9, 14, 40, 90, 400, 1994, 2024, 3999} {
		r, err := ParaRomano(n)
		if err != nil {
			fmt.Println("  erro:", err)
			continue
		}
		fmt.Printf("  %4d -> %s\n", n, r)
	}
	fmt.Println("2) Romano para inteiro:")
	for _, r := range []string{"III", "XLII", "XCIX", "CDXLIV", "MMXXVI", "mcmxciv"} {
		n, err := ParaInteiro(r)
		if err != nil {
			fmt.Println("  erro:", err)
			continue
		}
		fmt.Printf("  %-8s -> %d\n", strings.ToUpper(r), n)
	}
	fmt.Println("3) Entradas inválidas sendo rejeitadas:")
	for _, r := range []string{"IIII", "VX", "IC", "MMMM", "ABC", ""} {
		if _, err := ParaInteiro(r); err != nil {
			fmt.Printf("  %-6q rejeitado: %v\n", r, err)
		}
	}
	if _, err := ParaRomano(0); err != nil {
		fmt.Printf("  %-6d rejeitado: %v\n", 0, err)
	}
	fmt.Println("4) Tabela:")
	imprimirTabela()
}

func uso() {
	fmt.Println("Conversor de números romanos (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/18_numeros_romanos --inteiro 1994")
	fmt.Println("  go run ./sources/18_numeros_romanos --romano MCMXCIV")
	fmt.Println("  go run ./sources/18_numeros_romanos 42          (descobre o sentido sozinho)")
	fmt.Println("  go run ./sources/18_numeros_romanos --tabela")
	fmt.Println("  go run ./sources/18_numeros_romanos --demo")
	flag.PrintDefaults()
}

func main() {
	inteiro := flag.Int("inteiro", 0, "inteiro de 1 a 3999 para converter em romano")
	romano := flag.String("romano", "", "número romano para converter em inteiro")
	tabelaFlag := flag.Bool("tabela", false, "imprime a tabela de referência")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	var saida string
	var err error
	switch {
	case *demoFlag:
		demo()
		return
	case *tabelaFlag:
		imprimirTabela()
		return
	case *inteiro != 0:
		saida, err = detectar(strconv.Itoa(*inteiro))
	case *romano != "":
		saida, err = detectar(*romano)
	case flag.NArg() > 0:
		saida, err = detectar(flag.Arg(0))
	default:
		uso()
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
	fmt.Println(saida)
}

// ------------------------------------------------------------
// Para ir além
// 1. Estenda o intervalo acima de 3999 usando a notação com barra em cima
//    (represente a barra com um sinal, por exemplo "_V" para 5000).
// 2. Converta uma data inteira (dia, mês e ano) para romanos, como nas
//    fachadas de prédios antigos.
// 3. Escreva testes com o pacote testing que convertem 1 a 3999 para
//    romano e de volta, conferindo que o valor se mantém.
// ------------------------------------------------------------
