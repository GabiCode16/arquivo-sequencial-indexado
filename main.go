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

const (
	produtosPath = "dados/produtos.dat"
	indiceProdutosPath = "dados/indice_prod.idx"
)

func main() {
	entrada := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n=== Arquivo Sequencial-Indexado ===")
		fmt.Println("1 - Consultar produto por ID")
		fmt.Println("0 - Sair")
		fmt.Print("Opção: ")

		if !entrada.Scan() {
			if err := entrada.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao ler opção: %v\n", err)
			}
			return
		}

		switch strings.TrimSpace(entrada.Text()) {
		case "1":
			consultarProduto(entrada)
		case "0":
			return
		default:
			fmt.Println("Opção inválida. Escolha 1 ou 0.")
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

	productID, err := strconv.ParseInt(strings.TrimSpace(entrada.Text()), 10, 64)
	if err != nil {
		fmt.Println("ID inválido. Informe um número inteiro de 64 bits.")
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
		strings.Trim(string(produto.Brand[:]), " \x00"),
		strings.Trim(string(produto.CategoryCode[:]), " \x00"),
		produto.Price,
	)
}

func buscarProduto(productID int64) (core.ProdutoRecord, bool, error) {
	arquivoProdutos, err := core.AbrirArquivo(produtosPath)
	if err != nil {
		return core.ProdutoRecord{}, false, fmt.Errorf("abrir %s: %w", produtosPath, err)
	}
	defer arquivoProdutos.Close()

	arquivoIndice, err := core.AbrirArquivo(indiceProdutosPath)
	if err != nil {
		return core.ProdutoRecord{}, false, fmt.Errorf("abrir %s: %w", indiceProdutosPath, err)
	}
	defer arquivoIndice.Close()

	return core.BuscarProdutoPorID(arquivoProdutos, arquivoIndice, productID)
}
