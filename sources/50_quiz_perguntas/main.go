// ============================================================
// CodeNomad Academy  |  Sources Go  |  50. Quiz de perguntas e respostas
// ------------------------------------------------------------
// O que faz: aplica um quiz de múltipla escolha lido de um arquivo texto
//            simples (ou do banco embutido sobre programação), embaralha as
//            perguntas, pontua e revisa as erradas ao fim. Também grava um
//            arquivo de exemplo para você montar o seu próprio quiz.
// Como usar: go run ./sources/50_quiz_perguntas --arquivo meu_quiz.txt --n 5
//            go run ./sources/50_quiz_perguntas --exemplo modelo.txt
//            go run ./sources/50_quiz_perguntas --demo
// O que você aprende: bufio.Scanner linha a linha com número da linha nos
//            erros, strings.CutPrefix e strings.Cut, slices de structs,
//            rand.Shuffle com semente, map de erros para a revisão e uma
//            função que lê respostas do teclado ou de um gabarito automático.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

// Pergunta tem o enunciado, as alternativas na ordem e a posição da certa.
type Pergunta struct {
	Enunciado string
	Opcoes    []string
	Certa     int
}

// bancoEmbutido usa o mesmo formato do arquivo: o parser é testado a cada execução.
const bancoEmbutido = `# Linhas com # são comentários. Cada pergunta: P:, de duas a seis
# alternativas A) a F), e R: com a letra certa. Linha em branco separa.
P: O que um compilador faz?
A) Executa o programa linha por linha
B) Traduz o código-fonte para código de máquina
C) Corrige erros de lógica sozinho
R: B

P: Qual estrutura guarda pares de chave e valor?
A) Lista
B) Pilha
C) Mapa (dicionário)
R: C

P: O que significa "bug" num programa?
A) Um recurso escondido
B) Um erro de comportamento
C) Um comentário longo
R: B

P: Qual destes é um sistema de controle de versão?
A) Git
B) HTML
C) JSON
R: A

P: O que uma função recursiva faz?
A) Roda em outra máquina
B) Chama a si mesma
C) Nunca termina
R: B

P: Qual a base do sistema que usa só os dígitos 0 e 1?
A) 8
B) 10
C) 2
R: C

P: Em quase todas as linguagens, a contagem de posições numa lista começa em:
A) 0
B) 1
C) -1
R: A

P: O que é um laço (loop)?
A) Um bloco que repete enquanto uma condição vale
B) Um tipo de variável
C) Um arquivo de configuração
R: A
`

// lerPerguntas interpreta o formato. Cada erro traz o número da linha.
func lerPerguntas(texto string) ([]Pergunta, error) {
	var lista []Pergunta
	var atual *Pergunta
	scanner := bufio.NewScanner(strings.NewReader(texto))
	for n := 1; scanner.Scan(); n++ {
		linha := strings.TrimSpace(scanner.Text())
		switch {
		case linha == "" || strings.HasPrefix(linha, "#"):
			continue
		case strings.HasPrefix(linha, "P:"):
			lista = append(lista, Pergunta{Enunciado: strings.TrimSpace(linha[2:]), Certa: -1})
			atual = &lista[len(lista)-1] // ponteiro para o último item do slice
		case atual == nil:
			return nil, fmt.Errorf("linha %d: esperava uma pergunta começando com P:", n)
		case len(linha) > 2 && linha[1] == ')' && linha[0] >= 'A' && linha[0] <= 'F':
			if int(linha[0]-'A') != len(atual.Opcoes) {
				return nil, fmt.Errorf("linha %d: as alternativas precisam seguir a ordem A), B), C)...", n)
			}
			atual.Opcoes = append(atual.Opcoes, strings.TrimSpace(linha[2:]))
		case strings.HasPrefix(linha, "R:"):
			letra := strings.ToUpper(strings.TrimSpace(linha[2:]))
			if len(letra) != 1 || int(letra[0]-'A') >= len(atual.Opcoes) {
				return nil, fmt.Errorf("linha %d: a resposta %q não corresponde a nenhuma alternativa", n, letra)
			}
			atual.Certa = int(letra[0] - 'A')
		default:
			return nil, fmt.Errorf("linha %d: não entendi %q", n, linha)
		}
	}
	for _, p := range lista {
		if len(p.Opcoes) < 2 || p.Certa < 0 {
			return nil, fmt.Errorf("a pergunta %q precisa de duas ou mais alternativas e de uma linha R:", p.Enunciado)
		}
	}
	if len(lista) == 0 {
		return nil, fmt.Errorf("nenhuma pergunta encontrada")
	}
	return lista, nil
}

// aplicar faz as perguntas, corrige na hora e devolve a pontuação e as erradas.
// A função responder recebe a pergunta e devolve a posição escolhida.
func aplicar(perguntas []Pergunta, responder func(int, Pergunta) int) (int, []int) {
	acertos := 0
	var erradas []int
	for i, p := range perguntas {
		fmt.Printf("\n%d) %s\n", i+1, p.Enunciado)
		for j, o := range p.Opcoes {
			fmt.Printf("   %c) %s\n", 'A'+j, o)
		}
		escolha := responder(i, p)
		if escolha == p.Certa {
			acertos++
			fmt.Println("   Certo!")
		} else {
			erradas = append(erradas, i)
			fmt.Printf("   Errado. A resposta certa é %c.\n", 'A'+p.Certa)
		}
	}
	return acertos, erradas
}

