// ============================================================
// CodeNomad Academy  |  Sources Go  |  32. Agenda de contatos em CSV
// ------------------------------------------------------------
// O que faz: guarda contatos (nome, telefone, e-mail e etiqueta) num arquivo
//            CSV. Você adiciona, busca por qualquer campo, ordena, exporta um
//            filtro para outro CSV e descobre contatos duplicados.
// Como usar: go run ./sources/32_agenda_contatos -arquivo agenda.csv add "Ana Souza" "(11) 99876-1234" ana@exemplo.com amigos
//            go run ./sources/32_agenda_contatos -arquivo agenda.csv -ordem etiqueta listar
//            go run ./sources/32_agenda_contatos --demo
// O que você aprende: encoding/csv para ler e gravar, strings.Map para limpar
//            texto, maps para agrupar, sort, switch sem expressão, defer.
// Requisitos: Go 1.22 ou superior, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Contato é uma linha do CSV. Os campos ficam na mesma ordem das colunas.
type Contato struct {
	Nome, Telefone, Email, Etiqueta string
}

// campos devolve o contato como linha de CSV; a busca e a ordenação também usam.
func (c Contato) campos() []string {
	return []string{c.Nome, c.Telefone, c.Email, c.Etiqueta}
}

// apenasDigitos tira tudo que não é número: "(11) 9.8765-4321" vira "11987654321".
func apenasDigitos(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1 // em strings.Map, devolver -1 descarta o caractere
	}, s)
}

func novoContato(nome, telefone, email, etiqueta string) (Contato, error) {
	arroba := strings.Index(email, "@")
	// switch sem expressão: cada case é uma condição, lida de cima para baixo.
	switch n := len(apenasDigitos(telefone)); {
	case strings.TrimSpace(nome) == "":
		return Contato{}, errors.New("o nome não pode ficar vazio")
	case n < 8 || n > 13:
		return Contato{}, fmt.Errorf("telefone %q inválido: precisa ter de 8 a 13 dígitos", telefone)
	case arroba < 1 || !strings.Contains(email[arroba:], "."):
		return Contato{}, fmt.Errorf("e-mail %q inválido: use o formato nome@dominio.com", email)
	}
	if etiqueta = strings.ToLower(strings.TrimSpace(etiqueta)); etiqueta == "" {
		etiqueta = "geral"
	}
	return Contato{strings.TrimSpace(nome), strings.TrimSpace(telefone), strings.ToLower(email), etiqueta}, nil
}

func carregar(caminho string) ([]Contato, error) {
	f, err := os.Open(caminho)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil // agenda nova: começa vazia
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = 4 // qualquer linha com outro número de colunas vira erro
	linhas, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV inválido em %s: %w", caminho, err)
	}
	var contatos []Contato
	for i, l := range linhas {
		if i == 0 && strings.EqualFold(l[0], "nome") {
			continue // pula o cabeçalho
		}
		contatos = append(contatos, Contato{l[0], l[1], l[2], l[3]})
	}
	return contatos, nil
}

func gravar(caminho string, contatos []Contato) error {
	f, err := os.Create(caminho)
	if err != nil {
		return err
	}
	defer f.Close()
	linhas := [][]string{{"nome", "telefone", "email", "etiqueta"}}
	for _, c := range contatos {
		linhas = append(linhas, c.campos())
	}
	// WriteAll põe aspas quando um campo tem vírgula e já faz o Flush no fim.
	return csv.NewWriter(f).WriteAll(linhas)
}

// buscar devolve os contatos em que qualquer campo contém o termo.
func buscar(contatos []Contato, termo string) []Contato {
	var achados []Contato
	for _, c := range contatos {
		if strings.Contains(strings.ToLower(strings.Join(c.campos(), " ")), strings.ToLower(termo)) {
			achados = append(achados, c)
		}
	}
	return achados
}

func imprimir(contatos []Contato) {
	if len(contatos) == 0 {
		fmt.Println("  (nenhum contato)")
		return
	}
	for _, c := range contatos {
		fmt.Printf("  %-18s %-17s %-26s %s\n", c.Nome, c.Telefone, c.Email, c.Etiqueta)
	}
}

// duplicados imprime os grupos com o mesmo telefone (só dígitos) ou e-mail e devolve quantos achou.
func duplicados(contatos []Contato) int {
	grupos := map[string][]Contato{}
	for _, c := range contatos {
		for _, k := range []string{"telefone " + apenasDigitos(c.Telefone), "e-mail " + strings.ToLower(c.Email)} {
			grupos[k] = append(grupos[k], c)
		}
	}
	chaves := make([]string, 0, len(grupos))
	for k := range grupos {
		chaves = append(chaves, k)
	}
	sort.Strings(chaves) // a ordem de um map é aleatória; ordenamos as chaves
	n := 0
	for _, k := range chaves {
		if len(grupos[k]) > 1 {
			n++
			fmt.Println("  Mesmo", k+":")
			imprimir(grupos[k])
		}
	}
	return n
}

