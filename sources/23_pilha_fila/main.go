// ============================================================
// CodeNomad Academy  |  Sources Go  |  23. Pilha e fila na prática
// ------------------------------------------------------------
// O que faz: implementa pilha (LIFO) e fila (FIFO) com as operações clássicas
//            e usa as duas em três problemas reais: parênteses balanceados,
//            desfazer/refazer de um editor e fila de atendimento com espera.
// Como usar: go run ./sources/23_pilha_fila --parenteses "{[(1+2)*3]}"
//            go run ./sources/23_pilha_fila --editor "escrever:Olá,escrever:mundo,desfazer,refazer"
//            go run ./sources/23_pilha_fila --fila "Ana:4,Bruno:2,Carla:6"
//            go run ./sources/23_pilha_fila --demo
// O que você aprende: tipos genéricos (type Pilha[T any]), métodos com ponteiro, o
//            padrão "valor, ok", strings.Cut e Split, e a fila com slice que segura memória.
// Requisitos: Go 1.22+, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Pilha é LIFO: o último que entra é o primeiro que sai. O parâmetro T any
// torna o tipo genérico: a mesma Pilha serve para runas, textos ou structs.
type Pilha[T any] struct {
	itens []T
}

func (p *Pilha[T]) Empilhar(x T) { p.itens = append(p.itens, x) }

// Desempilhar devolve o topo e true, ou o valor zero de T e false se vazia.
// É o padrão "valor, ok" que você já viu em maps e em conversões de tipo.
func (p *Pilha[T]) Desempilhar() (T, bool) {
	x, ok := p.Topo()
	if ok {
		p.itens = p.itens[:len(p.itens)-1]
	}
	return x, ok
}

func (p *Pilha[T]) Topo() (T, bool) {
	var zero T
	if len(p.itens) == 0 {
		return zero, false
	}
	return p.itens[len(p.itens)-1], true
}

func (p *Pilha[T]) Tamanho() int { return len(p.itens) }

// Fila é FIFO: o primeiro que entra é o primeiro que sai.
type Fila[T any] struct {
	itens []T
}

func (f *Fila[T]) Enfileirar(x T) { f.itens = append(f.itens, x) }

// Desenfileirar tira da frente. Fatiar com [1:] segura os itens antigos até realocar; em filas longas, use um índice.
func (f *Fila[T]) Desenfileirar() (T, bool) {
	var zero T
	if len(f.itens) == 0 {
		return zero, false
	}
	x := f.itens[0]
	f.itens = f.itens[1:]
	return x, true
}

func (f *Fila[T]) Tamanho() int { return len(f.itens) }

// Problema 1: parênteses balanceados. Cada abertura entra na pilha; cada
// fechamento precisa casar com o topo. No fim, a pilha tem de estar vazia.
var pares = map[rune]rune{')': '(', ']': '[', '}': '{'}

type abertura struct {
	simbolo rune
	pos     int
}

// balanceado devolve se fechou tudo na ordem certa, a posição do problema (ou -1) e o motivo.
func balanceado(texto string) (bool, int, string) {
	var p Pilha[abertura]
	for i, r := range []rune(texto) { // []rune: a posição é contada em runas, não em bytes
		switch r {
		case '(', '[', '{':
			p.Empilhar(abertura{r, i})
		case ')', ']', '}':
			topo, ok := p.Desempilhar()
			if !ok {
				return false, i, fmt.Sprintf("'%c' fecha sem ninguém aberto", r)
			}
			if topo.simbolo != pares[r] {
				return false, i, fmt.Sprintf("'%c' tenta fechar '%c' aberto na posição %d", r, topo.simbolo, topo.pos)
			}
		}
	}
	if topo, ok := p.Topo(); ok {
		return false, topo.pos, fmt.Sprintf("'%c' aberto e nunca fechado", topo.simbolo)
	}
	return true, -1, "tudo fechado na ordem certa"
}

func conferir(texto string) {
	ok, pos, motivo := balanceado(texto)
	veredito := "balanceado"
	if !ok {
		veredito = fmt.Sprintf("erro na posição %d", pos)
	}
	fmt.Printf("  %-22q %-20s %s\n", texto, veredito, motivo)
}

// Problema 2: editor com desfazer e refazer. Duas pilhas: o passado (para
// desfazer) e o futuro (para refazer). Escrever algo novo apaga o futuro.
type Editor struct {
	texto   string
	passado Pilha[string]
	futuro  Pilha[string]
}

func (e *Editor) Escrever(palavra string) {
	e.passado.Empilhar(e.texto)
	e.futuro = Pilha[string]{}
	e.texto = strings.TrimSpace(e.texto + " " + palavra)
}

// mover tira um estado da pilha "de", guarda o texto atual na pilha "para" e
// adota o estado tirado. Desfazer e refazer são a mesma dança, em sentidos opostos.
func (e *Editor) mover(de, para *Pilha[string]) bool {
	estado, ok := de.Desempilhar()
	if !ok {
		return false
	}
	para.Empilhar(e.texto)
	e.texto = estado
	return true
}

func (e *Editor) Desfazer() bool { return e.mover(&e.passado, &e.futuro) }
func (e *Editor) Refazer() bool  { return e.mover(&e.futuro, &e.passado) }

