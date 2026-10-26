package core

// ---------------------------------------------------------
// 1. CHAMADAS PÚBLICAS PRINCIPAIS
// ---------------------------------------------------------

// BuscarProdutoPorID busca um produto pelo ID, preenchendo o ponteiro fornecido se encontrado e ativo.
func BuscarProdutoPorID(id int64, p *ProdutoRecord) (bool, string) {
	offset, msg := localizarProduto(id)
	if offset == -1 {
		return false, msg
	}

	arquivo, err := AbrirArquivo(ProdPath)
	if err != nil {
		return false, "[Erro] Falha ao abrir arquivo de produtos"
	}
	defer arquivo.Close()

	if _, err := SeekStart(arquivo, offset); err != nil {
		return false, "[Erro] Falha ao posicionar ponteiro no arquivo de produtos"
	}

	registro, err := LerLinhaProduto(arquivo)
	if err != nil {
		return false, "[Erro] Falha ao ler linha de produto"
	}

	if registro.Excluido {
		return false, "[Erro] Produto excluído"
	}

	*p = registro
	return true, "[Sucesso] Produto encontrado"
}

// BuscarAcessoPorID busca um acesso pelo ID, preenchendo o ponteiro fornecido se encontrado e ativo.
func BuscarAcessoPorID(id int64, a *AcessoRecord) (bool, string) {
	offset, msg := localizarAcesso(id)
	if offset == -1 {
		return false, msg
	}

	arquivo, err := AbrirArquivo(AcessPath)
	if err != nil {
		return false, "[Erro] Falha ao abrir arquivo de acessos"
	}
	defer arquivo.Close()

	if _, err := SeekStart(arquivo, offset); err != nil {
		return false, "[Erro] Falha ao posicionar ponteiro no arquivo de acessos"
	}

	registro, err := LerLinhaAcesso(arquivo)
	if err != nil {
		return false, "[Erro] Falha ao ler linha de acesso"
	}

	if registro.Excluido {
		return false, "[Erro] Acesso excluído"
	}

	*a = registro
	return true, "[Sucesso] Acesso encontrado"
}

// ---------------------------------------------------------
// 2. FUNÇÕES DE LOCALIZAÇÃO (PESQUISA BINÁRIA + BLOCO + EXTENSÃO)
// ---------------------------------------------------------

// localizarIdxProduto realiza a pesquisa binária diretamente no arquivo de índice de produtos.
func localizarIdxProduto(id int64) (int64, string) {
	indice, err := AbrirArquivo(ProdIdxPath)
	if err != nil {
		return -1, "[Erro] Falha ao abrir arquivo de índice"
	}
	defer indice.Close()

	tamIndice := ObterTamanhoArquivo(indice)
	if tamIndice < 0 || tamIndice%IdxSize != 0 {
		return -1, "[Erro] Tamanho do arquivo de índice inválido"
	}

	blocos := tamIndice / IdxSize
	if blocos == 0 {
		return -1, "[Erro] Arquivo de índice vazio"
	}

	// Pesquisa binária no arquivo de índice (.idx)
	var offsetEncontrado int64 = -1
	inicio, fim := int64(0), blocos-1

	for inicio <= fim {
		meio := inicio + (fim-inicio)/2

		if _, err := SeekStart(indice, meio*IdxSize); err != nil {
			return -1, "[Erro] Falha ao buscar no arquivo de índice"
		}

		entrada, err := LerLinhaProdutoIdx(indice)
		if err != nil {
			return -1, "[Erro] Falha ao ler entrada do arquivo de índice"
		}

		if entrada.ProductID == id {
			return entrada.Offset, "[Sucesso] Produto encontrado"
		}

		if entrada.ProductID < id {
			// Salva o candidato válido e tenta buscar uma chave maior à direita
			offsetEncontrado = entrada.Offset
			inicio = meio + 1
		} else {
			// A chave do índice é maior, vai para a esquerda
			fim = meio - 1
		}
	}

	return offsetEncontrado, "[Sucesso] Produto encontrado"
}

