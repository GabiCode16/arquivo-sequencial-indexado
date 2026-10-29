package core

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// ListarProdutos exibe registros de produtos do arquivo principal e das extensões.
// limite igual a zero significa que todos os registros serão exibidos.
func ListarProdutos(saida io.Writer, limite int64) (int64, error) {
	if limite < 0 {
		return 0, fmt.Errorf("o limite não pode ser negativo")
	}

	dados, err := AbrirArquivo(ProdPath, ModR, false)
	if err != nil {
		return 0, fmt.Errorf("abrir produtos: %w", err)
	}
	defer dados.Close()

	exibidos, err := listarProdutosArquivo(saida, dados, "principal", limite, 0)
	if err != nil {
		return exibidos, err
	}
	if limite > 0 && exibidos >= limite {
		return exibidos, nil
	}

	extensoes, err := filepath.Glob(ProdPath + "_*.ext")
	if err != nil {
		return exibidos, fmt.Errorf("localizar extensões de produtos: %w", err)
	}
	for _, caminho := range extensoes {
		ext, err := AbrirArquivo(caminho, ModR, false)
		if err != nil {
			return exibidos, fmt.Errorf("abrir extensão de produtos %s: %w", caminho, err)
		}
		restante := limite
		if limite > 0 {
			restante = limite - exibidos
		}
		lidos, leituraErr := listarProdutosArquivo(saida, ext, filepath.Base(caminho), restante, exibidos)
		closeErr := ext.Close()
		exibidos += lidos
		if leituraErr != nil {
			return exibidos, leituraErr
		}
		if closeErr != nil {
			return exibidos, fmt.Errorf("fechar extensão de produtos %s: %w", caminho, closeErr)
		}
		if limite > 0 && exibidos >= limite {
			break
		}
	}

	return exibidos, nil
}

func listarProdutosArquivo(
	saida io.Writer,
	arquivo *File,
	origem string,
	limite int64,
	inicio int64,
) (int64, error) {
	var exibidos int64
	for limite == 0 || exibidos < limite {
		produto, err := LerLinhaProduto(arquivo.Reader())
		if err == EOF {
			return exibidos, nil
		}
		if err != nil {
			return exibidos, fmt.Errorf("ler produto em %s: %w", origem, err)
		}
		status := "ativo"
		if produto.Excluido {
			status = "excluído"
		}
		fmt.Fprintf(saida,
			"[%d] origem=%s id=%d marca=%q categoria=%q preço=%.2f status=%s\n",
			inicio+exibidos+1,
			origem,
			produto.ProductID,
			strings.Trim(string(produto.Brand[:]), " \x00"),
			strings.Trim(string(produto.CategoryCode[:]), " \x00"),
			produto.Price,
			status,
		)
		exibidos++
	}
	return exibidos, nil
}

// ListarAcessos exibe registros de acessos do arquivo principal e das extensões.
// limite igual a zero significa que todos os registros serão exibidos.
func ListarAcessos(saida io.Writer, limite int64) (int64, error) {
	if limite < 0 {
		return 0, fmt.Errorf("o limite não pode ser negativo")
	}

	dados, err := AbrirArquivo(AcessPath, ModR, false)
	if err != nil {
		return 0, fmt.Errorf("abrir acessos: %w", err)
	}
	defer dados.Close()

	exibidos, err := listarAcessosArquivo(saida, dados, "principal", limite, 0)
	if err != nil {
		return exibidos, err
	}
	if limite > 0 && exibidos >= limite {
		return exibidos, nil
	}

	extensoes, err := filepath.Glob(AcessPath + "_*.ext")
	if err != nil {
		return exibidos, fmt.Errorf("localizar extensões de acessos: %w", err)
	}
	for _, caminho := range extensoes {
		ext, err := AbrirArquivo(caminho, ModR, false)
		if err != nil {
			return exibidos, fmt.Errorf("abrir extensão de acessos %s: %w", caminho, err)
		}
		restante := limite
		if limite > 0 {
			restante = limite - exibidos
		}
		lidos, leituraErr := listarAcessosArquivo(saida, ext, filepath.Base(caminho), restante, exibidos)
		closeErr := ext.Close()
		exibidos += lidos
		if leituraErr != nil {
			return exibidos, leituraErr
		}
		if closeErr != nil {
			return exibidos, fmt.Errorf("fechar extensão de acessos %s: %w", caminho, closeErr)
		}
		if limite > 0 && exibidos >= limite {
			break
		}
	}

	return exibidos, nil
}

func listarAcessosArquivo(
	saida io.Writer,
	arquivo *File,
	origem string,
	limite int64,
	inicio int64,
) (int64, error) {
	var exibidos int64
	for limite == 0 || exibidos < limite {
		acesso, err := LerLinhaAcesso(arquivo.Reader())
		if err == EOF {
			return exibidos, nil
		}
		if err != nil {
			return exibidos, fmt.Errorf("ler acesso em %s: %w", origem, err)
		}
		status := "ativo"
		if acesso.Excluido {
			status = "excluído"
		}
		fmt.Fprintf(saida,
			"[%d] origem=%s evento=%d usuário=%d produto=%d tipo=%q data=%q status=%s\n",
			inicio+exibidos+1,
			origem,
			acesso.EventID,
			acesso.UserID,
			acesso.ProductID,
			strings.Trim(string(acesso.EventType[:]), " \x00"),
			strings.Trim(string(acesso.EventTime[:]), " \x00"),
			status,
		)
		exibidos++
	}
	return exibidos, nil
}
