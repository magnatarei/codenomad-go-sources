// ============================================================
// CodeNomad Academy  |  Sources Go  |  41. Cliente HTTP sem dependências
// ------------------------------------------------------------
// O que faz: faz um GET em qualquer URL e mostra status, cabeçalhos e corpo.
//            Se o corpo for JSON, interpreta e extrai os campos que você pedir
//            (loja.nome, produtos.0.preco). Timeout e erro de rede viram mensagens claras.
// Como usar: go run ./sources/41_cliente_http --url https://exemplo.com/api/loja --campo loja.nome,total
//            go run ./sources/41_cliente_http --demo
// O que você aprende: net/http com pedido e contexto com prazo, io.ReadAll,
//            encoding/json em map[string]any, errors.As e errors.Is para reconhecer
//            o tipo do erro, e http.ReadResponse para interpretar uma resposta crua.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// resposta guarda só o que interessa depois que o corpo já foi lido: ele precisa
// ser lido antes de o contexto ser cancelado, por isso devolvemos bytes prontos.
type resposta struct {
	status     string
	cabecalhos http.Header
	corpo      []byte
}

func main() {
	endereco := flag.String("url", "", "endereço para consultar (http:// ou https://)")
	campos := flag.String("campo", "", "campos do JSON a extrair, separados por vírgula (ex.: loja.nome,produtos.0.preco)")
	timeout := flag.Int("timeout", 10, "segundos de espera antes de desistir")
	limite := flag.Int("max", 2000, "quantos símbolos do corpo mostrar")
	demo := flag.Bool("demo", false, "demonstração completa sem acesso à rede")
	flag.Parse()

	if *demo {
		rodarDemo(*limite)
		return
	}
	if *endereco == "" {
		fmt.Fprintln(os.Stderr, "Uso: cliente_http --url ENDEREÇO [--campo a.b,c] [--timeout N] [--max N]   ou   --demo")
		flag.PrintDefaults()
		os.Exit(2)
	}
	if *timeout <= 0 {
		falhar("Erro: --timeout precisa ser maior que zero.")
	}
	prazo := time.Duration(*timeout) * time.Second
	r, err := buscar(*endereco, prazo)
	if err != nil {
		falhar("Erro: " + explicarErro(err, prazo))
	}
	// FieldsFunc separa por vírgula ou espaço e já descarta pedaços vazios.
	lista := strings.FieldsFunc(*campos, func(c rune) bool { return c == ',' || c == ' ' })
	mostrar(r, lista, *limite)
}

// falhar imprime a mensagem na saída de erro e encerra com código 1.
func falhar(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// buscar valida a URL, monta o pedido com um prazo e lê a resposta inteira.
func buscar(endereco string, prazo time.Duration) (resposta, error) {
	u, err := url.Parse(endereco)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return resposta{}, fmt.Errorf("URL inválida %q: use algo como https://servidor/caminho", endereco)
	}
	// O prazo cobre conexão, espera e leitura do corpo. Como o cancelamento
	// roda ao sair desta função, o corpo é lido aqui dentro, em lerResposta.
	ctx, cancelar := context.WithTimeout(context.Background(), prazo)
	defer cancelar()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endereco, nil)
	if err != nil {
		return resposta{}, err
	}
	req.Header.Set("User-Agent", "CodeNomad-Cliente/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resposta{}, err
	}
	return lerResposta(resp)
}

// lerResposta consome o corpo e fecha a conexão; serve para a rede e para o texto embutido.
func lerResposta(resp *http.Response) (resposta, error) {
	defer resp.Body.Close()
	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return resposta{}, fmt.Errorf("falha ao ler o corpo: %w", err)
	}
	return resposta{status: resp.Status, cabecalhos: resp.Header, corpo: corpo}, nil
}

// explicarErro traduz os erros mais comuns de rede para uma frase útil:
// errors.As procura, na cadeia de erros embrulhados, um erro do tipo pedido.
func explicarErro(err error, prazo time.Duration) string {
	var erroDNS *net.DNSError
	var erroURL *url.Error
	switch {
	case errors.As(err, &erroDNS):
		return fmt.Sprintf("não consegui resolver o nome %q: confira a URL ou a sua conexão", erroDNS.Name)
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &erroURL) && erroURL.Timeout():
		return fmt.Sprintf("tempo esgotado depois de %v: o servidor não respondeu a tempo", prazo)
	case errors.As(err, &erroURL):
		return fmt.Sprintf("falha de rede ao acessar %s: %v (o servidor está no ar? a porta está certa?)", erroURL.URL, erroURL.Err)
	}
	return err.Error()
}

