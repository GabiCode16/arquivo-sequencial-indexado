package main

import (
	"asi/core"
)

func generateIdxProdFile() (bool, string) {
	fileOrig, err := core.AbrirArquivo(core.ProdPath, core.ModR, false)
	if err != nil {
		return false, "Falha ao abrir o arquivo de produtos"
	}

	defer fileOrig.Close()

	fileDest, err := core.AbrirArquivo(core.ProdIdxPath, core.ModW, true)

	if err != nil {
		return false, "Falha ao criar o arquivo de índices de produtos"
	}

	defer fileDest.Close()

	idx := 0
	var produto core.ProdutoRecord

	for err == nil {
		_, err = core.SeekStart(fileOrig, int64(idx)*core.ProdSize*core.BlockSize)

		if err != nil {
			break
		}

		produto, err = core.LerLinhaProduto(fileOrig.Reader())

		if err == nil {
			productId := produto.ProductID

			idxStruct := core.ProdutoIdx{
				ProductID: productId,
				Offset:    int64(idx) * core.ProdSize * core.BlockSize,
			}

			err = core.GravarProdutoIdx(fileDest, idxStruct)

			idx++
		}
	}

	if err != core.EOF {
		fileDest.Close()
		core.ApagarArquivo(core.ProdIdxPath)
		return false, "Erro ao ler o arquivo de produtos, " +
			"arquivo de índices de produtos não gerado"
	}

	return true, "Arquivo de índices de produtos gerado com sucesso"
}

func generateIdxAcessFile() (bool, string) {
	fileOrig, err := core.AbrirArquivo(core.AcessPath, core.ModR, false)

	if err != nil {
		return false, "Falha ao abrir o arquivo de acessos"
	}

	defer fileOrig.Close()

	fileDest, err := core.AbrirArquivo(core.AcessIdxPath, core.ModW, true)

	if err != nil {
		return false, "Falha ao criar o arquivo de índices de acessos"
	}

	defer fileDest.Close()

	idx := 0
	var acesso core.AcessoRecord

	for err == nil {
		_, err = core.SeekStart(fileOrig, int64(idx)*core.AcessSize*core.BlockSize)

		if err != nil {
			break
		}

		acesso, err = core.LerLinhaAcesso(fileOrig.Reader())

		if err == nil {
			eventID := acesso.EventID

			idxStruct := core.AcessoIdx{
				EventID: eventID,
				Offset:  int64(idx) * core.AcessSize * core.BlockSize,
			}

			err = core.GravarAcessoIdx(fileDest, idxStruct)

			idx++
		}
	}

	if err != core.EOF {
		fileDest.Close()
		core.ApagarArquivo(core.AcessIdxPath)
		return false, "Erro ao ler o arquivo de acessos, " +
			"arquivo de índices de acessos não gerado"
	}

	return true, "Arquivo de índices de acessos gerado com sucesso"
}

func main() {

	existsProdutos, existsAcessos :=
		core.Exists(core.ProdIdxPath), core.Exists(core.AcessIdxPath)

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
