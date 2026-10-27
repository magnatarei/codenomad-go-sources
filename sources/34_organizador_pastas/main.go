// ============================================================
// CodeNomad Academy  |  Sources Go  |  34. Organizador de pasta por tipo
// ------------------------------------------------------------
// O que faz: move os arquivos soltos de uma pasta para subpastas por
//            categoria (imagens, documentos, planilhas, áudio, vídeo, código
//            e outros) olhando a extensão. Mostra a prévia, imprime um relatório
//            e grava um registro que permite desfazer tudo com -desfazer.
// Como usar: go run ./sources/34_organizador_pastas -pasta ./Downloads
//            go run ./sources/34_organizador_pastas -pasta ./Downloads -aplicar
//            go run ./sources/34_organizador_pastas -pasta ./Downloads -desfazer
//            go run ./sources/34_organizador_pastas --demo
// O que você aprende: maps de extensão para categoria, os.ReadDir, os.MkdirAll,
//            os.Rename, encoding/json para gravar o registro do que foi
//            movido, sort para um relatório estável e flags que se excluem.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
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
	"strings"
)

// registro é o nome do arquivo que guarda o que foi movido, para desfazer.
const registro = ".organizador.json"

// categorias liga cada extensão (em minúsculas, com ponto) a uma pasta.
// Um map é a estrutura certa aqui: a busca é por chave exata e instantânea.
var categorias = map[string]string{
	".jpg": "imagens", ".jpeg": "imagens", ".png": "imagens", ".gif": "imagens", ".webp": "imagens", ".svg": "imagens",
	".pdf": "documentos", ".doc": "documentos", ".docx": "documentos", ".txt": "documentos", ".md": "documentos", ".odt": "documentos",
	".xls": "planilhas", ".xlsx": "planilhas", ".csv": "planilhas", ".ods": "planilhas",
	".mp3": "audio", ".wav": "audio", ".ogg": "audio", ".flac": "audio",
	".mp4": "video", ".mkv": "video", ".avi": "video", ".mov": "video",
	".go": "codigo", ".py": "codigo", ".js": "codigo", ".ts": "codigo", ".java": "codigo", ".rs": "codigo", ".sh": "codigo",
}

// Movimento é uma linha do plano e também uma linha do registro em JSON.
type Movimento struct {
	De   string `json:"de"`   // caminho relativo à pasta, antes
	Para string `json:"para"` // caminho relativo à pasta, depois
}

// categoriaDe decide a subpasta de um arquivo. Extensão desconhecida vai
// para "outros", e assim nenhum arquivo fica para trás.
func categoriaDe(nome string) string {
	if c, ok := categorias[strings.ToLower(filepath.Ext(nome))]; ok {
		return c
	}
	return "outros"
}

// planejar lista os arquivos soltos da pasta (subpastas não entram) e diz
// para onde cada um iria. Nada é movido aqui.
func planejar(pasta string) ([]Movimento, error) {
	entradas, err := os.ReadDir(pasta)
	if err != nil {
		return nil, fmt.Errorf("não consegui ler a pasta: %w", err)
	}
	var plano []Movimento
	for _, e := range entradas {
		nome := e.Name()
		if e.IsDir() || nome == registro || strings.HasPrefix(nome, ".") {
			continue // pastas e arquivos ocultos ficam onde estão
		}
		plano = append(plano, Movimento{De: nome, Para: filepath.Join(categoriaDe(nome), nome)})
	}
	if len(plano) == 0 {
		return nil, errors.New("não há arquivos soltos para organizar nessa pasta")
	}
	return plano, nil
}

// mover executa cada movimento, criando a subpasta se preciso, e recusa
// sobrescrever um arquivo que já exista no destino.
func mover(pasta string, plano []Movimento) ([]Movimento, error) {
	var feitos []Movimento
	for _, m := range plano {
		destino := filepath.Join(pasta, m.Para)
		if _, err := os.Stat(destino); err == nil {
			return feitos, fmt.Errorf("já existe %s; parei antes de sobrescrever", m.Para)
		}
		// MkdirAll não reclama se a pasta já existe: ideal para "garanta que exista".
		if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
			return feitos, err
		}
		if err := os.Rename(filepath.Join(pasta, m.De), destino); err != nil {
			return feitos, fmt.Errorf("falha ao mover %s: %w", m.De, err)
		}
		feitos = append(feitos, m)
	}
	return feitos, nil
}

// relatorio agrupa o plano por categoria e imprime em ordem alfabética.
func relatorio(plano []Movimento) {
	porCategoria := map[string][]string{}
	for _, m := range plano {
		cat := filepath.Dir(m.Para)
		porCategoria[cat] = append(porCategoria[cat], m.De)
	}
	nomes := make([]string, 0, len(porCategoria))
	for cat := range porCategoria {
		nomes = append(nomes, cat)
	}
	sort.Strings(nomes) // a ordem de iteração de um map muda a cada execução
	for _, cat := range nomes {
		arquivos := porCategoria[cat]
		fmt.Printf("  %-11s %2d arquivo(s): %s\n", cat+"/", len(arquivos), strings.Join(arquivos, ", "))
	}
	fmt.Printf("  Total: %d arquivo(s) em %d categoria(s).\n", len(plano), len(nomes))
}

