// ============================================================
// CodeNomad Academy  |  Sources Go  |  45. Monitor de alterações numa pasta
// ------------------------------------------------------------
// O que faz: observa uma pasta (e as subpastas) consultando o disco em
//            intervalos regulares e registra, com hora, cada arquivo criado,
//            alterado ou removido. Ao encerrar, mostra um resumo por tipo.
// Como usar: go run ./sources/45_monitor_arquivos --pasta ./projeto --intervalo 1s
//            go run ./sources/45_monitor_arquivos --pasta ./projeto --duracao 30s
//            go run ./sources/45_monitor_arquivos --demo
// O que você aprende: filepath.WalkDir com fs.DirEntry, comparação de dois
//            mapas para achar diferenças, time.Ticker com select e contexto,
//            Ctrl+C via signal.NotifyContext, goroutine com canal de término
//            e os.MkdirTemp para testar sem sujar o disco.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"time"
)

// Estado é o que lembramos de cada arquivo entre duas consultas: se o tamanho
// ou a data de modificação mudar, o arquivo foi alterado.
type Estado struct {
	Tamanho    int64
	Modificado time.Time
}

// Evento é uma mudança detectada. Tipo é "criado", "alterado" ou "removido".
type Evento struct {
	Tipo    string
	Caminho string
	Hora    time.Time
}

// Monitor guarda a pasta vigiada, a fotografia mais recente e a contagem por tipo.
type Monitor struct {
	Pasta string
	atual map[string]Estado
	total map[string]int
}

func main() {
	pasta := flag.String("pasta", "", "pasta a observar")
	intervalo := flag.Duration("intervalo", time.Second, "tempo entre consultas (ex.: 500ms, 2s)")
	duracao := flag.Duration("duracao", 0, "tempo até encerrar sozinho (0 = até Ctrl+C)")
	demo := flag.Bool("demo", false, "cria uma pasta temporária, provoca três eventos e mostra a detecção")
	flag.Parse()

	if *demo {
		if err := rodarDemo(); err != nil {
			falhar("Erro na demonstração: " + err.Error())
		}
		return
	}
	if *pasta == "" {
		fmt.Fprintln(os.Stderr, "Uso: monitor_arquivos --pasta CAMINHO [--intervalo 1s] [--duracao 30s]   ou   --demo")
		flag.PrintDefaults()
		os.Exit(2)
	}
	if *intervalo < 100*time.Millisecond {
		falhar("Erro: --intervalo precisa ter pelo menos 100ms; consultar mais depressa só gasta CPU.")
	}
	if *duracao < 0 {
		falhar("Erro: --duracao não pode ser negativa (use 0 para rodar até Ctrl+C).")
	}
	m, err := novoMonitor(*pasta)
	if err != nil {
		falhar("Erro: " + err.Error())
	}
	// NotifyContext cancela o contexto no Ctrl+C; o WithTimeout, se houver,
	// cancela sozinho depois de --duracao. O laço de vigia observa os dois.
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer parar()
	if *duracao > 0 {
		var cancelar context.CancelFunc
		ctx, cancelar = context.WithTimeout(ctx, *duracao)
		defer cancelar()
	}
	fmt.Printf("Observando %s a cada %v (%d arquivo(s) no início). Ctrl+C encerra.\n", m.Pasta, *intervalo, len(m.atual))
	m.vigiar(ctx, *intervalo, imprimirEvento)
	m.resumo()
}

// falhar imprime a mensagem na saída de erro e encerra com código 1.
func falhar(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// novoMonitor valida a pasta e tira a primeira fotografia dela.
func novoMonitor(pasta string) (*Monitor, error) {
	absoluta, err := filepath.Abs(pasta)
	if err != nil {
		return nil, err
	}
	if info, err := os.Stat(absoluta); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("a pasta %q não existe ou não é uma pasta", pasta)
	}
	inicial, err := fotografar(absoluta)
	if err != nil {
		return nil, err
	}
	return &Monitor{Pasta: absoluta, atual: inicial, total: map[string]int{}}, nil
}

