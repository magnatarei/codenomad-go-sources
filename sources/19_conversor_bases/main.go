// ============================================================
// CodeNomad Academy  |  Sources Go  |  19. Conversor de bases numéricas
// ------------------------------------------------------------
// O que faz: converte um número entre qualquer base de 2 a 36 (binário,
//            octal, base 10, hexadecimal), explica o cálculo passo a passo
//            e converte texto para binário e de volta, byte a byte.
// Como usar: go run ./sources/19_conversor_bases --valor 255 --de 10 --para 2
//            go run ./sources/19_conversor_bases --valor ff --de 16 --para 10 --passos
//            go run ./sources/19_conversor_bases --texto "Go!"
//            go run ./sources/19_conversor_bases --bits "01000111 01101111 00100001"
//            go run ./sources/19_conversor_bases --demo
// O que você aprende: strconv.ParseUint e FormatUint, bytes contra runas em
//            UTF-8, o verbo %08b do fmt, laços de divisão e potência, erros claros.
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
	"unicode/utf8"
)

// digitos são os símbolos das bases até 36: 0 a 9 e depois A a Z.
const digitos = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func validarBase(nome string, base int) error {
	if base < 2 || base > 36 {
		return fmt.Errorf("a base %s precisa estar entre 2 e 36, recebi %d", nome, base)
	}
	return nil
}

// Converter lê o valor na base de origem e escreve na base de destino.
// strconv faz o trabalho pesado; a validação é que deixa a mensagem clara.
func Converter(valor string, de, para int) (string, error) {
	if err := validarBase("de origem", de); err != nil {
		return "", err
	}
	if err := validarBase("de destino", para); err != nil {
		return "", err
	}
	valor = strings.ToUpper(strings.TrimSpace(valor))
	if valor == "" {
		return "", errors.New("o valor está vazio")
	}
	for _, r := range valor {
		pos := strings.IndexRune(digitos, r)
		if pos < 0 || pos >= de {
			return "", fmt.Errorf("o dígito %q não existe na base %d (válidos: %s)", r, de, digitos[:de])
		}
	}
	n, err := strconv.ParseUint(valor, de, 64)
	if err != nil {
		return "", fmt.Errorf("%s não cabe em 64 bits sem sinal", valor)
	}
	return strings.ToUpper(strconv.FormatUint(n, para)), nil
}

// Explicar imprime as duas etapas do cálculo: primeiro cada dígito vezes a
// potência da base de origem (chegando à base 10), depois divisões sucessivas
// pela base de destino, lendo os restos de trás para a frente.
func Explicar(valor string, de, para int) error {
	resultado, err := Converter(valor, de, para)
	if err != nil {
		return err
	}
	valor = strings.ToUpper(strings.TrimSpace(valor))
	fmt.Printf("Passo a passo de %s (base %d) para base %d:\n", valor, de, para)
	fmt.Printf("  1) Cada dígito vale dígito x %d^posição, contando posições da direita:\n", de)
	var total uint64
	potencia := uint64(1)
	for i := len(valor) - 1; i >= 0; i-- {
		d := uint64(strings.IndexByte(digitos, valor[i]))
		fmt.Printf("     %c x %d^%d = %d x %d = %d\n", valor[i], de, len(valor)-1-i, d, potencia, d*potencia)
		total += d * potencia
		potencia *= uint64(de)
	}
	fmt.Printf("     soma = %d (base 10)\n", total)
	fmt.Printf("  2) Divide por %d até chegar a zero e guarda os restos:\n", para)
	n := total
	if n == 0 {
		fmt.Println("     0 / qualquer base = 0, resto 0")
	}
	for n > 0 {
		fmt.Printf("     %d / %d = %d, resto %d (dígito %c)\n", n, para, n/uint64(para), n%uint64(para), digitos[n%uint64(para)])
		n /= uint64(para)
	}
	fmt.Printf("     restos lidos de baixo para cima: %s\n", resultado)
	return nil
}

// TextoParaBinario mostra cada byte em 8 bits. Letras acentuadas ocupam
// dois bytes em UTF-8, por isso "ç" vira dois grupos.
func TextoParaBinario(texto string) string {
	grupos := make([]string, 0, len(texto))
	for i := 0; i < len(texto); i++ { // índice por byte, de propósito
		grupos = append(grupos, fmt.Sprintf("%08b", texto[i]))
	}
	return strings.Join(grupos, " ")
}

