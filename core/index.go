package core

// ---------------------------------------------------------
// 1. CHAMADAS PÚBLICAS PRINCIPAIS
// ---------------------------------------------------------

// BuscarProdutoPorID busca um produto pelo ID, preenchendo o ponteiro fornecido se encontrado e ativo.
func BuscarProdutoPorID(id int64, p *ProdutoRecord) (bool, string) {
	achou, _, msg := LocalizarProduto(id, p)
	if !achou {
		return false, msg
	}

	if p.Excluido {
		return false, "[Erro] Produto excluído"
	}

	return true, "[Sucesso] Produto encontrado"
}

// BuscarAcessoPorID busca um acesso pelo ID, preenchendo o ponteiro fornecido se encontrado e ativo.
func BuscarAcessoPorID(id int64, a *AcessoRecord) (bool, string) {
	achou, _, msg := LocalizarAcesso(id, a)
	if !achou {
		return false, msg
	}

	if a.Excluido {
		return false, "[Erro] Acesso excluído"
	}

	return true, "[Sucesso] Acesso encontrado"
}

// BuscarProdutoDoAcesso localiza um acesso e, em seguida, o produto associado.
func BuscarProdutoDoAcesso(eventID int64, acesso *AcessoRecord, produto *ProdutoRecord) (bool, string) {
	if encontrado, mensagem := BuscarAcessoPorID(eventID, acesso); !encontrado {
		return false, mensagem
	}

	if encontrado, mensagem := BuscarProdutoPorID(acesso.ProductID, produto); !encontrado {
		return false, mensagem
	}

	return true, "[Sucesso] Produto associado ao acesso encontrado"
}

// ---------------------------------------------------------
// 2. FUNÇÕES DE LOCALIZAÇÃO (PESQUISA BINÁRIA + BLOCO + EXTENSÃO)
// ---------------------------------------------------------

