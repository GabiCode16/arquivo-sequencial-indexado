package core

import "fmt"

// BuscarProdutoBinario busca um produto diretamente em produtos.dat,
// sem utilizar o índice parcial.
func BuscarProdutoBinario(id int64, produto *ProdutoRecord) (bool, error) {
	dados, err := AbrirArquivo(ProdPath, ModR, false)
	if err != nil {
		return false, fmt.Errorf("abrir arquivo de produtos: %w", err)
	}
	defer dados.Close()

	tamanho := ObterTamanhoArquivo(dados)
	if tamanho < 0 || tamanho%ProdSize != 0 {
		return false, fmt.Errorf("tamanho do arquivo de produtos inválido")
	}

	inicio := int64(0)
	fim := tamanho/ProdSize - 1
	for inicio <= fim {
		meio := inicio + (fim-inicio)/2
		if _, err := SeekStart(dados, meio*ProdSize); err != nil {
			return false, fmt.Errorf("posicionar busca de produtos: %w", err)
		}

		registro, err := LerLinhaProduto(dados.Reader())
		if err != nil {
			return false, fmt.Errorf("ler registro de produto: %w", err)
		}

		if registro.ProductID == id {
			*produto = registro
			return true, nil
		}
		if registro.ProductID < id {
			inicio = meio + 1
		} else {
			fim = meio - 1
		}
	}

	return false, nil
}

// BuscarAcessoBinario busca um acesso diretamente em acessos.dat,
// sem utilizar o índice parcial.
func BuscarAcessoBinario(id int64, acesso *AcessoRecord) (bool, error) {
	dados, err := AbrirArquivo(AcessPath, ModR, false)
	if err != nil {
		return false, fmt.Errorf("abrir arquivo de acessos: %w", err)
	}
	defer dados.Close()

	tamanho := ObterTamanhoArquivo(dados)
	if tamanho < 0 || tamanho%AcessSize != 0 {
		return false, fmt.Errorf("tamanho do arquivo de acessos inválido")
	}

	inicio := int64(0)
	fim := tamanho/AcessSize - 1
	for inicio <= fim {
		meio := inicio + (fim-inicio)/2
		if _, err := SeekStart(dados, meio*AcessSize); err != nil {
			return false, fmt.Errorf("posicionar busca de acessos: %w", err)
		}

		registro, err := LerLinhaAcesso(dados.Reader())
		if err != nil {
			return false, fmt.Errorf("ler registro de acesso: %w", err)
		}

		if registro.EventID == id {
			*acesso = registro
			return true, nil
		}
		if registro.EventID < id {
			inicio = meio + 1
		} else {
			fim = meio - 1
		}
	}

	return false, nil
}
