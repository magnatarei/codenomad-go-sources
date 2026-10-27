// ============================================================
// CodeNomad Academy  |  Sources Go  |  40. Backup de pasta com data
// ------------------------------------------------------------
// O que faz: copia uma pasta para um destino com carimbo de data e hora no
//            nome (projeto_2026-10-21_090000), ignorando padrões. Mostra o que
//            copiou e compara com o backup anterior por SHA-256 (o que mudou).
// Como usar: go run ./sources/40_backup_pasta -origem ./projeto -destino ./backups
//            go run ./sources/40_backup_pasta -origem ./projeto -destino ./backups -ignorar "*.tmp,node_modules,.git"
//            go run ./sources/40_backup_pasta --demo
// O que você aprende: filepath.WalkDir e Match, cópia com io.Copy, os.MkdirAll,
//            time.Format para carimbos, crypto/sha256, maps para comparar e sort.
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
	"time"
)

// Inventario liga o caminho relativo de cada arquivo ao hash do conteúdo.
type Inventario map[string]string

// ignorado testa cada padrão contra o nome; filepath.Match entende curingas como "*.tmp".
func ignorado(nome string, padroes []string) bool {
	for _, p := range padroes {
		if ok, _ := filepath.Match(p, nome); ok {
			return true
		}
	}
	return false
}

// inventariar percorre a pasta pulando o que for ignorado (fs.SkipDir corta a subpasta
// inteira). O hash lê o arquivo todo na memória; para arquivos enormes, veja o source 35.
func inventariar(pasta string, padroes []string) (Inventario, error) {
	inv := Inventario{}
	err := filepath.WalkDir(pasta, func(caminho string, d fs.DirEntry, err error) error {
		if err != nil || caminho == pasta {
			return err
		}
		if ignorado(d.Name(), padroes) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		dados, err := os.ReadFile(caminho)
		if err != nil {
			return err
		}
		soma := sha256.Sum256(dados)
		rel, _ := filepath.Rel(pasta, caminho)
		inv[filepath.ToSlash(rel)] = hex.EncodeToString(soma[:])
		return nil
	})
	return inv, err
}

// copiarArquivo copia com io.Copy; O_EXCL faz o OpenFile falhar se o destino já existir.
func copiarArquivo(origem, destino string) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
		return 0, err
	}
	in, err := os.Open(origem)
	if err != nil {
		return 0, err
	}
	defer in.Close()
	out, err := os.OpenFile(destino, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	return io.Copy(out, in)
}

func ordenadas(inv Inventario) []string {
	chaves := make([]string, 0, len(inv))
	for k := range inv {
		chaves = append(chaves, k)
	}
	sort.Strings(chaves)
	return chaves
}

// anterior acha o backup mais recente do mesmo prefixo: o carimbo começa pelo ano, então a ordem alfabética é a cronológica.
func anterior(destino, prefixo string) string {
	entradas, _ := os.ReadDir(destino) // destino inexistente: sem anterior
	ultimo := ""
	for _, e := range entradas {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefixo+"_") && e.Name() > ultimo {
			ultimo = e.Name()
		}
	}
	return ultimo
}

// comparar lista o que mudou entre o backup anterior e o atual.
func comparar(antigo, novo Inventario) (adicionados, alterados, removidos []string) {
	for _, k := range ordenadas(novo) {
		h, existia := antigo[k]
		switch {
		case !existia:
			adicionados = append(adicionados, k)
		case h != novo[k]:
			alterados = append(alterados, k)
		}
	}
	for _, k := range ordenadas(antigo) {
		if _, existe := novo[k]; !existe {
			removidos = append(removidos, k)
		}
	}
	return adicionados, alterados, removidos
}

