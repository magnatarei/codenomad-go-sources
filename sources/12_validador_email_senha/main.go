// ============================================================
// CodeNomad Academy  |  Sources Go  |  12. Validador de e-mail e força de senha
// ------------------------------------------------------------
// O que faz: valida um e-mail com regras claras (um único @, domínio com
//            ponto, sem espaços, tamanhos) e pontua a força de uma senha de 0 a
//            100 (comprimento, classes, sequências e repetições), com dicas.
// Como usar: go run ./sources/12_validador_email_senha --email ana@escola.com.br
//            go run ./sources/12_validador_email_senha --senha "Tr0ca-de-Oleo!"
//            go run ./sources/12_validador_email_senha --demo
// O que você aprende: strings.Cut e Split, o pacote unicode, closures, min e
//            max, e regras pequenas e nomeadas no lugar de uma regex gigante.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// ValidarEmail devolve a lista de problemas encontrados. Lista vazia = ok.
func ValidarEmail(email string) []string {
	var problemas []string
	if email == "" {
		return []string{"e-mail vazio"}
	}
	if len(email) > 254 {
		problemas = append(problemas, "passa de 254 caracteres")
	}
	if strings.ContainsAny(email, " \t\n") {
		problemas = append(problemas, "contém espaço")
	}
	if n := strings.Count(email, "@"); n != 1 {
		return append(problemas, fmt.Sprintf("precisa de exatamente um @, achei %d", n))
	}
	local, dominio, _ := strings.Cut(email, "@") // divide no @ e devolve as duas metades
	switch {
	case local == "":
		problemas = append(problemas, "nada antes do @")
	case len(local) > 64:
		problemas = append(problemas, "a parte antes do @ passa de 64 caracteres")
	case strings.HasPrefix(local, ".") || strings.HasSuffix(local, "."):
		problemas = append(problemas, "a parte antes do @ começa ou termina com ponto")
	case strings.Contains(local, ".."):
		problemas = append(problemas, "dois pontos seguidos antes do @")
	}
	for _, r := range local {
		if !ehAscii(r) && !strings.ContainsRune("._%+-", r) {
			problemas = append(problemas, fmt.Sprintf("símbolo não permitido antes do @: %q", r))
			break
		}
	}
	return append(problemas, validarDominio(dominio)...)
}

// ehAscii aceita só letras e algarismos sem acento, como a maioria dos provedores.
func ehAscii(r rune) bool {
	return r < 128 && (unicode.IsLetter(r) || unicode.IsDigit(r))
}

func validarDominio(dominio string) []string {
	if dominio == "" {
		return []string{"nada depois do @"}
	}
	if !strings.Contains(dominio, ".") {
		return []string{"domínio sem ponto (ex.: escola.com.br)"}
	}
	var problemas []string
	rotulos := strings.Split(dominio, ".")
	for _, rotulo := range rotulos {
		if rotulo == "" {
			problemas = append(problemas, "domínio com ponto no início, no fim ou dois seguidos")
			break
		}
		for _, r := range rotulo {
			if !ehAscii(r) && r != '-' {
				problemas = append(problemas, fmt.Sprintf("símbolo não permitido no domínio: %q", r))
				break
			}
		}
	}
	// A terminação (com, br, org) tem pelo menos duas letras e nenhum algarismo.
	if tld := rotulos[len(rotulos)-1]; len(tld) < 2 || strings.ContainsAny(tld, "0123456789") {
		problemas = append(problemas, fmt.Sprintf("terminação %q inválida: precisa de 2+ letras", tld))
	}
	return problemas
}

// Avaliacao reúne a pontuação de uma senha e as dicas para melhorá-la.
type Avaliacao struct {
	Pontos int
	Dicas  []string
}

// Veredito traduz a pontuação em uma palavra.
func (a Avaliacao) Veredito() string {
	switch {
	case a.Pontos >= 75:
		return "muito forte"
	case a.Pontos >= 55:
		return "forte"
	case a.Pontos >= 35:
		return "razoável"
	}
	return "fraca"
}

// senhasComuns são as primeiras que qualquer ataque tenta. Lista curta,
// só para ilustrar a ideia; uma real teria milhões de linhas.
var senhasComuns = []string{"123456", "123456789", "12345678", "senha", "password",
	"qwerty", "abc123", "111111", "admin", "senha123", "asdfgh"}