// BinarioParaTexto faz o caminho inverso: cada grupo de 8 bits vira um byte,
// e a sequência de bytes é validada como UTF-8 antes de virar texto.
func BinarioParaTexto(bits string) (string, error) {
	grupos := strings.Fields(bits)
	if len(grupos) == 0 {
		return "", errors.New("nenhum grupo de bits informado")
	}
	saida := make([]byte, 0, len(grupos))
	for _, g := range grupos {
		if len(g) != 8 {
			return "", fmt.Errorf("o grupo %q precisa ter exatamente 8 bits, tem %d", g, len(g))
		}
		b, err := strconv.ParseUint(g, 2, 8)
		if err != nil {
			return "", fmt.Errorf("o grupo %q contém algo que não é 0 ou 1", g)
		}
		saida = append(saida, byte(b))
	}
	if !utf8.Valid(saida) {
		return "", errors.New("os bytes não formam texto UTF-8 válido (grupo cortado no meio de um acento?)")
	}
	return string(saida), nil
}

func nomeBase(b int) string {
	nomes := map[int]string{2: "binário", 8: "octal", 10: "base 10", 16: "hexadecimal"}
	if n, ok := nomes[b]; ok {
		return n
	}
	return fmt.Sprintf("base %d", b)
}

func demo() {
	fmt.Println("=== Conversor de bases numéricas: demonstração ===")
	fmt.Println("1) Conversões entre bases:")
	casos := []struct {
		valor    string
		de, para int
	}{
		{"255", 10, 2}, {"255", 10, 16}, {"11111111", 2, 10}, {"377", 8, 10},
		{"ff", 16, 2}, {"2026", 10, 8}, {"1000", 10, 36}, {"zz", 36, 10}, {"777", 8, 16},
	}
	for _, c := range casos {
		r, err := Converter(c.valor, c.de, c.para)
		if err != nil {
			fmt.Println("  erro:", err)
			continue
		}
		fmt.Printf("  %-10s %-14s -> %-12s (%s)\n", strings.ToUpper(c.valor), "("+nomeBase(c.de)+")", r, nomeBase(c.para))
	}
	fmt.Println("2) Explicando um dos cálculos:")
	if err := Explicar("2026", 10, 16); err != nil {
		fmt.Println("  erro:", err)
	}
	fmt.Println("3) Texto em binário e de volta:")
	palavra := "Go!"
	bits := TextoParaBinario(palavra)
	fmt.Printf("  %q -> %s\n", palavra, bits)
	volta, err := BinarioParaTexto(bits)
	if err != nil {
		fmt.Println("  erro:", err)
	} else {
		fmt.Printf("  %s -> %q\n", bits, volta)
	}
	fmt.Printf("  %q -> %s  (o ç ocupa dois bytes em UTF-8)\n", "ação", TextoParaBinario("ação"))
	fmt.Println("4) Um dígito inválido para a base:")
	if _, err := Converter("129", 8, 10); err != nil {
		fmt.Println("  rejeitado:", err)
	}
}

func uso() {
	fmt.Println("Conversor de bases numéricas (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/19_conversor_bases --valor 255 --de 10 --para 2")
	fmt.Println("  go run ./sources/19_conversor_bases --valor ff --de 16 --para 10 --passos")
	fmt.Println("  go run ./sources/19_conversor_bases --texto \"Go!\"")
	fmt.Println("  go run ./sources/19_conversor_bases --bits \"01000111 01101111 00100001\"")
	fmt.Println("  go run ./sources/19_conversor_bases --demo")
	flag.PrintDefaults()
}

func main() {
	valor := flag.String("valor", "", "número a converter, escrito na base de origem")
	de := flag.Int("de", 10, "base de origem (2 a 36)")
	para := flag.Int("para", 2, "base de destino (2 a 36)")
	passos := flag.Bool("passos", false, "explica o cálculo passo a passo")
	texto := flag.String("texto", "", "texto para converter em binário")
	bits := flag.String("bits", "", "grupos de 8 bits separados por espaço para converter em texto")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	var err error
	switch {
	case *demoFlag:
		demo()
	case *texto != "":
		fmt.Println(TextoParaBinario(*texto))
	case *bits != "":
		var saida string
		if saida, err = BinarioParaTexto(*bits); err == nil {
			fmt.Println(saida)
		}
	case *valor != "" && *passos:
		err = Explicar(*valor, *de, *para)
	case *valor != "":
		var saida string
		if saida, err = Converter(*valor, *de, *para); err == nil {
			fmt.Printf("%s (base %d) = %s (base %d)\n", strings.ToUpper(*valor), *de, saida, *para)
		}
	default:
		uso()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Aceite números negativos e mostre também a representação em
//    complemento de dois com 8, 16 e 32 bits.
// 2. Converta a parte fracionária (por exemplo 0.75 em base 10 = 0.11 em binário)
//    multiplicando pela base e guardando a parte inteira a cada passo.
// 3. Use math/big para converter números maiores que 64 bits.
// ------------------------------------------------------------
