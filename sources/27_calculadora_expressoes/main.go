// ============================================================
// CodeNomad Academy  |  Sources Go  |  27. Calculadora de expressões matemáticas
// ------------------------------------------------------------
// O que faz: lê uma expressão em texto com + - * / ^, parênteses e números
//            decimais, respeita a precedência com o algoritmo shunting-yard,
//            mostra a notação posfixa gerada e avalia o resultado, com erros
//            claros para expressão mal formada e divisão por zero.
// Como usar: go run ./sources/27_calculadora_expressoes "2 + 3 * (4 - 1) ^ 2"
//            go run ./sources/27_calculadora_expressoes --posfixa -- "-2 ^ 2"
//            go run ./sources/27_calculadora_expressoes --demo
// O que você aprende: tokenização com um laço sobre runas, pilha com slice,
//            tabela de precedência em map, erros com fmt.Errorf e errors.New,
//            strconv.ParseFloat, math.Pow e o padrão "ler, converter, avaliar".
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Token é um pedaço da expressão: um número, um operador ou um parêntese.
type Token struct {
	tipo  string // "num", "op", "(" ou ")"
	texto string
	valor float64 // só faz sentido quando tipo == "num"
}

// precedencia diz quem opera primeiro. O "~" é o menos unário (-2), que fica
// acima de * e abaixo de ^, para -2^2 valer -4 como na matemática.
var precedencia = map[string]int{"+": 1, "-": 1, "*": 2, "/": 2, "~": 3, "^": 4}

// tokenizar quebra o texto em tokens. O sinal de menos vira unário ("~")
// quando aparece no começo, após um operador ou após "(".
func tokenizar(texto string) ([]Token, error) {
	var tokens []Token
	runas := []rune(texto)
	for i := 0; i < len(runas); {
		r := runas[i]
		switch {
		case unicode.IsSpace(r):
			i++
		case unicode.IsDigit(r) || r == '.':
			inicio := i
			for i < len(runas) && (unicode.IsDigit(runas[i]) || runas[i] == '.') {
				i++
			}
			v, err := strconv.ParseFloat(string(runas[inicio:i]), 64)
			if err != nil {
				return nil, fmt.Errorf("número mal formado: '%s'", string(runas[inicio:i]))
			}
			tokens = append(tokens, Token{"num", string(runas[inicio:i]), v})
		case strings.ContainsRune("+-*/^", r):
			op := string(r)
			if r == '-' && (len(tokens) == 0 || tokens[len(tokens)-1].tipo == "op" || tokens[len(tokens)-1].tipo == "(") {
				op = "~"
			}
			tokens = append(tokens, Token{"op", op, 0})
			i++
		case r == '(' || r == ')':
			tokens = append(tokens, Token{string(r), string(r), 0})
			i++
		default:
			return nil, fmt.Errorf("caractere inválido '%c' na posição %d", r, i+1)
		}
	}
	if len(tokens) == 0 {
		return nil, errors.New("expressão vazia")
	}
	return tokens, nil
}

// paraPosfixa aplica o shunting-yard de Dijkstra: números vão direto para a
// saída; operadores esperam numa pilha até aparecer um de precedência menor.
func paraPosfixa(tokens []Token) ([]Token, error) {
	var saida, pilha []Token
	for _, t := range tokens {
		switch t.tipo {
		case "num":
			saida = append(saida, t)
		case "op":
			if t.texto == "~" { // prefixo: nunca tira ninguém da pilha
				pilha = append(pilha, t)
				continue
			}
			for len(pilha) > 0 && pilha[len(pilha)-1].tipo == "op" {
				topo := pilha[len(pilha)-1]
				pTopo, pAtual := precedencia[topo.texto], precedencia[t.texto]
				// ^ associa pela direita: 2^3^2 é 2^(3^2) = 512, então um ^ não tira outro ^ da pilha.
				if pTopo < pAtual || (pTopo == pAtual && t.texto == "^") {
					break
				}
				saida = append(saida, topo)
				pilha = pilha[:len(pilha)-1]
			}
			pilha = append(pilha, t)
		case "(":
			pilha = append(pilha, t)
		case ")":
			for len(pilha) > 0 && pilha[len(pilha)-1].tipo != "(" {
				saida = append(saida, pilha[len(pilha)-1])
				pilha = pilha[:len(pilha)-1]
			}
			if len(pilha) == 0 {
				return nil, errors.New("parêntese ')' sem o '(' correspondente")
			}
			pilha = pilha[:len(pilha)-1] // descarta o "(": ele não vai para a saída
		}
	}
	for len(pilha) > 0 { // o que sobrou na pilha vai para a saída
		topo := pilha[len(pilha)-1]
		pilha = pilha[:len(pilha)-1]
		if topo.tipo == "(" {
			return nil, errors.New("parêntese '(' aberto e nunca fechado")
		}
		saida = append(saida, topo)
	}
	return saida, nil
}