// fazerBackup é o fluxo completo: inventaria, acha o anterior, copia e compara.
func fazerBackup(origem, destino string, padroes []string, carimbo time.Time) error {
	info, err := os.Stat(origem)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("a origem %q não é uma pasta acessível", origem)
	}
	prefixo := filepath.Base(filepath.Clean(origem))
	anteriorNome := anterior(destino, prefixo)
	pastaNova := filepath.Join(destino, prefixo+"_"+carimbo.Format("2006-01-02_150405"))
	if _, err := os.Stat(pastaNova); err == nil {
		return fmt.Errorf("já existe um backup em %s; espere um segundo e tente de novo", pastaNova)
	}
	inv, err := inventariar(origem, padroes)
	if err != nil {
		return err
	}
	if len(inv) == 0 {
		return errors.New("não há nada para copiar depois de aplicar os padrões ignorados")
	}
	var bytes int64
	for _, rel := range ordenadas(inv) {
		n, err := copiarArquivo(filepath.Join(origem, filepath.FromSlash(rel)), filepath.Join(pastaNova, filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("falha ao copiar %s: %w", rel, err)
		}
		bytes += n
		fmt.Printf("  copiado  %s (%d bytes)\n", rel, n)
	}
	fmt.Printf("Backup pronto: %d arquivo(s), %d bytes em %s\n", len(inv), bytes, pastaNova)
	if anteriorNome == "" {
		fmt.Println("Primeiro backup desta pasta: não há com o que comparar.")
		return nil
	}
	invAntigo, err := inventariar(filepath.Join(destino, anteriorNome), nil)
	if err != nil {
		return err
	}
	adicionados, alterados, removidos := comparar(invAntigo, inv)
	fmt.Printf("Comparado com %s:\n", anteriorNome)
	if len(adicionados)+len(alterados)+len(removidos) == 0 {
		fmt.Println("  Nada mudou desde o backup anterior.")
		return nil
	}
	titulos := []string{"Adicionados", "Alterados", "Removidos"}
	for i, lista := range [][]string{adicionados, alterados, removidos} {
		fmt.Printf("  %s (%d)\n", titulos[i], len(lista))
		for _, item := range lista {
			fmt.Println("    ", item)
		}
	}
	return nil
}

// separarPadroes transforma "*.tmp, .git" em ["*.tmp", ".git"]: Fields ignora espaços sobrando.
func separarPadroes(lista string) []string {
	return strings.Fields(strings.ReplaceAll(lista, ",", " "))
}

func demo() error {
	dir, err := os.MkdirTemp("", "codenomad-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	origem, destino := filepath.Join(dir, "projeto"), filepath.Join(dir, "backups")
	arquivos := map[string]string{
		"README.md": "# Projeto\n", "main.go": "package main\n", "docs/guia.txt": "passo 1\n",
		"build/saida.tmp": "lixo", "cache/x.bin": "cache",
	}
	for rel, c := range arquivos {
		caminho := filepath.Join(origem, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(caminho, []byte(c), 0o644); err != nil {
			return err
		}
	}
	padroes := separarPadroes("*.tmp, cache")
	fmt.Println("== Demonstração: origem", origem)
	fmt.Println("Ignorando:", strings.Join(padroes, ", "))
	fmt.Println("\n== Backup 1")
	carimbo := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	if err := fazerBackup(origem, destino, padroes, carimbo); err != nil {
		return err
	}
	fmt.Println("\n== Alterações: main.go muda, docs/novo.txt aparece, README.md some")
	_ = os.WriteFile(filepath.Join(origem, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644)
	_ = os.WriteFile(filepath.Join(origem, "docs", "novo.txt"), []byte("novidade\n"), 0o644)
	_ = os.Remove(filepath.Join(origem, "README.md"))
	fmt.Println("\n== Backup 2 (um dia depois)")
	if err := fazerBackup(origem, destino, padroes, carimbo.AddDate(0, 0, 1)); err != nil {
		return err
	}
	fmt.Println("\n== Backup 3 com o mesmo carimbo do 2 (deve ser recusado)")
	if err := fazerBackup(origem, destino, padroes, carimbo.AddDate(0, 0, 1)); err != nil {
		fmt.Println("  Erro esperado:", err)
	}
	return nil
}

func main() {
	origem := flag.String("origem", "", "pasta a copiar")
	destino := flag.String("destino", "", "pasta onde os backups são criados")
	ignorar := flag.String("ignorar", "", "padrões separados por vírgula, ex.: \"*.tmp,node_modules,.git\"")
	modoDemo := flag.Bool("demo", false, "roda uma demonstração completa em pastas temporárias")
	flag.Parse()
	var err error
	switch {
	case *modoDemo:
		err = demo()
	case *origem == "" || *destino == "":
		fmt.Fprintln(os.Stderr, "Uso: backup -origem <pasta> -destino <pasta> [-ignorar \"*.tmp,.git\"]")
		flag.PrintDefaults()
		os.Exit(2)
	default:
		err = fazerBackup(*origem, *destino, separarPadroes(*ignorar), time.Now())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Acrescente -manter N para apagar os backups mais antigos e deixar só os N últimos.
// 2. Compacte o backup num .zip com archive/zip em vez de copiar pasta por pasta.
// 3. Crie -restaurar <nome_do_backup> que copia de volta para a origem sem sobrescrever o que é mais novo.
