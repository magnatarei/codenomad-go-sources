// ============================================================
// CodeNomad Academy  |  Sources Go  |  26. Fibonacci, fatorial e primos
// ------------------------------------------------------------
// O que faz: calcula Fibonacci de três jeitos (recursivo puro, recursivo com
//            memória e iterativo) medindo tempo e chamadas, fatorial, teste de
//            primalidade e o crivo de Eratóstenes até N, explicando por que a
//            recursão pura explode.
// Como usar: go run ./sources/26_fibonacci_primos --fib 35
//            go run ./sources/26_fibonacci_primos --fatorial 20 --primo 97 --crivo 50
//            go run ./sources/26_fibonacci_primos --demo
// O que você aprende: recursão e caso base, memoização com map, uint64 e o limite
//            de estouro, time.Since para medir, slice de bool como marcação,
//            math.Sqrt para cortar o laço e strings.Builder para montar saída.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

// Limites que cabem num uint64 sem estourar: fib(93) e 20! são os últimos.
const (
	maxFib      = 93
	maxFatorial = 20
	maxFibPuro  = 40 // acima disso a versão recursiva pura leva vários segundos
)

// chamadas conta quantas vezes cada versão de Fibonacci foi chamada. Uma
// variável global é o jeito mais simples de medir sem mudar as assinaturas.
var chamadas int

// fibPuro é a definição matemática ao pé da letra. Ela recalcula os mesmos
// valores sem parar: fib(30) chama fib(2) centenas de milhares de vezes.
func fibPuro(n int) uint64 {
	chamadas++
	if n < 2 {
		return uint64(n) // caso base: sem ele a recursão nunca termina
	}
	return fibPuro(n-1) + fibPuro(n-2)
}

// fibMemo guarda cada resultado num map: cada n é calculado uma única vez.
func fibMemo(n int, memo map[int]uint64) uint64 {
	chamadas++
	if n < 2 {
		return uint64(n)
	}
	if v, ok := memo[n]; ok {
		return v
	}
	memo[n] = fibMemo(n-1, memo) + fibMemo(n-2, memo)
	return memo[n]
}

// fibIterativo anda para a frente com duas variáveis: nem recursão nem map.
func fibIterativo(n int) uint64 {
	chamadas++
	a, b := uint64(0), uint64(1)
	for i := 0; i < n; i++ {
		a, b = b, a+b // troca em paralelo: o Go avalia a direita inteira antes de atribuir
	}
	return a
}

// fatorial multiplica de 1 até n. Em uint64, 21! já não cabe.
func fatorial(n int) uint64 {
	resultado := uint64(1)
	for i := 2; i <= n; i++ {
		resultado *= uint64(i)
	}
	return resultado
}

// ehPrimo testa divisores só até a raiz quadrada: se n = a*b, um dos dois
// é no máximo sqrt(n). Pulamos os pares depois de testar o 2.
func ehPrimo(n int) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 {
		return n == 2
	}
	limite := int(math.Sqrt(float64(n)))
	for d := 3; d <= limite; d += 2 {
		if n%d == 0 {
			return false
		}
	}
	return true
}

// crivo marca os múltiplos de cada primo como compostos e devolve o que sobrou.
// É muito mais rápido que testar número por número quando se quer todos até N.
func crivo(n int) []int {
	composto := make([]bool, n+1) // false = ainda pode ser primo
	var primos []int
	for i := 2; i <= n; i++ {
		if composto[i] {
			continue
		}
		primos = append(primos, i)
		for m := i * i; m <= n; m += i { // começa em i*i: os múltiplos menores já foram marcados
			composto[m] = true
		}
	}
	return primos
}

// medir executa f, zera e lê o contador de chamadas e devolve o tempo gasto.
func medir(f func() uint64) (valor uint64, n int, gasto time.Duration) {
	chamadas = 0
	inicio := time.Now()
	valor = f()
	return valor, chamadas, time.Since(inicio)
}

// compararFib imprime a tabela das três versões para o mesmo n.
func compararFib(n int) {
	fmt.Printf("Fibonacci de %d pelas três versões:\n", n)
	fmt.Printf("  %-18s %22s %12s %12s\n", "versão", "resultado", "chamadas", "tempo")
	versoes := []struct {
		nome string
		f    func() uint64
	}{
		{"recursiva pura", func() uint64 { return fibPuro(n) }},
		{"recursiva + memo", func() uint64 { return fibMemo(n, map[int]uint64{}) }},
		{"iterativa", func() uint64 { return fibIterativo(n) }},
	}
	for _, v := range versoes {
		if v.nome == "recursiva pura" && n > maxFibPuro {
			fmt.Printf("  %-18s %22s\n", v.nome, "pulada: levaria minutos")
			continue
		}
		valor, qtd, gasto := medir(v.f)
		fmt.Printf("  %-18s %22d %12d %12s\n", v.nome, valor, qtd, gasto.Round(time.Microsecond))
	}
}

