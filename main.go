package main

import (
	"asi/core"
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	entrada := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n=== Arquivo Sequencial-Indexado ===")
		fmt.Printf("%s - Consultar produto por ID\n", core.OptProd)
		fmt.Printf("%s - Sair\n", core.OptExit)
		fmt.Print("Opção: ")

		if !entrada.Scan() {
			if err := entrada.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler opção: %v\n", err)
			}
			return
		}

		switch strings.TrimSpace(entrada.Text()) {
		case core.OptProd:
			consultarProduto(entrada)
		case core.OptExit:
			return
		default:
			fmt.Printf("Opção inválida. Escolha %s ou %s.\n", core.OptProd, core.OptExit)
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

	productID, err := strconv.ParseInt(strings.TrimSpace(entrada.Text()), core.IDBase, core.IDBits)
	if err != nil {
		fmt.Printf("ID inválido. Informe um número inteiro de %d bits.\n", core.IDBits)
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
		strings.Trim(string(produto.Brand[:]), core.TextPad),
		strings.Trim(string(produto.CategoryCode[:]), core.TextPad),
		produto.Price,
	)
}

func buscarProduto(productID int64) (core.ProdutoRecord, bool, error) {
	arquivoProdutos, err := core.AbrirArquivo(core.ProdPath)
	if err != nil {
		return core.ProdutoRecord{}, false, fmt.Errorf("abrir %s: %w", core.ProdPath, err)
	}
	defer arquivoProdutos.Close()

	arquivoIndice, err := core.AbrirArquivo(core.ProdIdxPath)
	if err != nil {
		return core.ProdutoRecord{}, false, fmt.Errorf("abrir %s: %w", core.ProdIdxPath, err)
	}
	defer arquivoIndice.Close()

	return core.BuscarProdutoPorID(arquivoProdutos, arquivoIndice, productID)
}