// executarEditor interpreta comandos como "escrever:Olá", "desfazer" e "refazer".
func executarEditor(roteiro string) error {
	var e Editor
	for _, comando := range strings.Split(roteiro, ",") {
		nome, arg, _ := strings.Cut(strings.TrimSpace(comando), ":")
		switch nome {
		case "escrever":
			if arg == "" {
				return fmt.Errorf("'escrever' precisa de um texto, assim: escrever:palavra")
			}
			e.Escrever(arg)
		case "desfazer":
			if !e.Desfazer() {
				fmt.Println("  (nada para desfazer)")
			}
		case "refazer":
			if !e.Refazer() {
				fmt.Println("  (nada para refazer)")
			}
		default:
			return fmt.Errorf("comando desconhecido '%s' (use escrever:texto, desfazer ou refazer)", comando)
		}
		fmt.Printf("  %-18s texto: %-22q passado: %d  futuro: %d\n", comando, e.texto, e.passado.Tamanho(), e.futuro.Tamanho())
	}
	return nil
}

// Problema 3: fila de atendimento com um único guichê. Quem chega espera a
// soma dos atendimentos de quem está na frente.
type Cliente struct {
	nome    string
	minutos int
}

func lerClientes(texto string) ([]Cliente, error) {
	var clientes []Cliente
	for _, item := range strings.Split(texto, ",") {
		nome, min, achou := strings.Cut(strings.TrimSpace(item), ":")
		minutos, err := strconv.Atoi(min)
		if !achou || nome == "" || err != nil || minutos <= 0 {
			return nil, fmt.Errorf("item inválido '%s': use nome:minutos, com minutos maior que zero", item)
		}
		clientes = append(clientes, Cliente{nome, minutos})
	}
	return clientes, nil
}

func atender(clientes []Cliente) {
	var fila Fila[Cliente]
	for _, c := range clientes {
		fila.Enfileirar(c)
	}
	fmt.Printf("  %-10s %11s %8s %8s\n", "cliente", "atendimento", "espera", "sai aos")
	relogio, somaEspera := 0, 0
	for fila.Tamanho() > 0 {
		c, _ := fila.Desenfileirar() // o laço garante que há alguém, por isso ignoramos o ok
		somaEspera += relogio
		fmt.Printf("  %-10s %7d min %4d min %4d min\n", c.nome, c.minutos, relogio, relogio+c.minutos)
		relogio += c.minutos
	}
	fmt.Printf("  espera média: %.1f min; o guichê fecha aos %d min\n", float64(somaEspera)/float64(len(clientes)), relogio)
}

func demo() {
	fmt.Println("=== Pilha e fila na prática: CodeNomad Academy ===")
	fmt.Println("\n1) Parênteses balanceados (pilha de aberturas):")
	for _, t := range []string{"{[(1+2)*3]}", "(a+b]", "((x)", "fim)", "sem nenhum"} {
		conferir(t)
	}
	fmt.Println("\n2) Editor com desfazer e refazer (duas pilhas de estados):")
	if err := executarEditor("escrever:Olá,escrever:mundo,escrever:cruel,desfazer,desfazer,refazer,escrever:novo,refazer"); err != nil {
		fmt.Println("  erro:", err)
	}
	fmt.Println("\n3) Fila de atendimento (quem chega primeiro é atendido primeiro):")
	clientes, _ := lerClientes("Ana:4,Bruno:12,Carla:2,Diego:7,Elisa:3")
	atender(clientes)
	fmt.Println("\nLeitura: pilha serve quando o mais recente importa (fechar, desfazer);\n" +
		"fila serve quando a ordem de chegada importa (atender, processar).")
}

func main() {
	parenteses := flag.String("parenteses", "", "texto para conferir parênteses, colchetes e chaves")
	editor := flag.String("editor", "", "roteiro do editor: escrever:texto, desfazer e refazer separados por vírgula")
	fila := flag.String("fila", "", "clientes nome:minutos separados por vírgula")
	demoFlag := flag.Bool("demo", false, "roda a demonstração completa, sem pedir nada")
	flag.Parse()
	switch {
	case *demoFlag:
		demo()
	case *parenteses != "":
		conferir(*parenteses)
	case *editor != "":
		if err := executarEditor(*editor); err != nil {
			fmt.Fprintln(os.Stderr, "erro:", err)
			os.Exit(2)
		}
	case *fila != "":
		clientes, err := lerClientes(*fila)
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro:", err)
			os.Exit(2)
		}
		atender(clientes)
	default:
		fmt.Fprintln(os.Stderr, "Uso: go run ./sources/23_pilha_fila --parenteses TEXTO | --editor ROTEIRO | --fila CLIENTES | --demo")
		flag.PrintDefaults()
		os.Exit(2)
	}
}

// ------------------------------------------------------------
// Para ir além
// 1. Faça a Fila guardar um índice de início em vez de fatiar e meça a memória com runtime.MemStats.
// 2. No editor, limite o histórico a 10 estados: o que descartar, e de onde?
// 3. Abra dois guichês na fila: cada cliente vai para o que desocupa primeiro. Quanto cai a espera?
// ------------------------------------------------------------
