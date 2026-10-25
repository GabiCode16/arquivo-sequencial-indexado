package core

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	tamanhoRegistroProduto int64 = 97
	tamanhoEntradaProdutoIdx int64 = 16
	registrosPorBlocoProduto int64 = 1024
)

// BuscarProdutoPorID usa busca binária no índice parcial para localizar o
// bloco de produtos e então percorre no máximo 1024 registros desse bloco.
// Os dois arquivos devem estar abertos para leitura e permanecer sem alterações
// durante a chamada.
func BuscarProdutoPorID(arquivoProdutos, arquivoIndice *os.File, productID int64) (ProdutoRecord, bool, error) {
	var produto ProdutoRecord

	if arquivoProdutos == nil || arquivoIndice == nil {
		return produto, false, fmt.Errorf("arquivos de produtos e índice são obrigatórios")
	}

	tamanhoProdutos, err := tamanhoArquivo(arquivoProdutos)
	if err != nil {
		return produto, false, fmt.Errorf("consultar tamanho de produtos.dat: %w", err)
	}
	tamanhoIndice, err := tamanhoArquivo(arquivoIndice)
	if err != nil {
		return produto, false, fmt.Errorf("consultar tamanho do índice de produtos: %w", err)
	}

	if tamanhoProdutos%tamanhoRegistroProduto != 0 {
		return produto, false, fmt.Errorf("produtos.dat tem tamanho inválido: %d bytes não é múltiplo de %d", tamanhoProdutos, tamanhoRegistroProduto)
	}
	if tamanhoIndice%tamanhoEntradaProdutoIdx != 0 {
		return produto, false, fmt.Errorf("índice de produtos tem tamanho inválido: %d bytes não é múltiplo de %d", tamanhoIndice, tamanhoEntradaProdutoIdx)
	}

	quantidadeProdutos := tamanhoProdutos / tamanhoRegistroProduto
	quantidadeIndices := tamanhoIndice / tamanhoEntradaProdutoIdx
	indicesEsperados := (quantidadeProdutos + registrosPorBlocoProduto - 1) / registrosPorBlocoProduto
	if quantidadeIndices != indicesEsperados {
		return produto, false, fmt.Errorf("índice de produtos desatualizado ou incompleto: contém %d entradas; esperadas %d", quantidadeIndices, indicesEsperados)
	}
	if quantidadeProdutos == 0 {
		return produto, false, nil
	}

	// Encontra a última entrada cuja chave não ultrapassa a chave procurada.
	var indiceBloco int64 = -1
	for inicio, fim := int64(0), quantidadeIndices-1; inicio <= fim; {
		meio := inicio + (fim-inicio)/2
		entrada, err := lerEntradaProdutoIndice(arquivoIndice, meio)
		if err != nil {
			return produto, false, err
		}
		if entrada.Offset != meio*registrosPorBlocoProduto*tamanhoRegistroProduto {
			return produto, false, fmt.Errorf("offset inválido na entrada %d do índice de produtos", meio)
		}
		registroIndice, err := lerProdutoNoOffset(arquivoProdutos, entrada.Offset)
		if err != nil {
			return produto, false, fmt.Errorf("validar entrada %d do índice de produtos: %w", meio, err)
		}
		if registroIndice.ProductID != entrada.ProductID || registroIndice.Delimiter != '\n' {
			return produto, false, fmt.Errorf("entrada %d do índice de produtos não corresponde aos dados", meio)
		}

		if entrada.ProductID <= productID {
			indiceBloco = meio
			inicio = meio + 1
		} else {
			fim = meio - 1
		}
	}
	if indiceBloco < 0 {
		return produto, false, nil
	}

	entrada, err := lerEntradaProdutoIndice(arquivoIndice, indiceBloco)
	if err != nil {
		return produto, false, err
	}
	registroInicial := entrada.Offset / tamanhoRegistroProduto
	produto, err = lerProdutoNoOffset(arquivoProdutos, entrada.Offset)
	if err != nil {
		return produto, false, fmt.Errorf("ler início do bloco de produtos %d: %w", indiceBloco, err)
	}
	if produto.ProductID != entrada.ProductID {
		return ProdutoRecord{}, false, fmt.Errorf("índice de produtos desatualizado no bloco %d", indiceBloco)
	}

	quantidadeNoBloco := registrosPorBlocoProduto
	if restante := quantidadeProdutos - registroInicial; restante < quantidadeNoBloco {
		quantidadeNoBloco = restante
	}

	var idAnterior int64
	for i := int64(0); i < quantidadeNoBloco; i++ {
		offset := entrada.Offset + i*tamanhoRegistroProduto
		produto, err = lerProdutoNoOffset(arquivoProdutos, offset)
		if err != nil {
			return ProdutoRecord{}, false, fmt.Errorf("ler produto no offset %d: %w", offset, err)
		}
		if produto.Delimiter != '\n' {
			return ProdutoRecord{}, false, fmt.Errorf("delimitador inválido no registro de produto no offset %d", offset)
		}
		if i > 0 && produto.ProductID <= idAnterior {
			return ProdutoRecord{}, false, fmt.Errorf("produtos fora de ordem no bloco iniciado no offset %d", entrada.Offset)
		}
		if produto.ProductID == productID {
			return produto, true, nil
		}
		if produto.ProductID > productID {
			return ProdutoRecord{}, false, nil
		}
		idAnterior = produto.ProductID
	}

	return ProdutoRecord{}, false, nil
}

func lerEntradaProdutoIndice(arquivo *os.File, posicao int64) (ProdutoIdx, error) {
	var entrada ProdutoIdx
	offset := posicao * tamanhoEntradaProdutoIdx
	if _, err := arquivo.Seek(offset, io.SeekStart); err != nil {
		return entrada, fmt.Errorf("posicionar no índice de produtos no offset %d: %w", offset, err)
	}
	if err := binary.Read(arquivo, binary.LittleEndian, &entrada); err != nil {
		return entrada, fmt.Errorf("ler entrada %d do índice de produtos: %w", posicao, err)
	}
	return entrada, nil
}

func lerProdutoNoOffset(arquivo *os.File, offset int64) (ProdutoRecord, error) {
	var produto ProdutoRecord
	if _, err := arquivo.Seek(offset, io.SeekStart); err != nil {
		return produto, fmt.Errorf("posicionar em produtos.dat no offset %d: %w", offset, err)
	}
	if err := binary.Read(arquivo, binary.LittleEndian, &produto); err != nil {
		return produto, fmt.Errorf("ler registro de produtos no offset %d: %w", offset, err)
	}
	return produto, nil
}

func tamanhoArquivo(arquivo *os.File) (int64, error) {
	info, err := arquivo.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
