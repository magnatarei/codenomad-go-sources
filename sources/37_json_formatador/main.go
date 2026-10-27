// ============================================================
// CodeNomad Academy  |  Sources Go  |  37. Validador e formatador de JSON
// ------------------------------------------------------------
// O que faz: valida um arquivo JSON apontando linha e coluna do erro,
//            formata com recuo, compacta numa linha só, extrai um valor por
//            caminho (ex.: cliente.pedidos[0].total) e lista todas as chaves
//            com o tipo de cada valor.
// Como usar: go run ./sources/37_json_formatador -arquivo dados.json -formatar
//            go run ./sources/37_json_formatador -arquivo dados.json -caminho cliente.pedidos[0].total
//            go run ./sources/37_json_formatador -arquivo dados.json -chaves
//            go run ./sources/37_json_formatador --demo
// O que você aprende: json.Indent e json.Compact com bytes.Buffer, o tipo
//            *json.SyntaxError e errors.As, Unmarshal em "any" com type switch,
//            recursão para percorrer mapas e slices, e sort para saída estável.
// Requisitos: Go 1.22 ou mais novo, só biblioteca padrão
// Licença MIT. Um projeto da CodeNomad Academy (codenomad.org)
// ============================================================

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// validar devolve nil se o JSON é válido. Em caso de erro, converte o
// deslocamento em bytes (Offset) para linha e coluna, que é o que ajuda a achar.
func validar(dados []byte) error {
	var sintaxe *json.SyntaxError
	err := json.Unmarshal(dados, new(any))
	if errors.As(err, &sintaxe) { // errors.As "desembrulha" até achar o tipo pedido
		linha, coluna := linhaColuna(dados, sintaxe.Offset)
		return fmt.Errorf("JSON inválido na linha %d, coluna %d: %s", linha, coluna, sintaxe.Error())
	}
	return err
}

// linhaColuna conta as quebras de linha antes do deslocamento informado.
func linhaColuna(dados []byte, offset int64) (int, int) {
	antes := dados[:min(int(offset), len(dados))]
	linha := bytes.Count(antes, []byte("\n")) + 1
	coluna := len(antes) - bytes.LastIndex(antes, []byte("\n"))
	return linha, coluna
}

// formatar devolve o JSON com recuo de dois espaços; compactar tira tudo
// que é só espaço. As duas funções da biblioteca escrevem num bytes.Buffer.
func formatar(dados []byte) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, dados, "", "  "); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func compactar(dados []byte) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, dados); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// extrair caminha pelo valor seguindo um caminho como "cliente.pedidos[0].total".
// Cada pedaço é uma chave de objeto ou um índice entre colchetes.
func extrair(valor any, caminho string) (any, error) {
	for _, parte := range strings.Split(caminho, ".") {
		chave, indices, _ := strings.Cut(parte, "[")
		if chave != "" {
			obj, ok := valor.(map[string]any) // asserção de tipo: é um objeto?
			if !ok {
				return nil, fmt.Errorf("%q não é um objeto, não dá para pegar a chave %q", parte, chave)
			}
			if valor, ok = obj[chave]; !ok {
				return nil, fmt.Errorf("a chave %q não existe", chave)
			}
		}
		// "pedidos[0][1]" vira os índices "0" e "1" depois de tirar os colchetes.
		for _, idx := range strings.Split(strings.TrimSuffix(indices, "]"), "]") {
			if idx == "" {
				continue
			}
			i, err := strconv.Atoi(strings.TrimPrefix(idx, "["))
			lista, ok := valor.([]any)
			if err != nil || !ok {
				return nil, fmt.Errorf("%q não é uma lista indexável", parte)
			}
			if i < 0 || i >= len(lista) {
				return nil, fmt.Errorf("índice %d fora da lista (ela tem %d itens)", i, len(lista))
			}
			valor = lista[i]
		}
	}
	return valor, nil
}

// tipoDe dá um nome legível ao tipo dinâmico de um valor vindo do Unmarshal.
func tipoDe(v any) string {
	switch v.(type) { // type switch: o case escolhe pelo tipo, não pelo valor
	case map[string]any:
		return "objeto"
	case []any:
		return "lista"
	case string:
		return "texto"
	case float64:
		return "número" // todo número JSON vira float64 quando o destino é any
	case bool:
		return "booleano"
	default:
		return "nulo"
	}
}

