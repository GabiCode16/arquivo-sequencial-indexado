package core

import (
	"fmt"
)

// InserirProduto insere um produto na organização indexado-sequencial.
func InserirProduto(dados, indice *File, produto ProdutoRecord) string {
	tamDados := ObterTamanhoArquivo(dados)
	tamIndice := ObterTamanhoArquivo(indice)
	if tamDados < 0 || tamDados%int64(ProdSize) != 0 {
		return "[Erro] Tamanho do arquivo de produtos inválido"
	}
	if tamIndice < 0 || tamIndice%IdxSize != 0 ||
		tamIndice/IdxSize != (tamDados/int64(ProdSize)+BlockSize-1)/BlockSize {
		return "[Erro] Índice de produtos incompatível com o arquivo de dados"
	}

	produto.Excluido = false
	produto.Delimiter = '\n'

	if tamDados == 0 {
		produto.Elo = -1
		if _, err := SeekStart(dados, 0); err != nil {
			return "[Erro] " + err.Error()
		}
		if err := GravarProduto(dados, produto); err != nil {
			return "[Erro] " + err.Error()
		}
		if _, err := SeekStart(indice, 0); err != nil {
			return "[Erro] " + err.Error()
		}
		entrada := ProdutoIdx{ProductID: produto.ProductID, Offset: 0}
		if err := GravarProdutoIdx(indice, entrada); err != nil {
			return "[Erro] " + err.Error()
		}
		return "[Sucesso] Produto inserido"
	}

	var anterior ProdutoRecord
	achou, pos, msg := LocalizarProduto(produto.ProductID, &anterior)
	if !achou && pos == -1 {
		return msg
	}

	if achou {
		if !anterior.Excluido {
			return fmt.Sprintf("[Erro] O produto ID %d já existe", produto.ProductID)
		}
		// Reativação de registro excluído
		produto.Elo = anterior.Elo
		produto.Excluido = false
		produto.Delimiter = '\n'

		arquivo := dados
		if pos < -1 {
			offsetBloco, msg := localizarIdxProduto(produto.ProductID)
			if offsetBloco == -1 {
				if msg != "[Sucesso] Produto encontrado" {
					return msg
				}
				offsetBloco = 0
			}
			ext, err := AbrirArquivo(
				CaminhoExtensao(dados.Name(), offsetBloco/(BlockSize*int64(ProdSize))),
				ModRW, false,
			)
			if err != nil {
				return "[Erro] " + err.Error()
			}
			defer ext.Close()
			arquivo = ext
			pos = -pos - 2
		}
		if _, err := SeekStart(arquivo, pos); err != nil {
			return "[Erro] " + err.Error()
		}

		err := GravarProduto(arquivo, produto)
		if err != nil {
			return "[Erro] " + err.Error()
		}

		return "[Sucesso] Produto reativado"
	}

	// O primeiro registro do bloco guarda o início da cadeia de extensão.
	if _, err := SeekStart(dados, pos); err != nil {
		return "[Erro] " + err.Error()
	}
	anterior, err := LerLinhaProduto(dados.Reader())
	if err != nil {
		return "[Erro] " + err.Error()
	}
	ext, err := AbrirArquivo(
		CaminhoExtensao(dados.Name(), pos/(BlockSize*int64(ProdSize))),
		ModRW,
		true,
	)
	if err != nil {
		return "[Erro] " + err.Error()
	}
	defer ext.Close()

	tamExt := ObterTamanhoArquivo(ext)
	if tamExt < 0 || tamExt%int64(ProdSize) != 0 {
		return "[Erro] Tamanho do arquivo de extensão inválido"
	}
	produto.Elo = anterior.Elo
	if _, err := SeekStart(ext, tamExt); err != nil {
		return "[Erro] " + err.Error()
	}
	if err := GravarProduto(ext, produto); err != nil {
		return "[Erro] " + err.Error()
	}
	// Atualiza o elo lendo a struct inteira, modificando o campo e regravando-a
	if _, err := SeekStart(dados, pos); err != nil {
		return "[Erro] " + err.Error()
	}

	// Lê o registro principal atual
	registro, err := LerLinhaProduto(dados.Reader())
	if err != nil {
		return "[Erro] " + err.Error()
	}

	// Atualiza o elo apontando para a nova extensão
	registro.Elo = tamExt

	// Retorna para a posição do registro e grava a struct inteira atualizada
	if _, err := SeekStart(dados, pos); err != nil {
		return "[Erro] " + err.Error()
	}
	if err := GravarProduto(dados, registro); err != nil {
		return "[Erro] " + err.Error()
	}

	return "[Sucesso] Produto inserido"
}

