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

func generateIdxProdFile() (bool, string) {
	fileOrig, err := core.AbrirArquivo("dados/produtos.dat")

	if err != nil {
		return false, "Falha ao abrir o arquivo de produtos"
	}

	defer fileOrig.Close()

	fileDest, err := core.CriarArquivo(idxProdPath)

	if err != nil {
		return false, "Falha ao criar o arquivo de índices de produtos"
	}

	defer fileDest.Close()

	idx := 0
	var produto core.ProdutoRecord

	for err == nil {
		_, err = core.SeekStart(fileOrig, int64(idx*97)*int64(blockSize))

		if err != nil {
			break
		}

		produto, err = core.LerLinhaProduto(fileOrig)

		if err == nil {
			productId := produto.ProductID

			idxStruct := core.ProdutoIdx{
				ProductID: productId,
				Offset:    int64(idx*97) * int64(blockSize),
			}

			err = binary.Write(fileDest, binary.LittleEndian, &idxStruct)

			idx++
		}
	}

	if err != io.EOF {
		fileDest.Close()
		os.Remove(idxProdPath)
		return false, "Erro ao ler o arquivo de produtos, " +
			"arquivo de índices de produtos não gerado"
	}

	return true, "Arquivo de índices de produtos gerado com sucesso"
}

func generateIdxAcessFile() (bool, string) {
	fileOrig, err := core.AbrirArquivo("dados/acessos.dat")

	if err != nil {
		return false, "Falha ao abrir o arquivo de acessos"
	}

	defer fileOrig.Close()

	fileDest, err := core.CriarArquivo(idxAcessPath)

	if err != nil {
		return false, "Falha ao criar o arquivo de índices de acessos"
	}

	defer fileDest.Close()

	idx := 0
	var acesso core.AcessoRecord

	for err == nil {
		_, err = core.SeekStart(fileOrig, int64(idx*61)*int64(blockSize))

		if err != nil {
			break
		}

		acesso, err = core.LerLinhaAcesso(fileOrig)

		if err == nil {
			eventID := acesso.EventID

			idxStruct := core.AcessoIdx{
				EventID: eventID,
				Offset:  int64(idx*61) * int64(blockSize),
			}

			err = binary.Write(fileDest, binary.LittleEndian, &idxStruct)

			idx++
		}
	}

	if err != io.EOF {
		fileDest.Close()
		os.Remove(idxAcessPath)
		return false, "Erro ao ler o arquivo de acessos, " +
			"arquivo de índices de acessos não gerado"
	}

	return true, "Arquivo de índices de acessos gerado com sucesso"
}

func main() {

	existsProdutos, existsAcessos :=
		core.Exists(idxProdPath), core.Exists(idxAcessPath)

	if !existsProdutos {
		success, msg := generateIdxProdFile()

		if success {
			println("[Sucesso] " + msg)
		} else {
			println("[Erro] " + msg)
		}
	} else {
		println("[Aviso] O arquivo de índices de produtos já existe")
	}

	if !existsAcessos {
		success, msg := generateIdxAcessFile()

		if success {
			println("[Sucesso] " + msg)
		} else {
			println("[Erro] " + msg)
		}
	} else {
		println("[Aviso] O arquivo de índices de acessos já existe")
	}
}