// mostrar imprime status, cabeçalhos e corpo; se o corpo for JSON, extrai campos.
func mostrar(r resposta, campos []string, limite int) {
	fmt.Println("Status:", r.status)
	fmt.Println("Cabeçalhos:")
	for _, nome := range chavesOrdenadas(r.cabecalhos) {
		fmt.Printf("  %s: %s\n", nome, strings.Join(r.cabecalhos[nome], ", "))
	}
	fmt.Printf("Corpo (%d bytes):\n", len(r.corpo))
	texto := strings.TrimRight(string(r.corpo), "\n")
	if runas := []rune(texto); len(runas) > limite { // runas, para não cortar uma letra acentuada ao meio
		texto = string(runas[:limite]) + fmt.Sprintf("\n... (%d símbolos omitidos; use --max para ver mais)", len(runas)-limite)
	}
	fmt.Println(texto)

	var dado any
	if err := json.Unmarshal(r.corpo, &dado); err != nil {
		fmt.Println("O corpo não é JSON válido, então não há campos para extrair.")
		return
	}
	fmt.Println("JSON interpretado:", descrever(dado))
	for _, caminho := range campos {
		valor, err := extrair(dado, caminho)
		if err != nil {
			fmt.Printf("  %-18s erro: %v\n", caminho, err)
			continue
		}
		b, _ := json.Marshal(valor) // o que saiu de Unmarshal sempre volta a ser JSON
		fmt.Printf("  %-18s = %s\n", caminho, b)
	}
}

// extrair caminha pelo JSON seguindo um caminho como "produtos.1.preco":
// cada parte é uma chave (em objeto) ou um índice numérico (em lista).
func extrair(dado any, caminho string) (any, error) {
	atual := dado
	for _, parte := range strings.Split(caminho, ".") {
		switch t := atual.(type) {
		case map[string]any:
			valor, ok := t[parte]
			if !ok {
				return nil, fmt.Errorf("a chave %q não existe (há: %s)", parte, strings.Join(chavesOrdenadas(t), ", "))
			}
			atual = valor
		case []any:
			i, err := strconv.Atoi(parte)
			if err != nil || i < 0 || i >= len(t) {
				return nil, fmt.Errorf("%q não é um índice válido para uma lista de %d itens", parte, len(t))
			}
			atual = t[i]
		default:
			return nil, fmt.Errorf("não dá para entrar em %q: o valor anterior é %s", parte, descrever(atual))
		}
	}
	return atual, nil
}

// chavesOrdenadas devolve as chaves de qualquer mapa com chave string, em ordem (mapas não têm ordem).
func chavesOrdenadas[V any](m map[string]V) []string {
	chaves := make([]string, 0, len(m))
	for k := range m {
		chaves = append(chaves, k)
	}
	sort.Strings(chaves)
	return chaves
}

// descrever diz, em palavras, que tipo de valor o JSON trouxe.
func descrever(v any) string {
	switch t := v.(type) {
	case map[string]any:
		return fmt.Sprintf("objeto com %d chaves (%s)", len(t), strings.Join(chavesOrdenadas(t), ", "))
	case []any:
		return fmt.Sprintf("lista com %d itens", len(t))
	case string:
		return "texto"
	case float64: // todo número do JSON vira float64 quando se decodifica em any
		return "número"
	case bool:
		return "booleano"
	}
	return "nulo"
}

// respostaExemplo monta uma resposta HTTP crua, byte a byte como chegaria
// pela rede, para o modo de demonstração.
func respostaExemplo() string {
	corpo := `{
  "loja": {"nome": "Papelaria Horizonte", "cidade": "Curitiba", "aberta": true},
  "produtos": [
    {"nome": "Caderno pautado", "preco": 18.9, "estoque": 42},
    {"nome": "Caneta azul", "preco": 2.5, "estoque": 310},
    {"nome": "Mochila escolar", "preco": 129.0, "estoque": 7}
  ],
  "total": 3
}`
	return "HTTP/1.1 200 OK\r\nContent-Type: application/json; charset=utf-8\r\n" +
		"Server: CodeNomad-Exemplo\r\nCache-Control: no-cache\r\n" +
		"Content-Length: " + strconv.Itoa(len(corpo)) + "\r\n\r\n" + corpo
}

func rodarDemo(limite int) {
	fmt.Println("=== CodeNomad Academy | Cliente HTTP: demonstração sem rede ===")
	fmt.Println("A resposta abaixo está embutida no programa e passa por http.ReadResponse, o mesmo caminho da rede.")
	fmt.Println()
	resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(respostaExemplo())), nil)
	if err != nil {
		falhar("Erro ao interpretar a resposta de exemplo: " + err.Error())
	}
	r, err := lerResposta(resp)
	if err != nil {
		falhar("Erro: " + err.Error())
	}
	mostrar(r, []string{"loja.nome", "loja", "produtos.1.nome", "produtos.2.preco", "total", "loja.telefone", "produtos.9.nome"}, limite)

	fmt.Println()
	fmt.Println("Como o programa explica erros de rede (erros montados, sem abrir conexão):")
	exemplos := []error{
		&url.Error{Op: "Get", URL: "https://exemplo.invalido/api", Err: context.DeadlineExceeded},
		&url.Error{Op: "Get", URL: "https://servidor-inexistente.exemplo/", Err: &net.DNSError{Err: "no such host", Name: "servidor-inexistente.exemplo", IsNotFound: true}},
		&url.Error{Op: "Get", URL: "http://127.0.0.1:9/", Err: errors.New("connect: connection refused")},
	}
	for _, e := range exemplos {
		fmt.Println("  *", explicarErro(e, 10*time.Second))
	}
}

// ------------------------------------------------------------
// Para ir além (três exercícios sugeridos):
// 1. Aceite --metodo POST e --dados '{"a":1}', enviando Content-Type: application/json.
// 2. Mostre quanto tempo o pedido levou (time.Since) e o tamanho do corpo em KB.
// 3. Permita --arquivo resposta.json para gravar o corpo com os.WriteFile.
// ------------------------------------------------------------
