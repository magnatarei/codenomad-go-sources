// ============================================================
// CodeNomad Academy  |  Sources Go  |  35. Detector de arquivos duplicados
// ------------------------------------------------------------
// O que faz: percorre uma pasta (com subpastas), agrupa os arquivos primeiro
//            por tamanho e depois por SHA-256, lista cada grupo de duplicados
//            e calcula quanto espaço seria liberado mantendo uma cópia de cada.
//            Com -relatorio grava o resultado num arquivo de texto.
// Como usar: go run ./sources/35_duplicados_hash -pasta ./fotos
//            go run ./sources/35_duplicados_hash -pasta ./fotos -min 1024 -relatorio duplicados.txt
//            go run ./sources/35_duplicados_hash --demo
// O que você aprende: filepath.WalkDir, crypto/sha256 com io.Copy (sem carregar
//            o arquivo inteiro na memória), maps de slices para agrupar,
//            sort.Slice, formatação de bytes legíveis e strings.Builder.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Arquivo guarda o que precisamos saber de cada arquivo encontrado.
type Arquivo struct {
	Caminho string
	Tamanho int64
}

// Grupo é um conjunto de arquivos idênticos (mesmo tamanho e mesmo hash).
type Grupo struct {
	Hash     string
	Tamanho  int64
	Arquivos []Arquivo
}

// Desperdicio é o espaço ocupado pelas cópias extras: todas menos uma.
func (g Grupo) Desperdicio() int64 {
	return g.Tamanho * int64(len(g.Arquivos)-1)
}

