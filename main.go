package main

import (
	"asi/core"
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Opções de navegação do menu e configurações de formatação do console.
const (
	optProd     = "1"
	optInsProd  = "2"
	optRemProd  = "3"
	optInsAcess = "4"
	optRemAcess = "5"
	optExit     = "0"
	idBase      = 10
	idBits      = 64
	floatBits   = 64
	textPad     = " \x00"
)

func main() {
	entrada := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n=== Arquivo Sequencial-Indexado ===")
		fmt.Printf("%s - Consultar produto por ID\n", optProd)
		fmt.Printf("%s - Inserir produto\n", optInsProd)
		fmt.Printf("%s - Remover produto\n", optRemProd)
		fmt.Printf("%s - Inserir acesso\n", optInsAcess)
		fmt.Printf("%s - Remover acesso\n", optRemAcess)
		fmt.Printf("%s - Sair\n", optExit)
		fmt.Print("Opção: ")

		if !entrada.Scan() {
			if err := entrada.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler opção: %v\n", err)
			}
			return
		}

		switch strings.TrimSpace(entrada.Text()) {
		case optProd:
			consultarProduto(entrada)
		case optInsProd:
			inserir(entrada, true)
		case optRemProd:
			remover(entrada, true)
		case optInsAcess:
			inserir(entrada, false)
		case optRemAcess:
			remover(entrada, false)
		case optExit:
			return
		default:
			fmt.Println("Opção inválida.")
		}
	}
}

func consultarProduto(entrada *bufio.Scanner) {
	fmt.Print("ID do produto: ")
	if !entrada.Scan() {
		if err := entrada.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler o ID: %v\n", err)
		}
		return
	}

	productID, err := strconv.ParseInt(strings.TrimSpace(entrada.Text()), idBase, idBits)
	if err != nil {
		fmt.Printf("ID inválido. Informe um número inteiro de %d bits.\n", idBits)
		return
	}

	produto, encontrado, err := buscarProduto(productID)
	if err != nil {
		fmt.Printf("Falha na consulta: %v\n", err)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("Confira se os arquivos .dat existem e gere os índices com: go run ./cmd/indices")
		}
		return
	}
	if !encontrado {
		fmt.Printf("Produto com ID %d não encontrado.\n", productID)
		return
	}

	fmt.Printf("\nProduto encontrado\nID: %d\nMarca: %s\nCategoria: %s\nPreço: %.2f\n",
		produto.ProductID,
		strings.Trim(string(produto.Brand[:]), textPad),
		strings.Trim(string(produto.CategoryCode[:]), textPad),
		produto.Price,
	)
}

func buscarProduto(productID int64) (core.ProdutoRecord, bool, error) {
	arquivoProdutos, arquivoIndice, err := abrirArquivos(true, false)
	if err != nil {
		return core.ProdutoRecord{}, false, err
	}
	defer arquivoProdutos.Close()
	defer arquivoIndice.Close()

	return core.BuscarProdutoPorID(arquivoProdutos, arquivoIndice, productID)
}

func abrirArquivos(produto, escrita bool) (*os.File, *os.File, error) {
	path, idx := core.ProdPath, core.ProdIdxPath
	if !produto {
		path, idx = core.AcessPath, core.AcessIdxPath
	}
	modo := os.O_RDONLY
	if escrita {
		modo = os.O_RDWR
	}
	dados, err := os.OpenFile(path, modo, core.Perm)
	if err != nil {
		return nil, nil, fmt.Errorf("abrir %s: %w", path, err)
	}
	indice, err := os.OpenFile(idx, modo, core.Perm)
	if err != nil {
		dados.Close()
		return nil, nil, fmt.Errorf("abrir %s: %w", idx, err)
	}
	return dados, indice, nil
}