// doTeclado lê uma letra válida para a pergunta atual.
func doTeclado() func(int, Pergunta) int {
	scanner := bufio.NewScanner(os.Stdin)
	return func(_ int, p Pergunta) int {
		for {
			fmt.Printf("   Sua resposta (A a %c): ", 'A'+len(p.Opcoes)-1)
			if !scanner.Scan() {
				fmt.Println("\nEntrada encerrada: a pergunta conta como errada.")
				return -1
			}
			r := strings.ToUpper(strings.TrimSpace(scanner.Text()))
			if len(r) == 1 && int(r[0]-'A') < len(p.Opcoes) {
				return int(r[0] - 'A')
			}
			fmt.Println("   Digite só a letra da alternativa.")
		}
	}
}

// respostaAutomatica acerta tudo, menos as perguntas 2, 5 e 7, em que erra de propósito.
func respostaAutomatica(i int, p Pergunta) int {
	escolha := p.Certa
	if i == 1 || i == 4 || i == 6 {
		escolha = (p.Certa + 1) % len(p.Opcoes)
	}
	fmt.Printf("   Resposta automática: %c\n", 'A'+escolha)
	return escolha
}

func revisao(perguntas []Pergunta, acertos int, erradas []int) {
	total := len(perguntas)
	fmt.Printf("\nResultado: %d de %d (%.0f%%).\n", acertos, total, 100*float64(acertos)/float64(total))
	if len(erradas) == 0 {
		fmt.Println("Gabarito limpo, nada a revisar.")
		return
	}
	fmt.Println("Revisão das erradas:")
	for _, i := range erradas {
		p := perguntas[i]
		fmt.Printf("  %d) %s\n     certa: %c) %s\n", i+1, p.Enunciado, 'A'+p.Certa, p.Opcoes[p.Certa])
	}
}

func main() {
	arquivo := flag.String("arquivo", "", "arquivo de perguntas (vazio = banco embutido)")
	quantas := flag.Int("n", 0, "quantas perguntas aplicar (0 = todas)")
	semente := flag.Int64("semente", 0, "semente do embaralhamento (0 = aleatória)")
	semEmbaralhar := flag.Bool("sem-embaralhar", false, "mantém a ordem do arquivo")
	exemplo := flag.String("exemplo", "", "grava o banco embutido neste arquivo, como modelo, e sai")
	demo := flag.Bool("demo", false, "responde 8 perguntas automaticamente e mostra a revisão")
	flag.Parse()

	if *exemplo != "" {
		if err := os.WriteFile(*exemplo, []byte(bancoEmbutido), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "Erro ao gravar o modelo:", err)
			os.Exit(1)
		}
		fmt.Println("Modelo gravado em", *exemplo)
		return
	}
	texto := bancoEmbutido
	if *arquivo != "" {
		dados, err := os.ReadFile(*arquivo)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro: não consegui ler o arquivo:", err)
			os.Exit(1)
		}
		texto = string(dados)
	}
	perguntas, err := lerPerguntas(texto)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro no quiz:", err)
		os.Exit(1)
	}
	responder := doTeclado()
	if *demo {
		*quantas, *semente, responder = 8, 7, respostaAutomatica
		fmt.Println("=== CodeNomad Academy | Quiz: 8 perguntas respondidas automaticamente ===")
		fmt.Println("As perguntas 2, 5 e 7 serão erradas de propósito, para mostrar a revisão.")
	}
	if *quantas < 0 || *quantas > len(perguntas) {
		fmt.Fprintf(os.Stderr, "Erro: --n precisa estar entre 0 e %d (o quiz tem %d perguntas).\n", len(perguntas), len(perguntas))
		os.Exit(1)
	}
	if !*semEmbaralhar {
		gerador := rand.New(rand.NewSource(*semente))
		if *semente == 0 {
			gerador = rand.New(rand.NewSource(rand.Int63()))
		}
		gerador.Shuffle(len(perguntas), func(i, j int) { perguntas[i], perguntas[j] = perguntas[j], perguntas[i] })
	}
	if *quantas > 0 {
		perguntas = perguntas[:*quantas]
	}
	fmt.Printf("Quiz com %d pergunta(s). Boa sorte!\n", len(perguntas))
	acertos, erradas := aplicar(perguntas, responder)
	revisao(perguntas, acertos, erradas)
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. Embaralhe também as alternativas de cada pergunta, mantendo a certa
//    apontada para a posição nova.
// 2. Cronometre cada resposta com time.Now e mostre a média ao fim.
// 3. Grave o resultado (data, acertos, total) em resultados.csv com
//    encoding/csv e mostre a evolução nas últimas cinco tentativas.
// ------------------------------------------------------------
