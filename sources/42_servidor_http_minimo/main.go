// ============================================================
// CodeNomad Academy  |  Sources Go  |  42. Servidor HTTP mínimo
// ------------------------------------------------------------
// O que faz: sobe um servidor web que serve os arquivos de uma pasta e
//            responde JSON em /api/hora (hora do servidor) e /api/eco (devolve
//            método, caminho, parâmetros e corpo do pedido). Cada pedido vira
//            uma linha de log com método, caminho, status e duração.
// Como usar: go run ./sources/42_servidor_http_minimo --porta 8080 --pasta ./site
//            go run ./sources/42_servidor_http_minimo --porta 8080 --duracao 60
//            go run ./sources/42_servidor_http_minimo --demo
// O que você aprende: http.ServeMux com os padrões "GET /rota" do Go 1.22,
//            http.FileServer, middleware que embrulha um Handler, encoding/json
//            na resposta, encerramento limpo com signal.NotifyContext e
//            srv.Shutdown, e net.Listen na porta 0 para ganhar uma porta livre.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	porta := flag.Int("porta", 8080, "porta TCP para escutar")
	host := flag.String("host", "127.0.0.1", "endereço para escutar (0.0.0.0 abre para a rede local)")
	pasta := flag.String("pasta", ".", "pasta cujos arquivos serão servidos")
	duracao := flag.Int("duracao", 0, "segundos até encerrar sozinho (0 = até Ctrl+C)")
	demo := flag.Bool("demo", false, "sobe numa porta livre, faz pedidos a si mesmo e encerra")
	flag.Parse()
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stdout) // o log do servidor sai junto com o resto, na ordem certa

	if *demo {
		if err := rodarDemo(); err != nil {
			falhar("Erro na demonstração: " + err.Error())
		}
		return
	}
	if flag.NFlag() == 0 {
		fmt.Fprintln(os.Stderr, "Uso: servidor_http_minimo --porta 8080 [--host 127.0.0.1] [--pasta ./site] [--duracao N]   ou   --demo")
		flag.PrintDefaults()
		os.Exit(2)
	}
	if *porta < 1 || *porta > 65535 {
		falhar("Erro: --porta precisa estar entre 1 e 65535.")
	}
	if info, err := os.Stat(*pasta); err != nil || !info.IsDir() {
		falhar(fmt.Sprintf("Erro: a pasta %q não existe ou não é uma pasta.", *pasta))
	}
	ouvinte, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*porta)))
	if err != nil {
		falhar("Erro: não consegui abrir a porta (ela já está em uso?): " + err.Error())
	}
	// NotifyContext cancela o contexto quando chega Ctrl+C; o WithTimeout
	// opcional cancela sozinho depois de --duracao segundos.
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer parar()
	if *duracao > 0 {
		var cancelar context.CancelFunc
		ctx, cancelar = context.WithTimeout(ctx, time.Duration(*duracao)*time.Second)
		defer cancelar()
	}
	absoluta, _ := filepath.Abs(*pasta)
	fmt.Printf("Servindo %s em http://%s (Ctrl+C encerra)\n", absoluta, ouvinte.Addr())
	fmt.Println("Rotas: /api/hora, /api/eco?nome=Ana e os arquivos da pasta")
	if err := servir(ctx, ouvinte, montar(*pasta)); err != nil {
		falhar("Erro: o servidor parou: " + err.Error())
	}
	fmt.Println("Servidor encerrado.")
}

// falhar imprime a mensagem na saída de erro e encerra com código 1.
func falhar(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// montar cria o roteador. Nos padrões do Go 1.22 o método faz parte da rota:
// "GET /api/hora" só atende GET. Quando vários padrões casam, vence o mais
// específico: um POST em /api/hora não casa com a rota GET, casa com "/api/"
// (mais específico que "/") e recebe o 404 em JSON, não um arquivo.
func montar(pasta string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/hora", hora)
	mux.HandleFunc("/api/eco", eco) // qualquer método: o eco mostra qual chegou
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		responderJSON(w, http.StatusNotFound, map[string]string{"erro": "rota desconhecida: " + r.Method + " " + r.URL.Path})
	})
	mux.Handle("/", http.FileServer(http.Dir(pasta)))
	return comLog(mux)
}

// registrador embrulha o ResponseWriter só para anotar o status escrito pelo handler.
type registrador struct {
	http.ResponseWriter
	status int
}

func (r *registrador) WriteHeader(codigo int) {
	r.status = codigo
	r.ResponseWriter.WriteHeader(codigo)
}

// comLog é um middleware: recebe um Handler e devolve outro que, além de
// chamar o original, registra método, caminho, status e duração.
func comLog(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		reg := &registrador{ResponseWriter: w, status: http.StatusOK}
		proximo.ServeHTTP(reg, r)
		ms := float64(time.Since(inicio)) / float64(time.Millisecond)
		log.Printf("%s %s -> %d em %.1fms", r.Method, r.URL.RequestURI(), reg.status, ms)
	})
}

