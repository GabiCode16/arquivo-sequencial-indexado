package main

import (
	"asi/core"
	"bufio"
	"fmt"
)

func validateProductFile() (bool, string) {
	file, e := core.AbrirArquivo(core.ProdPath, core.ModR, false)

	if e != nil {
		return false, "Erro ao abrir o arquivo de produtos"
	}

	defer file.Close()

	tamanhoTotal := core.ObterTamanhoArquivo(file)

	if tamanhoTotal < 0 || tamanhoTotal%core.ProdSize != 0 {
		return false, "Tamanho do arquivo de produtos inválido"
	}

	qtdProdutos := tamanhoTotal / core.ProdSize
	lidos := 0

	product, e := core.LerLinhaProduto(file.Reader())

	if e != nil {
		return false, "Erro ao ler o primeiro registro de produtos"
	}

	if product.Delimiter != '\n' {
		return false, "Delimitador de produtos inválido"
	}

	id := product.ProductID
	lidos++

	for {
		product, e = core.LerLinhaProduto(file.Reader())
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

		if int64(lidos)%core.ProdStep == 0 {
			fmt.Printf("Validando produtos: %d/%d\n", lidos, qtdProdutos)
		}
	}

	if int64(lidos) != qtdProdutos {
		return false, "Quantidade de registros de produtos lidos não corresponde ao tamanho do arquivo"
	}

	if e == core.EOF {
		fmt.Printf("Validando produtos: %d/%d\n", lidos, qtdProdutos)
		return true, "Sucesso: Arquivo de produtos válido"
	}

	return false, "Erro ao ler o arquivo de produtos"
}

func validateAccessFile() (bool, string) {
	file, e := core.AbrirArquivo(core.AcessPath, core.ModR, false)

	if e != nil {
		return false, "Erro ao abrir o arquivo de acessos"
	}

	defer file.Close()

	reader := bufio.NewReaderSize(file.Reader(), int(core.BufSize)) // Buffer de 4 MB

	tamanhoTotal := core.ObterTamanhoArquivo(file)

	if tamanhoTotal < 0 || tamanhoTotal%core.AcessSize != 0 {
		return false, "Tamanho do arquivo de acessos inválido"
	}

	qtdAcessos := tamanhoTotal / core.AcessSize
	lidos := 0

	access, e := core.LerLinhaAcesso(reader)

	if e != nil {
		return false, "Erro ao ler o primeiro registro de acessos"
	}

	if access.Delimiter != '\n' {
		return false, "Delimitador de acessos inválido"
	}

	id := access.EventID
	lidos++

	for {
		access, e = core.LerLinhaAcesso(reader)
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

		if int64(lidos)%core.AcessStep == 0 {
			fmt.Printf("Validando acessos: %d/%d\n", lidos, qtdAcessos)
		}
	}

	if int64(lidos) != qtdAcessos {
		return false, "Quantidade de registros de acessos lidos não corresponde ao tamanho do arquivo"
	}

	if e == core.EOF {
		fmt.Printf("Validando acessos: %d/%d\n", lidos, qtdAcessos)
		return true, "Sucesso: Arquivo de acessos válido"
	}

	return false, "Erro ao ler o arquivo de acessos"
}

func validateProductIndexFile() (bool, string) {
	file, e := core.AbrirArquivo(core.ProdIdxPath, core.ModR, false)

	if e != nil {
		return false, "Erro ao abrir o arquivo de índices de produtos"
	}

	defer file.Close()

	tamanhoTotal := core.ObterTamanhoArquivo(file)

	if tamanhoTotal < 0 || tamanhoTotal%core.IdxSize != 0 {
		return false, "Tamanho do arquivo de índices de produtos inválido"
	}

	qtdIndices := tamanhoTotal / core.IdxSize
	lidos := 0

	productIdx, e := core.LerLinhaProdutoIdx(file.Reader())

	if e != nil {
		return false, "Erro ao ler o primeiro registro de índices de produtos"
	}

	id := productIdx.ProductID
	lidos++

	for {
		productIdx, e = core.LerLinhaProdutoIdx(file.Reader())
		newId := productIdx.ProductID

		if e != nil {
			break
		}

		if newId <= id {
			return false, "IDs de produtos não estão em ordem crescente"
		} else {
			id = newId
		}

		lidos++

		if int64(lidos)%core.ProdIdxStep == 0 {
			fmt.Printf("Validando índices de produtos: %d/%d\n", lidos, qtdIndices)
		}
	}

	if int64(lidos) != qtdIndices {
		return false, "Quantidade de registros de índices de produtos lidos " +
			"não corresponde ao tamanho do arquivo"
	}

	if e == core.EOF {
		fmt.Printf("Validando índices de produtos: %d/%d\n", lidos, qtdIndices)
		return true, "Sucesso: Arquivo de índices de produtos válido"
	}

	return false, "Erro ao ler o arquivo de índices de produtos"
}

func validateAccessIndexFile() (bool, string) {
	file, e := core.AbrirArquivo(core.AcessIdxPath, core.ModR, false)

	if e != nil {
		return false, "Erro ao abrir o arquivo de índices de acessos"
	}

	defer file.Close()

	tamanhoTotal := core.ObterTamanhoArquivo(file)

	if tamanhoTotal < 0 || tamanhoTotal%core.IdxSize != 0 {
		return false, "Tamanho do arquivo de índices de acessos inválido"
	}

	qtdIndices := tamanhoTotal / core.IdxSize
	lidos := 0

	accessIdx, e := core.LerLinhaAcessoIdx(file.Reader())

	if e != nil {
		return false, "Erro ao ler o primeiro registro de índices de acessos"
	}

	id := accessIdx.EventID
	lidos++

	for {
		accessIdx, e = core.LerLinhaAcessoIdx(file.Reader())
		newId := accessIdx.EventID

		if e != nil {
			break
		}

		if newId <= id {
			return false, "IDs de acessos não estão em ordem crescente"
		} else {
			id = newId
		}

		lidos++

		if int64(lidos)%core.AcessIdxStep == 0 {
			fmt.Printf("Validando índices de acessos: %d/%d\n", lidos, qtdIndices)
		}
	}

	if int64(lidos) != qtdIndices {
		return false, "Quantidade de registros de índices de acessos lidos " +
			"não corresponde ao tamanho do arquivo"
	}

	if e == core.EOF {
		fmt.Printf("Validando índices de acessos: %d/%d\n", lidos, qtdIndices)
		return true, "Sucesso: Arquivo de índices de acessos válido"
	}

	return false, "Erro ao ler o arquivo de índices de acessos"
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

	productIdxValid, productIdxMsg := validateProductIndexFile()

	if productIdxValid {
		fmt.Println("Arquivo de índices de produtos: Integro")
	} else {
		fmt.Printf("Arquivo de índices de produtos: Corrompido\n%s\n", productIdxMsg)
	}

	accessIdxValid, accessIdxMsg := validateAccessIndexFile()

	if accessIdxValid {
		fmt.Println("Arquivo de índices de acessos: Integro")
	} else {
		fmt.Printf("Arquivo de índices de acessos: Corrompido\n%s\n", accessIdxMsg)
	}
}