// chaves percorre o valor recursivamente e devolve "caminho: tipo" de cada folha.
func chaves(v any, prefixo string, saida *[]string) {
	switch t := v.(type) {
	case map[string]any:
		nomes := make([]string, 0, len(t))
		for k := range t {
			nomes = append(nomes, k)
		}
		sort.Strings(nomes)
		for _, k := range nomes {
			caminho := k
			if prefixo != "" {
				caminho = prefixo + "." + k
			}
			*saida = append(*saida, caminho+": "+tipoDe(t[k]))
			chaves(t[k], caminho, saida)
		}
	case []any:
		for i, item := range t {
			chaves(item, fmt.Sprintf("%s[%d]", prefixo, i), saida)
		}
	}
}

func executar(dados []byte, fmtar, compact, listar bool, caminho string) error {
	if err := validar(dados); err != nil {
		return err
	}
	fmt.Println("JSON válido.")
	if fmtar {
		s, err := formatar(dados)
		if err != nil {
			return err
		}
		fmt.Println(s)
	}
	if compact {
		s, err := compactar(dados)
		if err != nil {
			return err
		}
		fmt.Println(s)
	}
	if listar || caminho != "" {
		var valor any
		if err := json.Unmarshal(dados, &valor); err != nil {
			return err
		}
		if listar {
			var linhas []string
			chaves(valor, "", &linhas)
			fmt.Printf("%d chave(s):\n", len(linhas))
			for _, l := range linhas {
				fmt.Println("  " + l)
			}
		}
		if caminho != "" {
			v, err := extrair(valor, caminho)
			if err != nil {
				return fmt.Errorf("caminho %q: %w", caminho, err)
			}
			b, _ := json.Marshal(v) // qualquer valor volta no formato JSON, inclusive objetos e listas
			fmt.Printf("%s = %s (%s)\n", caminho, b, tipoDe(v))
		}
	}
	return nil
}

const exemplo = `{"cliente": {"nome": "Loja do Bairro", "ativo": true, "cidade": "Curitiba",
  "pedidos": [{"id": 101, "total": 249.90, "itens": ["caderno", "caneta"]},
              {"id": 102, "total": 89.50, "itens": []}]}, "observacao": null}`

func demo() error {
	dados := []byte(exemplo)
	fmt.Println("== JSON de exemplo (como veio, mal alinhado):")
	fmt.Println(exemplo)
	fmt.Println("\n== Formatado:")
	if err := executar(dados, true, false, false, ""); err != nil {
		return err
	}
	fmt.Println("\n== Compactado:")
	if err := executar(dados, false, true, false, ""); err != nil {
		return err
	}
	fmt.Println("\n== Todas as chaves:")
	if err := executar(dados, false, false, true, ""); err != nil {
		return err
	}
	fmt.Println("\n== Consultas por caminho:")
	for _, c := range []string{"cliente.nome", "cliente.pedidos[0].total", "cliente.pedidos[1].itens", "cliente.pedidos[5].id", "cliente.cep"} {
		if err := executar(dados, false, false, false, c); err != nil {
			fmt.Println("  Erro esperado:", err)
		}
	}
	fmt.Println("\n== JSON inválido (vírgula sobrando antes do fechamento):")
	ruim := "{\n  \"nome\": \"Ana\",\n  \"idade\": 30,\n}"
	fmt.Println(ruim)
	if err := validar([]byte(ruim)); err != nil {
		fmt.Println("  Erro esperado:", err)
	}
	return nil
}

func main() {
	arquivo := flag.String("arquivo", "", "arquivo JSON de entrada")
	fmtar := flag.Bool("formatar", false, "imprime o JSON com recuo")
	compact := flag.Bool("compactar", false, "imprime o JSON numa linha só")
	listar := flag.Bool("chaves", false, "lista todas as chaves com o tipo do valor")
	caminho := flag.String("caminho", "", "extrai um valor, ex.: cliente.pedidos[0].total")
	modoDemo := flag.Bool("demo", false, "roda uma demonstração com um JSON embutido")
	flag.Parse()

	var err error
	switch {
	case *modoDemo:
		err = demo()
	case *arquivo == "":
		fmt.Fprintln(os.Stderr, "Uso: jsonfmt -arquivo dados.json [-formatar] [-compactar] [-chaves] [-caminho a.b[0].c]")
		flag.PrintDefaults()
		os.Exit(2)
	default:
		var dados []byte
		if dados, err = os.ReadFile(*arquivo); err == nil {
			err = executar(dados, *fmtar, *compact, *listar, *caminho)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro:", err)
		os.Exit(1)
	}
}

// Para ir além:
// 1. Acrescente -gravar para salvar o resultado formatado por cima do arquivo original.
// 2. Mostre, ao lado do erro de sintaxe, a linha problemática com um "^" sob a coluna.
// 3. Aceite um segundo arquivo e liste as chaves que existem em um e não no outro.
