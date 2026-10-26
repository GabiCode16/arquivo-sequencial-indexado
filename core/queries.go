package core

import (
	"errors"
	"fmt"
	"os"
)

// InserirProduto usa a extensão do bloco, sem deslocar registros da base.
// Os arquivos devem estar abertos para leitura e escrita.
func InserirProduto(dados, indice *os.File, produto ProdutoRecord) error {
	return inserir(dados, indice, produto, prodFmt)
}

// InserirAcesso usa a extensão do bloco, sem deslocar registros da base.
func InserirAcesso(dados, indice *os.File, acesso AcessoRecord) error {
	return inserir(dados, indice, acesso, acessFmt)
}

// RemoverProduto marca o registro como excluído, preservando o elo e o índice.
func RemoverProduto(dados, indice *os.File, id int64) (bool, error) {
	return remover(dados, indice, id, prodFmt)
}

// RemoverAcesso marca o registro como excluído, preservando o elo e o índice.
func RemoverAcesso(dados, indice *os.File, id int64) (bool, error) {
	return remover(dados, indice, id, acessFmt)
}

func inserir[T any](dados, indice *os.File, registro T, f formato[T]) error {
	anterior, pos, achou, err := localizar(dados, indice, f.id(registro), f)
	if err != nil {
		return err
	}
	if achou {
		if !f.excluido(anterior) {
			return fmt.Errorf("o ID %d já existe", f.id(registro))
		}
		// A reinserção reativa a mesma posição e preserva seus encadeamentos.
		f.preparar(&registro, f.elo(anterior))
		arquivo, err := abrirLocal(dados, pos)
		if err != nil {
			return err
		}
		if pos.ext {
			defer arquivo.Close()
		}
		return gravarNoOffset(arquivo, pos.offset, registro)
	}

	if !pos.ext {
		// Uma base vazia precisa de uma âncora antes de receber extensões.
		f.preparar(&registro, EloNil)
		if err := gravarNoOffset(dados, 0, registro); err != nil {
			return errors.Join(err, dados.Truncate(0))
		}
		entrada := entradaIdx{Chave: f.id(registro), Offset: 0}
		if err := gravarNoOffset(indice, 0, entrada); err != nil {
			return errors.Join(err, dados.Truncate(0), indice.Truncate(0))
		}
		return nil
	}

	ext, err := os.OpenFile(caminhoExt(dados, pos.bloco), os.O_RDWR|os.O_CREATE, Perm)
	if err != nil {
		return fmt.Errorf("abrir extensão: %w", err)
	}
	defer ext.Close()
	tamanho, err := tamanhoArquivo(ext)
	if err != nil {
		return err
	}
	if tamanho%int64(f.size) != 0 {
		return fmt.Errorf("tamanho inválido da extensão do bloco %d", pos.bloco)
	}
	f.preparar(&registro, pos.offset)
	if err := gravarNoOffset(ext, tamanho, registro); err != nil {
		return errors.Join(err, ext.Truncate(tamanho))
	}

	// Mantém a cadeia ordenada por ID, embora a escrita física seja no final.
	arquivoElo, offsetElo := ext, pos.ant
	if pos.ant == EloNil {
		arquivoElo = dados
		offsetElo = pos.bloco * BlockSize * int64(f.size)
	}
	if err := gravarElo(arquivoElo, offsetElo, f.size, tamanho); err != nil {
		if restaurar := gravarElo(arquivoElo, offsetElo, f.size, pos.offset); restaurar != nil {
			return errors.Join(err, restaurar)
		}
		return errors.Join(err, ext.Truncate(tamanho))
	}
	return nil
}

func remover[T any](dados, indice *os.File, id int64, f formato[T]) (bool, error) {
	registro, pos, achou, err := localizar(dados, indice, id, f)
	if err != nil || !achou || f.excluido(registro) {
		return false, err
	}
	arquivo, err := abrirLocal(dados, pos)
	if err != nil {
		return false, err
	}
	if pos.ext {
		defer arquivo.Close()
	}
	if err := marcarExcluido(arquivo, pos.offset, f.size); err != nil {
		return false, err
	}
	return true, nil
}

func abrirLocal(dados *os.File, pos local) (*os.File, error) {
	if !pos.ext {
		return dados, nil
	}
	return os.OpenFile(caminhoExt(dados, pos.bloco), os.O_RDWR, Perm)
}