func inserir(entrada *bufio.Scanner, produto bool) {
	var p core.ProdutoRecord
	var a core.AcessoRecord
	var err error
	if produto {
		p, err = lerProduto(entrada)
	} else {
		a, err = lerAcesso(entrada)
	}
	if err != nil {
		fmt.Printf("Não foi possível inserir: %v\n", err)
		return
	}
	dados, indice, err := abrirArquivos(produto, true)
	if err != nil {
		fmt.Printf("Não foi possível inserir: %v\n", err)
		return
	}
	defer dados.Close()
	defer indice.Close()
	if produto {
		err = core.InserirProduto(dados, indice, p)
	} else {
		err = core.InserirAcesso(dados, indice, a)
	}
	if err != nil {
		fmt.Printf("Não foi possível inserir: %v\n", err)
		return
	}
	fmt.Println("Registro inserido.")
}

func remover(entrada *bufio.Scanner, produto bool) {
	rotulo := "ID do produto: "
	if !produto {
		rotulo = "ID do evento: "
	}
	id, err := lerID(entrada, rotulo)
	if err != nil {
		fmt.Printf("Não foi possível remover: %v\n", err)
		return
	}
	dados, indice, err := abrirArquivos(produto, true)
	if err != nil {
		fmt.Printf("Não foi possível remover: %v\n", err)
		return
	}
	defer dados.Close()
	defer indice.Close()
	var removido bool
	if produto {
		removido, err = core.RemoverProduto(dados, indice, id)
	} else {
		removido, err = core.RemoverAcesso(dados, indice, id)
	}
	if err != nil {
		fmt.Printf("Não foi possível remover: %v\n", err)
	} else if !removido {
		fmt.Println("Registro não encontrado ou já excluído.")
	} else {
		fmt.Println("Registro excluído logicamente.")
	}
}

func lerCampo(entrada *bufio.Scanner, rotulo string) (string, error) {
	fmt.Print(rotulo)
	if !entrada.Scan() {
		if err := entrada.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return strings.TrimSpace(entrada.Text()), nil
}

func lerID(entrada *bufio.Scanner, rotulo string) (int64, error) {
	texto, err := lerCampo(entrada, rotulo)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(texto, idBase, idBits)
}

func lerTexto(entrada *bufio.Scanner, rotulo string, destino []byte) error {
	texto, err := lerCampo(entrada, rotulo)
	if err != nil {
		return err
	}
	if len(texto) > len(destino) {
		return fmt.Errorf("%s máximo de %d bytes", rotulo, len(destino))
	}
	for i := range destino {
		destino[i] = ' '
	}
	copy(destino, texto)
	return nil
}

func lerProduto(entrada *bufio.Scanner) (core.ProdutoRecord, error) {
	var p core.ProdutoRecord
	var err error
	if p.ProductID, err = lerID(entrada, "ID do produto: "); err != nil {
		return p, err
	}
	if err = lerTexto(entrada, "Marca: ", p.Brand[:]); err != nil {
		return p, err
	}
	if err = lerTexto(entrada, "Categoria: ", p.CategoryCode[:]); err != nil {
		return p, err
	}
	preco, err := lerCampo(entrada, "Preço (use ponto decimal): ")
	if err != nil {
		return p, err
	}
	p.Price, err = strconv.ParseFloat(preco, floatBits)
	return p, err
}

func lerAcesso(entrada *bufio.Scanner) (core.AcessoRecord, error) {
	var a core.AcessoRecord
	var err error
	if a.EventID, err = lerID(entrada, "ID do evento: "); err != nil {
		return a, err
	}
	if a.UserID, err = lerID(entrada, "ID do usuário: "); err != nil {
		return a, err
	}
	if a.ProductID, err = lerID(entrada, "ID do produto: "); err != nil {
		return a, err
	}
	if err = lerTexto(entrada, "Tipo de evento: ", a.EventType[:]); err != nil {
		return a, err
	}
	err = lerTexto(entrada, "Data/hora (AAAA-MM-DD HH:MM:SS): ", a.EventTime[:])
	return a, err
}
