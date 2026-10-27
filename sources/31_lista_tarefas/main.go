// ============================================================
// CodeNomad Academy  |  Sources Go  |  31. Lista de tarefas em arquivo JSON
// ------------------------------------------------------------
// O que faz: gerencia uma lista de tarefas guardada num arquivo JSON. Você
//            adiciona, lista, conclui, remove e busca tarefas, cada uma com
//            prioridade (alta, média ou baixa) e prazo opcional.
// Como usar: go run ./sources/31_lista_tarefas -arquivo tarefas.json add "Estudar Go" alta 2026-11-05
//            go run ./sources/31_lista_tarefas -arquivo tarefas.json list
//            go run ./sources/31_lista_tarefas --demo
// O que você aprende: structs com tags JSON, encoding/json, métodos com receptor
//            ponteiro, sort.SliceStable, subcomandos com flag.Args, os.MkdirTemp.
// Requisitos: Go 1.22 ou superior, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Tarefa é um item da lista. As tags `json:"..."` definem o nome de cada
// campo no arquivo; sem elas, o JSON sairia com os nomes em maiúscula.
type Tarefa struct {
	ID         int    `json:"id"`
	Titulo     string `json:"titulo"`
	Prioridade string `json:"prioridade"`
	Prazo      string `json:"prazo,omitempty"`
	Feita      bool   `json:"feita"`
}

// Lista guarda as tarefas e o próximo ID livre, para nunca repetir número.
type Lista struct {
	ProximoID int      `json:"proximo_id"`
	Tarefas   []Tarefa `json:"tarefas"`
}

// peso ordena as prioridades (menor vem primeiro) e também as valida.
var peso = map[string]int{"alta": 0, "média": 1, "baixa": 2}

func carregar(caminho string) (*Lista, error) {
	l := &Lista{ProximoID: 1} // lista vazia, usada enquanto o arquivo não existe
	dados, err := os.ReadFile(caminho)
	if errors.Is(err, os.ErrNotExist) {
		return l, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(dados, l); err != nil {
		return nil, fmt.Errorf("o arquivo %s não é um JSON válido: %w", caminho, err)
	}
	return l, nil
}

func (l *Lista) adicionar(titulo, prioridade, prazo string) (Tarefa, error) {
	titulo = strings.TrimSpace(titulo)
	if titulo == "" {
		return Tarefa{}, errors.New("o título não pode ficar vazio")
	}
	p := strings.ToLower(strings.TrimSpace(prioridade))
	if p == "" || p == "media" {
		p = "média"
	}
	if _, ok := peso[p]; !ok { // "vírgula ok": pergunta se a chave existe no mapa
		return Tarefa{}, fmt.Errorf("prioridade %q inválida: use alta, média ou baixa", prioridade)
	}
	// time.Parse usa a data de referência fixa do Go (2006-01-02) como molde.
	if _, err := time.Parse("2006-01-02", prazo); prazo != "" && err != nil {
		return Tarefa{}, fmt.Errorf("prazo %q inválido: use o formato AAAA-MM-DD", prazo)
	}
	t := Tarefa{ID: l.ProximoID, Titulo: titulo, Prioridade: p, Prazo: prazo}
	l.ProximoID++
	l.Tarefas = append(l.Tarefas, t)
	return t, nil
}

// ordenadas devolve uma cópia: pendentes antes das feitas, depois por
// prioridade, depois pelo prazo mais próximo (sem prazo vai para o fim).
func (l *Lista) ordenadas() []Tarefa {
	c := append([]Tarefa(nil), l.Tarefas...)
	sort.SliceStable(c, func(i, j int) bool {
		a, b := c[i], c[j]
		if a.Feita != b.Feita {
			return !a.Feita
		}
		if peso[a.Prioridade] != peso[b.Prioridade] {
			return peso[a.Prioridade] < peso[b.Prioridade]
		}
		// O "~" vem depois dos dígitos na tabela ASCII: prazo vazio vai para o fim.
		return a.Prazo+"~" < b.Prazo+"~"
	})
	return c
}

func imprimir(tarefas []Tarefa) {
	if len(tarefas) == 0 {
		fmt.Println("  (nenhuma tarefa)")
		return
	}
	fmt.Printf("  %-3s %-4s %-6s %-10s %s\n", "ID", "Ok", "Prior.", "Prazo", "Título")
	for _, t := range tarefas {
		marca, prazo := "[ ]", "-"
		if t.Feita {
			marca = "[x]"
		}
		if t.Prazo != "" {
			prazo = t.Prazo
		}
		fmt.Printf("  %-3d %-4s %-6s %-10s %s\n", t.ID, marca, t.Prioridade, prazo, t.Titulo)
	}
}

// executar interpreta um comando e devolve true quando a lista mudou e
// precisa ser salva. Assim "list" e "find" nunca regravam o arquivo.
func executar(l *Lista, cmd string, args []string) (bool, error) {
	switch cmd {
	case "add":
		if len(args) < 1 {
			return false, errors.New("uso: add <título> [alta|média|baixa] [AAAA-MM-DD]")
		}
		for len(args) < 3 { // completa prioridade e prazo ausentes com ""
			args = append(args, "")
		}
		t, err := l.adicionar(args[0], args[1], args[2])
		if err != nil {
			return false, err
		}
		fmt.Printf("  Adicionada a tarefa %d: %s (prioridade %s)\n", t.ID, t.Titulo, t.Prioridade)
		return true, nil
	case "list":
		imprimir(l.ordenadas())
	case "find":
		if len(args) < 1 {
			return false, errors.New("uso: find <termo>")
		}
		var achadas []Tarefa
		for _, t := range l.Tarefas {
			if strings.Contains(strings.ToLower(t.Titulo), strings.ToLower(args[0])) {
				achadas = append(achadas, t)
			}
		}
		imprimir(achadas)
	case "done", "rm":
		if len(args) < 1 {
			return false, fmt.Errorf("uso: %s <id>", cmd)
		}
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return false, fmt.Errorf("ID %q inválido: precisa ser um número inteiro", args[0])
		}
		for i, t := range l.Tarefas {
			if t.ID != id {
				continue
			}
			if cmd == "rm" {
				// append(a[:i], a[i+1:]...) é o jeito idiomático de tirar o item i.
				l.Tarefas = append(l.Tarefas[:i], l.Tarefas[i+1:]...)
				fmt.Println("  Removida:", t.Titulo)
				return true, nil
			}
			if t.Feita {
				return false, fmt.Errorf("a tarefa %d já estava concluída", id)
			}
			l.Tarefas[i].Feita = true // pelo índice: t é uma cópia e não mudaria a lista
			fmt.Println("  Concluída:", t.Titulo)
			return true, nil
		}
		return false, fmt.Errorf("não existe tarefa com o ID %d", id)
	default:
		return false, fmt.Errorf("comando %q desconhecido (use add, list, done, rm ou find)", cmd)
	}
	return false, nil
}