// hashDe calcula o SHA-256 lendo o arquivo em blocos: io.Copy manda os bytes
// para o hasher e funciona igual para um arquivo de 1 KB ou de 10 GB.
func hashDe(caminho string) (string, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// listar percorre a pasta e devolve os arquivos com o tamanho mínimo pedido.
func listar(pasta string, minimo int64) ([]Arquivo, error) {
	info, err := os.Stat(pasta)
	if err != nil {
		return nil, fmt.Errorf("não consegui abrir a pasta: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s não é uma pasta", pasta)
	}
	var arquivos []Arquivo
	err = filepath.WalkDir(pasta, func(caminho string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() >= minimo {
			arquivos = append(arquivos, Arquivo{caminho, info.Size()})
		}
		return nil
	})
	return arquivos, err
}

// agrupar faz o trabalho em duas etapas. Tamanhos diferentes nunca são
// duplicados, então só calculamos o hash (caro) de quem tem um "gêmeo" de tamanho.
func agrupar(arquivos []Arquivo) ([]Grupo, int, error) {
	porTamanho := map[int64][]Arquivo{}
	for _, a := range arquivos {
		porTamanho[a.Tamanho] = append(porTamanho[a.Tamanho], a)
	}
	porHash := map[string]*Grupo{}
	hashes := 0
	for tamanho, lista := range porTamanho {
		if len(lista) < 2 {
			continue
		}
		for _, a := range lista {
			h, err := hashDe(a.Caminho)
			if err != nil {
				return nil, hashes, err
			}
			hashes++
			if porHash[h] == nil {
				porHash[h] = &Grupo{Hash: h, Tamanho: tamanho}
			}
			porHash[h].Arquivos = append(porHash[h].Arquivos, a)
		}
	}
	var grupos []Grupo
	for _, g := range porHash {
		if len(g.Arquivos) > 1 {
			sort.Slice(g.Arquivos, func(i, j int) bool { return g.Arquivos[i].Caminho < g.Arquivos[j].Caminho })
			grupos = append(grupos, *g)
		}
	}
	// Grupos que mais desperdiçam primeiro; empate decidido pelo hash para ser estável.
	sort.Slice(grupos, func(i, j int) bool {
		if grupos[i].Desperdicio() != grupos[j].Desperdicio() {
			return grupos[i].Desperdicio() > grupos[j].Desperdicio()
		}
		return grupos[i].Hash < grupos[j].Hash
	})
	return grupos, hashes, nil
}

// legivel converte bytes em B, KB, MB ou GB com uma casa depois da vírgula.
func legivel(n int64) string {
	const kb = 1024.0
	v := float64(n)
	for _, u := range []string{"B", "KB", "MB", "GB"} {
		if v < kb || u == "GB" {
			s := strings.Replace(fmt.Sprintf("%.1f %s", v, u), ".0 ", " ", 1)
			return strings.Replace(s, ".", ",", 1) // vírgula, como se escreve no Brasil
		}
		v /= kb
	}
	return ""
}

// montarRelatorio escreve o texto numa strings.Builder; serve tanto para a
// tela quanto para o arquivo, sem duplicar a lógica de formatação.
func montarRelatorio(pasta string, total int, hashes int, grupos []Grupo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Pasta analisada: %s\n", pasta)
	fmt.Fprintf(&b, "Arquivos examinados: %d (hash calculado em %d)\n", total, hashes)
	if len(grupos) == 0 {
		b.WriteString("Nenhum arquivo duplicado encontrado.\n")
		return b.String()
	}
	var liberavel int64
	for i, g := range grupos {
		liberavel += g.Desperdicio()
		fmt.Fprintf(&b, "\nGrupo %d: %d cópias de %s (sha256 %s...)\n", i+1, len(g.Arquivos), legivel(g.Tamanho), g.Hash[:12])
		for _, a := range g.Arquivos {
			rel, err := filepath.Rel(pasta, a.Caminho)
			if err != nil {
				rel = a.Caminho
			}
			fmt.Fprintf(&b, "  %s\n", filepath.ToSlash(rel))
		}
	}
	fmt.Fprintf(&b, "\n%d grupo(s) de duplicados. Mantendo uma cópia de cada, você libera %s.\n", len(grupos), legivel(liberavel))
	return b.String()
}

func executar(pasta string, minimo int64, relatorio string) error {
	arquivos, err := listar(pasta, minimo)
	if err != nil {
		return err
	}
	grupos, hashes, err := agrupar(arquivos)
	if err != nil {
		return err
	}
	texto := montarRelatorio(pasta, len(arquivos), hashes, grupos)
	fmt.Print(texto)
	if relatorio != "" {
		if err := os.WriteFile(relatorio, []byte(texto), 0o644); err != nil {
			return fmt.Errorf("não consegui gravar o relatório: %w", err)
		}
		fmt.Println("Relatório gravado em", relatorio)
	}
	return nil
}

// demo monta uma pasta temporária com cópias espalhadas em subpastas.
func demo() error {
	dir, err := os.MkdirTemp("", "codenomad-duplicados-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	foto, texto := strings.Repeat("pixel", 600), strings.Repeat("linha\n", 50) // 3000 e 300 bytes
	conteudos := map[string]string{
		"fotos/praia.jpg": foto, "fotos/2025/praia (1).jpg": foto, "backup/praia_copia.jpg": foto,
		"docs/notas.txt": texto, "backup/notas.txt": texto,
		"docs/outro.txt": strings.Repeat("outra\n", 50), // mesmo tamanho de notas.txt, conteúdo diferente
		"docs/unico.md":  "só existe um assim",
	}
	for nome, c := range conteudos {
		caminho := filepath.Join(dir, filepath.FromSlash(nome))
		if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(caminho, []byte(c), 0o644); err != nil {
			return err
		}
	}
	fmt.Println("== Demonstração: duplicados em", dir)
	fmt.Println("(outro.txt tem o mesmo tamanho de notas.txt, mas o hash mostra que é diferente)")
	relatorio := filepath.Join(dir, "relatorio.txt")
	if err := executar(dir, 1, relatorio); err != nil {
		return err
	}
	fmt.Println("\n== Mesma pasta ignorando arquivos menores que 1 KB:")
	return executar(dir, 1024, "")
}

func main() {
	pasta := flag.String("pasta", "", "pasta a examinar (inclui subpastas)")
	minimo := flag.Int64("min", 1, "ignora arquivos menores que este tamanho em bytes")
	relatorio := flag.String("relatorio", "", "grava o relatório neste arquivo de texto")
	modoDemo := flag.Bool("demo", false, "roda uma demonstração completa numa pasta temporária")
	flag.Parse()

	var err error
	switch {
	case *modoDemo:
		err = demo()
	case *pasta == "":
		fmt.Fprintln(os.Stderr, "Uso: duplicados -pasta <caminho> [-min bytes] [-relatorio saida.txt]")
		flag.PrintDefaults()
		os.Exit(2)
	case *minimo < 0:
		err = errors.New("o tamanho mínimo não pode ser negativo")
	default:
		err = executar(*pasta, *minimo, *relatorio)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Acrescente -apagar que remove as cópias extras de cada grupo (mantendo a primeira) com confirmação.
// 2. Calcule os hashes em paralelo com goroutines e um canal, e meça o ganho com time.Since.
// 3. Compare só os primeiros 4 KB antes do SHA-256 completo, para descartar rápido arquivos grandes.
