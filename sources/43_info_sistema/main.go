// ============================================================
// CodeNomad Academy  |  Sources Go  |  43. Informações do sistema
// ------------------------------------------------------------
// O que faz: mostra sistema operacional e arquitetura, versão do Go, máquina,
//            usuário, pasta inspecionada, CPUs, memória do processo, variáveis
//            de ambiente filtradas por prefixo e o espaço em disco da pasta.
//            A saída pode ser em texto alinhado ou em JSON.
// Como usar: go run ./sources/43_info_sistema .
//            go run ./sources/43_info_sistema --env GO,PATH --json C:/projetos
//            go run ./sources/43_info_sistema --demo
// O que você aprende: os pacotes runtime, os, os/user e os/exec, structs com
//            etiquetas json e ponteiro opcional (omitempty), strings.Cut para
//            separar NOME=valor, exec.CommandContext com prazo e o costume de
//            seguir em frente quando uma informação não está disponível.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Relatorio reúne tudo o que foi medido. As etiquetas `json:"..."` definem o
// nome de cada campo na saída em JSON; omitempty some com o campo vazio.
type Relatorio struct {
	Sistema      string   `json:"sistema"`
	Arquitetura  string   `json:"arquitetura"`
	VersaoGo     string   `json:"versao_go"`
	Maquina      string   `json:"maquina"`
	Usuario      string   `json:"usuario"`
	PastaUsuario string   `json:"pasta_usuario"`
	Hora         string   `json:"hora"`
	Pasta        string   `json:"pasta"`
	Executavel   string   `json:"executavel"`
	PID          int      `json:"pid"`
	CPUs         int      `json:"cpus"`
	MemoriaMB    float64  `json:"memoria_processo_mb"`
	ReservadaMB  float64  `json:"memoria_reservada_mb"`
	Disco        *Disco   `json:"disco,omitempty"`
	AvisoDisco   string   `json:"aviso_disco,omitempty"`
	Ambiente     []string `json:"ambiente"`
}

// Disco descreve o volume onde a pasta inspecionada está.
type Disco struct {
	TotalGB float64 `json:"total_gb"`
	LivreGB float64 `json:"livre_gb"`
	UsoPct  float64 `json:"uso_pct"`
}

func main() {
	prefixos := flag.String("env", "GO,PATH,HOME,USER,TEMP,LANG,OS,SHELL", "prefixos das variáveis de ambiente a mostrar, separados por vírgula")
	emJSON := flag.Bool("json", false, "imprime o relatório em JSON em vez de texto")
	demo := flag.Bool("demo", false, "mostra tudo o que conseguir do sistema atual")
	flag.Parse()

	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "Uso: info_sistema [--env GO,PATH] [--json] PASTA   ou   --demo")
		flag.PrintDefaults()
		os.Exit(2)
	}
	pasta := "."
	if flag.NArg() > 0 {
		pasta = flag.Arg(0)
	}
	rel, err := coletar(pasta, strings.Split(*prefixos, ","))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
	if *emJSON {
		texto, err := json.MarshalIndent(rel, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Erro ao gerar JSON:", err)
			os.Exit(1)
		}
		fmt.Println(string(texto))
		return
	}
	if *demo {
		fmt.Println("=== CodeNomad Academy | Informações do sistema: demonstração ===")
	}
	imprimir(rel)
	if *demo {
		fmt.Println("\nDica: rode com --json para receber o mesmo relatório pronto para outro programa ler.")
	}
}

// coletar mede tudo. Nem toda informação existe em todo sistema: quando uma
// chamada falha, guardamos um texto neutro em vez de derrubar o relatório.
func coletar(pasta string, prefixos []string) (Relatorio, error) {
	absoluta, err := filepath.Abs(pasta)
	if err != nil {
		return Relatorio{}, err
	}
	if info, err := os.Stat(absoluta); err != nil || !info.IsDir() {
		return Relatorio{}, fmt.Errorf("a pasta %q não existe ou não é uma pasta", pasta)
	}
	r := Relatorio{
		Sistema: runtime.GOOS, Arquitetura: runtime.GOARCH, VersaoGo: runtime.Version(),
		CPUs: runtime.NumCPU(), PID: os.Getpid(), Pasta: absoluta,
		Hora: time.Now().Format("2006-01-02 15:04:05 MST"),
	}
	// Uma função que devolve (valor, erro) pode ser passada direto para outra
	// que recebe exatamente esses dois parâmetros: é o que ouDesconhecido faz.
	r.Maquina = ouDesconhecido(os.Hostname())
	r.Executavel = ouDesconhecido(os.Executable())
	r.PastaUsuario = ouDesconhecido(os.UserHomeDir())
	r.Usuario = "desconhecido"
	if u, err := user.Current(); err == nil {
		r.Usuario = u.Username
	}
	// O runtime só conhece a memória do próprio processo; a memória total da
	// máquina não é exposta pela biblioteca padrão de forma portátil.
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	r.MemoriaMB = duasCasas(float64(m.Alloc) / (1 << 20))
	r.ReservadaMB = duasCasas(float64(m.Sys) / (1 << 20))
	r.Ambiente = ambienteFiltrado(prefixos)
	if d, err := espacoEmDisco(absoluta); err == nil {
		r.Disco = &d
	} else {
		r.AvisoDisco = err.Error()
	}
	return r, nil
}