// passo carrega a lista, executa um comando e regrava o arquivo se algo mudou.
func passo(caminho string, args []string) error {
	l, err := carregar(caminho)
	if err != nil {
		return err
	}
	mudou, err := executar(l, args[0], args[1:])
	if err != nil || !mudou {
		return err
	}
	dados, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(caminho, dados, 0o644)
}

// demo roda um roteiro completo num arquivo temporário e o apaga ao sair.
func demo() error {
	dir, err := os.MkdirTemp("", "codenomad-tarefas-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	caminho := filepath.Join(dir, "tarefas.json")
	fmt.Println("== Demonstração: lista de tarefas em", caminho)
	roteiro := [][]string{
		{"add", "Revisar capítulo de slices", "alta", "2026-11-03"}, {"add", "Comprar café", "baixa"},
		{"add", "Escrever testes do projeto", "alta", "2026-11-01"},
		{"add", "Ler sobre encoding/json", "média", "2026-11-10"}, {"list"}, {"done", "2"},
		{"find", "json"}, {"rm", "3"}, {"done", "99"}, {"add", "Sem prazo", "urgente"}, {"list"},
	}
	for _, args := range roteiro {
		fmt.Printf("\n$ tarefas %s\n", strings.Join(args, " "))
		if err := passo(caminho, args); err != nil {
			fmt.Println("  Erro:", err)
		}
	}
	dados, err := os.ReadFile(caminho)
	if err == nil {
		fmt.Println("\n== Conteúdo final do arquivo JSON:\n" + string(dados))
	}
	return err
}

func uso() {
	fmt.Fprint(os.Stderr, `Uso: tarefas [-arquivo tarefas.json] <comando> [argumentos]
Comandos: add <título> [alta|média|baixa] [AAAA-MM-DD] | list | find <termo> | done <id> | rm <id>
`)
	flag.PrintDefaults()
}

func main() {
	arquivo := flag.String("arquivo", "tarefas.json", "caminho do arquivo JSON com as tarefas")
	modoDemo := flag.Bool("demo", false, "roda uma demonstração completa num arquivo temporário")
	flag.Usage = uso
	flag.Parse()
	var err error
	if *modoDemo {
		err = demo()
	} else if flag.NArg() == 0 {
		uso()
		os.Exit(2)
	} else {
		err = passo(*arquivo, flag.Args())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Acrescente o comando "edit <id> <novo título>" reaproveitando o laço de done/rm.
// 2. Mostre "atrasada" na listagem quando o prazo já passou (compare com time.Now).
// 3. Crie o comando "stats" com o total de tarefas feitas e pendentes por prioridade.
