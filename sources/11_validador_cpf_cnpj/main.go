// ============================================================
// CodeNomad Academy  |  Sources Go  |  11. Validador de CPF e CNPJ
// ------------------------------------------------------------
// O que faz: valida os dígitos verificadores de CPF e CNPJ, aceita o
//            número com ou sem pontuação, rejeita sequências repetidas
//            (111.111.111-11), formata e gera números válidos para teste.
// Como usar: go run ./sources/11_validador_cpf_cnpj --valor 123.456.789-09
//            go run ./sources/11_validador_cpf_cnpj --gerar cnpj
//            go run ./sources/11_validador_cpf_cnpj --demo
// O que você aprende: percorrer strings como runas, slices de inteiros,
//            funções que devolvem (valor, erro), métodos em structs, flag.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

// Documento guarda só os algarismos e o tipo reconhecido (CPF ou CNPJ).
type Documento struct {
	Tipo    string
	Digitos []int
}

// Pesos fixos do CNPJ. O CPF usa pesos decrescentes, montados na hora.
var (
	pesosCNPJ1 = []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	pesosCNPJ2 = []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
)

// somenteDigitos extrai os algarismos, pulando a pontuação permitida.
// Qualquer outra letra é erro: "12a.456" não é um documento, é outra coisa.
func somenteDigitos(s string) ([]int, error) {
	var digitos []int
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digitos = append(digitos, int(r-'0'))
		case strings.ContainsRune(".-/ ", r):
			// pontuação: simplesmente ignoramos
		default:
			return nil, fmt.Errorf("caractere inesperado %q", r)
		}
	}
	return digitos, nil
}

// digitoVerificador aplica a regra do módulo 11: soma cada algarismo vezes
// o peso, olha o resto por 11; resto 0 ou 1 vira 0, senão vira 11 - resto.
func digitoVerificador(digitos, pesos []int) int {
	soma := 0
	for i, d := range digitos {
		soma += d * pesos[i]
	}
	resto := soma % 11
	if resto < 2 {
		return 0
	}
	return 11 - resto
}

// pesosDecrescentes monta a lista n, n-1, ..., 2 usada pelo CPF.
func pesosDecrescentes(n int) []int {
	pesos := make([]int, 0, n-1)
	for p := n; p >= 2; p-- {
		pesos = append(pesos, p)
	}
	return pesos
}

// todosIguais detecta sequências como 000.000.000-00, que passam na conta
// dos verificadores mas não são documentos reais.
func todosIguais(digitos []int) bool {
	for _, d := range digitos[1:] {
		if d != digitos[0] {
			return false
		}
	}
	return true
}

// Validar reconhece o tipo pelo tamanho e confere os dois verificadores.
// Mesmo recusando, devolve o Tipo quando deu para reconhecer: a mensagem fica melhor.
func Validar(entrada string) (Documento, error) {
	digitos, err := somenteDigitos(entrada)
	if err != nil {
		return Documento{}, err
	}
	var doc Documento
	var dv1, dv2 int
	switch len(digitos) {
	case 11:
		doc = Documento{Tipo: "CPF", Digitos: digitos}
		dv1 = digitoVerificador(digitos[:9], pesosDecrescentes(10))
		dv2 = digitoVerificador(digitos[:10], pesosDecrescentes(11))
	case 14:
		doc = Documento{Tipo: "CNPJ", Digitos: digitos}
		dv1 = digitoVerificador(digitos[:12], pesosCNPJ1)
		dv2 = digitoVerificador(digitos[:13], pesosCNPJ2)
	default:
		return Documento{}, fmt.Errorf("esperava 11 dígitos (CPF) ou 14 (CNPJ), recebi %d", len(digitos))
	}
	if todosIguais(digitos) {
		return doc, errors.New("sequência de dígitos repetidos não é um documento válido")
	}
	n := len(digitos)
	if digitos[n-2] != dv1 || digitos[n-1] != dv2 {
		return doc, fmt.Errorf("dígitos verificadores incorretos (esperava %d%d, recebi %d%d)", dv1, dv2, digitos[n-2], digitos[n-1])
	}
	return doc, nil
}

// String junta os algarismos. Por se chamar String, o fmt usa este método ao imprimir com %s ou %v.
func (d Documento) String() string {
	var sb strings.Builder
	for _, x := range d.Digitos {
		sb.WriteByte(byte('0' + x))
	}
	return sb.String()
}

// Formatado devolve o documento com a pontuação oficial.
func (d Documento) Formatado() string {
	s := d.String()
	if d.Tipo == "CPF" {
		return s[0:3] + "." + s[3:6] + "." + s[6:9] + "-" + s[9:11]
	}
	return s[0:2] + "." + s[2:5] + "." + s[5:8] + "/" + s[8:12] + "-" + s[12:14]
}