// fotografar percorre a pasta inteira e devolve um mapa caminho -> Estado.
// Os caminhos ficam relativos à pasta, para a saída não repetir o prefixo.
func fotografar(pasta string) (map[string]Estado, error) {
	foto := map[string]Estado{}
	err := filepath.WalkDir(pasta, func(caminho string, d fs.DirEntry, err error) error {
		if err != nil {
			if caminho == pasta {
				return err // a raiz sumiu ou não pode ser lida: isso é erro de verdade
			}
			return nil // um item que sumiu no meio do percurso aparece na próxima volta
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info() // pode falhar se o arquivo foi removido neste instante
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(pasta, caminho)
		foto[rel] = Estado{Tamanho: info.Size(), Modificado: info.ModTime()}
		return nil
	})
	return foto, err
}

// comparar cruza duas fotografias: o que só existe na nova foi criado, o que
// só existe na antiga foi removido, e o que existe nas duas com dados
// diferentes foi alterado. O resultado sai em ordem de caminho.
func comparar(antes, depois map[string]Estado, hora time.Time) []Evento {
	var eventos []Evento
	for caminho, novo := range depois {
		velho, existia := antes[caminho]
		switch {
		case !existia:
			eventos = append(eventos, Evento{"criado", caminho, hora})
		case velho.Tamanho != novo.Tamanho || !velho.Modificado.Equal(novo.Modificado):
			eventos = append(eventos, Evento{"alterado", caminho, hora})
		}
	}
	for caminho := range antes {
		if _, existe := depois[caminho]; !existe {
			eventos = append(eventos, Evento{"removido", caminho, hora})
		}
	}
	sort.Slice(eventos, func(i, j int) bool { return eventos[i].Caminho < eventos[j].Caminho })
	return eventos
}

// vigiar consulta a pasta a cada intervalo até o contexto ser cancelado e
// chama aoDetectar para cada evento. Quem chama decide o que fazer com ele.
func (m *Monitor) vigiar(ctx context.Context, intervalo time.Duration, aoDetectar func(Evento)) {
	ticker := time.NewTicker(intervalo)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		nova, err := fotografar(m.Pasta)
		if err != nil {
			fmt.Println("Aviso: não consegui ler a pasta nesta volta:", err)
			continue
		}
		for _, ev := range comparar(m.atual, nova, time.Now()) {
			m.total[ev.Tipo]++
			aoDetectar(ev)
		}
		m.atual = nova
	}
}

func imprimirEvento(ev Evento) {
	fmt.Printf("%s  %-9s %s\n", ev.Hora.Format("15:04:05.000"), ev.Tipo, ev.Caminho)
}

func (m *Monitor) resumo() {
	fmt.Printf("\nResumo: %d criado(s), %d alterado(s), %d removido(s). Arquivos agora: %d.\n",
		m.total["criado"], m.total["alterado"], m.total["removido"], len(m.atual))
}

// rodarDemo cria uma pasta temporária, liga o monitor numa goroutine e
// provoca os três tipos de evento com pausas maiores que o intervalo.
func rodarDemo() error {
	fmt.Println("=== CodeNomad Academy | Monitor de alterações: demonstração ===")
	pasta, err := os.MkdirTemp("", "codenomad-monitor-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(pasta) // a pasta some ao terminar, mesmo se algo falhar
	m, err := novoMonitor(pasta)
	if err != nil {
		return err
	}
	fmt.Printf("Pasta temporária: %s\nConsultando a cada 150ms. Eventos detectados:\n", pasta)

	ctx, parar := context.WithCancel(context.Background())
	terminou := make(chan struct{})
	go func() {
		m.vigiar(ctx, 150*time.Millisecond, imprimirEvento)
		close(terminou) // avisa que a goroutine acabou de verdade
	}()

	notas := filepath.Join(pasta, "notas.txt")
	passos := []struct {
		descricao string
		acao      func() error
	}{
		{"criar notas.txt", func() error { return os.WriteFile(notas, []byte("primeira linha\n"), 0o644) }},
		{"criar sub/lista.csv", func() error {
			if err := os.Mkdir(filepath.Join(pasta, "sub"), 0o755); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(pasta, "sub", "lista.csv"), []byte("nome,nota\nAna,9\n"), 0o644)
		}},
		{"alterar notas.txt", func() error { return os.WriteFile(notas, []byte("primeira linha\nsegunda linha\n"), 0o644) }},
		{"remover notas.txt", func() error { return os.Remove(notas) }},
	}
	for _, p := range passos {
		time.Sleep(400 * time.Millisecond) // mais que o intervalo: garante uma consulta entre os passos
		fmt.Println("  (ação:", p.descricao+")")
		if err := p.acao(); err != nil {
			parar()
			return err
		}
	}
	time.Sleep(400 * time.Millisecond)
	parar()
	<-terminou
	m.resumo()
	fmt.Println("Pasta temporária removida.")
	return nil
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. Aceite --ignorar ".git,node_modules" e pule essas pastas devolvendo
//    fs.SkipDir dentro do WalkDir.
// 2. Grave cada evento num arquivo de log (hora, tipo, caminho) usando
//    encoding/csv, além de imprimir na tela.
// 3. Guarde um crypto/sha256 de cada arquivo em vez da data: assim uma
//    gravação que não muda o conteúdo deixa de contar como alteração.
// ------------------------------------------------------------
