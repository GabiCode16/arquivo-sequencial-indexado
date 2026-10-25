package main

import (
	"asi/core"
	"encoding/binary"
	"io"
	"os"
)

const (
	idxProdPath  = "dados/indice_prod.idx"
	idxAcessPath = "dados/indice_acess.idx"
	blockSize    = 1024
)

func generateIdxProdFile() bool {
	// Implementation for generating the products index file
	fileOrig, err := core.AbrirArquivo("dados/produtos.dat")

	if err != nil {
		return false
	}

	defer fileOrig.Close()

	fileDest, err := core.CriarArquivo(idxProdPath)

	if err != nil {
		return false
	}

	defer fileDest.Close()

	idx := 0
	var produto core.ProdutoRecord

	for err == nil {
		_, err = core.SeekStart(fileOrig, int64(idx*97)*int64(blockSize))
		produto, err = core.LerLinhaProduto(fileOrig)

		if err == nil {
			productId := produto.ProductID

			idxStruct := core.ProdutoIdx{
				ProductID: productId,
				Offset:    int64(idx*97) * int64(blockSize),
			}

			binary.Write(fileDest, binary.LittleEndian, &idxStruct)

			idx++
		}
	}

	if err != io.EOF {
		fileDest.Close()
		os.Remove(idxProdPath)
		return false
	}

	return true
}

func generateIdxAcessFile() bool {
	// Implementation for generating the access index file
	fileOrig, err := core.AbrirArquivo("dados/acessos.dat")

	if err != nil {
		return false
	}

	defer fileOrig.Close()

	fileDest, err := core.CriarArquivo(idxAcessPath)

	if err != nil {
		return false
	}

	defer fileDest.Close()

	idx := 0
	var acesso core.AcessoRecord

	for err == nil {
		_, err = core.SeekStart(fileOrig, int64(idx*61)*int64(blockSize))
		acesso, err = core.LerLinhaAcesso(fileOrig)

		if err == nil {
			eventID := acesso.EventID

			idxStruct := core.AcessoIdx{
				EventID: eventID,
				Offset:  int64(idx*61) * int64(blockSize),
			}

			binary.Write(fileDest, binary.LittleEndian, &idxStruct)

			idx++
		}
	}

	if err != io.EOF {
		fileDest.Close()
		os.Remove(idxAcessPath)
		return false
	}

	return true
}

func main() {

	existsProdutos, existsAcessos :=
		core.Exists(idxProdPath), core.Exists(idxAcessPath)

	if !existsProdutos {
		// Handle the case where the products index doesn't exist
		generateIdxProdFile()

	}

	if !existsAcessos {
		// Handle the case where the access index doesn't exist
		generateIdxAcessFile()
	}
}