// AvaliarSenha pontua de 0 a 100. Cada regra soma ou desconta pontos e,
// quando desconta, registra uma dica para o usuário.
func AvaliarSenha(senha string) Avaliacao {
	runas := []rune(senha)
	n := len(runas)
	if n == 0 {
		return Avaliacao{Dicas: []string{"a senha está vazia"}}
	}
	var av Avaliacao
	// pontuar é uma closure: enxerga av e evita repetir o if/else quatro vezes.
	pontuar := func(ok bool, pontos int, dica string) {
		if ok {
			av.Pontos += pontos
		} else {
			av.Dicas = append(av.Dicas, dica)
		}
	}
	// 1) Comprimento: 2 pontos por símbolo, no máximo 40.
	av.Pontos += min(n, 20) * 2
	pontuar(n >= 8, 0, "use pelo menos 8 caracteres (12 é melhor)")
	// 2) Classes: minúsculas, maiúsculas, algarismos e símbolos (espaço conta).
	var minus, maius, digito, simbolo bool
	for _, r := range runas {
		switch {
		case unicode.IsLower(r):
			minus = true
		case unicode.IsUpper(r):
			maius = true
		case unicode.IsDigit(r):
			digito = true
		default:
			simbolo = true
		}
	}
	pontuar(minus, 10, "misture letras minúsculas")
	pontuar(maius, 10, "misture letras maiúsculas")
	pontuar(digito, 10, "inclua algarismos")
	pontuar(simbolo, 15, "inclua símbolos como ! - _ ou espaço")
	// 3) Sequências (abc, 123, cba) e repetições (aaa) são previsíveis.
	sequencias, repeticoes := 0, 0
	for i := 1; i+1 < n; i++ {
		d1, d2 := runas[i]-runas[i-1], runas[i+1]-runas[i]
		if (d1 == 1 && d2 == 1) || (d1 == -1 && d2 == -1) {
			sequencias++
		}
		if d1 == 0 && d2 == 0 {
			repeticoes++
		}
	}
	pontuar(sequencias == 0, 0, "evite sequências como abc ou 123")
	pontuar(repeticoes == 0, 0, "evite o mesmo símbolo três vezes seguidas")
	av.Pontos -= min(sequencias, 2)*15 + min(repeticoes, 2)*10
	// 4) Senha famosa: não importa o resto, vale no máximo 10.
	baixa := strings.ToLower(senha)
	for _, comum := range senhasComuns {
		if baixa == comum || (len(comum) >= 5 && strings.Contains(baixa, comum)) {
			av.Pontos = min(av.Pontos, 10)
			av.Dicas = append(av.Dicas, fmt.Sprintf("contém %q, que está nas listas de senhas vazadas", comum))
			break
		}
	}
	av.Pontos = max(0, min(100, av.Pontos))
	return av
}

func relatarEmail(email string) bool {
	problemas := ValidarEmail(email)
	if len(problemas) == 0 {
		fmt.Printf("  %-28s OK\n", email)
		return true
	}
	fmt.Printf("  %-28s INVÁLIDO: %s\n", email, strings.Join(problemas, "; "))
	return false
}

func relatarSenha(senha string, mostrar bool) {
	av := AvaliarSenha(senha)
	rotulo := strings.Repeat("*", len([]rune(senha))) // fora da demo, a senha fica mascarada
	if mostrar {
		rotulo = senha
	}
	fmt.Printf("  %-32s %3d/100  %s\n", rotulo, av.Pontos, av.Veredito())
	for _, d := range av.Dicas {
		fmt.Println("      dica:", d)
	}
}

func demo() {
	fmt.Println("=== Validador de e-mail e força de senha: demonstração ===")
	fmt.Println("1) E-mails (todos inventados):")
	for _, e := range []string{"ana.silva@escola.com.br", "joao@@exemplo.com", "maria@exemplo",
		"pedro silva@exemplo.com", ".lucas@exemplo.com.br"} {
		relatarEmail(e)
	}
	fmt.Println("\n2) Senhas (exemplos didáticos, não use nenhuma delas):")
	for _, s := range []string{"123456", "senha123", "Abacaxi2024", "Tr0ca-de-Oleo!",
		"Vaga-lume azul no telhado #7"} {
		relatarSenha(s, true)
	}
}

func uso() {
	fmt.Println("Validador de e-mail e força de senha (CodeNomad Academy)")
	fmt.Println("Uso:")
	fmt.Println("  go run ./sources/12_validador_email_senha --email ana@escola.com.br")
	fmt.Println("  go run ./sources/12_validador_email_senha --senha \"Tr0ca-de-Oleo!\"")
	fmt.Println("  go run ./sources/12_validador_email_senha --demo")
	flag.PrintDefaults()
}

func main() {
	email := flag.String("email", "", "e-mail para validar")
	senha := flag.String("senha", "", "senha para pontuar (aparece mascarada na tela)")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa")
	flag.Usage = uso
	flag.Parse()
	switch {
	case *demoFlag:
		demo()
	case *email == "" && *senha == "":
		uso()
	default:
		ok := true
		if *email != "" {
			ok = relatarEmail(*email)
		}
		if *senha != "" {
			relatarSenha(*senha, false)
		}
		if !ok {
			os.Exit(1) // e-mail recusado: código 1 para quem usa em scripts
		}
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Leia e-mails de um arquivo e separe válidos e inválidos em dois arquivos.
// 2. Aceite letras acentuadas no domínio (e-mails internacionalizados).
// 3. Carregue uma lista maior de senhas comuns de um arquivo e compare o
//    tempo da busca em slice contra a busca em map.
// ------------------------------------------------------------
