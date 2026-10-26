package core

import (
	"encoding/binary"
	"fmt"
	"os"
)

type entradaIdx struct {
	Chave  int64
	Offset int64
}

// local identifica um registro ou o ponto de inserção na cadeia de extensão.
type local struct {
	bloco  int64
	offset int64
	ant    int64
	ext    bool
}

// BuscarProdutoPorID consulta o bloco principal e sua extensão, ignorando excluídos.
func BuscarProdutoPorID(dados, indice *os.File, id int64) (ProdutoRecord, bool, error) {
	produto, _, achou, err := localizar(dados, indice, id, prodFmt)
	if err != nil || !achou || produto.Excluido {
		return ProdutoRecord{}, false, err
	}
	return produto, true, nil
}

// BuscarAcessoPorID consulta o bloco principal e sua extensão, ignorando excluídos.
func BuscarAcessoPorID(dados, indice *os.File, id int64) (AcessoRecord, bool, error) {
	acesso, _, achou, err := localizar(dados, indice, id, acessFmt)
	if err != nil || !achou || acesso.Excluido {
		return AcessoRecord{}, false, err
	}
	return acesso, true, nil
}

func conferirArquivos(dados, indice *os.File, size int) (int64, int64, error) {
	if dados == nil || indice == nil {
		return 0, 0, fmt.Errorf("arquivos de dados e índice são obrigatórios")
	}
	tamDados, err := tamanhoArquivo(dados)
	if err != nil {
		return 0, 0, err
	}
	tamIndice, err := tamanhoArquivo(indice)
	if err != nil {
		return 0, 0, err
	}
	if tamDados%int64(size) != 0 || tamIndice%IdxSize != 0 {
		return 0, 0, fmt.Errorf("formato inválido; regenere os arquivos de dados e índices")
	}
	total, blocos := tamDados/int64(size), tamIndice/IdxSize
	if esperado := (total + BlockSize - 1) / BlockSize; blocos != esperado {
		return 0, 0, fmt.Errorf("índice desatualizado: contém %d entradas; esperadas %d", blocos, esperado)
	}
	return total, blocos, nil
}

func localizar[T any](dados, indice *os.File, id int64, f formato[T]) (T, local, bool, error) {
	var vazio T
	pos := local{offset: EloNil, ant: EloNil}
	total, blocos, err := conferirArquivos(dados, indice, f.size)
	if err != nil || total == 0 {
		return vazio, pos, false, err
	}

	// A primeira faixa também recebe chaves menores que a chave inicial da base.
	bloco := int64(0)
	for inicio, fim := int64(0), blocos-1; inicio <= fim; {
		meio := inicio + (fim-inicio)/2
		entrada, err := lerEntradaIndice(indice, meio)
		if err != nil {
			return vazio, pos, false, err
		}
		if entrada.Offset != meio*BlockSize*int64(f.size) {
			return vazio, pos, false, fmt.Errorf("offset inválido no índice, entrada %d", meio)
		}
		ancora, err := lerNoOffset(dados, entrada.Offset, f)
		if err != nil {
			return vazio, pos, false, err
		}
		if f.id(ancora) != entrada.Chave {
			return vazio, pos, false, fmt.Errorf("entrada %d do índice não corresponde aos dados", meio)
		}
		if entrada.Chave <= id {
			bloco = meio
			inicio = meio + 1
		} else {
			fim = meio - 1
		}
	}

	pos.bloco = bloco
	inicio := bloco * BlockSize * int64(f.size)
	ancora, err := lerNoOffset(dados, inicio, f)
	if err != nil {
		return vazio, pos, false, err
	}
	limite := BlockSize
	if restante := total - bloco*BlockSize; restante < limite {
		limite = restante
	}
	var idAnterior int64
	for i := int64(0); i < limite; i++ {
		offset := inicio + i*int64(f.size)
		registro := ancora
		if i > 0 {
			registro, err = lerNoOffset(dados, offset, f)
			if err != nil {
				return vazio, pos, false, err
			}
		}
		if i > 0 && f.id(registro) <= idAnterior {
			return vazio, pos, false, fmt.Errorf("registros fora de ordem no bloco %d", bloco)
		}
		if f.id(registro) == id {
			pos.offset = offset
			return registro, pos, true, nil
		}
		if f.id(registro) > id {
			break
		}
		idAnterior = f.id(registro)
	}

	// O Elo da âncora continua sendo usado mesmo quando ela está excluída.
	pos.ext = true
	proximo := f.elo(ancora)
	if proximo == EloNil {
		return vazio, pos, false, nil
	}
	ext, err := AbrirArquivo(caminhoExt(dados, bloco))
	if err != nil {
		return vazio, pos, false, fmt.Errorf("abrir extensão do bloco %d: %w", bloco, err)
	}
	defer ext.Close()
	tamanho, err := tamanhoArquivo(ext)
	if err != nil {
		return vazio, pos, false, err
	}
	if tamanho%int64(f.size) != 0 {
		return vazio, pos, false, fmt.Errorf("extensão do bloco %d tem tamanho inválido", bloco)
	}
	var chaveAnterior int64
	for lidos := int64(0); proximo != EloNil; lidos++ {
		if lidos >= tamanho/int64(f.size) || proximo < 0 || proximo%int64(f.size) != 0 || proximo >= tamanho {
			return vazio, pos, false, fmt.Errorf("elo inválido ou ciclo na extensão do bloco %d", bloco)
		}
		registro, err := lerNoOffset(ext, proximo, f)
		if err != nil {
			return vazio, pos, false, err
		}
		if lidos > 0 && f.id(registro) <= chaveAnterior {
			return vazio, pos, false, fmt.Errorf("cadeia de extensão fora de ordem ou cíclica no bloco %d", bloco)
		}
		pos.offset = proximo
		if f.id(registro) >= id {
			return registro, pos, f.id(registro) == id, nil
		}
		chaveAnterior = f.id(registro)
		pos.ant = proximo
		proximo = f.elo(registro)
	}
	pos.offset = EloNil
	return vazio, pos, false, nil
}

func lerEntradaIndice(arquivo *os.File, posicao int64) (entradaIdx, error) {
	var entrada entradaIdx
	if _, err := SeekStart(arquivo, posicao*IdxSize); err != nil {
		return entrada, err
	}
	err := binary.Read(arquivo, binary.LittleEndian, &entrada)
	return entrada, err
}

func tamanhoArquivo(arquivo *os.File) (int64, error) {
	info, err := arquivo.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