// organizar é o fluxo principal: prévia sempre, movimentação só com aplicar.
func organizar(pasta string, aplicar bool) error {
	plano, err := planejar(pasta)
	if err != nil {
		return err
	}
	fmt.Println("Prévia da organização:")
	relatorio(plano)
	if !aplicar {
		fmt.Println("Rode com -aplicar para mover de verdade.")
		return nil
	}
	feitos, errMover := mover(pasta, plano)
	// Grava o registro mesmo se parou no meio: o que já foi movido pode ser desfeito.
	dados, err := json.MarshalIndent(feitos, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(pasta, registro), dados, 0o644); err != nil {
		return err
	}
	fmt.Printf("Movidos %d arquivo(s). Registro gravado em %s (use -desfazer para voltar).\n", len(feitos), registro)
	return errMover
}

// desfazer lê o registro, devolve cada arquivo ao lugar original e apaga as
// subpastas que ficaram vazias. Por fim remove o próprio registro.
func desfazer(pasta string) error {
	dados, err := os.ReadFile(filepath.Join(pasta, registro))
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("não há registro de organização nessa pasta; nada a desfazer")
	}
	if err != nil {
		return err
	}
	var feitos []Movimento
	if err := json.Unmarshal(dados, &feitos); err != nil {
		return fmt.Errorf("registro corrompido: %w", err)
	}
	for i := len(feitos) - 1; i >= 0; i-- { // ordem inversa, como num "ctrl+z"
		m := feitos[i]
		if err := os.Rename(filepath.Join(pasta, m.Para), filepath.Join(pasta, m.De)); err != nil {
			return fmt.Errorf("não consegui devolver %s: %w", m.Para, err)
		}
		// os.Remove só apaga a pasta se ela estiver vazia; se não, o erro é ignorado de propósito.
		_ = os.Remove(filepath.Join(pasta, filepath.Dir(m.Para)))
	}
	fmt.Printf("Desfeito: %d arquivo(s) voltaram para a raiz da pasta.\n", len(feitos))
	return os.Remove(filepath.Join(pasta, registro))
}

// arvore imprime a pasta e as subpastas de forma simples, para o demo.
func arvore(pasta string) {
	_ = filepath.WalkDir(pasta, func(caminho string, d os.DirEntry, err error) error {
		if err != nil || caminho == pasta {
			return err
		}
		rel, _ := filepath.Rel(pasta, caminho)
		if d.IsDir() {
			rel += "/"
		}
		fmt.Println("   ", filepath.ToSlash(rel))
		return nil
	})
}

func demo() error {
	dir, err := os.MkdirTemp("", "codenomad-organizar-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	nomes := []string{"ferias.JPG", "logo.png", "contrato.pdf", "notas.md", "vendas.xlsx", "clientes.csv",
		"musica.mp3", "aula.mp4", "main.go", "script.py", "backup.zip", "sem_extensao"}
	for _, n := range nomes {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("exemplo"), 0o644); err != nil {
			return err
		}
	}
	fmt.Println("== Demonstração: pasta bagunçada em", dir)
	arvore(dir)
	fmt.Println("\n== Passo 1: só prévia")
	if err := organizar(dir, false); err != nil {
		return err
	}
	fmt.Println("\n== Passo 2: aplicar")
	if err := organizar(dir, true); err != nil {
		return err
	}
	fmt.Println("Pasta depois de organizar:")
	arvore(dir)
	fmt.Println("\n== Passo 3: organizar de novo (não há mais arquivos soltos)")
	if err := organizar(dir, true); err != nil {
		fmt.Println("  Aviso:", err)
	}
	fmt.Println("\n== Passo 4: desfazer")
	if err := desfazer(dir); err != nil {
		return err
	}
	fmt.Println("Pasta depois de desfazer:")
	arvore(dir)
	return nil
}

func main() {
	pasta := flag.String("pasta", "", "pasta a organizar")
	aplicar := flag.Bool("aplicar", false, "move os arquivos (sem esta flag só mostra a prévia)")
	voltar := flag.Bool("desfazer", false, "desfaz a última organização usando o registro")
	modoDemo := flag.Bool("demo", false, "roda uma demonstração completa numa pasta temporária")
	flag.Parse()

	var err error
	switch {
	case *modoDemo:
		err = demo()
	case *pasta == "":
		fmt.Fprintln(os.Stderr, "Uso: organizador -pasta <caminho> [-aplicar | -desfazer]")
		flag.PrintDefaults()
		os.Exit(2)
	case *aplicar && *voltar:
		err = errors.New("escolha -aplicar ou -desfazer, não os dois")
	case *voltar:
		err = desfazer(*pasta)
	default:
		err = organizar(*pasta, *aplicar)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Deixe o usuário acrescentar categorias com -regra ".psd=design" (strings.Cut ajuda).
// 2. Separe as imagens por ano de modificação: imagens/2026/ferias.jpg (info.ModTime()).
// 3. Some o tamanho dos arquivos por categoria no relatório e mostre em KB ou MB.