// Gerar sorteia um documento válido para testar formulários. Os números não pertencem a ninguém.
func Gerar(tipo string) (Documento, error) {
	tipo = strings.ToUpper(strings.TrimSpace(tipo))
	var base int
	switch tipo {
	case "CPF":
		base = 9
	case "CNPJ":
		base = 8
	default:
		return Documento{}, fmt.Errorf("tipo desconhecido %q (use cpf ou cnpj)", tipo)
	}
	digitos := make([]int, 0, 14)
	for i := 0; i < base; i++ {
		digitos = append(digitos, rand.IntN(10))
	}
	if tipo == "CNPJ" {
		digitos = append(digitos, 0, 0, 0, 1) // sufixo 0001, o da matriz
	}
	if todosIguais(digitos) { // raríssimo, mas o sorteio pode repetir tudo
		digitos[0] = (digitos[0] + 1) % 10
	}
	if tipo == "CPF" {
		digitos = append(digitos, digitoVerificador(digitos, pesosDecrescentes(10)))
		digitos = append(digitos, digitoVerificador(digitos, pesosDecrescentes(11)))
	} else {
		digitos = append(digitos, digitoVerificador(digitos, pesosCNPJ1))
		digitos = append(digitos, digitoVerificador(digitos, pesosCNPJ2))
	}
	return Documento{Tipo: tipo, Digitos: digitos}, nil
}

// relatar imprime uma linha de resultado e diz se o documento passou.
func relatar(entrada string) bool {
	doc, err := Validar(entrada)
	if err != nil {
		tipo := doc.Tipo
		if tipo == "" {
			tipo = "?"
		}
		fmt.Printf("  %-20s %-5s INVÁLIDO: %v\n", entrada, tipo, err)
		return false
	}
	fmt.Printf("  %-20s %-5s válido   -> %s\n", entrada, doc.Tipo, doc.Formatado())
	return true
}

func demo() {
	fmt.Println("=== Validador de CPF e CNPJ: demonstração ===")
	fmt.Println("1) Validando uma lista de documentos fictícios:")
	casos := []string{
		"123.456.789-09", "12345678909", "111.444.777-35",
		"11.222.333/0001-81", "11222333000181",
		"123.456.789-00", "111.111.111-11", "11.222.333/0001-82",
		"12345", "12a.456.789-09",
	}
	validos := 0
	for _, c := range casos {
		if relatar(c) {
			validos++
		}
	}
	fmt.Printf("\n  Resumo: %d válidos, %d inválidos\n\n", validos, len(casos)-validos)
	fmt.Println("2) Gerando números de teste (sorteados, não pertencem a ninguém):")
	for _, tipo := range []string{"cpf", "cnpj"} {
		doc, err := Gerar(tipo)
		if err != nil {
			fmt.Println("  erro:", err)
			continue
		}
		_, errV := Validar(doc.String())
		fmt.Printf("  %-5s %-20s revalidado com sucesso: %v\n", doc.Tipo, doc.Formatado(), errV == nil)
	}
}

func uso() {
	fmt.Println("Validador de CPF e CNPJ (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/11_validador_cpf_cnpj --valor 123.456.789-09")
	fmt.Println("  go run ./sources/11_validador_cpf_cnpj --gerar cnpj")
	fmt.Println("  go run ./sources/11_validador_cpf_cnpj --demo")
	fmt.Println("Opções:")
	flag.PrintDefaults()
}

func main() {
	valor := flag.String("valor", "", "CPF ou CNPJ para validar, com ou sem pontuação")
	gerar := flag.String("gerar", "", "gera um documento válido de teste: cpf ou cnpj")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	switch {
	case *demoFlag:
		demo()
	case *gerar != "":
		doc, err := Gerar(*gerar)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			os.Exit(1)
		}
		fmt.Printf("%s gerado: %s  (sem pontuação: %s)\n", doc.Tipo, doc.Formatado(), doc)
	case *valor != "" || flag.NArg() > 0:
		entrada := *valor
		if entrada == "" {
			entrada = flag.Arg(0) // aceita também o valor solto, sem --valor
		}
		if !relatar(entrada) {
			os.Exit(1) // código 1 permite usar o programa em scripts
		}
	default:
		uso()
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Aceite vários documentos de uma vez (um por linha, lidos de um
//    arquivo) e imprima um resumo no fim.
// 2. Adapte o CNPJ para o formato alfanumérico: letras valem o valor
//    ASCII menos 48, e o resto da conta é igual.
// 3. Escreva testes com o pacote testing cobrindo os casos da demo.
// ------------------------------------------------------------