func ouDesconhecido(valor string, err error) string {
	if err != nil {
		return "desconhecido"
	}
	return valor
}

// ambienteFiltrado devolve "NOME=valor" das variáveis cujo nome começa com
// um dos prefixos, em ordem alfabética e com valores longos encurtados.
func ambienteFiltrado(prefixos []string) []string {
	var lista []string
	for _, par := range os.Environ() {
		chave, valor, _ := strings.Cut(par, "=")
		for _, p := range prefixos {
			p = strings.ToUpper(strings.TrimSpace(p))
			if p == "" || !strings.HasPrefix(strings.ToUpper(chave), p) {
				continue
			}
			if len(valor) > 70 {
				valor = valor[:67] + "..."
			}
			lista = append(lista, chave+"="+valor)
			break
		}
	}
	sort.Strings(lista)
	return lista
}

// espacoEmDisco pergunta ao sistema: a biblioteca padrão não tem uma chamada
// portátil para isso, então usamos os/exec com a ferramenta nativa de cada
// sistema, com prazo, e lemos os números que ela imprime.
func espacoEmDisco(pasta string) (Disco, error) {
	ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		unidade := strings.TrimSuffix(filepath.VolumeName(pasta), ":") // "C:" vira "C"
		script := fmt.Sprintf("$d = Get-PSDrive -Name %s; Write-Output ($d.Used + $d.Free); Write-Output $d.Free", unidade)
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", script)
	} else {
		cmd = exec.CommandContext(ctx, "df", "-kP", pasta)
	}
	saida, err := cmd.Output()
	if err != nil {
		return Disco{}, fmt.Errorf("não consegui medir o disco: %v", err)
	}
	texto := strings.TrimSpace(string(saida))
	var total, livre float64
	if runtime.GOOS == "windows" {
		if c := strings.Fields(texto); len(c) >= 2 {
			total, livre = lerNumero(c[0]), lerNumero(c[1])
		}
	} else if linhas := strings.Split(texto, "\n"); len(linhas) >= 2 {
		// A segunda linha do df é: sistema, blocos de 1K, usados, livres, uso, ponto.
		if c := strings.Fields(linhas[1]); len(c) >= 4 {
			total, livre = lerNumero(c[1])*1024, lerNumero(c[3])*1024
		}
	}
	if total <= 0 {
		return Disco{}, fmt.Errorf("não entendi a resposta do sistema: %q", texto)
	}
	return Disco{TotalGB: duasCasas(total / (1 << 30)), LivreGB: duasCasas(livre / (1 << 30)), UsoPct: duasCasas(100 * (total - livre) / total)}, nil
}

// duasCasas arredonda para duas casas: deixa o JSON legível sem perder o que importa.
func duasCasas(v float64) float64 { return math.Round(v*100) / 100 }

// lerNumero converte texto em float64 e devolve 0 quando o texto não é um número.
func lerNumero(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}

// imprimir mostra o relatório em texto, com os rótulos alinhados por %-22s.
func imprimir(r Relatorio) {
	linha := func(rotulo string, valor any) { fmt.Printf("  %-22s %v\n", rotulo+":", valor) }
	fmt.Println("Sistema")
	linha("Sistema operacional", r.Sistema+"/"+r.Arquitetura)
	linha("Versão do Go", r.VersaoGo)
	linha("Máquina", r.Maquina)
	linha("Usuário", r.Usuario)
	linha("Pasta do usuário", r.PastaUsuario)
	linha("Hora local", r.Hora)
	fmt.Println("Processo")
	linha("Pasta inspecionada", r.Pasta)
	linha("Executável", r.Executavel)
	linha("PID", r.PID)
	linha("CPUs lógicas", r.CPUs)
	linha("Memória em uso", fmt.Sprintf("%.2f MB (reservada do sistema: %.2f MB)", r.MemoriaMB, r.ReservadaMB))
	fmt.Println("Disco da pasta")
	if r.Disco != nil {
		linha("Total", fmt.Sprintf("%.1f GB", r.Disco.TotalGB))
		linha("Livre", fmt.Sprintf("%.1f GB (%.0f%% em uso)", r.Disco.LivreGB, r.Disco.UsoPct))
	} else {
		linha("Aviso", r.AvisoDisco)
	}
	fmt.Printf("Variáveis de ambiente filtradas (%d)\n", len(r.Ambiente))
	for _, v := range r.Ambiente {
		fmt.Println("  " + v)
	}
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. Acrescente --gravar relatorio.json e salve o JSON com os.WriteFile,
//    em vez de imprimir na tela.
// 2. Mostre o tamanho total da pasta inspecionada somando os arquivos com
//    filepath.WalkDir (cuidado com pastas enormes: conte também os arquivos).
// 3. Compare dois relatórios salvos em JSON e liste só o que mudou entre
//    eles (versão do Go, espaço livre, variáveis novas ou removidas).
// ------------------------------------------------------------
