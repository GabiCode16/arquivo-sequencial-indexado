package main

import (
	"asi/core"
	"fmt"
	"io"
)

func validateProductFile() (bool, string) {
	file, e := core.AbrirArquivo("dados/produtos.dat")

	if e != nil {
		return false, "Erro ao abrir o arquivo de produtos"
	}

	defer file.Close()

	tamanhoTotal := core.ObterTamanhoArquivo(file)

	if tamanhoTotal < 0 || tamanhoTotal%97 != 0 {
		return false, "Tamanho do arquivo de produtos inválido"
	}

	qtdProdutos := tamanhoTotal / 97
	lidos := 0

	product, e := core.LerLinhaProduto(file)

	if e != nil {
		return false, "Erro ao ler o primeiro registro de produtos"
	}

	id := product.ProductID
	lidos++

	for {
		product, e = core.LerLinhaProduto(file)
		newId := product.ProductID

		if e != nil {
			break
		}

		if newId <= id {
			return false, "IDs de produtos não estão em ordem crescente"
		} else {
			id = newId
		}

		if product.Delimiter != '\n' {
			return false, "Delimitador de produtos inválido"
		}

		lidos++

		if lidos%10000 == 0 {
			fmt.Printf("Validando produtos: %d/%d\n", lidos, qtdProdutos)
		}
	}

	if e == io.EOF {
		fmt.Printf("Validando produtos: %d/%d\n", lidos, qtdProdutos)
		return true, "Sucesso: Arquivo de produtos válido"
	}

	return false, "Erro ao ler o arquivo de produtos"
}

func validateAccessFile() (bool, string) {
	file, e := core.AbrirArquivo("dados/acessos.dat")

	if e != nil {
		return false, "Erro ao abrir o arquivo de acessos"
	}

	defer file.Close()

	tamanhoTotal := core.ObterTamanhoArquivo(file)

	if tamanhoTotal < 0 || tamanhoTotal%61 != 0 {
		return false, "Tamanho do arquivo de acessos inválido"
	}

	qtdAcessos := tamanhoTotal / 61
	lidos := 0

	access, e := core.LerLinhaAcesso(file)

	if e != nil {
		return false, "Erro ao ler o primeiro registro de acessos"
	}

	id := access.EventID
	lidos++

	for {
		access, e = core.LerLinhaAcesso(file)
		newId := access.EventID

		if e != nil {
			break
		}

		if newId <= id {
			return false, "IDs de acessos não estão em ordem crescente"
		} else {
			id = newId
		}

		if access.Delimiter != '\n' {
			return false, "Delimitador de acessos inválido"
		}

		lidos++

		if lidos%1000000 == 0 {
			fmt.Printf("Validando acessos: %d/%d\n", lidos, qtdAcessos)
		}
	}

	if e == io.EOF {
		fmt.Printf("Validando acessos: %d/%d\n", lidos, qtdAcessos)
		return true, "Sucesso: Arquivo de acessos válido"
	}

	return false, "Erro ao ler o arquivo de acessos"
}

func main() {
	fmt.Println("=== Validador de Integridade de Arquivos (.dat) ===")
	// Ponto de entrada do script validador em Go

	productValid, productMsg := validateProductFile()

	if productValid {
		fmt.Println("Arquivo de produtos: Integro")
	} else {
		fmt.Printf("Arquivo de produtos: Corrompido\n%s\n", productMsg)
	}

	accessValid, accessMsg := validateAccessFile()

	if accessValid {
		fmt.Println("Arquivo de acessos: Integro")
	} else {
		fmt.Printf("Arquivo de acessos: Corrompido\n%s\n", accessMsg)
	}
}
