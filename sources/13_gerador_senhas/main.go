// ============================================================
// CodeNomad Academy  |  Sources Go  |  13. Gerador de senhas seguras
// ------------------------------------------------------------
// O que faz: gera senhas com comprimento e classes de símbolos à escolha,
//            usando o gerador criptográfico da biblioteca padrão, garante um
//            símbolo de cada classe pedida e estima a entropia em bits.
// Como usar: go run ./sources/13_gerador_senhas --tamanho 16 --quantidade 3
//            go run ./sources/13_gerador_senhas --tamanho 20 --sem-simbolos --sem-ambiguos
//            go run ./sources/13_gerador_senhas --demo
// O que você aprende: crypto/rand contra math/rand, math/big, embaralhar
//            um slice (Fisher-Yates), structs de configuração, math.Log2.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
)

const (
	minusculas = "abcdefghijklmnopqrstuvwxyz"
	maiusculas = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	algarismos = "0123456789"
	simbolos   = "!@#$%&*()-_=+[]{};:,.?/"
	ambiguos   = "0O1lI|" // fáceis de confundir ao ler ou copiar à mão
)

// Perfil descreve quais classes entram na senha.
type Perfil struct {
	Minusculas, Maiusculas, Algarismos, Simbolos bool
	SemAmbiguos                                  bool
}

// ativas devolve o nome e o alfabeto de cada classe escolhida, já sem os ambíguos se o perfil pede.
func (p Perfil) ativas() (nomes, alfabetos []string) {
	for _, c := range []struct {
		ativa          bool
		nome, alfabeto string
	}{{p.Minusculas, "minúsculas", minusculas}, {p.Maiusculas, "maiúsculas", maiusculas},
		{p.Algarismos, "algarismos", algarismos}, {p.Simbolos, "símbolos", simbolos}} {
		if !c.ativa {
			continue
		}
		alfabeto := c.alfabeto
		if p.SemAmbiguos {
			alfabeto = strings.Map(func(r rune) rune {
				if strings.ContainsRune(ambiguos, r) {
					return -1 // devolver -1 no strings.Map apaga a runa
				}
				return r
			}, alfabeto)
		}
		nomes = append(nomes, c.nome)
		alfabetos = append(alfabetos, alfabeto)
	}
	return nomes, alfabetos
}

// Resumo descreve o perfil em uma frase curta para a tela.
func (p Perfil) Resumo() string {
	nomes, _ := p.ativas()
	s := strings.Join(nomes, " + ")
	if p.SemAmbiguos {
		s += " (sem ambíguos)"
	}
	return s
}

// sortear devolve um inteiro uniforme em [0, n) usando crypto/rand.
// math/rand serve para jogos e simulações; para senhas, só o criptográfico.
func sortear(n int) (int, error) {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}

// Gerar monta a senha: um símbolo garantido de cada classe, o resto sorteado do
// alfabeto inteiro, e um embaralhamento final para a posição dos garantidos não ser previsível.
func Gerar(tamanho int, p Perfil) (string, error) {
	_, classes := p.ativas()
	if len(classes) == 0 {
		return "", errors.New("escolha pelo menos uma classe de símbolos")
	}
	if tamanho < len(classes) {
		return "", fmt.Errorf("tamanho %d não cabe um símbolo de cada uma das %d classes", tamanho, len(classes))
	}
	if tamanho > 256 {
		return "", fmt.Errorf("tamanho %d é exagerado; o limite é 256", tamanho)
	}
	todos := strings.Join(classes, "")
	senha := make([]byte, 0, tamanho)
	for _, c := range classes {
		i, err := sortear(len(c))
		if err != nil {
			return "", err
		}
		senha = append(senha, c[i])
	}
	for len(senha) < tamanho {
		i, err := sortear(len(todos))
		if err != nil {
			return "", err
		}
		senha = append(senha, todos[i])
	}
	// Fisher-Yates: do fim para o começo, troca cada posição com uma sorteada antes dela.
	for i := len(senha) - 1; i > 0; i-- {
		j, err := sortear(i + 1)
		if err != nil {
			return "", err
		}
		senha[i], senha[j] = senha[j], senha[i]
	}
	return string(senha), nil
}

// Entropia estima os bits de aleatoriedade: tamanho vezes log2 do alfabeto.
// A garantia de um símbolo por classe tira uma fração pequena disso; a fórmula simples basta.
func Entropia(tamanho int, p Perfil) float64 {
	_, classes := p.ativas()
	alfabeto := len(strings.Join(classes, ""))
	if alfabeto == 0 {
		return 0
	}
	return float64(tamanho) * math.Log2(float64(alfabeto))
}

