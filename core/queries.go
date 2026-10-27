package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// InserirProduto insere um produto na organização indexado-sequencial.
func InserirProduto(dados, indice *os.File, produto ProdutoRecord) string {
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
	produto.Delimiter = Delim
	if tamDados == 0 {
		produto.Elo = -1
		if _, err := SeekStart(dados, 0); err != nil {
			return "[Erro] " + err.Error()
		}
		if err := GravarProduto(dados, produto); err != nil {
			return "[Erro] " + errors.Join(err, dados.Truncate(0)).Error()
		}
		if _, err := SeekStart(indice, 0); err != nil {
			return "[Erro] " + errors.Join(err, dados.Truncate(0)).Error()
		}
		entrada := ProdutoIdx{ProductID: produto.ProductID, Offset: 0}
		if err := binary.Write(indice, binary.LittleEndian, &entrada); err != nil {
			return "[Erro] " + errors.Join(err, dados.Truncate(0), indice.Truncate(0)).Error()
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
		produto.Delimiter = Delim

		arquivo := dados
		if pos < -1 {
			offsetBloco, msg := localizarIdxProduto(produto.ProductID)
			if offsetBloco == -1 {
				if msg != "[Sucesso] Produto encontrado" {
					return msg
				}
				offsetBloco = 0
			}
			ext, err := os.OpenFile(CaminhoExtensao(dados.Name(), offsetBloco/(BlockSize*int64(ProdSize))), os.O_RDWR, Perm)
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
	anterior, err := LerLinhaProduto(dados)
	if err != nil {
		return "[Erro] " + err.Error()
	}
	ext, err := os.OpenFile(CaminhoExtensao(dados.Name(), pos/(BlockSize*int64(ProdSize))), os.O_RDWR|os.O_CREATE, Perm)
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
		return "[Erro] " + errors.Join(err, ext.Truncate(tamExt)).Error()
	}
	// Atualiza somente o elo, preservando os demais campos do registro principal.
	if _, err := SeekStart(dados, pos+int64(ProdSize-EloSize-FlagSize-DelimSize)); err != nil {
		return "[Erro] " + errors.Join(err, ext.Truncate(tamExt)).Error()
	}
	if err := binary.Write(dados, binary.LittleEndian, tamExt); err != nil {
		if _, restaurar := SeekStart(dados, pos+int64(ProdSize-EloSize-FlagSize-DelimSize)); restaurar != nil {
			return "[Erro] " + errors.Join(err, restaurar).Error()
		}
		var eloAtual int64
		if leitura := binary.Read(dados, binary.LittleEndian, &eloAtual); leitura != nil {
			return "[Erro] " + errors.Join(err, leitura).Error()
		}
		if eloAtual == anterior.Elo {
			return "[Erro] " + errors.Join(err, ext.Truncate(tamExt)).Error()
		}
		if _, restaurar := SeekStart(dados, pos+int64(ProdSize-EloSize-FlagSize-DelimSize)); restaurar != nil {
			return "[Erro] " + errors.Join(err, restaurar).Error()
		}
		if restaurar := binary.Write(dados, binary.LittleEndian, anterior.Elo); restaurar != nil {
			return "[Erro] " + errors.Join(err, restaurar).Error()
		}
		return "[Erro] " + errors.Join(err, ext.Truncate(tamExt)).Error()
	}
	return "[Sucesso] Produto inserido"
}

// InserirAcesso insere um acesso na organização indexado-sequencial.
func InserirAcesso(dados, indice *os.File, acesso AcessoRecord) string {
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
	acesso.Delimiter = Delim
	if tamDados == 0 {
		acesso.Elo = -1
		if _, err := SeekStart(dados, 0); err != nil {
			return "[Erro] " + err.Error()
		}
		if err := GravarAcesso(dados, acesso); err != nil {
			return "[Erro] " + errors.Join(err, dados.Truncate(0)).Error()
		}
		if _, err := SeekStart(indice, 0); err != nil {
			return "[Erro] " + errors.Join(err, dados.Truncate(0)).Error()
		}
		entrada := AcessoIdx{EventID: acesso.EventID, Offset: 0}
		if err := binary.Write(indice, binary.LittleEndian, &entrada); err != nil {
			return "[Erro] " + errors.Join(err, dados.Truncate(0), indice.Truncate(0)).Error()
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
		acesso.Delimiter = Delim

		arquivo := dados
		if pos < -1 {
			offsetBloco, msg := localizarIdxAcesso(acesso.EventID)
			if offsetBloco == -1 {
				if msg != "[Sucesso] Acesso encontrado" {
					return msg
				}
				offsetBloco = 0
			}
			ext, err := os.OpenFile(CaminhoExtensao(dados.Name(), offsetBloco/(BlockSize*int64(AcessSize))), os.O_RDWR, Perm)
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
	anterior, err := LerLinhaAcesso(dados)
	if err != nil {
		return "[Erro] " + err.Error()
	}
	ext, err := os.OpenFile(CaminhoExtensao(dados.Name(), pos/(BlockSize*int64(AcessSize))), os.O_RDWR|os.O_CREATE, Perm)
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
		return "[Erro] " + errors.Join(err, ext.Truncate(tamExt)).Error()
	}
	// Atualiza somente o elo, preservando os demais campos do registro principal.
	if _, err := SeekStart(dados, pos+int64(AcessSize-EloSize-FlagSize-DelimSize)); err != nil {
		return "[Erro] " + errors.Join(err, ext.Truncate(tamExt)).Error()
	}
	if err := binary.Write(dados, binary.LittleEndian, tamExt); err != nil {
		if _, restaurar := SeekStart(dados, pos+int64(AcessSize-EloSize-FlagSize-DelimSize)); restaurar != nil {
			return "[Erro] " + errors.Join(err, restaurar).Error()
		}
		var eloAtual int64
		if leitura := binary.Read(dados, binary.LittleEndian, &eloAtual); leitura != nil {
			return "[Erro] " + errors.Join(err, leitura).Error()
		}
		if eloAtual == anterior.Elo {
			return "[Erro] " + errors.Join(err, ext.Truncate(tamExt)).Error()
		}
		if _, restaurar := SeekStart(dados, pos+int64(AcessSize-EloSize-FlagSize-DelimSize)); restaurar != nil {
			return "[Erro] " + errors.Join(err, restaurar).Error()
		}
		if restaurar := binary.Write(dados, binary.LittleEndian, anterior.Elo); restaurar != nil {
			return "[Erro] " + errors.Join(err, restaurar).Error()
		}
		return "[Erro] " + errors.Join(err, ext.Truncate(tamExt)).Error()
	}
	return "[Sucesso] Acesso inserido"
}

// RemoverProduto marca o registro do produto como logicamente excluído.
func RemoverProduto(dados, indice *os.File, id int64) (bool, error) {
	registro, pos, achou, err := LocalizarProduto(dados, indice, id)
	if err != nil || !achou || registro.Excluido {
		return false, err
	}
	arquivo, err := abrirLocal(dados, pos)
	if err != nil {
		return false, err
	}
	if pos.NaExt {
		defer arquivo.Close()
	}
	if err := MarcarExcluido(arquivo, pos.Offset, ProdSize); err != nil {
		return false, err
	}
	return true, nil
}

// RemoverAcesso marca o registro de acesso como logicamente excluído.
func RemoverAcesso(dados, indice *os.File, id int64) (bool, error) {
	registro, pos, achou, err := LocalizarAcesso(dados, indice, id)
	if err != nil || !achou || registro.Excluido {
		return false, err
	}
	arquivo, err := abrirLocal(dados, pos)
	if err != nil {
		return false, err
	}
	if pos.NaExt {
		defer arquivo.Close()
	}
	if err := MarcarExcluido(arquivo, pos.Offset, AcessSize); err != nil {
		return false, err
	}
	return true, nil
}