// localizarIdxProduto realiza a pesquisa binária diretamente no arquivo de índice de produtos.
func localizarIdxProduto(id int64) (int64, string) {
	indice, err := AbrirArquivo(ProdIdxPath, ModR, false)
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

		entrada, err := LerLinhaProdutoIdx(indice.Reader())
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

// LocalizarProduto localiza um produto no arquivo principal ou na extensão.
// A localização retornada informa o offset e o bloco, além de indicar se o
// registro está em uma extensão.
func LocalizarProduto(id int64, p *ProdutoRecord) (bool, Localizacao, string) {
	// 1. Localiza o offset inicial do bloco usando a pesquisa binária no índice
	offsetBloco, msg := localizarIdxProduto(id)
	if offsetBloco == -1 {
		if msg != "[Sucesso] Produto encontrado" {
			return false, Localizacao{Offset: -1}, msg
		}
		// Uma chave menor que a primeira do índice pertence à extensão do primeiro bloco.
		offsetBloco = 0
	}

	// 2. Abre o arquivo de dados principais
	dados, err := AbrirArquivo(ProdPath, ModR, false)
	if err != nil {
		return false, Localizacao{Offset: -1}, "[Erro] Falha ao abrir arquivo de dados de produtos"
	}
	defer dados.Close()

	tamDados := ObterTamanhoArquivo(dados)
	if tamDados < 0 || tamDados%int64(ProdSize) != 0 {
		return false, Localizacao{Offset: -1}, "[Erro] Tamanho do arquivo de dados inválido"
	}
	totalRegistros := tamDados / int64(ProdSize)
	if offsetBloco < 0 || offsetBloco >= tamDados || offsetBloco%(BlockSize*int64(ProdSize)) != 0 {
		return false, Localizacao{Offset: -1}, "[Erro] Offset de bloco inválido no índice de produtos"
	}

	// Posiciona o cursor no início do bloco encontrado pelo índice
	if _, err := SeekStart(dados, offsetBloco); err != nil {
		return false, Localizacao{Offset: -1}, "[Erro] Falha ao posicionar ponteiro no bloco de dados"
	}

	// Calcula quantos registros faltam a partir do offset do bloco ou usa o tamanho completo do bloco
	// Como cada bloco possui BlockSize registros, limitamos a varredura ao final do bloco ou ao fim do arquivo.
	offsetAtual := offsetBloco
	registrosLidos := int64(0)
	var elo int64 = -1

	for registrosLidos < BlockSize && (offsetAtual/int64(ProdSize)) < totalRegistros {
		if _, err := SeekStart(dados, offsetAtual); err != nil {
			return false, Localizacao{Offset: -1}, "[Erro] Falha ao ler registro no bloco"
		}

		registro, err := LerLinhaProduto(dados.Reader())
		if err != nil {
			return false, Localizacao{Offset: -1}, "[Erro] Falha ao ler linha de produto"
		}
		if registrosLidos == 0 {
			elo = registro.Elo
		}

		// Se achou o ID exato
		if registro.ProductID == id {
			*p = registro
			return true, Localizacao{Offset: offsetAtual, Bloco: offsetBloco}, "[Sucesso] Produto localizado no bloco de dados"
		}

		// Como os registros dentro do bloco estão ordenados, se passarmos de ID, ele não existe neste bloco
		if registro.ProductID > id {
			break
		}

		offsetAtual += int64(ProdSize)
		registrosLidos++
	}

	if elo != -1 {
		ext, err := AbrirArquivo(
			CaminhoExtensao(dados.Name(), offsetBloco/(BlockSize*int64(ProdSize))),
			ModR, false,
		)
		if err != nil {
			return false, Localizacao{Offset: -1}, "[Erro] Falha ao abrir arquivo de extensão de produtos"
		}
		defer ext.Close()

		tamExt := ObterTamanhoArquivo(ext)
		if tamExt < 0 || tamExt%int64(ProdSize) != 0 {
			return false, Localizacao{Offset: -1}, "[Erro] Tamanho do arquivo de extensão inválido"
		}
		// Limitar as leituras à quantidade de registros também impede ciclos nos elos.
		for lidos := int64(0); elo != -1; lidos++ {
			if elo < 0 || elo%int64(ProdSize) != 0 || elo >= tamExt || lidos >= tamExt/int64(ProdSize) {
				return false, Localizacao{Offset: -1}, "[Erro] Elo inválido na extensão de produtos"
			}
			if _, err := SeekStart(ext, elo); err != nil {
				return false, Localizacao{Offset: -1}, "[Erro] Falha ao posicionar ponteiro na extensão de produtos"
			}
			registro, err := LerLinhaProduto(ext.Reader())
			if err != nil {
				return false, Localizacao{Offset: -1}, "[Erro] Falha ao ler produto na extensão"
			}
			if registro.ProductID == id {
				*p = registro
				return true, Localizacao{Offset: elo, Bloco: offsetBloco, EmExtensao: true}, "[Sucesso] Produto localizado na extensão"
			}
			elo = registro.Elo
		}
	}

	return false, Localizacao{Offset: offsetBloco, Bloco: offsetBloco}, "[Sucesso] Produto não encontrado"
}

// localizarIdxAcesso realiza a pesquisa binária diretamente no arquivo de índice de acessos.
func localizarIdxAcesso(id int64) (int64, string) {
	indice, err := AbrirArquivo(AcessIdxPath, ModR, false)
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

		entrada, err := LerLinhaAcessoIdx(indice.Reader())
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

// LocalizarAcesso localiza um acesso no arquivo principal ou na extensão.
// A localização retornada informa o offset e o bloco, além de indicar se o
// registro está em uma extensão.
func LocalizarAcesso(id int64, a *AcessoRecord) (bool, Localizacao, string) {
	// 1. Localiza o offset inicial do bloco usando a pesquisa binária no índice
	offsetBloco, msg := localizarIdxAcesso(id)
	if offsetBloco == -1 {
		if msg != "[Sucesso] Acesso encontrado" {
			return false, Localizacao{Offset: -1}, msg
		}
		// Uma chave menor que a primeira do índice pertence à extensão do primeiro bloco.
		offsetBloco = 0
	}

	// 2. Abre o arquivo de dados principais
	dados, err := AbrirArquivo(AcessPath, ModR, false)
	if err != nil {
		return false, Localizacao{Offset: -1}, "[Erro] Falha ao abrir arquivo de dados de acessos"
	}
	defer dados.Close()

	tamDados := ObterTamanhoArquivo(dados)
	if tamDados < 0 || tamDados%int64(AcessSize) != 0 {
		return false, Localizacao{Offset: -1}, "[Erro] Tamanho do arquivo de dados inválido"
	}
	totalRegistros := tamDados / int64(AcessSize)
	if offsetBloco < 0 || offsetBloco >= tamDados || offsetBloco%(BlockSize*int64(AcessSize)) != 0 {
		return false, Localizacao{Offset: -1}, "[Erro] Offset de bloco inválido no índice de acessos"
	}

	// Posiciona o cursor no início do bloco encontrado pelo índice
	if _, err := SeekStart(dados, offsetBloco); err != nil {
		return false, Localizacao{Offset: -1}, "[Erro] Falha ao posicionar ponteiro no bloco de dados"
	}

	// Calcula quantos registros faltam a partir do offset do bloco ou usa o tamanho completo do bloco
	// Como cada bloco possui BlockSize registros, limitamos a varredura ao final do bloco ou ao fim do arquivo.
	offsetAtual := offsetBloco
	registrosLidos := int64(0)
	var elo int64 = -1

	for registrosLidos < BlockSize && (offsetAtual/int64(AcessSize)) < totalRegistros {
		if _, err := SeekStart(dados, offsetAtual); err != nil {
			return false, Localizacao{Offset: -1}, "[Erro] Falha ao ler registro no bloco"
		}

		registro, err := LerLinhaAcesso(dados.Reader())
		if err != nil {
			return false, Localizacao{Offset: -1}, "[Erro] Falha ao ler linha de acesso"
		}
		if registrosLidos == 0 {
			elo = registro.Elo
		}

		// Se achou o ID exato
		if registro.EventID == id {
			*a = registro
			return true, Localizacao{Offset: offsetAtual, Bloco: offsetBloco}, "[Sucesso] Acesso localizado no bloco de dados"
		}

		// Como os registros dentro do bloco estão ordenados, se passarmos de ID, ele não existe neste bloco
		if registro.EventID > id {
			break
		}

		offsetAtual += int64(AcessSize)
		registrosLidos++
	}

	if elo != -1 {
		ext, err := AbrirArquivo(
			CaminhoExtensao(dados.Name(), offsetBloco/(BlockSize*int64(AcessSize))),
			ModR, false,
		)
		if err != nil {
			return false, Localizacao{Offset: -1}, "[Erro] Falha ao abrir arquivo de extensão de acessos"
		}
		defer ext.Close()

		tamExt := ObterTamanhoArquivo(ext)
		if tamExt < 0 || tamExt%int64(AcessSize) != 0 {
			return false, Localizacao{Offset: -1}, "[Erro] Tamanho do arquivo de extensão inválido"
		}
		// Limitar as leituras à quantidade de registros também impede ciclos nos elos.
		for lidos := int64(0); elo != -1; lidos++ {
			if elo < 0 || elo%int64(AcessSize) != 0 || elo >= tamExt || lidos >= tamExt/int64(AcessSize) {
				return false, Localizacao{Offset: -1}, "[Erro] Elo inválido na extensão de acessos"
			}
			if _, err := SeekStart(ext, elo); err != nil {
				return false, Localizacao{Offset: -1}, "[Erro] Falha ao posicionar ponteiro na extensão de acessos"
			}
			registro, err := LerLinhaAcesso(ext.Reader())
			if err != nil {
				return false, Localizacao{Offset: -1}, "[Erro] Falha ao ler acesso na extensão"
			}
			if registro.EventID == id {
				*a = registro
				return true, Localizacao{Offset: elo, Bloco: offsetBloco, EmExtensao: true}, "[Sucesso] Acesso localizado na extensão"
			}
			elo = registro.Elo
		}
	}

	return false, Localizacao{Offset: offsetBloco, Bloco: offsetBloco}, "[Sucesso] Acesso não encontrado"
}
