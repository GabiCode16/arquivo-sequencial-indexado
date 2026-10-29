package main

import (
	"asi/core"
	"fmt"
	"os"
)

func generateIdxProdFile() (bool, string) {
	fileOrig, err := core.AbrirArquivo(core.ProdPath, core.ModR, false)
	if err != nil {
		return false, "Falha ao abrir o arquivo de produtos"
	}

	defer fileOrig.Close()

	fileDest, err := core.AbrirArquivo(core.ProdIdxPath, core.ModW|core.ModTrunc, true)

	if err != nil {
		return false, "Falha ao criar o arquivo de índices de produtos"
	}

	defer fileDest.Close()

	tamanho := core.ObterTamanhoArquivo(fileOrig)
	if tamanho < 0 || tamanho%core.ProdSize != 0 {
		return false, "Tamanho do arquivo de produtos inválido"
	}
	for offset := int64(0); offset < tamanho; offset += core.ProdSize * core.BlockSize {
		if _, err := core.SeekStart(fileOrig, offset); err != nil {
			return false, fmt.Sprintf("Erro ao posicionar arquivo de produtos: %v", err)
		}
		produto, err := core.LerLinhaProduto(fileOrig.Reader())
		if err != nil {
			return false, fmt.Sprintf("Erro ao ler arquivo de produtos: %v", err)
		}
		if err := core.GravarProdutoIdx(fileDest, core.ProdutoIdx{
			ProductID: produto.ProductID,
			Offset:    offset,
		}); err != nil {
			return false, fmt.Sprintf("Erro ao gravar índice de produtos: %v", err)
		}
	}

	return true, "Arquivo de índices de produtos gerado com sucesso"
}

func generateIdxAcessFile() (bool, string) {
	fileOrig, err := core.AbrirArquivo(core.AcessPath, core.ModR, false)

	if err != nil {
		return false, "Falha ao abrir o arquivo de acessos"
	}

	defer fileOrig.Close()

	fileDest, err := core.AbrirArquivo(core.AcessIdxPath, core.ModW|core.ModTrunc, true)

	if err != nil {
		return false, "Falha ao criar o arquivo de índices de acessos"
	}

	defer fileDest.Close()

	tamanho := core.ObterTamanhoArquivo(fileOrig)
	if tamanho < 0 || tamanho%core.AcessSize != 0 {
		return false, "Tamanho do arquivo de acessos inválido"
	}
	for offset := int64(0); offset < tamanho; offset += core.AcessSize * core.BlockSize {
		if _, err := core.SeekStart(fileOrig, offset); err != nil {
			return false, fmt.Sprintf("Erro ao posicionar arquivo de acessos: %v", err)
		}
		acesso, err := core.LerLinhaAcesso(fileOrig.Reader())
		if err != nil {
			return false, fmt.Sprintf("Erro ao ler arquivo de acessos: %v", err)
		}
		if err := core.GravarAcessoIdx(fileDest, core.AcessoIdx{
			EventID: acesso.EventID,
			Offset:  offset,
		}); err != nil {
			return false, fmt.Sprintf("Erro ao gravar índice de acessos: %v", err)
		}
	}

	return true, "Arquivo de índices de acessos gerado com sucesso"
}

func main() {
	falhou := false

	existsProdutos, existsAcessos :=
		core.Exists(core.ProdIdxPath), core.Exists(core.AcessIdxPath)

	if !existsProdutos {
		success, msg := generateIdxProdFile()

		if success {
			println("[Sucesso] " + msg)
		} else {
			falhou = true
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
			falhou = true
			println("[Erro] " + msg)
		}
	} else {
		println("[Aviso] O arquivo de índices de acessos já existe")
	}
	if falhou {
		os.Exit(1)
	}
}