// tempoParaQuebrar supõe um atacante testando 10 bilhões de senhas por
// segundo e precisando, em média, de metade do espaço de busca.
func tempoParaQuebrar(bits float64) string {
	segundos := math.Pow(2, bits-1) / 1e10
	anos := segundos / (365 * 86400)
	switch {
	case segundos < 1:
		return "menos de um segundo"
	case segundos < 3600:
		return fmt.Sprintf("%.0f minutos", math.Ceil(segundos/60))
	case segundos < 86400:
		return fmt.Sprintf("%.0f horas", segundos/3600)
	case anos < 1:
		return fmt.Sprintf("%.0f dias", segundos/86400)
	case anos < 1e6:
		return fmt.Sprintf("%.0f anos", anos)
	case anos < 1e9:
		return fmt.Sprintf("%.0f milhões de anos", anos/1e6)
	}
	return "mais de um bilhão de anos"
}

func classificar(bits float64) string {
	switch {
	case bits < 36:
		return "fraca"
	case bits < 60:
		return "razoável"
	case bits < 100:
		return "forte"
	}
	return "muito forte"
}

// mostrar gera e imprime uma senha com a sua ficha técnica.
func mostrar(tamanho int, p Perfil) error {
	senha, err := Gerar(tamanho, p)
	if err != nil {
		return err
	}
	bits := Entropia(tamanho, p)
	fmt.Printf("  %-28s %3d símbolos  %5.1f bits  %-11s ataque: %s\n",
		senha, tamanho, bits, classificar(bits), tempoParaQuebrar(bits))
	return nil
}

func demo() {
	fmt.Println("=== Gerador de senhas seguras: demonstração ===")
	tudo := Perfil{Minusculas: true, Maiusculas: true, Algarismos: true, Simbolos: true}
	legivel := Perfil{Minusculas: true, Maiusculas: true, Algarismos: true, Simbolos: true, SemAmbiguos: true}
	perfis := []struct {
		titulo  string
		tamanho int
		perfil  Perfil
	}{
		{"1) 12 símbolos, só letras e algarismos", 12, Perfil{Minusculas: true, Maiusculas: true, Algarismos: true}},
		{"2) 16 símbolos, todas as classes", 16, tudo},
		{"3) 24 símbolos, todas as classes, sem ambíguos", 24, legivel},
		{"4) 8 algarismos, como um PIN (veja como é fraco)", 8, Perfil{Algarismos: true}},
	}
	for _, c := range perfis {
		fmt.Printf("%s  [%s]\n", c.titulo, c.perfil.Resumo())
		for i := 0; i < 2; i++ {
			if err := mostrar(c.tamanho, c.perfil); err != nil {
				fmt.Println("  erro:", err)
			}
		}
	}
	fmt.Println("Entropia = tamanho x log2(tamanho do alfabeto). Cada bit a mais dobra o trabalho do ataque.")
}

func uso() {
	fmt.Println("Gerador de senhas seguras (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/13_gerador_senhas --tamanho 16 --quantidade 3")
	fmt.Println("  go run ./sources/13_gerador_senhas --tamanho 20 --sem-simbolos --sem-ambiguos")
	fmt.Println("  go run ./sources/13_gerador_senhas --demo")
	flag.PrintDefaults()
}

func main() {
	tamanho := flag.Int("tamanho", 16, "quantidade de símbolos da senha")
	quantidade := flag.Int("quantidade", 1, "quantas senhas gerar")
	semMinus := flag.Bool("sem-minusculas", false, "não usar letras minúsculas")
	semMaius := flag.Bool("sem-maiusculas", false, "não usar letras maiúsculas")
	semNum := flag.Bool("sem-numeros", false, "não usar algarismos")
	semSimb := flag.Bool("sem-simbolos", false, "não usar símbolos")
	semAmb := flag.Bool("sem-ambiguos", false, "tirar 0 O 1 l I |, fáceis de confundir")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	if *demoFlag {
		demo()
		return
	}
	if len(os.Args) == 1 {
		uso()
		return
	}
	// As flags são "sem-..." para o perfil completo ser o comportamento inicial.
	p := Perfil{Minusculas: !*semMinus, Maiusculas: !*semMaius, Algarismos: !*semNum,
		Simbolos: !*semSimb, SemAmbiguos: *semAmb}
	if *quantidade < 1 || *quantidade > 100 {
		fmt.Fprintln(os.Stderr, "Erro: a quantidade precisa estar entre 1 e 100")
		os.Exit(1)
	}
	fmt.Printf("Perfil: %s\n", p.Resumo())
	for i := 0; i < *quantidade; i++ {
		if err := mostrar(*tamanho, p); err != nil {
			fmt.Fprintln(os.Stderr, "Erro:", err)
			os.Exit(1)
		}
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Acrescente a flag --frase para gerar senhas de quatro palavras sorteadas de uma lista.
// 2. Calcule a entropia exata descontando a garantia de uma letra por classe.
// 3. Copie a senha para a área de transferência com um comando do sistema.
// ------------------------------------------------------------
