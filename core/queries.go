package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// InserirProduto insere um produto na organização indexado-sequencial.
func InserirProduto(dados, indice *os.File, produto ProdutoRecord) string {
	var anterior ProdutoRecord
	achou, pos, _ := LocalizarProduto(produto.ProductID, &anterior)

	if achou {
		if !anterior.Excluido {
			return fmt.Sprintf("[Erro] O produto ID %d já existe", produto.ProductID)
		}
		// Reativação de registro excluído
		produto.Elo = anterior.Elo
		produto.Excluido = false
		produto.Delimiter = Delim

		SeekStart(dados, pos*BlockSize*int64(ProdSize))

		err := GravarProduto(dados, produto)
		if err != nil {
			return "[Erro] " + err.Error()
		}

		return "[Sucesso] Produto reativado"
	}

	// Inserção de novo registro ( IMPLEMENTAR )
}

// InserirAcesso insere um acesso na organização indexado-sequencial.
func InserirAcesso(dados, indice *os.File, acesso AcessoRecord) error {
	anterior, pos, achou, err := LocalizarAcesso(dados, indice, acesso.EventID)
	if err != nil {
		return err
	}
	if achou {
		if !anterior.Excluido {
			return fmt.Errorf("o ID %d já existe", acesso.EventID)
		}
		acesso.Elo = anterior.Elo
		acesso.Excluido = false
		acesso.Delimiter = Delim
		arquivo, err := abrirLocal(dados, pos)
		if err != nil {
			return err
		}
		if pos.NaExt {
			defer arquivo.Close()
		}
		return GravarAcessoNoOffset(arquivo, pos.Offset, acesso)
	}

	if !pos.NaExt {
		acesso.Elo = EloNil
		acesso.Excluido = false
		acesso.Delimiter = Delim
		if err := GravarAcessoNoOffset(dados, 0, acesso); err != nil {
			return errors.Join(err, dados.Truncate(0))
		}
		entrada := AcessoIdx{EventID: acesso.EventID, Offset: 0}
		if _, err := SeekStart(indice, 0); err != nil {
			return err
		}
		if err := binary.Write(indice, binary.LittleEndian, &entrada); err != nil {
			return errors.Join(err, dados.Truncate(0), indice.Truncate(0))
		}
		return nil
	}

	ext, err := os.OpenFile(CaminhoExtensao(dados, pos.Bloco), os.O_RDWR|os.O_CREATE, Perm)
	if err != nil {
		return fmt.Errorf("abrir extensão: %w", err)
	}
	defer ext.Close()

	tamExt := ObterTamanhoArquivo(ext)
	if tamExt < 0 || tamExt%int64(AcessSize) != 0 {
		return fmt.Errorf("tamanho inválido da extensão do bloco %d", pos.Bloco)
	}

	acesso.Elo = pos.Offset
	acesso.Excluido = false
	acesso.Delimiter = Delim
	if err := GravarAcessoNoOffset(ext, tamExt, acesso); err != nil {
		return errors.Join(err, ext.Truncate(tamExt))
	}

	arquivoElo, offsetElo := ext, pos.Ant
	if pos.Ant == EloNil {
		arquivoElo = dados
		offsetElo = pos.Bloco * BlockSize * int64(AcessSize)
	}
	if err := GravarElo(arquivoElo, offsetElo, AcessSize, tamExt); err != nil {
		if restaurar := GravarElo(arquivoElo, offsetElo, AcessSize, pos.Offset); restaurar != nil {
			return errors.Join(err, restaurar)
		}
		return errors.Join(err, ext.Truncate(tamExt))
	}
	return nil
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