// avaliar percorre a posfixa com uma pilha de números: número empilha,
// operador tira dois (ou um, no unário), calcula e empilha o resultado.
func avaliar(posfixa []Token) (float64, error) {
	var pilha []float64
	tirar := func() (float64, error) {
		if len(pilha) == 0 {
			return 0, errors.New("faltou um número para o operador")
		}
		v := pilha[len(pilha)-1]
		pilha = pilha[:len(pilha)-1]
		return v, nil
	}
	for _, t := range posfixa {
		if t.tipo == "num" {
			pilha = append(pilha, t.valor)
			continue
		}
		b, err := tirar()
		if err != nil {
			return 0, err
		}
		if t.texto == "~" {
			pilha = append(pilha, -b)
			continue
		}
		a, err := tirar()
		if err != nil {
			return 0, err
		}
		switch t.texto {
		case "+":
			pilha = append(pilha, a+b)
		case "-":
			pilha = append(pilha, a-b)
		case "*":
			pilha = append(pilha, a*b)
		case "/":
			if b == 0 {
				return 0, errors.New("divisão por zero")
			}
			pilha = append(pilha, a/b)
		case "^":
			pilha = append(pilha, math.Pow(a, b))
		}
	}
	if len(pilha) != 1 {
		return 0, errors.New("sobraram números sem operador entre eles")
	}
	return pilha[0], nil
}

// calcular junta as três etapas e devolve o resultado e a posfixa em texto.
func calcular(expr string) (float64, string, error) {
	tokens, err := tokenizar(expr)
	if err != nil {
		return 0, "", err
	}
	posfixa, err := paraPosfixa(tokens)
	if err != nil {
		return 0, "", err
	}
	partes := make([]string, len(posfixa))
	for i, t := range posfixa {
		partes[i] = strings.ReplaceAll(t.texto, "~", "neg")
	}
	v, err := avaliar(posfixa)
	return v, strings.Join(partes, " "), err
}

// formatar usa precisão -1: o menor número de casas que representa o valor,
// então 14 sai como "14" e 3.5 como "3.5".
func formatar(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func demo() {
	fmt.Println("=== Calculadora de expressões: CodeNomad Academy ===")
	fmt.Println("\nCinco expressões avaliadas com precedência correta:")
	for _, e := range []string{"2 + 3 * 4", "(2 + 3) * 4", "2 ^ 3 ^ 2", "-(1.5 + 2.5) * 2", "10 / 4 - 0.5"} {
		v, _, err := calcular(e)
		if err != nil {
			fmt.Printf("  %-22s erro: %v\n", e, err)
			continue
		}
		fmt.Printf("  %-22s = %s\n", e, formatar(v))
	}
	e := "3 + 4 * (2 - 1) ^ 2 / 8"
	v, posfixa, _ := calcular(e)
	fmt.Printf("\nA posfixa de %q:\n  %s\n  = %s\n", e, posfixa, formatar(v))
	fmt.Println("Na posfixa não há parênteses: a ordem dos operadores já carrega a precedência.")
	fmt.Println("\nDois erros tratados sem derrubar o programa:")
	for _, e := range []string{"(2 + 3", "8 / (4 - 4)"} {
		if _, _, err := calcular(e); err != nil {
			fmt.Printf("  %-14s -> %v\n", e, err)
		}
	}
}

func main() {
	posfixa := flag.Bool("posfixa", false, "mostra também a notação posfixa gerada")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Uso: go run ./sources/27_calculadora_expressoes [--posfixa] [--] \"EXPRESSÃO\"")
		fmt.Fprintln(os.Stderr, "Aceita + - * / ^, parênteses e decimais com ponto. Expressão que começa com menos: use -- antes dela.")
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
	expr := strings.Join(flag.Args(), " ") // permite escrever sem aspas: 2 + 3
	v, pf, err := calcular(expr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
	if *posfixa {
		fmt.Println("posfixa:", pf)
	}
	fmt.Println(formatar(v))
}

// ------------------------------------------------------------
// Para ir além
// 1. Aceite funções como sqrt(x) e max(a, b): um token novo e uma pilha de argumentos.
// 2. Permita variáveis (x = 3 depois 2*x) guardadas num map entre expressões.
// 3. Reescreva o avaliador como descida recursiva (uma função por nível de precedência).
// ------------------------------------------------------------
