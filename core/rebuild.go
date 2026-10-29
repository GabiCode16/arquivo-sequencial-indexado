package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type registroReconstrucao struct {
	chave        int64
	ativo        bool
	produto      ProdutoRecord
	acesso       AcessoRecord
	gravarDados  func(*File) error
	gravarIndice func(*File, int64) error
}

func reconstruirProdutos() error {
	return reconstruirArquivo(
		ProdPath,
		ProdIdxPath,
		ProdSize,
		lerProdutoReconstrucao,
	)
}

func reconstruirAcessos() error {
	return reconstruirArquivo(
		AcessPath,
		AcessIdxPath,
		AcessSize,
		lerAcessoReconstrucao,
	)
}

func lerProdutoReconstrucao(arquivo *File) (registroReconstrucao, error) {
	registro, err := LerLinhaProduto(arquivo.Reader())
	if err != nil {
		return registroReconstrucao{}, err
	}
	return registroReconstrucao{
		chave:   registro.ProductID,
		ativo:   !registro.Excluido,
		produto: registro,
		gravarDados: func(destino *File) error {
			registro.Elo = -1
			return GravarProduto(destino, registro)
		},
		gravarIndice: func(destino *File, offset int64) error {
			return GravarProdutoIdx(destino, ProdutoIdx{
				ProductID: registro.ProductID,
				Offset:    offset,
			})
		},
	}, nil
}

func lerAcessoReconstrucao(arquivo *File) (registroReconstrucao, error) {
	registro, err := LerLinhaAcesso(arquivo.Reader())
	if err != nil {
		return registroReconstrucao{}, err
	}
	return registroReconstrucao{
		chave:  registro.EventID,
		ativo:  !registro.Excluido,
		acesso: registro,
		gravarDados: func(destino *File) error {
			registro.Elo = -1
			return GravarAcesso(destino, registro)
		},
		gravarIndice: func(destino *File, offset int64) error {
			return GravarAcessoIdx(destino, AcessoIdx{
				EventID: registro.EventID,
				Offset:  offset,
			})
		},
	}, nil
}

func reconstruirArquivo(
	dadosPath string,
	indicePath string,
	tamanhoRegistro int64,
	ler func(*File) (registroReconstrucao, error),
) error {
	extensoes, err := filepath.Glob(dadosPath + "_*.ext")
	if err != nil {
		return fmt.Errorf("localizar extensões: %w", err)
	}

	extensaoRegistros := make([]registroReconstrucao, 0)
	for _, caminho := range extensoes {
		arquivo, err := AbrirArquivo(caminho, ModR, false)
		if err != nil {
			return fmt.Errorf("abrir extensão %s: %w", caminho, err)
		}
		for {
			registro, err := ler(arquivo)
			if err == EOF {
				break
			}
			if err != nil {
				arquivo.Close()
				return fmt.Errorf("ler extensão %s: %w", caminho, err)
			}
			if registro.ativo {
				extensaoRegistros = append(extensaoRegistros, registro)
			}
		}
		if err := arquivo.Close(); err != nil {
			return fmt.Errorf("fechar extensão %s: %w", caminho, err)
		}
	}
	sort.Slice(extensaoRegistros, func(i, j int) bool {
		return extensaoRegistros[i].chave < extensaoRegistros[j].chave
	})

	origem, err := AbrirArquivo(dadosPath, ModR, false)
	if err != nil {
		return fmt.Errorf("abrir dados: %w", err)
	}

	tempDadosPath := dadosPath + ".tmp"
	tempIndicePath := indicePath + ".tmp"
	_ = os.Remove(tempDadosPath)
	_ = os.Remove(tempIndicePath)
	tempDados, err := abrirTemporario(tempDadosPath)
	if err != nil {
		origem.Close()
		return fmt.Errorf("criar dados temporários: %w", err)
	}
	tempIndice, err := abrirTemporario(tempIndicePath)
	if err != nil {
		origem.Close()
		tempDados.Close()
		os.Remove(tempDadosPath)
		return fmt.Errorf("criar índice temporário: %w", err)
	}

	cleanup := func() {
		origem.Close()
		tempDados.Close()
		tempIndice.Close()
		os.Remove(tempDadosPath)
		os.Remove(tempIndicePath)
	}

	var principal registroReconstrucao
	principalAtivo := false
	lerPrincipal := func() error {
		for {
			registro, err := ler(origem)
			if err == EOF {
				principalAtivo = false
				return nil
			}
			if err != nil {
				return err
			}
			if registro.ativo {
				principal = registro
				principalAtivo = true
				return nil
			}
		}
	}
	if err := lerPrincipal(); err != nil {
		cleanup()
		return fmt.Errorf("ler dados principais: %w", err)
	}

	extensaoIndex := 0
	var quantidade int64
	for principalAtivo || extensaoIndex < len(extensaoRegistros) {
		var registro registroReconstrucao
		if !principalAtivo {
			registro = extensaoRegistros[extensaoIndex]
			extensaoIndex++
		} else if extensaoIndex >= len(extensaoRegistros) ||
			principal.chave <= extensaoRegistros[extensaoIndex].chave {
			registro = principal
			if err := lerPrincipal(); err != nil {
				cleanup()
				return fmt.Errorf("ler dados principais: %w", err)
			}
		} else {
			registro = extensaoRegistros[extensaoIndex]
			extensaoIndex++
		}

		if quantidade%BlockSize == 0 {
			if err := registro.gravarIndice(tempIndice, quantidade*tamanhoRegistro); err != nil {
				cleanup()
				return fmt.Errorf("gravar índice temporário: %w", err)
			}
		}
		if err := registro.gravarDados(tempDados); err != nil {
			cleanup()
			return fmt.Errorf("gravar dados temporários: %w", err)
		}
		quantidade++
	}

	if err := origem.Close(); err != nil {
		cleanup()
		return fmt.Errorf("fechar dados de origem: %w", err)
	}
	if err := tempDados.Close(); err != nil {
		cleanup()
		return fmt.Errorf("fechar dados temporários: %w", err)
	}
	if err := tempIndice.Close(); err != nil {
		cleanup()
		return fmt.Errorf("fechar índice temporário: %w", err)
	}
	if err := substituirArquivo(tempDadosPath, dadosPath); err != nil {
		cleanup()
		return fmt.Errorf("substituir dados: %w", err)
	}
	if err := substituirArquivo(tempIndicePath, indicePath); err != nil {
		return fmt.Errorf("substituir índice: %w", err)
	}
	for _, caminho := range extensoes {
		if err := os.Remove(caminho); err != nil {
			return fmt.Errorf("remover extensão %s: %w", caminho, err)
		}
	}
	return nil
}

func abrirTemporario(caminho string) (*File, error) {
	arquivo, err := os.OpenFile(caminho, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(Perm))
	if err != nil {
		return nil, err
	}
	return (*File)(arquivo), nil
}

func substituirArquivo(origem, destino string) error {
	backup := destino + ".bak"
	_ = os.Remove(backup)
	if err := os.Rename(destino, backup); err != nil {
		return err
	}
	if err := os.Rename(origem, destino); err != nil {
		_ = os.Rename(backup, destino)
		return err
	}
	return os.Remove(backup)
}