// LocalizarProduto localiza o offset exato de um produto no arquivo de dados.
// Utiliza a pesquisa binária no índice para encontrar o bloco e, em seguida,
// faz a varredura linear dentro daquele bloco principal.
func localizarProduto(id int64) (int64, string) {
	// 1. Localiza o offset inicial do bloco usando a pesquisa binária no índice
	offsetBloco, msg := localizarIdxProduto(id)
	if offsetBloco == -1 {
		return -1, msg
	}

	// 2. Abre o arquivo de dados principais
	dados, err := AbrirArquivo(ProdPath)
	if err != nil {
		return -1, "[Erro] Falha ao abrir arquivo de dados de produtos"
	}
	defer dados.Close()

	tamDados := ObterTamanhoArquivo(dados)
	if tamDados < 0 || tamDados%int64(ProdSize) != 0 {
		return -1, "[Erro] Tamanho do arquivo de dados inválido"
	}
	totalRegistros := tamDados / int64(ProdSize)

	// Posiciona o cursor no início do bloco encontrado pelo índice
	if _, err := SeekStart(dados, offsetBloco); err != nil {
		return -1, "[Erro] Falha ao posicionar ponteiro no bloco de dados"
	}

	// Calcula quantos registros faltam a partir do offset do bloco ou usa o tamanho completo do bloco
	// Como cada bloco possui BlockSize registros, limitamos a varredura ao final do bloco ou ao fim do arquivo.
	offsetAtual := offsetBloco
	registrosLidos := int64(0)

	for registrosLidos < BlockSize && (offsetAtual/int64(ProdSize)) < totalRegistros {
		if _, err := SeekStart(dados, offsetAtual); err != nil {
			return -1, "[Erro] Falha ao ler registro no bloco"
		}

		registro, err := LerLinhaProduto(dados)
		if err != nil {
			return -1, "[Erro] Falha ao ler linha de produto"
		}

		// Se achou o ID exato
		if registro.ProductID == id {
			return offsetAtual, "[Sucesso] Produto localizado no bloco de dados"
		}

		// Como os registros dentro do bloco estão ordenados, se passarmos de ID, ele não existe neste bloco
		if registro.ProductID > id {
			break
		}

		offsetAtual += int64(ProdSize)
		registrosLidos++
	}

	return -1, "[Erro] Produto não encontrado no bloco de dados principal"
}

// localizarIdxAcesso realiza a pesquisa binária diretamente no arquivo de índice de acessos.
func localizarIdxAcesso(id int64) (int64, string) {
	indice, err := AbrirArquivo(AcessIdxPath)
	if err != nil {
		return -1, "[Erro] Falha ao abrir arquivo de índice de acessos"
	}
	defer indice.Close()

	tamIndice := ObterTamanhoArquivo(indice)
	if tamIndice < 0 || tamIndice%IdxSize != 0 {
		return -1, "[Erro] Tamanho do arquivo de índice de acessos inválido"
	}

	blocos := tamIndice / IdxSize
	if blocos == 0 {
		return -1, "[Erro] Arquivo de índice de acessos vazio"
	}

	// Pesquisa binária no arquivo de índice (.idx)
	var offsetEncontrado int64 = -1
	inicio, fim := int64(0), blocos-1

	for inicio <= fim {
		meio := inicio + (fim-inicio)/2

		if _, err := SeekStart(indice, meio*IdxSize); err != nil {
			return -1, "[Erro] Falha ao buscar no arquivo de índice de acessos"
		}

		entrada, err := LerLinhaAcessoIdx(indice)
		if err != nil {
			return -1, "[Erro] Falha ao ler entrada do arquivo de índice de acessos"
		}

		if entrada.EventID == id {
			return entrada.Offset, "[Sucesso] Acesso encontrado"
		}

		if entrada.EventID < id {
			// Salva o candidato válido e tenta buscar uma chave maior à direita
			offsetEncontrado = entrada.Offset
			inicio = meio + 1
		} else {
			// A chave do índice é maior, vai para a esquerda
			fim = meio - 1
		}
	}

	return offsetEncontrado, "[Sucesso] Acesso encontrado"
}

// LocalizarAcesso localiza o offset exato de um acesso no arquivo de dados.
// Utiliza a pesquisa binária no índice para encontrar o bloco e, em seguida,
// faz a varredura linear dentro daquele bloco principal.
func localizarAcesso(id int64) (int64, string) {
	// 1. Localiza o offset inicial do bloco usando a pesquisa binária no índice
	offsetBloco, msg := localizarIdxAcesso(id)
	if offsetBloco == -1 {
		return -1, msg
	}

	// 2. Abre o arquivo de dados principais de acessos
	dados, err := AbrirArquivo(AcessPath)
	if err != nil {
		return -1, "[Erro] Falha ao abrir arquivo de dados de acessos"
	}
	defer dados.Close()

	tamDados := ObterTamanhoArquivo(dados)
	if tamDados < 0 || tamDados%int64(AcessSize) != 0 {
		return -1, "[Erro] Tamanho do arquivo de dados de acessos inválido"
	}
	totalRegistros := tamDados / int64(AcessSize)

	// Posiciona o cursor no início do bloco encontrado pelo índice
	if _, err := SeekStart(dados, offsetBloco); err != nil {
		return -1, "[Erro] Falha ao posicionar ponteiro no bloco de dados de acessos"
	}

	offsetAtual := offsetBloco
	registrosLidos := int64(0)

	for registrosLidos < BlockSize && (offsetAtual/int64(AcessSize)) < totalRegistros {
		if _, err := SeekStart(dados, offsetAtual); err != nil {
			return -1, "[Erro] Falha ao ler registro no bloco de acessos"
		}

		registro, err := LerLinhaAcesso(dados)
		if err != nil {
			return -1, "[Erro] Falha ao ler linha de acesso"
		}

		// Se achou o ID exato
		if registro.EventID == id {
			return offsetAtual, "[Sucesso] Acesso localizado no bloco de dados"
		}

		// Como os registros dentro do bloco estão ordenados, se passarmos do ID, ele não existe neste bloco
		if registro.EventID > id {
			break
		}

		offsetAtual += int64(AcessSize)
		registrosLidos++
	}

	return -1, "[Erro] Acesso não encontrado no bloco de dados principal"
}