// InserirAcesso insere um acesso na organização indexado-sequencial.
func InserirAcesso(dados, indice *File, acesso AcessoRecord) string {
	tamDados := ObterTamanhoArquivo(dados)
	tamIndice := ObterTamanhoArquivo(indice)
	if tamDados < 0 || tamDados%int64(AcessSize) != 0 {
		return "[Erro] Tamanho do arquivo de acessos inválido"
	}
	if tamIndice < 0 || tamIndice%IdxSize != 0 ||
		tamIndice/IdxSize != (tamDados/int64(AcessSize)+BlockSize-1)/BlockSize {
		return "[Erro] Índice de acessos incompatível com o arquivo de dados"
	}

	acesso.Excluido = false
	acesso.Delimiter = '\n'
	if tamDados == 0 {
		acesso.Elo = -1
		if _, err := SeekStart(dados, 0); err != nil {
			return "[Erro] " + err.Error()
		}
		if err := GravarAcesso(dados, acesso); err != nil {
			return "[Erro] " + err.Error()
		}
		if _, err := SeekStart(indice, 0); err != nil {
			return "[Erro] " + err.Error()
		}
		entrada := AcessoIdx{EventID: acesso.EventID, Offset: 0}
		if err := GravarAcessoIdx(indice, entrada); err != nil {
			return "[Erro] " + err.Error()
		}
		return "[Sucesso] Acesso inserido"
	}

	var anterior AcessoRecord
	achou, pos, msg := LocalizarAcesso(acesso.EventID, &anterior)
	if !achou && pos == -1 {
		return msg
	}

	if achou {
		if !anterior.Excluido {
			return fmt.Sprintf("[Erro] O acesso ID %d já existe", acesso.EventID)
		}
		// Reativação de registro excluído
		acesso.Elo = anterior.Elo
		acesso.Excluido = false
		acesso.Delimiter = '\n'

		arquivo := dados
		if pos < -1 {
			offsetBloco, msg := localizarIdxAcesso(acesso.EventID)
			if offsetBloco == -1 {
				if msg != "[Sucesso] Acesso encontrado" {
					return msg
				}
				offsetBloco = 0
			}
			ext, err := AbrirArquivo(
				CaminhoExtensao(dados.Name(), offsetBloco/(BlockSize*int64(AcessSize))),
				ModRW, false,
			)
			if err != nil {
				return "[Erro] " + err.Error()
			}
			defer ext.Close()
			arquivo = ext
			pos = -pos - 2
		}
		if _, err := SeekStart(arquivo, pos); err != nil {
			return "[Erro] " + err.Error()
		}

		err := GravarAcesso(arquivo, acesso)
		if err != nil {
			return "[Erro] " + err.Error()
		}

		return "[Sucesso] Acesso reativado"
	}

	// O primeiro registro do bloco guarda o início da cadeia de extensão.
	if _, err := SeekStart(dados, pos); err != nil {
		return "[Erro] " + err.Error()
	}
	anterior, err := LerLinhaAcesso(dados.Reader())
	if err != nil {
		return "[Erro] " + err.Error()
	}
	ext, err := AbrirArquivo(
		CaminhoExtensao(dados.Name(), pos/(BlockSize*int64(AcessSize))),
		ModRW, true,
	)
	if err != nil {
		return "[Erro] " + err.Error()
	}
	defer ext.Close()

	tamExt := ObterTamanhoArquivo(ext)
	if tamExt < 0 || tamExt%int64(AcessSize) != 0 {
		return "[Erro] Tamanho do arquivo de extensão inválido"
	}
	acesso.Elo = anterior.Elo
	if _, err := SeekStart(ext, tamExt); err != nil {
		return "[Erro] " + err.Error()
	}
	if err := GravarAcesso(ext, acesso); err != nil {
		return "[Erro] " + err.Error()
	}
	// Atualiza o elo lendo a struct inteira, modificando o campo e regravando-a
	if _, err := SeekStart(dados, pos); err != nil {
		return "[Erro] " + err.Error()
	}

	// Lê o registro principal atual
	registro, err := LerLinhaAcesso(dados.Reader())
	if err != nil {
		return "[Erro] " + err.Error()
	}

	// Atualiza o elo apontando para a nova extensão
	registro.Elo = tamExt

	// Retorna para a posição do registro e grava a struct inteira atualizada
	if _, err := SeekStart(dados, pos); err != nil {
		return "[Erro] " + err.Error()
	}
	if err := GravarAcesso(dados, registro); err != nil {
		return "[Erro] " + err.Error()
	}

	return "[Sucesso] Acesso inserido"
}

// RemoverProduto marca o registro do produto como logicamente excluído.
func RemoverProduto(dados *File, id int64) (bool, string) {
	var registro ProdutoRecord
	achou, _, msg := LocalizarProduto(id, &registro)
	if !achou || registro.Excluido {
		return false, msg
	}
	registro.Excluido = true
	if err := GravarProduto(dados, registro); err != nil {
		return false, "[Erro] " + err.Error()
	}
	return true, "[Sucesso] Produto removido"
}

// RemoverAcesso marca o registro de acesso como logicamente excluído.
func RemoverAcesso(dados *File, id int64) (bool, string) {
	var registro AcessoRecord
	achou, _, msg := LocalizarAcesso(id, &registro)
	if !achou || registro.Excluido {
		return false, msg
	}
	registro.Excluido = true
	if err := GravarAcesso(dados, registro); err != nil {
		return false, "[Erro] " + err.Error()
	}
	return true, "[Sucesso] Acesso removido"
}