// executar carrega a agenda e roda um comando; só "add" e "exportar" gravam arquivo.
func executar(caminho, ordem, cmd string, args []string) error {
	contatos, err := carregar(caminho)
	if err != nil {
		return err
	}
	switch cmd {
	case "add":
		if len(args) < 3 {
			return errors.New("uso: add <nome> <telefone> <e-mail> [etiqueta]")
		}
		c, err := novoContato(args[0], args[1], args[2], strings.Join(args[3:], " "))
		if err != nil {
			return err
		}
		fmt.Printf("  Adicionado: %s (%s)\n", c.Nome, c.Etiqueta)
		return gravar(caminho, append(contatos, c))
	case "listar":
		pos := map[string]int{"nome": 0, "telefone": 1, "email": 2, "etiqueta": 3}
		i, ok := pos[ordem]
		if !ok {
			return fmt.Errorf("campo de ordem %q inválido: use nome, telefone, email ou etiqueta", ordem)
		}
		// SliceStable mantém a ordem original entre contatos com o mesmo valor.
		sort.SliceStable(contatos, func(a, b int) bool {
			return strings.ToLower(contatos[a].campos()[i]) < strings.ToLower(contatos[b].campos()[i])
		})
		imprimir(contatos)
	case "buscar":
		if len(args) < 1 {
			return errors.New("uso: buscar <termo>")
		}
		imprimir(buscar(contatos, args[0]))
	case "exportar":
		if len(args) < 2 {
			return errors.New("uso: exportar <termo> <saida.csv>")
		}
		achados := buscar(contatos, args[0])
		fmt.Printf("  %d contato(s) com %q exportado(s) para %s\n", len(achados), args[0], args[1])
		return gravar(args[1], achados)
	case "duplicados":
		if duplicados(contatos) == 0 {
			fmt.Println("  Nenhum duplicado.")
		}
	default:
		return fmt.Errorf("comando %q desconhecido (use add, listar, buscar, exportar ou duplicados)", cmd)
	}
	return nil
}

func demo() error {
	dir, err := os.MkdirTemp("", "codenomad-agenda-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	caminho := filepath.Join(dir, "agenda.csv")
	exemplo := []Contato{
		{"Ana Souza", "(11) 99876-1234", "ana.souza@exemplo.com", "amigos"},
		{"Bruno Lima", "(21) 98765-4321", "bruno@exemplo.com", "trabalho"},
		{"Ana Souza", "11998761234", "ana.souza@exemplo.com", "família"},
		{"Diego Rocha", "(41) 99111-2222", "Diego.Rocha@exemplo.com", "amigos"},
		{"Diego Rocha", "(41) 99111-2223", "diego.rocha@exemplo.com", "academia"},
	}
	if err := gravar(caminho, exemplo); err != nil {
		return err
	}
	fmt.Println("== Demonstração: agenda em", caminho)
	roteiro := [][]string{
		{"listar"}, {"buscar", "99111"}, {"add", "Elisa Prado", "(51) 3030-4040", "elisa@exemplo.com", "trabalho"},
		{"add", "Fulano", "123", "sem-arroba", "x"}, {"buscar", "trabalho"}, {"duplicados"},
		{"exportar", "trabalho", filepath.Join(dir, "trabalho.csv")},
	}
	for _, args := range roteiro {
		fmt.Printf("\n$ agenda %s\n", strings.Join(args, " "))
		if err := executar(caminho, "etiqueta", args[0], args[1:]); err != nil {
			fmt.Println("  Erro:", err)
		}
	}
	return nil
}

func main() {
	arquivo := flag.String("arquivo", "agenda.csv", "caminho do arquivo CSV da agenda")
	ordem := flag.String("ordem", "nome", "campo de ordenação em listar: nome, telefone, email ou etiqueta")
	modoDemo := flag.Bool("demo", false, "roda uma demonstração completa num arquivo temporário")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Uso: agenda [-arquivo agenda.csv] [-ordem campo] <comando> [argumentos]")
		fmt.Fprintln(os.Stderr, "Comandos: add <nome> <telefone> <e-mail> [etiqueta] | listar | buscar <termo> | exportar <termo> <saida.csv> | duplicados")
		flag.PrintDefaults()
	}
	flag.Parse()
	var err error
	if *modoDemo {
		err = demo()
	} else if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	} else {
		err = executar(*arquivo, *ordem, flag.Arg(0), flag.Args()[1:])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Crie o comando "remover <e-mail>" e trate o caso de e-mail não encontrado.
// 2. Em "duplicados", junte cada grupo num só contato e grave a agenda limpa.
// 3. Aceite um quinto campo "aniversário" e liste quem faz anos no mês atual.
