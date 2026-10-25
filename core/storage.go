package core

import (
	"encoding/binary"
	"io"
	"os"
)

// AbrirArquivo abre um arquivo no caminho especificado em modo de leitura.
func AbrirArquivo(caminho string) (*os.File, error) {
	return os.Open(caminho)
}

// LerLinhaProduto lê o próximo registro de tamanho fixo (97 bytes) de produtos.dat.
func LerLinhaProduto(r io.Reader) (ProdutoRecord, error) {
	var p ProdutoRecord
	err := binary.Read(r, binary.LittleEndian, &p)
	return p, err
}

// LerLinhaAcesso lê o próximo registro de tamanho fixo (61 bytes) de acessos.dat.
func LerLinhaAcesso(r io.Reader) (AcessoRecord, error) {
	var a AcessoRecord
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
