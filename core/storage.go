package core

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AbrirArquivo abre um arquivo no caminho especificado em modo de leitura.
func AbrirArquivo(caminho string) (*os.File, error) {
	return os.Open(caminho)
}

// CriarArquivo cria um arquivo no caminho especificado em modo de escrita.
func CriarArquivo(caminho string) (*os.File, error) {
	return os.Create(caminho)
}

// LerLinhaProduto lê o próximo registro de tamanho fixo de produtos.dat.
func LerLinhaProduto(r io.Reader) (ProdutoRecord, error) {
	var p ProdutoRecord
	err := binary.Read(r, binary.LittleEndian, &p)
	return p, err
}

// LerLinhaProdutoIdx lê o próximo registro de tamanho fixo (16 bytes) de indice_prod.idx.
func LerLinhaProdutoIdx(r io.Reader) (ProdutoIdx, error) {
	var p ProdutoIdx
	err := binary.Read(r, binary.LittleEndian, &p)
	return p, err
}

// LerLinhaAcesso lê o próximo registro de tamanho fixo de acessos.dat.
func LerLinhaAcesso(r io.Reader) (AcessoRecord, error) {
	var a AcessoRecord
	err := binary.Read(r, binary.LittleEndian, &a)
	return a, err
}

// LerLinhaAcessoIdx lê o próximo registro de tamanho fixo (16 bytes) de indice_acess.idx.
func LerLinhaAcessoIdx(r io.Reader) (AcessoIdx, error) {
	var a AcessoIdx
	err := binary.Read(r, binary.LittleEndian, &a)
	return a, err
}

// ObterTamanhoArquivo retorna o tamanho do arquivo em bytes.
func ObterTamanhoArquivo(file *os.File) int64 {
	info, err := file.Stat()

	if err != nil {
		return -1
	}

	return info.Size()
}

// Exists verifica se o arquivo no caminho especificado existe.
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// SeekStart move o ponteiro do arquivo para o início mais o deslocamento especificado.
func SeekStart(file *os.File, offset int64) (int64, error) {
	ret, err := file.Seek(offset, io.SeekStart)
	return ret, err
}

// O mesmo formato é usado no arquivo principal e na extensão de cada bloco.
type formato[T any] struct {
	size     int
	ler      func(io.Reader) (T, error)
	id       func(T) int64
	elo      func(T) int64
	excluido func(T) bool
	delim    func(T) byte
	preparar func(*T, int64)
}

var prodFmt = formato[ProdutoRecord]{
	size:     ProdSize,
	ler:      LerLinhaProduto,
	id:       func(p ProdutoRecord) int64 { return p.ProductID },
	elo:      func(p ProdutoRecord) int64 { return p.Elo },
	excluido: func(p ProdutoRecord) bool { return p.Excluido },
	delim:    func(p ProdutoRecord) byte { return p.Delimiter },
	preparar: func(p *ProdutoRecord, elo int64) {
		p.Elo, p.Excluido, p.Delimiter = elo, false, Delim
	},
}

var acessFmt = formato[AcessoRecord]{
	size:     AcessSize,
	ler:      LerLinhaAcesso,
	id:       func(a AcessoRecord) int64 { return a.EventID },
	elo:      func(a AcessoRecord) int64 { return a.Elo },
	excluido: func(a AcessoRecord) bool { return a.Excluido },
	delim:    func(a AcessoRecord) byte { return a.Delimiter },
	preparar: func(a *AcessoRecord, elo int64) {
		a.Elo, a.Excluido, a.Delimiter = elo, false, Delim
	},
}

func lerNoOffset[T any](arquivo *os.File, offset int64, f formato[T]) (T, error) {
	var vazio T
	if _, err := SeekStart(arquivo, offset); err != nil {
		return vazio, fmt.Errorf("posicionar registro no offset %d: %w", offset, err)
	}
	registro, err := f.ler(arquivo)
	if err != nil {
		return vazio, fmt.Errorf("ler registro no offset %d: %w", offset, err)
	}
	if f.delim(registro) != Delim || f.elo(registro) < EloNil {
		return vazio, fmt.Errorf("registro inválido no offset %d; confira o formato da base", offset)
	}
	return registro, nil
}

func gravarNoOffset[T any](arquivo *os.File, offset int64, registro T) error {
	if _, err := SeekStart(arquivo, offset); err != nil {
		return err
	}
	return binary.Write(arquivo, binary.LittleEndian, registro)
}

func caminhoExt(arquivo *os.File, bloco int64) string {
	nome := arquivo.Name()
	return fmt.Sprintf(ExtFmt, strings.TrimSuffix(nome, filepath.Ext(nome)), bloco)
}

// Atualiza somente os oito bytes do elo, preservando o restante do registro.
func gravarElo(arquivo *os.File, offset int64, size int, elo int64) error {
	var dados [EloSize]byte
	binary.LittleEndian.PutUint64(dados[:], uint64(elo))
	_, err := arquivo.WriteAt(dados[:], offset+int64(size-EloSize-FlagSize-DelimSize))
	return err
}

func marcarExcluido(arquivo *os.File, offset int64, size int) error {
	_, err := arquivo.WriteAt([]byte{FlagOn}, offset+int64(size-FlagSize-DelimSize))
	return err
}
