// ============================================================
// CodeNomad Academy  |  Sources Go  |  33. Renomeador de arquivos em lote
// ------------------------------------------------------------
// O que faz: renomeia todos os arquivos de uma pasta seguindo um padrão:
//            prefixo, sufixo, numeração sequencial, troca de texto e
//            minúsculas. Sempre mostra a prévia (dry-run) e só aplica com
//            -aplicar; nunca sobrescreve um arquivo que já existe.
// Como usar: go run ./sources/33_renomeador_lote -pasta ./fotos -prefixo viagem_ -numerar -minusculas
//            go run ./sources/33_renomeador_lote -pasta ./docs -trocar "rascunho=final" -aplicar
//            go run ./sources/33_renomeador_lote --demo
// O que você aprende: os.ReadDir e os.Rename, path/filepath para separar
//            nome e extensão, strings.NewReplacer, structs com métodos,
//            maps como conjunto para detectar conflitos e flags combinadas.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Regras reúne tudo o que o usuário pediu. Cada campo vazio ou falso
// significa "não mexer nessa parte do nome".
type Regras struct {
	Prefixo    string
	Sufixo     string
	Trocar     string // no formato "antigo=novo"
	Numerar    bool
	Minusculas bool
}

// Plano é um par "nome atual, nome novo" calculado antes de tocar no disco.
type Plano struct {
	Antigo, Novo string
}

// novoNome aplica as regras a um único nome. A extensão fica de fora da
// numeração e do sufixo para "foto.JPG" virar "01_foto_x.jpg", não "foto.JPG_x".
func (r Regras) novoNome(nome string, indice, total int) (string, error) {
	ext := filepath.Ext(nome)
	base := strings.TrimSuffix(nome, ext)
	if r.Trocar != "" {
		antigo, novo, ok := strings.Cut(r.Trocar, "=")
		if !ok || antigo == "" {
			return "", fmt.Errorf("a troca %q precisa ter o formato antigo=novo", r.Trocar)
		}
		base = strings.NewReplacer(antigo, novo).Replace(base)
	}
	if r.Numerar {
		// A largura do número acompanha o total: 9 arquivos usam 1 dígito, 120 usam 3.
		largura := len(fmt.Sprint(total))
		base = fmt.Sprintf("%0*d_%s", largura, indice, base)
	}
	base = r.Prefixo + base + r.Sufixo
	if r.Minusculas {
		base, ext = strings.ToLower(base), strings.ToLower(ext)
	}
	// Símbolos que o sistema de arquivos não aceita no nome.
	if strings.ContainsAny(base, `/\:*?"<>|`) {
		return "", fmt.Errorf("o nome %q teria símbolos proibidos", base+ext)
	}
	return base + ext, nil
}

// planejar lê a pasta e monta a lista de renomeações sem alterar nada.
// Também recusa o plano se dois arquivos cairiam no mesmo nome final.
func planejar(pasta string, r Regras) ([]Plano, error) {
	entradas, err := os.ReadDir(pasta)
	if err != nil {
		return nil, fmt.Errorf("não consegui ler a pasta: %w", err)
	}
	var nomes []string
	for _, e := range entradas {
		if !e.IsDir() { // subpastas ficam intactas
			nomes = append(nomes, e.Name())
		}
	}
	sort.Strings(nomes) // ReadDir já ordena, mas deixamos explícito para a numeração
	if len(nomes) == 0 {
		return nil, errors.New("a pasta não tem arquivos para renomear")
	}
	usados := map[string]bool{} // conjunto: nome final -> já reservado
	var planos []Plano
	for i, nome := range nomes {
		novo, err := r.novoNome(nome, i+1, len(nomes))
		if err != nil {
			return nil, err
		}
		if novo == nome {
			continue // nada muda neste arquivo
		}
		if usados[novo] {
			return nil, fmt.Errorf("dois arquivos cairiam no mesmo nome %q; ajuste as regras", novo)
		}
		usados[novo] = true
		planos = append(planos, Plano{nome, novo})
	}
	return planos, nil
}

// aplicar executa o plano. Antes de cada os.Rename confere que o destino
// não existe: os.Rename sobrescreveria em silêncio.
func aplicar(pasta string, planos []Plano) (int, error) {
	feitos := 0
	for _, p := range planos {
		destino := filepath.Join(pasta, p.Novo)
		if _, err := os.Stat(destino); err == nil {
			return feitos, fmt.Errorf("parei em %q: já existe um arquivo chamado %q", p.Antigo, p.Novo)
		}
		if err := os.Rename(filepath.Join(pasta, p.Antigo), destino); err != nil {
			return feitos, fmt.Errorf("falha ao renomear %q: %w", p.Antigo, err)
		}
		feitos++
	}
	return feitos, nil
}