// simNao traduz o bool para a saída em português.
func simNao(b bool) string {
	if b {
		return "sim"
	}
	return "não"
}

// listar formata números em linhas de 10, para a lista de primos não virar uma salsicha.
func listar(nums []int) string {
	var sb strings.Builder
	for i, p := range nums {
		if i > 0 && i%10 == 0 {
			sb.WriteString("\n")
		}
		fmt.Fprintf(&sb, " %4d", p)
	}
	return sb.String()
}

func demo() {
	fmt.Println("=== Fibonacci, fatorial e primos: CodeNomad Academy ===")
	fmt.Println()
	compararFib(30)
	fmt.Println("\nPor que a recursão pura explode: cada chamada gera outras duas, então o")
	fmt.Println("número de chamadas dobra a cada passo (cerca de 1,6^n). A memória lembra o")
	fmt.Println("que já calculou e a iterativa nem precisa lembrar: só duas variáveis.")
	fmt.Println("\nChamadas da versão pura conforme n cresce:")
	for _, n := range []int{10, 20, 25, 30} {
		_, qtd, gasto := medir(func() uint64 { return fibPuro(n) })
		fmt.Printf("  fib(%2d): %10d chamadas em %s\n", n, qtd, gasto.Round(time.Microsecond))
	}
	fmt.Printf("\nFatorial de 20 = %d (o maior que cabe em uint64; 21! já estoura)\n", fatorial(maxFatorial))
	fmt.Println("\nTeste de primalidade por divisão até a raiz quadrada:")
	for _, n := range []int{2, 15, 97, 561, 7919} {
		fmt.Printf("  %5d é primo? %s\n", n, simNao(ehPrimo(n)))
	}
	primos := crivo(100)
	fmt.Printf("\nCrivo de Eratóstenes até 100 (%d primos):\n%s\n", len(primos), listar(primos))
	grande := crivo(1000000)
	fmt.Printf("\nAté 1.000.000 o crivo acha %d primos; o último é %d.\n", len(grande), grande[len(grande)-1])
}

func main() {
	fib := flag.Int("fib", -1, "compara as três versões de Fibonacci para este n (0 a 93)")
	fat := flag.Int("fatorial", -1, "fatorial de n (0 a 20)")
	primo := flag.Int("primo", -1, "testa se n é primo")
	ate := flag.Int("crivo", -1, "lista os primos até n com o crivo (2 a 10.000.000)")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	if flag.NFlag() == 0 {
		fmt.Fprintln(os.Stderr, "Uso: go run ./sources/26_fibonacci_primos --fib N | --fatorial N | --primo N | --crivo N | --demo")
		flag.PrintDefaults()
		os.Exit(2)
	}
	// Cada opção é validada antes de rodar: um erro de faixa não derruba as outras.
	if *fib >= 0 {
		if *fib > maxFib {
			fmt.Fprintf(os.Stderr, "erro: --fib aceita de 0 a %d (fib(%d) não cabe em uint64)\n", maxFib, maxFib+1)
			os.Exit(2)
		}
		compararFib(*fib)
	}
	if *fat >= 0 {
		if *fat > maxFatorial {
			fmt.Fprintf(os.Stderr, "erro: --fatorial aceita de 0 a %d (%d! não cabe em uint64)\n", maxFatorial, maxFatorial+1)
			os.Exit(2)
		}
		fmt.Printf("%d! = %d\n", *fat, fatorial(*fat))
	}
	if *primo >= 0 {
		fmt.Printf("%d é primo? %s\n", *primo, simNao(ehPrimo(*primo)))
	}
	if *ate >= 0 {
		if *ate < 2 || *ate > 10000000 {
			fmt.Fprintln(os.Stderr, "erro: --crivo aceita de 2 a 10.000.000")
			os.Exit(2)
		}
		primos := crivo(*ate)
		fmt.Printf("%d primos até %d:\n%s\n", len(primos), *ate, listar(primos))
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Use math/big para calcular fib(1000) e 100! sem estourar.
// 2. Faça o crivo guardar só os ímpares e meça quanto de memória economiza.
// 3. Escreva o teste de primalidade de Miller-Rabin e compare com ehPrimo
//    em números de 15 dígitos.
// ------------------------------------------------------------