// responderJSON serializa qualquer valor e manda com o cabeçalho certo.
func responderJSON(w http.ResponseWriter, status int, valor any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(valor); err != nil {
		log.Println("falha ao escrever a resposta:", err)
	}
}

func hora(w http.ResponseWriter, _ *http.Request) {
	agora := time.Now()
	fuso, _ := agora.Zone()
	responderJSON(w, http.StatusOK, map[string]any{
		"hora": agora.Format("2006-01-02 15:04:05"),
		"fuso": fuso,
		"unix": agora.Unix(),
	})
}

func eco(w http.ResponseWriter, r *http.Request) {
	// LimitReader protege o servidor: nunca confie no tamanho que o cliente manda.
	corpo, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		responderJSON(w, http.StatusBadRequest, map[string]string{"erro": "não consegui ler o corpo"})
		return
	}
	responderJSON(w, http.StatusOK, map[string]any{
		"metodo":     r.Method,
		"caminho":    r.URL.Path,
		"parametros": r.URL.Query(),
		"corpo":      string(corpo),
		"origem":     r.RemoteAddr,
	})
}

// servir atende no ouvinte até o contexto ser cancelado e então encerra com
// calma: Shutdown para de aceitar conexões e espera os pedidos em andamento.
func servir(ctx context.Context, ouvinte net.Listener, h http.Handler) error {
	srv := &http.Server{Handler: h, ReadHeaderTimeout: 5 * time.Second}
	falhou := make(chan error, 1)
	go func() { falhou <- srv.Serve(ouvinte) }()
	select {
	case err := <-falhou:
		return err // Serve só volta antes do Shutdown se algo deu errado
	case <-ctx.Done():
	}
	prazo, cancelar := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelar()
	return srv.Shutdown(prazo)
}

// pedir faz um pedido ao próprio servidor e imprime a resposta resumida.
func pedir(base, metodo, caminho, corpo string) error {
	fmt.Printf("\n> %s %s\n", metodo, caminho)
	req, err := http.NewRequest(metodo, base+caminho, strings.NewReader(corpo))
	if err != nil {
		return err
	}
	cliente := &http.Client{Timeout: 5 * time.Second}
	resp, err := cliente.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dados, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("< %s | %s\n%s\n", resp.Status, resp.Header.Get("Content-Type"), strings.TrimSpace(string(dados)))
	return nil
}

func rodarDemo() error {
	fmt.Println("=== CodeNomad Academy | Servidor HTTP mínimo: demonstração ===")
	pasta, err := os.MkdirTemp("", "codenomad-servidor-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(pasta) // a pasta some ao terminar, mesmo se algo falhar
	conteudo := "Olá! Este arquivo veio da pasta servida pelo http.FileServer.\n"
	if err := os.WriteFile(filepath.Join(pasta, "ola.txt"), []byte(conteudo), 0o644); err != nil {
		return err
	}
	ouvinte, err := net.Listen("tcp", "127.0.0.1:0") // porta 0: o sistema escolhe uma livre
	if err != nil {
		return err
	}
	base := "http://" + ouvinte.Addr().String()
	fmt.Println("Servidor no ar em", base)
	fmt.Println("Pasta servida:", pasta)

	ctx, parar := context.WithCancel(context.Background())
	defer parar()
	terminou := make(chan error, 1)
	go func() { terminou <- servir(ctx, ouvinte, montar(pasta)) }()

	pedidos := []struct{ metodo, caminho, corpo string }{
		{"GET", "/api/hora", ""},
		{"GET", "/api/eco?nome=Ana&curso=go", ""},
		{"POST", "/api/eco", `{"mensagem": "oi, servidor"}`},
		{"GET", "/ola.txt", ""},
		{"GET", "/nao-existe.txt", ""}, // 404 vindo do FileServer
		{"POST", "/api/hora", ""},      // não casa com "GET /api/hora": cai em "/api/" e recebe 404 em JSON
	}
	for _, p := range pedidos {
		if err := pedir(base, p.metodo, p.caminho, p.corpo); err != nil {
			return err
		}
	}
	parar() // cancela o contexto: servir chama Shutdown e devolve
	if err := <-terminou; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	fmt.Println("\nServidor encerrado e pasta temporária removida.")
	return nil
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. Crie a rota GET /api/arquivos que lista a pasta servida em JSON (nome,
//    tamanho e data) usando os.ReadDir.
// 2. Conte os pedidos por rota num map protegido por sync.Mutex e exponha
//    o total em GET /api/estatisticas.
// 3. Registre "GET /api/arquivo/{nome}" e use r.PathValue("nome"), novidade
//    do Go 1.22, para devolver o tamanho e a data daquele arquivo da pasta.
// ------------------------------------------------------------