func imprimirPlano(planos []Plano) {
	if len(planos) == 0 {
		fmt.Println("  Nenhum arquivo mudaria de nome com essas regras.")
		return
	}
	largura := 0
	for _, p := range planos {
		largura = max(largura, len(p.Antigo)) // max é embutido desde o Go 1.21
	}
	for _, p := range planos {
		fmt.Printf("  %-*s  ->  %s\n", largura, p.Antigo, p.Novo)
	}
	fmt.Printf("  %d arquivo(s) no plano.\n", len(planos))
}

// executar é o fluxo completo: planeja, mostra a prévia e, se pedido, aplica.
func executar(pasta string, r Regras, aplicarMesmo bool) error {
	planos, err := planejar(pasta, r)
	if err != nil {
		return err
	}
	fmt.Println("Prévia (nada foi alterado ainda):")
	imprimirPlano(planos)
	if !aplicarMesmo {
		fmt.Println("Rode de novo com -aplicar para efetivar.")
		return nil
	}
	n, err := aplicar(pasta, planos)
	fmt.Printf("Aplicado: %d arquivo(s) renomeado(s).\n", n)
	return err
}

func listarPasta(pasta string) {
	entradas, _ := os.ReadDir(pasta)
	for _, e := range entradas {
		fmt.Println("   ", e.Name())
	}
}

// demo cria uma pasta temporária com arquivos de exemplo e roda quatro
// cenários: conflito no plano, aplicação, proteção contra sobrescrita e sufixo.
func demo() error {
	dir, err := os.MkdirTemp("", "codenomad-renomear-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // apaga a pasta temporária ao sair, mesmo com erro
	for _, nome := range []string{"IMG_0001.JPG", "IMG_0002.JPG", "IMG_0003.jpg", "foto0001.JPG", "Rascunho Relatorio.TXT", "notas.md"} {
		if err := os.WriteFile(filepath.Join(dir, nome), []byte("exemplo"), 0o644); err != nil {
			return err
		}
	}
	fmt.Println("== Demonstração: renomeador em", dir)
	fmt.Println("Arquivos originais:")
	listarPasta(dir)

	fmt.Println("\n== Cenário 1: prefixo, troca IMG_=foto e minúsculas (dois arquivos colidem)")
	regras := Regras{Prefixo: "viagem_", Trocar: "IMG_=foto", Minusculas: true}
	if err := executar(dir, regras, false); err != nil {
		fmt.Println("  Erro esperado:", err)
	}
	fmt.Println("\n== Cenário 2: as mesmas regras com -numerar desfazem o conflito; aplicado")
	regras.Numerar = true
	if err := executar(dir, regras, true); err != nil {
		return err
	}
	fmt.Println("Arquivos depois:")
	listarPasta(dir)

	fmt.Println("\n== Cenário 3: troca que apontaria para um arquivo que já existe")
	if err := executar(dir, Regras{Trocar: "5_foto0001=1_foto0001"}, true); err != nil {
		fmt.Println("  Erro esperado:", err)
	}
	fmt.Println("\n== Cenário 4: só sufixo _ok, aplicado")
	return executar(dir, Regras{Sufixo: "_ok"}, true)
}

func main() {
	pasta := flag.String("pasta", "", "pasta cujos arquivos serão renomeados")
	var r Regras
	flag.StringVar(&r.Prefixo, "prefixo", "", "texto acrescentado no início do nome")
	flag.StringVar(&r.Sufixo, "sufixo", "", "texto acrescentado antes da extensão")
	flag.StringVar(&r.Trocar, "trocar", "", "troca de texto no formato antigo=novo")
	flag.BoolVar(&r.Numerar, "numerar", false, "numera os arquivos em ordem alfabética (01_, 02_, ...)")
	flag.BoolVar(&r.Minusculas, "minusculas", false, "converte nome e extensão para minúsculas")
	aplicarMesmo := flag.Bool("aplicar", false, "efetiva a renomeação (sem esta flag só mostra a prévia)")
	modoDemo := flag.Bool("demo", false, "roda uma demonstração completa numa pasta temporária")
	flag.Parse()

	var err error
	switch {
	case *modoDemo:
		err = demo()
	case *pasta == "":
		fmt.Fprintln(os.Stderr, "Uso: renomeador -pasta <caminho> [-prefixo x] [-sufixo y] [-trocar a=b] [-numerar] [-minusculas] [-aplicar]")
		flag.PrintDefaults()
		os.Exit(2)
	case r == Regras{}: // structs com campos comparáveis podem ser comparados com ==
		err = errors.New("informe pelo menos uma regra: -prefixo, -sufixo, -trocar, -numerar ou -minusculas")
	default:
		err = executar(*pasta, r, *aplicarMesmo)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Acrescente -recursivo para entrar nas subpastas com filepath.WalkDir.
// 2. Grave um arquivo "desfazer.txt" com os pares antigo/novo e crie a flag -desfazer.
// 3. Aceite -data para prefixar cada nome com a data de modificação (info.ModTime()).
