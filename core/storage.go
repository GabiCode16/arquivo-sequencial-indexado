package core

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type ModoArquivo int
type File os.File
type Reader io.Reader

var (
	ModR     ModoArquivo = ModoArquivo(os.O_RDONLY)
	ModW     ModoArquivo = ModoArquivo(os.O_WRONLY)
	ModRW    ModoArquivo = ModoArquivo(os.O_RDWR)
	ModTrunc ModoArquivo = ModoArquivo(os.O_TRUNC)
	EOF      error       = io.EOF
)

// AbrirArquivo abre um arquivo no caminho especificado.
func AbrirArquivo(caminho string, modo ModoArquivo, criar bool) (*File, error) {
	var flag int

	if criar {
		flag = int(modo) | os.O_CREATE
	} else {
		flag = int(modo)
	}

	file, err := os.OpenFile(caminho, flag, 0644)
	if err != nil {
		return nil, err
	}

	return (*File)(file), nil
}

// ApagarArquivo remove o arquivo no caminho especificado.
func ApagarArquivo(caminho string) error {
	err := os.Remove(caminho)
	if err != nil {
		return err
	}
	return nil
}

// LerLinhaProduto lê o próximo registro de tamanho fixo de produtos.dat.
func LerLinhaProduto(r Reader) (ProdutoRecord, error) {
	var p ProdutoRecord
	err := binary.Read(r, binary.LittleEndian, &p)
	return p, err
}

// LerLinhaProdutoIdx lê o próximo registro de tamanho fixo (16 bytes) de indice_prod.idx.
func LerLinhaProdutoIdx(r Reader) (ProdutoIdx, error) {
	var p ProdutoIdx
	err := binary.Read(r, binary.LittleEndian, &p)
	return p, err
}

// LerLinhaAcesso lê o próximo registro de tamanho fixo de acessos.dat.
func LerLinhaAcesso(r Reader) (AcessoRecord, error) {
	var a AcessoRecord
	err := binary.Read(r, binary.LittleEndian, &a)
	return a, err
}

// LerLinhaAcessoIdx lê o próximo registro de tamanho fixo (16 bytes) de indice_acess.idx.
func LerLinhaAcessoIdx(r Reader) (AcessoIdx, error) {
	var a AcessoIdx
	err := binary.Read(r, binary.LittleEndian, &a)
	return a, err
}

// ObterTamanhoArquivo retorna o tamanho do arquivo em bytes.
func ObterTamanhoArquivo(file *File) int64 {
	f := (*os.File)(file)
	info, err := f.Stat()

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
func SeekStart(file *File, offset int64) (int64, error) {
	f := (*os.File)(file)
	ret, err := f.Seek(offset, io.SeekStart)
	return ret, err
}

// GravarProduto grava um ProdutoRecord.
func GravarProduto(arquivo *File, p ProdutoRecord) error {
	arq := (*os.File)(arquivo)
	return binary.Write(arq, binary.LittleEndian, &p)
}

// GravarProdutoIdx grava um ProdutoIdx.
func GravarProdutoIdx(arquivo *File, p ProdutoIdx) error {
	arq := (*os.File)(arquivo)
	return binary.Write(arq, binary.LittleEndian, &p)
}

// GravarAcesso grava um AcessoRecord.
func GravarAcesso(arquivo *File, a AcessoRecord) error {
	arq := (*os.File)(arquivo)
	return binary.Write(arq, binary.LittleEndian, &a)
}

// GravarAcessoIdx grava um AcessoIdx.
func GravarAcessoIdx(arquivo *File, a AcessoIdx) error {
	arq := (*os.File)(arquivo)
	return binary.Write(arq, binary.LittleEndian, &a)
}

// CaminhoExtensao gera o caminho do arquivo de extensão para um determinado bloco.
func CaminhoExtensao(nome string, bloco int64) string {
	return fmt.Sprintf("%s_%d.ext", nome, bloco)
}

// Close fecha o arquivo.
func (f *File) Close() error {
	return (*os.File)(f).Close()
}

// Reader retorna um Reader para o arquivo.
func (f *File) Reader() Reader {
	return (*os.File)(f)
}

// Name retorna o nome do arquivo.
func (f *File) Name() string {
	return (*os.File)(f).Name()
}
